package lobby

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	// TODO: どこまで絞る？
	roomSelectQuery = "SELECT * FROM room WHERE app_id = ? ORDER BY created DESC LIMIT 10000"
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
}

type AppID string

type RoomService struct {
	db   *sqlx.DB
	conf *config.LobbyConf
	apps []pb.App

	muRoomQueries sync.Mutex
	roomQueries   map[AppID]*CachedQuery // by appid
}

func NewRoomService(db *sqlx.DB, conf *config.LobbyConf) (*RoomService, error) {
	query := "SELECT id, `key` FROM app"
	var apps []pb.App
	err := db.Select(&apps, query)
	if err != nil {
		return nil, xerrors.Errorf("select apps error: %w", err)
	}
	rs := &RoomService{
		db:          db,
		conf:        conf,
		apps:        apps,
		roomQueries: make(map[AppID]*CachedQuery),
	}
	return rs, nil
}

type gameServer struct {
	Id            uint32
	Hostname      string
	PublicName    string `db:"public_name"`
	GRPCPort      int    `db:"grpc_port"`
	WebSocketPort int    `db:"ws_port"`
}

func (rs *RoomService) getGameServers() ([]gameServer, error) {
	query := "SELECT id, hostname, public_name, grpc_port, ws_port FROM game_server WHERE status=1 AND heartbeat >= ?"
	lastbeat := time.Now().Unix() - rs.conf.ValidHeartBeat

	var gameServers []gameServer
	err := rs.db.Select(&gameServers, query, lastbeat)
	if err != nil {
		return nil, xerrors.Errorf("getGameServers: %w", err)
	}
	if len(gameServers) == 0 {
		return nil, xerrors.New("getGameServers: No game server available")
	}
	log.Debugf("Alive game servers: %+v", gameServers)
	return gameServers, nil
}

func (rs *RoomService) getGameServer(id uint32) (*gameServer, error) {
	gameServers, err := rs.getGameServers()
	if err != nil {
		return nil, err
	}
	for _, game := range gameServers {
		if game.Id == id {
			return &game, nil
		}
	}
	return nil, xerrors.Errorf("getGameServer: game server not found, id=%v", id)
}

func (rs *RoomService) scanRooms(rows *sqlx.Rows) (interface{}, error) {
	defer rows.Close()

	rooms := map[string]*pb.RoomInfo{}
	for rows.Next() {
		var room pb.RoomInfo
		err := rows.StructScan(&room)
		if err != nil {
			return rooms, err
		}
		rooms[room.Id] = &room
	}
	return rooms, nil
}

func (rs *RoomService) getRooms(appId string) (map[string]*pb.RoomInfo, error) {
	rs.muRoomQueries.Lock()
	defer rs.muRoomQueries.Unlock()
	q := rs.roomQueries[AppID(appId)]
	if q == nil {
		// TODO: どのくらいキャッシュしておく？
		q = NewCachedQuery(rs.db, time.Millisecond*10, rs.scanRooms, roomSelectQuery, appId)
		rs.roomQueries[AppID(appId)] = q
	}
	rooms, err := q.Query()
	if err != nil {
		return nil, xerrors.Errorf("cachedquery error: %w", err)
	}
	return rooms.(map[string]*pb.RoomInfo), nil
}

func (rs *RoomService) getRoom(appId, roomId string) (*pb.RoomInfo, error) {
	rooms, err := rs.getRooms(appId)
	if err != nil {
		return nil, err
	}
	room, ok := rooms[roomId]
	if !ok {
		return nil, xerrors.Errorf("room not found")
	}
	return room, nil
}

func (rs *RoomService) Create(appId string, roomOption *pb.RoomOption, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	appExists := false
	for _, app := range rs.apps {
		if app.Id == appId {
			appExists = true
			break
		}
	}
	if !appExists {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	gameServers, err := rs.getGameServers()
	if err != nil {
		return nil, xerrors.Errorf("Join: failed to get game server: %w", err)
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
		AppId:      appId,
		RoomOption: roomOption,
		MasterInfo: clientInfo,
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		fmt.Printf("create room error: %v", err)
		return nil, err
	}

	log.Infof("Created room: %v", res)

	return res, nil
}

func (rs *RoomService) Join(appId, roomId string, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	appExists := false
	for _, app := range rs.apps {
		if app.Id == appId {
			appExists = true
			break
		}
	}
	if !appExists {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	room, err := rs.getRoom(appId, roomId)
	if err != nil {
		return nil, xerrors.Errorf("Join: failed to get room: %w", err)
	}

	game, err := rs.getGameServer(room.HostId)
	if err != nil {
		return nil, xerrors.Errorf("Join: failed to get game server: %w", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort)
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("client connection error: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewGameClient(conn)

	req := &pb.JoinRoomReq{
		AppId:      appId,
		RoomId:     roomId,
		ClientInfo: clientInfo,
	}

	res, err := client.Join(context.TODO(), req)
	if err != nil {
		fmt.Printf("join room error: %v", err)
		return nil, err
	}

	log.Infof("Joined room: %v", res)

	return res, nil
}
