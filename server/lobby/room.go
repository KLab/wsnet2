package lobby

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/pb"
)

type RoomService struct {
	db       *sqlx.DB
	conf     *config.LobbyConf
	apps     map[string]*pb.App
	grpcPool *common.GrpcPool

	roomCache *RoomCache
	gameCache *common.GameCache
}

func NewRoomService(db *sqlx.DB, conf *config.LobbyConf) (*RoomService, error) {
	query := "SELECT id, `key` FROM app"
	var apps []pb.App
	err := db.Select(&apps, query)
	if err != nil {
		return nil, xerrors.Errorf("select apps error: %w", err)
	}
	rs := &RoomService{
		db:        db,
		conf:      conf,
		apps:      make(map[string]*pb.App),
		grpcPool:  common.NewGrpcPool(grpc.WithInsecure()),
		roomCache: NewRoomCache(db, time.Millisecond*10),
		gameCache: common.NewGameCache(db, time.Second*1, time.Duration(conf.ValidHeartBeat)),
	}
	for i, app := range apps {
		rs.apps[app.Id] = &apps[i]
	}
	return rs, nil
}

func (rs *RoomService) GetAppKey(appId string) (string, bool) {
	app, found := rs.apps[appId]
	if !found {
		return "", false
	}
	return app.Key, true
}

func (rs *RoomService) Create(appId string, roomOption *pb.RoomOption, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, withStatus(xerrors.Errorf("Unknown appId: %v", appId), http.StatusBadRequest, "")
	}

	game, err := rs.gameCache.Rand()
	if err != nil {
		return nil, withStatus(
			xerrors.Errorf("Create: failed to get game server: %w", err),
			http.StatusInternalServerError, "No game server found")
	}

	grpcAddr := fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort)
	conn, err := rs.grpcPool.Get(grpcAddr)
	if err != nil {
		return nil, xerrors.Errorf("Create: gRPC client connection error: %w", err)
	}

	client := pb.NewGameClient(conn)

	req := &pb.CreateRoomReq{
		AppId:      appId,
		RoomOption: roomOption,
		MasterInfo: clientInfo,
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		err = xerrors.Errorf("Create: gRPC Create: %w", err)
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				err = withStatus(err, http.StatusBadRequest, "Invalid argument")
			case codes.ResourceExhausted:
				err = withStatus(err, http.StatusServiceUnavailable, "Reached to the max room number")
			}
		}
		return nil, err
	}

	log.Infof("Created room: %v", res)

	return res, nil
}

func filter(rooms []pb.RoomInfo, props []binary.Dict, queries []PropQueries, limit int) []pb.RoomInfo {
	if limit == 0 || limit > len(rooms) {
		limit = len(rooms)
	}
	// queriesが空の場合には全件マッチさせる
	if len(queries) == 0 {
		return rooms[0:limit]
	}
	filtered := make([]pb.RoomInfo, 0, limit)
	for i, r := range rooms {
		// queriesの何れかとマッチするか判定（OR）
		for _, q := range queries {
			if q.match(props[i]) {
				filtered = append(filtered, r)
				break
			}
		}
		if len(filtered) >= limit {
			break
		}
	}
	return filtered
}

func (rs *RoomService) join(appId, roomId string, clientInfo *pb.ClientInfo, hostId uint32) (*pb.JoinedRoomRes, error) {
	game, err := rs.gameCache.Get(hostId)
	if err != nil {
		return nil, xerrors.Errorf("join: failed to get game server: %w", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort)
	conn, err := rs.grpcPool.Get(grpcAddr)
	if err != nil {
		return nil, xerrors.Errorf("join: gRPC client connection error: %w", err)
	}

	client := pb.NewGameClient(conn)

	req := &pb.JoinRoomReq{
		AppId:      appId,
		RoomId:     roomId,
		ClientInfo: clientInfo,
	}

	res, err := client.Join(context.TODO(), req)
	if err != nil {
		err = xerrors.Errorf("join: gRPC Join: %w", err)
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound: // roomが既に消えた
				err = withStatus(err, http.StatusOK, "No joinable room found")
			case codes.FailedPrecondition: // joinableでなくなっていた
				err = withStatus(err, http.StatusOK, "No joinable room found")
			case codes.ResourceExhausted: // 満室
				err = withStatus(err, http.StatusOK, "Room full")
			case codes.AlreadyExists: // 既に入室している
				err = withStatus(err, http.StatusConflict, "")
			case codes.InvalidArgument:
				err = withStatus(err, http.StatusBadRequest, "Invalid argument")
			}
		}
		return nil, err
	}

	log.Infof("Joined room: %v", res)

	return res, nil
}

func (rs *RoomService) JoinById(appId, roomId string, queries []PropQueries, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND id = ? AND joinable = 1", appId, roomId)
	if err != nil {
		return nil, withStatus(
			xerrors.Errorf("JoinById: Failed to get room: app=%v room=%v: %w", appId, roomId, err),
			http.StatusOK, "No joinable room found")
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("JoinById: unmarshalProps: %w", err)
	}

	filtered := filter([]pb.RoomInfo{room}, []binary.Dict{props}, queries, 1)
	if len(filtered) == 0 {
		return nil, withStatus(
			xerrors.Errorf("JoinById: filter result is empty: app=%v room=%v", appId, roomId),
			http.StatusOK, "No joinable room found")
	}
	room = filtered[0]

	return rs.join(appId, room.Id, clientInfo, room.HostId)
}

func (rs *RoomService) JoinByNumber(appId string, roomNumber int32, queries []PropQueries, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND number = ? AND joinable = 1", appId, roomNumber)
	if err != nil {
		return nil, withStatus(
			xerrors.Errorf("JoinByNumber: Failed to get room: app=%v number=%v: %w", appId, roomNumber, err),
			http.StatusOK, "No joinable room found")
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("JoinByNumber: unmarshalProps: %w", err)
	}

	filtered := filter([]pb.RoomInfo{room}, []binary.Dict{props}, queries, 1)
	if len(filtered) == 0 {
		return nil, withStatus(
			xerrors.Errorf("JoinByNumber: filter result is empty: app=%v number=%v: %w", appId, roomNumber, err),
			http.StatusOK, "No joinable room found")
	}
	room = filtered[0]

	return rs.join(appId, room.Id, clientInfo, room.HostId)
}

func (rs *RoomService) JoinAtRandom(appId string, searchGroup uint32, queries []PropQueries, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	rooms, err := rs.Search(appId, searchGroup, queries, 1000)
	if err != nil {
		return nil, xerrors.Errorf("JoinAtRandom: %w", err)
	}

	rand.Shuffle(len(rooms), func(i, j int) { rooms[i], rooms[j] = rooms[j], rooms[i] })

	for _, room := range rooms {
		res, err := rs.join(appId, room.Id, clientInfo, room.HostId)
		if err == nil {
			return res, nil
		}
		log.Debugf("JoinAtRandom: failed to join room: room=%v %v", room.Id, err)
	}

	return nil, withStatus(
		xerrors.Errorf("JoinAtRandom: Failed to join all rooms"),
		http.StatusOK, "No joinable room found")
}

func (rs *RoomService) Search(appId string, searchGroup uint32, queries []PropQueries, limit int) ([]pb.RoomInfo, error) {
	rooms, props, err := rs.roomCache.GetRooms(appId, searchGroup)
	if err != nil {
		return nil, xerrors.Errorf("RoomCache error: %w", err)
	}

	return filter(rooms, props, queries, limit), nil
}

func (rs *RoomService) watch(appId, roomId string, clientInfo *pb.ClientInfo, hostId uint32) (*pb.JoinedRoomRes, error) {
	game, err := rs.gameCache.Get(hostId)
	if err != nil {
		return nil, xerrors.Errorf("watch: failed to get game server: %w", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort)
	conn, err := rs.grpcPool.Get(grpcAddr)
	if err != nil {
		log.Errorf("client connection error: %v", err)
		return nil, err
	}

	client := pb.NewGameClient(conn)

	req := &pb.JoinRoomReq{
		AppId:      appId,
		RoomId:     roomId,
		ClientInfo: clientInfo,
	}

	res, err := client.Watch(context.TODO(), req)
	if err != nil {
		log.Errorf("watch room error: %v", err)
		return nil, err
	}

	log.Infof("Joined room: %v", res)

	return res, nil
}

func (rs *RoomService) WatchById(appId, roomId string, queries []PropQueries, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND id = ? AND watchable = 1", appId, roomId)
	if err != nil {
		return nil, xerrors.Errorf("WatchById: failed to get room: %w", err)
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("WatchById: unmarshalProps: %w", err)
	}

	filtered := filter([]pb.RoomInfo{room}, []binary.Dict{props}, queries, 1)
	if len(filtered) == 0 {
		return nil, xerrors.Errorf("WatchById: filter result is empty")
	}
	room = filtered[0]

	return rs.watch(appId, room.Id, clientInfo, room.HostId)
}

func (rs *RoomService) WatchByNumber(appId string, roomNumber int32, queries []PropQueries, clientInfo *pb.ClientInfo) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}
	if roomNumber == 0 {
		return nil, xerrors.Errorf("Invalid room number: %v", roomNumber)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND number = ? AND watchable = 1", appId, roomNumber)
	if err != nil {
		return nil, xerrors.Errorf("WatchByNumber: Failed to get room: %w", err)
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("WatchByNumber: unmarshalProps: %w", err)
	}

	filtered := filter([]pb.RoomInfo{room}, []binary.Dict{props}, queries, 1)
	if len(filtered) == 0 {
		return nil, xerrors.Errorf("WatchByNumber: filter result is empty")
	}
	room = filtered[0]

	return rs.watch(appId, room.Id, clientInfo, room.HostId)
}
