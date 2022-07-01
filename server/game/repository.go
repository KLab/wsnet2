package game

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"

	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	// RoomID文字列長
	lenId = 16
)

var (
	roomInsertQuery        string
	roomUpdateQuery        string
	roomHistoryInsertQuery string
)

func init() {
	initQueries()
}

func dbCols(t reflect.Type) []string {
	cols := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		if c := t.Field(i).Tag.Get("db"); c != "" {
			cols = append(cols, c)
		}
	}
	return cols
}

func initQueries() {
	// room_info queries
	{
		cols := dbCols(reflect.TypeOf(pb.RoomInfo{}))
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

	// room_history
	{
		cols := dbCols(reflect.TypeOf(roomHistory{}))
		roomHistoryInsertQuery = fmt.Sprintf("INSERT INTO room_history (%s) VALUES (:%s)",
			strings.Join(cols, ","), strings.Join(cols, ",:"))
	}
}

func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b) // rand.Read always success.
	return hex.EncodeToString(b)
}

type Repository struct {
	hostId uint32

	app  *pb.App
	conf *config.GameConf
	db   *sqlx.DB

	mu      sync.RWMutex
	rooms   map[RoomID]*Room
	clients map[ClientID]map[RoomID]*Client
}

func NewRepos(db *sqlx.DB, conf *config.GameConf, hostId uint32) (map[pb.AppId]*Repository, error) {
	if _, err := db.Exec("DELETE FROM `room` WHERE host_id=?", hostId); err != nil {
		return nil, xerrors.Errorf("delete room error: %w", err)
	}
	query := "SELECT id, `key` FROM app"
	var apps []*pb.App
	err := db.Select(&apps, query)
	if err != nil {
		return nil, xerrors.Errorf("select apps error: %w", err)
	}
	log.Debugf("new repos: apps=%v", apps)
	repos := make(map[pb.AppId]*Repository, len(apps))
	for _, app := range apps {
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

func (repo *Repository) CreateRoom(ctx context.Context, op *pb.RoomOption, master *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, ErrorWithCode) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	repo.mu.RLock()
	rooms := len(repo.rooms)
	clients := len(repo.clients)
	repo.mu.RUnlock()
	if rooms >= repo.conf.MaxRooms {
		return nil, WithCode(
			xerrors.Errorf("reached to the max_rooms"), codes.ResourceExhausted)
	}
	if clients >= repo.conf.MaxClients {
		return nil, WithCode(
			xerrors.Errorf("reached to the max_clients"), codes.ResourceExhausted)
	}

	tx, err := repo.db.Beginx()
	if err != nil {
		return nil, WithCode(xerrors.Errorf("db.Beginx: %w", err), codes.Internal)
	}

	info, ewc := repo.newRoomInfo(ctx, tx, op)
	if ewc != nil {
		tx.Rollback()
		return nil, ewc
	}

	loglevel := log.CurrentLevel()
	if op.LogLevel > 0 {
		loglevel = log.Level(op.LogLevel)
	}
	logger := log.Get(loglevel).With(log.KeyApp, repo.app.Id, log.KeyRoom, info.Id)
	logger.Infof("new room: %v, num=%v, master=%v", info.Id, info.Number.Number, master.Id)

	room, joined, ewc := NewRoom(ctx, repo, info, master, macKey, op.ClientDeadline, repo.conf, logger)
	if ewc != nil {
		tx.Rollback()
		return nil, WithCode(xerrors.Errorf("NewRoom: %w", ewc), ewc.Code())
	}

	if err := tx.Commit(); err != nil {
		return nil, WithCode(
			xerrors.Errorf("commit new room: %w", err), codes.Internal)
	}

	cli := joined.Client

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if len(repo.rooms) >= repo.conf.MaxRooms {
		logger.Warn("reached to the max_rooms. delete room: %v", room.Id)
		// 履歴は残さずに部屋を削除
		_, err := repo.db.Exec("DELETE FROM room WHERE id=?", room.Id)
		if err != nil {
			logger.Errorf("delete room (%v): %+v", room.Id, err)
		}
		return nil, WithCode(
			xerrors.Errorf("reached to the max_rooms"), codes.ResourceExhausted)
	}

	repo.rooms[room.ID()] = room
	if _, ok := repo.clients[cli.ID()]; !ok {
		repo.clients[cli.ID()] = make(map[RoomID]*Client)
	}
	repo.clients[cli.ID()][room.ID()] = cli

	return &pb.JoinedRoomRes{
		RoomInfo: joined.Room,
		Players:  joined.Players,
		AuthKey:  cli.authKey,
		MasterId: string(joined.MasterId),
		Deadline: uint32(joined.Deadline / time.Second),
	}, nil
}

func (repo *Repository) JoinRoom(ctx context.Context, id string, client *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, ErrorWithCode) {
	return repo.joinRoom(ctx, id, client, macKey, true)
}

func (repo *Repository) WatchRoom(ctx context.Context, id string, client *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, ErrorWithCode) {
	return repo.joinRoom(ctx, id, client, macKey, false)
}

func (repo *Repository) joinRoom(ctx context.Context, id string, client *pb.ClientInfo, macKey string, isPlayer bool) (*pb.JoinedRoomRes, ErrorWithCode) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	repo.mu.RLock()
	clients := len(repo.clients)
	repo.mu.RUnlock()
	if clients >= repo.conf.MaxClients && !client.IsHub { // 上限に達していてもHubからの接続は受け付ける
		return nil, WithCode(
			xerrors.Errorf("reached to the max_clients"), codes.ResourceExhausted)
	}

	room, err := repo.GetRoom(id)
	if err != nil {
		return nil, WithCode(xerrors.Errorf("repo.GetRoom: %w", err), codes.NotFound)
	}

	jch := make(chan *JoinedInfo, 1)
	errch := make(chan ErrorWithCode, 1)
	var msg Msg
	if isPlayer {
		msg = &MsgJoin{client, macKey, jch, errch}
	} else {
		msg = &MsgWatch{client, macKey, jch, errch}
	}

	select {
	case <-ctx.Done():
		return nil, WithCode(
			xerrors.Errorf("context done: room=%v", room.Id),
			codes.DeadlineExceeded)
	case room.msgCh <- msg:
	}

	var joined *JoinedInfo
	select {
	case <-ctx.Done():
		return nil, WithCode(
			xerrors.Errorf("context done: room=%v", room.Id),
			codes.DeadlineExceeded)
	case ewc := <-errch:
		return nil, WithCode(ewc, ewc.Code())
	case joined = <-jch:
	}

	cli := joined.Client

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if _, ok := repo.clients[cli.ID()]; !ok {
		repo.clients[cli.ID()] = make(map[RoomID]*Client)
	}
	repo.clients[cli.ID()][room.ID()] = cli

	return &pb.JoinedRoomRes{
		RoomInfo: joined.Room,
		Players:  joined.Players,
		AuthKey:  cli.authKey,
		MasterId: string(joined.MasterId),
		Deadline: uint32(joined.Deadline / time.Second),
	}, nil
}

func (repo *Repository) newRoomInfo(ctx context.Context, tx *sqlx.Tx, op *pb.RoomOption) (*pb.RoomInfo, ErrorWithCode) {
	ri := &pb.RoomInfo{
		AppId:        repo.app.Id,
		HostId:       repo.hostId,
		Visible:      op.Visible,
		Joinable:     op.Joinable,
		Watchable:    op.Watchable,
		Number:       &pb.RoomNumber{},
		SearchGroup:  op.SearchGroup,
		MaxPlayers:   op.MaxPlayers,
		Players:      1,
		PublicProps:  op.PublicProps,
		PrivateProps: op.PrivateProps,
	}
	ri.SetCreated(time.Now())

	maxNumber := int32(repo.conf.MaxRoomNum)
	retryCount := repo.conf.RetryCount
	var err error
	for n := 0; n < retryCount; n++ {
		select {
		case <-ctx.Done():
			return nil, WithCode(xerrors.Errorf("ctx done: %w", ctx.Err()), codes.DeadlineExceeded)
		default:
		}

		ri.Id = RandomHex(lenId)
		if op.WithNumber {
			ri.Number.Number = rand.Int31n(maxNumber) + 1 // [1..maxNumber]
		}

		_, err = tx.NamedExecContext(ctx, roomInsertQuery, ri)
		if err == nil {
			return ri, nil
		}
	}

	return nil, WithCode(xerrors.Errorf("NewRoomInfo try %d times: %w", retryCount, err), codes.Internal)
}

func (repo *Repository) updateRoomInfo(ri *pb.RoomInfo, conn *sqlx.Conn, logger log.Logger) {
	// DBへの反映は遅延して良い
	q, args, err := sqlx.Named(roomUpdateQuery, ri)
	if err != nil {
		logger.Errorf("update roominfo query: q=%v, ri=%v, err=%+v", roomUpdateQuery, ri, err)
		return
	}

	if _, err := conn.ExecContext(context.Background(), q, args...); err != nil {
		logger.Errorf("update roominfo: %v %+v", ri.Id, err)
	}
}

type roomHistory struct {
	AppID        string        `db:"app_id"`
	HostID       uint32        `db:"host_id"`
	RoomID       string        `db:"room_id"`
	Number       sql.NullInt32 `db:"number"`
	SearchGroup  uint32        `db:"search_group"`
	MaxPlayers   uint32        `db:"max_players"`
	PublicProps  []byte        `db:"public_props"`
	PrivateProps []byte        `db:"private_props"`
	PlayerLogs   []byte        `db:"player_logs"` // JSON Type
	Created      time.Time     `db:"created"`
	Closed       time.Time     `db:"closed"`
}

func (repo *Repository) deleteRoom(room *Room) {
	var err error
	_, err = repo.db.Exec("DELETE FROM room WHERE id=?", room.Id)
	if err != nil {
		room.logger.Errorf("delete room record (%v): %+v", room.Id, err)
		return
	}

	// room_history テーブルに クローズしたルーム情報を保存する
	// Room number は nil の可能性があるので場合分け
	number := sql.NullInt32{Int32: 0, Valid: false}
	if room.Number != nil {
		number = sql.NullInt32{Int32: room.Number.Number, Valid: true}
	}

	playerLogs, err := json.Marshal(room.playerLogs)
	if err != nil {
		room.logger.Errorf("marshal player logs: %+v", err)
	}

	history := roomHistory{
		AppID:        room.AppId,
		HostID:       room.HostId,
		RoomID:       room.Id,
		Number:       number,
		SearchGroup:  room.SearchGroup,
		MaxPlayers:   room.MaxPlayers,
		PublicProps:  room.PublicProps,
		PrivateProps: room.PrivateProps,
		PlayerLogs:   playerLogs,
		Created:      room.Created.Time(),
		Closed:       time.Now(),
	}

	_, err = repo.db.NamedExec(roomHistoryInsertQuery, history)
	if err != nil {
		room.logger.Errorf("insert to room_history: %+v", err)
	}
}

func (repo *Repository) RemoveRoom(room *Room) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	rid := room.ID()
	delete(repo.rooms, rid)

	repo.deleteRoom(room)
	room.logger.Debugf("room removed from repository: %v", rid)
}

func (repo *Repository) RemoveClient(cli *Client) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	cid := cli.ID()
	rid := cli.room.ID()
	if cmap, ok := repo.clients[cid]; ok {
		// IDが同じでも別クライアントの場合には削除しない
		if c, ok := cmap[rid]; ok && c != cli {
			c.logger.Debugf("cannot remove client from repository (already replaced new client): room=%v, client=%v", rid, cid)
			return
		}
		delete(cmap, rid)
		if len(cmap) == 0 {
			delete(repo.clients, cid)
		}
	}
	cli.logger.Debugf("client removed from repository: room=%v, client=%v", rid, cid)
}

func (repo *Repository) GetRoom(roomId string) (*Room, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	room, ok := repo.rooms[RoomID(roomId)]
	if !ok {
		return nil, xerrors.Errorf("room not found: room=%v", roomId)
	}
	return room, nil
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

func (repo *Repository) GetRoomCount() int {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	return len(repo.rooms)
}

func (repo *Repository) GetRoomInfo(ctx context.Context, id string) (*pb.GetRoomInfoRes, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	room, err := repo.GetRoom(id)
	if err != nil {
		return nil, WithCode(xerrors.Errorf("GetRoomInfo: %w", err), codes.NotFound)
	}

	ch := make(chan *pb.GetRoomInfoRes, 1)
	msg := &MsgGetRoomInfo{Res: ch}
	select {
	case <-ctx.Done():
		return nil, WithCode(
			xerrors.Errorf("GetRoomInfo write msg timeout or context done: room=%v", room.Id),
			codes.DeadlineExceeded)
	case room.msgCh <- msg:
	}

	var res *pb.GetRoomInfoRes
	select {
	case <-ctx.Done():
		return nil, WithCode(
			xerrors.Errorf("GetRoomInfo response timeout or context done: room=%v", room.Id),
			codes.DeadlineExceeded)
	case res = <-ch:
	}

	return res, nil
}

func (repo *Repository) AdminKick(ctx context.Context, roomID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	room, err := repo.GetRoom(roomID)
	if err != nil {
		return WithCode(xerrors.Errorf("AdminKick: can not find room %q; %w", roomID, err), codes.NotFound)
	}

	ch := make(chan error, 1)
	msg := &MsgAdminKick{
		Target: ClientID(userID),
		Res:    ch,
	}
	select {
	case <-ctx.Done():
		return WithCode(
			xerrors.Errorf("AdminKick write msg timeout or context done: room=%v", room.Id),
			codes.DeadlineExceeded)
	case room.msgCh <- msg:
	}

	select {
	case <-ctx.Done():
		return WithCode(
			xerrors.Errorf("GetRoomInfo response timeout or context done: room=%v", room.Id),
			codes.DeadlineExceeded)
	case err = <-ch:
		return err
	}
}
