package game

import (
	"context"
	"errors"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/config"
	"wsnet2/pb"
)

func TestQueries(t *testing.T) {
	ok, err := regexp.MatchString(
		`INSERT INTO room \((.+,|)id(,.+|)\) VALUES \((.+,|):id(,.+|)\)`,
		roomInsertQuery)
	if err != nil {
		t.Fatalf("roomInsertQuery match error: %+v", err)
	}
	if !ok {
		t.Fatalf("roomInsertQuery not match: %v, %v", ok, roomInsertQuery)
	}

	ok, err = regexp.MatchString(
		`UPDATE room SET (.+,|)app_id=:app_id(,.+|) WHERE id=:id`,
		roomUpdateQuery)
	if err != nil {
		t.Fatalf("roomUpdateQuery match error: %+v", err)
	}
	if !ok {
		t.Fatalf("roomUpdateQuery not match: %v, %v", ok, roomUpdateQuery)
	}
}

func newDbMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %+v", err)
	}
	return sqlx.NewDb(db, "mysql"), mock
}

func TestNewRoomInfo(t *testing.T) {
	ctx := context.Background()
	db, mock := newDbMock(t)
	retryCount := 3
	maxNumber := 999

	repo := &Repository{
		app: pb.App{Id: "testing"},
		conf: &config.GameConf{
			RetryCount: retryCount,
			MaxRoomNum: maxNumber,
		},
		db: db,
	}

	dupErr := xerrors.Errorf("Duplicate entry")

	op := &pb.RoomOption{
		Visible:        true,
		Watchable:      false,
		WithNumber:     true,
		SearchGroup:    5,
		ClientDeadline: 30,
		MaxPlayers:     10,
		PublicProps:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
		PrivateProps:   []byte{11, 12, 13, 14, 15, 16, 17, 18},
	}

	// 生成されるはずの値
	seed := time.Now().Unix()
	rand.Seed(seed)
	id1 := RandomHex(lenId)
	num1 := rand.Int31n(int32(maxNumber)) + 1
	id2 := RandomHex(lenId)
	num2 := rand.Int31n(int32(maxNumber)) + 1

	insQuery := "INSERT INTO room "
	mock.ExpectBegin()
	mock.ExpectExec(insQuery).WillReturnError(dupErr)
	mock.ExpectExec(insQuery).WillReturnResult(sqlmock.NewResult(1, 1))

	rand.Seed(seed)
	tx, _ := db.Beginx()
	ri, err := repo.newRoomInfo(ctx, tx, op)
	if err != nil {
		t.Fatalf("NewRoomInfo fail: %v", err)
	}

	if ri.Id == id1 || ri.Id != id2 {
		t.Fatalf("ri.Id = %v, wants %v", ri.Id, id2)
	}
	if ri.Number.Number == num1 || ri.Number.Number != num2 {
		t.Fatalf("ri.Number = %v, wants %v", ri.Number.Number, num2)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// リトライ回数オーバーでエラーになるはず
	for i := 0; i < retryCount; i++ {
		mock.ExpectExec(insQuery).WillReturnError(dupErr)
	}
	_, err = repo.newRoomInfo(ctx, tx, op)
	if !errors.Is(err, dupErr) {
		t.Fatalf("NewRoomInfo error: %v wants %v", err, dupErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
