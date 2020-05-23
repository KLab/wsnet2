package lobby

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"wsnet2/cachedquery"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	roomCountQueryString = "SELECT COUNT(*) FROM room INNER JOIN host ON room.host_id = host.id"
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
}

type Room struct {
	*pb.RoomInfo
	Host string `json:"host"`
	URL  string `json:"url"`
}

type RoomService struct {
	db             *sqlx.DB
	maxRooms       int
	appQuery       *cachedquery.Query
	roomCountQuery *cachedquery.Query
}

func NewRoom(info *pb.RoomInfo) *Room {
	return &Room{
		RoomInfo: info,
	}
}

func NewRoomService(db *sqlx.DB, maxRooms int) *RoomService {
	rs := &RoomService{
		db:       db,
		maxRooms: maxRooms,
	}
	rs.appQuery = cachedquery.New(db, time.Second*5, scanApp, appQueryString)
	rs.roomCountQuery = cachedquery.New(db, time.Millisecond, scanRoomCount, roomCountQueryString)
	return rs
}

func (rs *RoomService) getTotalRooms() (int, error) {
	ir, err := rs.roomCountQuery.Query()
	if err != nil {
		return 0, err
	}
	return ir.(int), nil
}

func scanRoomCount(rows *sqlx.Rows) (interface{}, error) {
	if !rows.Next() {
		panic("failed to fetch room count")
	}
	var roomCount int
	err := rows.Scan(&roomCount)
	return roomCount, err
}

func makeRoomURL(host, roomID string, port int) string {
	return fmt.Sprintf("ws://%s:%d/room/%s", host, port, roomID)
}

func makeRoomHost(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

type gameServer struct {
	Hostname      string
	PublicName    string `db:"public_name"`
	GRPCPort      int    `db:"grpc_port"`
	WebSocketPort int    `db:"ws_port"`
}

func (rs *RoomService) getGameServers() ([]gameServer, error) {
	var gameServers []gameServer
	// 5秒以内に更新通知がないサーバーは除外する
	err := rs.db.Select(&gameServers, "SELECT hostname, public_name, grpc_port, ws_port FROM host WHERE status = 1 AND heartbeat > unix_timestamp(now()) - ?", 5)
	if err != nil {
		return nil, err
	}
	log.Infof("Alive game servers: %+v", gameServers)
	return gameServers, nil
}

func (rs *RoomService) Create(appID string, roomOption *pb.RoomOption, clientInfo *pb.ClientInfo) (*Room, error) {
	apps, err := rs.appQuery.Query()
	if err != nil {
		return nil, err
	}
	appExists := false
	for _, app := range apps.([]pb.App) {
		if app.Id == appID {
			appExists = true
			break
		}
	}
	if !appExists {
		return nil, xerrors.Errorf("Unknown appID: %v", appID)
	}

	nRooms, err := rs.getTotalRooms()
	if err != nil {
		return nil, xerrors.Errorf("failed to fetch room count: %w", err)
	}
	if nRooms >= rs.maxRooms {
		return nil, xerrors.Errorf("maximum number of rooms has been exceeded")
	}

	gameServers, err := rs.getGameServers()
	if err != nil {
		return nil, xerrors.Errorf("failed to get game servers: %w", err)
	}
	if len(gameServers) == 0 {
		return nil, xerrors.Errorf("no game server")
	}
	game := gameServers[rand.Intn(len(gameServers))]

	grpcAddr := fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort)
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("client connection error: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewGameClient(conn)

	req := &pb.CreateRoomReq{
		AppId:      appID,
		RoomOption: roomOption,
		MasterInfo: clientInfo,
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		fmt.Printf("create room error: %v", err)
		return nil, err
	}

	// TODO: check response

	log.Infof("Created room: %v", res.RoomInfo)

	room := &Room{
		RoomInfo: res.RoomInfo,
	}
	room.Host = makeRoomHost(game.PublicName, game.WebSocketPort)
	room.URL = makeRoomURL(game.PublicName, room.RoomInfo.Id, game.WebSocketPort)

	return room, nil
}
