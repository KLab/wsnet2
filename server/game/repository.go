package game

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"wsnet2/pb"
)

const (
	// RoomID文字列長
	lenId = 16

	// キー衝突時のリトライ回数
	retryCount = 5
)

var (
	maxNumber int32 = 999999 // 部屋番号桁数. TODO: config化

	hostId uint32

	roomInsertQuery string
	roomUpdateQuery string
)

func init() {
	// TODO: get host id
	hostId = 1

	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())

	// room_info queries
	t := reflect.TypeOf(pb.RoomInfo{})
	cols := make([]string, 0, 16)
	for i := 0; i < t.NumField(); i++ {
		if c := t.Field(i).Tag.Get("db"); c != "" {
			cols = append(cols, c)
		}
	}
	roomInsertQuery = fmt.Sprintf("INSERT INTO room (%s) VALUES (:%s)",
		strings.Join(cols, ","), strings.Join(cols, ",:"))

	var sets []string
	for _, c := range cols {
		if c != "id" {
			sets = append(sets, c+"=:"+c)
		}
	}
	roomUpdateQuery = fmt.Sprintf("UPDATE room SET %s WHERE id=:id", strings.Join(sets, ","))
}

func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b) // rand.Read always success.
	return hex.EncodeToString(b)
}

type RoomRepository struct{}

func (repo *RoomRepository) NewRoomInfo(ctx context.Context, tx *sqlx.Tx, appId string, op *pb.RoomOption) (*pb.RoomInfo, error) {
	ri := &pb.RoomInfo{
		AppId:             appId,
		HostId:            hostId,
		Visible:           op.Visible,
		Watchable:         op.Watchable,
		SearchGroup:       op.SearchGroup,
		ClientDeadline:    op.ClientDeadline,
		MaxPlayers:        op.MaxPlayers,
		Players:           1,
		PublicProperties:  op.PublicProperties,
		PrivateProperties: op.PrivateProperties,
	}
	ri.SetCreated(time.Now())

	var err error
	for n := 0; n < retryCount; n++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ctx done: %w", ctx.Err())
		default:
		}

		ri.Id = RandomHex(lenId)
		if op.WithNumber {
			ri.Number = rand.Int31n(maxNumber) + 1 // [1..maxNumber]
		}

		_, err = tx.NamedExecContext(ctx, roomInsertQuery, ri)
		if err == nil {
			return ri, nil
		}
	}

	return nil, fmt.Errorf("NewRoomInfo try %d times: %w", retryCount, err)
}
