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
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	// RoomID文字列長
	lenId = 16
)

var (
	roomInsertQuery string
	roomUpdateQuery string
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())

	initQueries()
}

func initQueries() {
	// room_info queries
	t := reflect.TypeOf(pb.RoomInfo{})
	cols := make([]string, 0, t.NumField())
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

type Repository struct {
	hostId uint32

	app  pb.App
	conf *config.GameConf
	db   *sqlx.DB

	mu      sync.RWMutex
	rooms   map[RoomID]*Room
	clients map[ClientID]map[RoomID]*Client
}

func NewRepos(db *sqlx.DB, conf *config.GameConf) (map[pb.AppId]*Repository, error) {
	hostId := uint32(1) // TODO: ちゃんとした値を取得

	query := "SELECT id, `key` FROM app"
	var apps []pb.App
	err := db.Select(&apps, query)
	if err != nil {
		return nil, xerrors.Errorf("select apps error: %w", err)
	}
	repos := make(map[pb.AppId]*Repository, len(apps))
	for _, app := range apps {
		log.Debugf("new repository: app=%#v", app.Id)
		repos[app.Id] = &Repository{
			hostId: hostId,
			app:    app,
			conf:   conf,
			db:     db,

			rooms:   make(map[RoomID]*Room),
			clients: make(map[ClientID]map[RoomID]*Client),
		}
	}
	return repos, nil
}

func (repo *Repository) CreateRoom(ctx context.Context, op *pb.RoomOption, master *pb.ClientInfo) (*pb.RoomInfo, *pb.ClientInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := repo.db.Beginx()
	if err != nil {
		return nil, nil, xerrors.Errorf("begin error: %w", err)
	}

	info, err := repo.newRoomInfo(ctx, tx, op)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	loglevel := log.CurrentLevel()
	if op.LogLevel > 0 {
		loglevel = log.Level(op.LogLevel)
	}

	room, cli, ch, err := NewRoom(repo, info, master, repo.conf, loglevel)
	if err != nil {
		tx.Rollback()
		return nil, nil, xerrors.Errorf("NewRoom error: %w", err)
	}

	var joined JoinedInfo
	select {
	case j, ok := <-ch:
		if !ok {
			tx.Rollback()
			return nil, nil, xerrors.Errorf("CreateRoom joind chan closed: room=%v", room.ID())
		}
		joined = j
	case <-ctx.Done():
		tx.Rollback()
		return nil, nil, xerrors.Errorf("CreateRoom timeout or context done: room=%v", room.ID())
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.rooms[room.ID()] = room
	if _, ok := repo.clients[cli.ID()]; !ok {
		repo.clients[cli.ID()] = make(map[RoomID]*Client)
	}
	repo.clients[cli.ID()][room.ID()] = cli

	tx.Commit()
	return joined.Room, joined.Client, nil
}

func (repo *Repository) newRoomInfo(ctx context.Context, tx *sqlx.Tx, op *pb.RoomOption) (*pb.RoomInfo, error) {
	ri := &pb.RoomInfo{
		AppId:          repo.app.Id,
		HostId:         repo.hostId,
		Visible:        op.Visible,
		Watchable:      op.Watchable,
		SearchGroup:    op.SearchGroup,
		ClientDeadline: op.ClientDeadline,
		MaxPlayers:     op.MaxPlayers,
		Players:        1,
		PublicProps:    op.PublicProps,
		PrivateProps:   op.PrivateProps,
	}
	ri.SetCreated(time.Now())

	maxNumber := int32(repo.conf.MaxRoomNum)
	retryCount := repo.conf.RetryCount
	var err error
	for n := 0; n < retryCount; n++ {
		select {
		case <-ctx.Done():
			return nil, xerrors.Errorf("ctx done: %w", ctx.Err())
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

	return nil, xerrors.Errorf("NewRoomInfo try %d times: %w", retryCount, err)
}

func (repo *Repository) updateRoomInfo(room *Room) {
	// DBへの反映は遅延して良い
	ri := room.RoomInfo.Clone()
	go func() {
		if _, err := repo.db.NamedExec(roomUpdateQuery, ri); err != nil {
			log.Errorf("Repository updateRoomInfo error: %v", err)
		}
	}()
}

func (repo *Repository) deleteRoom(id RoomID) {
	var err error
	// TODO: 部屋の履歴を残す必要あり？
	_, err = repo.db.Query("DELETE FROM room WHERE id=?", id)
	if err != nil {
		log.Errorf("deleteRoom: %w", err)
	}
}

func (repo *Repository) RemoveRoom(room *Room) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	rid := room.ID()
	delete(repo.rooms, rid)
	repo.deleteRoom(rid)
	log.Debugf("room removed from repository: room=%v", rid)
}

func (repo *Repository) RemoveClient(cli *Client) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	cid := cli.ID()
	rid := cli.room.ID()
	if cmap, ok := repo.clients[cid]; ok {
		delete(cmap, rid)
		if len(cmap) == 0 {
			delete(repo.clients, cid)
		}
	}
	log.Debugf("client removed from repository: room=%v, client=%v", rid, cid)
}

func (repo *Repository) GetClient(roomId, userId string) (*Client, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	cli, ok := repo.clients[ClientID(userId)][RoomID(roomId)]
	if !ok {
		return nil, xerrors.Errorf("client not found: room=%v, client=%v", roomId, userId)
	}
	return cli, nil
}
