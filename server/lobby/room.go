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

	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/pb"
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
}

type RoomService struct {
	db   *sqlx.DB
	conf *config.LobbyConf
	apps []pb.App
}

func NewRoomService(db *sqlx.DB, conf *config.LobbyConf) (*RoomService, error) {
	query := "SELECT id, `key` FROM app"
	var apps []pb.App
	err := db.Select(&apps, query)
	if err != nil {
		return nil, xerrors.Errorf("select apps error: %w", err)
	}
	rs := &RoomService{
		db:   db,
		conf: conf,
		apps: apps,
	}
	return rs, nil
}

type gameServer struct {
	Hostname      string
	PublicName    string `db:"public_name"`
	GRPCPort      int    `db:"grpc_port"`
	WebSocketPort int    `db:"ws_port"`
}

func (rs *RoomService) getGameServers() ([]gameServer, error) {
	query := "SELECT hostname, public_name, grpc_port, ws_port FROM game_server WHERE status=1 AND heartbeat >= ?"
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
		return nil, xerrors.Errorf("Create: failed to get game servers: %w", err)
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
	query := "SELECT hostname, public_name, grpc_port, ws_port FROM game_server INNER JOIN room ON game_server.id = room.host_id WHERE status=1 AND heartbeat >= ? AND room.id = ?"
	lastbeat := time.Now().Unix() - rs.conf.ValidHeartBeat

	var game gameServer
	err := rs.db.Get(&game, query, lastbeat, roomId)
	if err != nil {
		return nil, xerrors.Errorf("Join: failed to get game server: %w", err)
	}
	log.Debugf("game: %v", game)

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
