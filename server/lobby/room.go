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

type JoinedResponse struct {
	Room    *pb.RoomInfo
	Players []*pb.ClientInfo
	URL     string
}

func NewJoinedResponse(game *gameServer, room *pb.RoomInfo, players []*pb.ClientInfo) *JoinedResponse {
	url := fmt.Sprintf("ws://%s:%d/room/%s", game.PublicName, game.WebSocketPort, room.Id)
	return &JoinedResponse{
		Room:    room,
		Players: players,
		URL:     url,
	}
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

func (rs *RoomService) Create(appID string, roomOption *pb.RoomOption, clientInfo *pb.ClientInfo) (*JoinedResponse, error) {
	appExists := false
	for _, app := range rs.apps {
		if app.Id == appID {
			appExists = true
			break
		}
	}
	if !appExists {
		return nil, xerrors.Errorf("Unknown appID: %v", appID)
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
		AppId:      appID,
		RoomOption: roomOption,
		MasterInfo: clientInfo,
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		fmt.Printf("create room error: %v", err)
		return nil, err
	}

	log.Infof("Created room: %v", res.RoomInfo)

	return NewJoinedResponse(&game, res.RoomInfo, []*pb.ClientInfo{res.MasterInfo}), nil
}
