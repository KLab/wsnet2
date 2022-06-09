package lobby

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
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
	hubCache  *common.HubCache
}

func NewRoomService(db *sqlx.DB, conf *config.LobbyConf) (*RoomService, error) {
	query := "SELECT id, `key` FROM app"
	var apps []*pb.App
	err := db.Select(&apps, query)
	if err != nil {
		return nil, xerrors.Errorf("select apps: %w", err)
	}
	rs := &RoomService{
		db:        db,
		conf:      conf,
		apps:      make(map[string]*pb.App),
		grpcPool:  common.NewGrpcPool(grpc.WithTransportCredentials(insecure.NewCredentials())),
		roomCache: NewRoomCache(db, time.Millisecond*10),
		gameCache: common.NewGameCache(db, time.Second*1, time.Duration(conf.ValidHeartBeat)),
		hubCache:  common.NewHubCache(db, time.Second*1, time.Duration(conf.ValidHeartBeat)),
	}
	for i, app := range apps {
		rs.apps[app.Id] = apps[i]
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

func (rs *RoomService) Create(ctx context.Context, appId string, roomOption *pb.RoomOption, clientInfo *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	game, err := rs.gameCache.Rand()
	if err != nil {
		return nil, xerrors.Errorf("get game server: %w", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort)
	conn, err := rs.grpcPool.Get(grpcAddr)
	if err != nil {
		return nil, xerrors.Errorf("get gRPC client(%s): %w", grpcAddr, err)
	}

	client := pb.NewGameClient(conn)

	req := &pb.CreateRoomReq{
		AppId:      appId,
		RoomOption: roomOption,
		MasterInfo: clientInfo,
		MacKey:     macKey,
	}

	res, err := client.Create(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		err = xerrors.Errorf("gRPC Create: %w", err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				err = withType(err, ErrArgument)
			case codes.ResourceExhausted:
				err = withType(err, ErrRoomLimit)
			}
		}
		return nil, err
	}

	return res, nil
}

func filter(rooms []*pb.RoomInfo, props []binary.Dict, queries []PropQueries, limit int, checkJoinable, checkWatchable bool) ([]*pb.RoomInfo, error) {
	if limit == 0 || limit > len(rooms) {
		limit = len(rooms)
	}
	filtered := make([]*pb.RoomInfo, 0, limit)
	for i := range rooms {
		if checkJoinable && !rooms[i].Joinable {
			continue
		}
		if checkWatchable && !rooms[i].Watchable {
			continue
		}
		if len(queries) == 0 {
			// queriesが空の場合にはマッチさせる
			filtered = append(filtered, rooms[i])
		} else {
			// queriesの何れかとマッチするか判定（OR）
			for _, q := range queries {
				match, err := q.match(props[i])
				if err != nil {
					return nil, err
				}
				if match {
					filtered = append(filtered, rooms[i])
					break
				}
			}
		}
		if len(filtered) >= limit {
			break
		}
	}
	return filtered, nil
}

func (rs *RoomService) join(ctx context.Context, appId, roomId string, clientInfo *pb.ClientInfo, macKey string, hostId uint32) (*pb.JoinedRoomRes, error) {
	game, err := rs.gameCache.Get(hostId)
	if err != nil {
		return nil, xerrors.Errorf("get game server(%s): %w", hostId, err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort)
	conn, err := rs.grpcPool.Get(grpcAddr)
	if err != nil {
		return nil, xerrors.Errorf("grpcPool.Get(%s): %w", grpcAddr, err)
	}

	client := pb.NewGameClient(conn)

	req := &pb.JoinRoomReq{
		AppId:      appId,
		RoomId:     roomId,
		ClientInfo: clientInfo,
		MacKey:     macKey,
	}

	res, err := client.Join(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		err = xerrors.Errorf("gRPC Join: %w", err)
		if ok {
			switch st.Code() {
			case codes.NotFound: // roomが既に消えた
				err = withType(err, ErrNoJoinableRoom)
			case codes.FailedPrecondition: // joinableでなくなっていた
				err = withType(err, ErrNoJoinableRoom)
			case codes.ResourceExhausted: // 満室
				err = withType(err, ErrRoomFull)
			case codes.AlreadyExists: // 既に入室している
				err = withType(err, ErrAlreadyJoined)
			case codes.InvalidArgument:
				err = withType(err, ErrArgument)
			}
		}
		return nil, err
	}

	return res, nil
}

func (rs *RoomService) JoinById(ctx context.Context, appId, roomId string, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND id = ? AND joinable = 1", appId, roomId)
	if err != nil {
		return nil, withType(
			xerrors.Errorf("select room (id=%v): %w", roomId, err),
			ErrNoJoinableRoom)
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("unmarshalProps: %w", err)
	}

	filtered, err := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, true, false)
	if err != nil {
		return nil, xerrors.Errorf("filter: %w", err)
	}
	if len(filtered) == 0 {
		return nil, withType(
			xerrors.Errorf("filter result is empty: room=%v", roomId),
			ErrNoJoinableRoom)
	}

	return rs.join(ctx, appId, filtered[0].Id, clientInfo, macKey, filtered[0].HostId)
}

func (rs *RoomService) JoinByNumber(ctx context.Context, appId string, roomNumber int32, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND number = ? AND joinable = 1", appId, roomNumber)
	if err != nil {
		return nil, withType(
			xerrors.Errorf("select room (num=%v): %w", roomNumber, err),
			ErrNoJoinableRoom)
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("unmarshalProps: %w", err)
	}

	filtered, err := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, true, false)
	if err != nil {
		return nil, xerrors.Errorf("filter: %w", err)
	}
	if len(filtered) == 0 {
		return nil, withType(
			xerrors.Errorf("filter result is empty: number=%v: %w", roomNumber, err),
			ErrNoJoinableRoom)
	}

	return rs.join(ctx, appId, filtered[0].Id, clientInfo, macKey, filtered[0].HostId)
}

func (rs *RoomService) JoinAtRandom(ctx context.Context, appId string, searchGroup uint32, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string, logger log.Logger) (*pb.JoinedRoomRes, error) {
	rooms, props, err := rs.roomCache.GetRooms(ctx, appId, searchGroup)
	if err != nil {
		return nil, xerrors.Errorf("get rooms (group=%v): %w", searchGroup, err)
	}
	filtered, err := filter(rooms, props, queries, 1000, true, false)
	if err != nil {
		return nil, xerrors.Errorf("filter: %w", err)
	}

	rand.Shuffle(len(filtered), func(i, j int) { filtered[i], filtered[j] = filtered[j], filtered[i] })

	for _, room := range filtered {
		select {
		case <-ctx.Done():
			return nil, xerrors.Errorf("context done")
		default:
		}

		res, err := rs.join(ctx, appId, room.Id, clientInfo, macKey, room.HostId)
		if err == nil {
			return res, nil
		}
		if e, ok := err.(ErrorWithType); ok {
			switch e.ErrType() {
			case ErrArgument:
				// 別の部屋でも引数エラーになるので打ち切る
				return nil, e
			}
		}
		logger.Debugf("try join %v: %v", room.Id, err)
	}

	return nil, withType(
		xerrors.Errorf("Failed to join all rooms"),
		ErrNoJoinableRoom)
}

func (rs *RoomService) Search(ctx context.Context, appId string, searchGroup uint32, queries []PropQueries, limit int, joinable, watchable bool) ([]*pb.RoomInfo, error) {
	rooms, props, err := rs.roomCache.GetRooms(ctx, appId, searchGroup)
	if err != nil {
		return nil, xerrors.Errorf("get rooms (group=%v): %w", searchGroup, err)
	}

	filtered, err := filter(rooms, props, queries, limit, joinable, watchable)
	if err != nil {
		return nil, xerrors.Errorf("filter: %w", err)
	}
	return filtered, nil
}

func (rs *RoomService) SearchByIds(ctx context.Context, appId string, roomIds []string, queries []PropQueries) ([]*pb.RoomInfo, error) {
	if len(roomIds) == 0 {
		return []*pb.RoomInfo{}, nil
	}

	sql, params, err := sqlx.In("SELECT * FROM room WHERE app_id = ? AND id IN (?)", appId, roomIds)
	if err != nil {
		return nil, xerrors.Errorf("sqlx.In: %w", err)
	}

	var rooms []*pb.RoomInfo
	err = rs.db.SelectContext(ctx, &rooms, sql, params...)
	if err != nil {
		return nil, xerrors.Errorf("Select: %w", err)
	}

	props := make([]binary.Dict, len(rooms))
	for i, r := range rooms {
		props[i], err = unmarshalProps(r.PublicProps)
		if err != nil {
			return nil, xerrors.Errorf("unmarshalProps(room=%v): %w", r.Id, err)
		}
	}
	filtered, err := filter(rooms, props, queries, len(rooms), false, false)
	if err != nil {
		return nil, xerrors.Errorf("filter: %w", err)
	}
	return filtered, nil
}

func (rs *RoomService) watch(ctx context.Context, appId, roomId string, clientInfo *pb.ClientInfo, macKey string, hostId uint32) (*pb.JoinedRoomRes, error) {
	var hostIDs []uint32
	err := rs.db.Select(&hostIDs, "SELECT `host_id` FROM `hub` WHERE `room_id`=? AND `watchers`<?", roomId, rs.conf.HubMaxWatchers)
	if err != nil {
		return nil, xerrors.Errorf("select hub: %w", err)
	}

	var hub *common.HubServer
	if len(hostIDs) > 0 {
		n := rand.Intn(len(hostIDs))
		hub, err = rs.hubCache.Get(hostIDs[n])
	} else {
		hub, err = rs.hubCache.Rand()
	}
	if err != nil {
		return nil, xerrors.Errorf("get hub server: %w", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", hub.Hostname, hub.GRPCPort)
	conn, err := rs.grpcPool.Get(grpcAddr)
	if err != nil {
		return nil, xerrors.Errorf("get gRPC client: %w", err)
	}

	client := pb.NewGameClient(conn)

	req := &pb.JoinRoomReq{
		AppId:      appId,
		RoomId:     roomId,
		ClientInfo: clientInfo,
		MacKey:     macKey,
	}

	res, err := client.Watch(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		err = xerrors.Errorf("gRPC Watch: %w", err)
		if ok {
			switch st.Code() {
			case codes.NotFound: // roomが既に消えた
				err = withType(err, ErrNoWatchableRoom)
			case codes.FailedPrecondition: // watchableでなくなっていた
				err = withType(err, ErrNoWatchableRoom)
			case codes.AlreadyExists: // 既に入室している
				err = withType(err, ErrAlreadyJoined)
			case codes.InvalidArgument:
				err = withType(err, ErrArgument)
			}
		}
		return nil, err
	}

	return res, nil
}

func (rs *RoomService) WatchById(ctx context.Context, appId, roomId string, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND id = ? AND watchable = 1", appId, roomId)
	if err != nil {
		return nil, withType(
			xerrors.Errorf("select room (id=%v): %w", roomId, err),
			ErrNoWatchableRoom)
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("unmarshalProps: %w", err)
	}

	filtered, err := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, false, true)
	if err != nil {
		return nil, xerrors.Errorf("filter: %w", err)
	}
	if len(filtered) == 0 {
		return nil, withType(
			xerrors.Errorf("filter result is empty: room=%v", roomId),
			ErrNoWatchableRoom)
	}

	return rs.watch(ctx, appId, filtered[0].Id, clientInfo, macKey, filtered[0].HostId)
}

func (rs *RoomService) WatchByNumber(ctx context.Context, appId string, roomNumber int32, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, error) {
	if _, found := rs.apps[appId]; !found {
		return nil, xerrors.Errorf("Unknown appId: %v", appId)
	}

	var room pb.RoomInfo
	err := rs.db.Get(&room, "SELECT * FROM room WHERE app_id = ? AND number = ? AND watchable = 1", appId, roomNumber)
	if err != nil {
		return nil, withType(
			xerrors.Errorf("select room (num=%v): %w", roomNumber, err),
			ErrNoWatchableRoom)
	}

	props, err := unmarshalProps(room.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("unmarshalProps: %w", err)
	}

	filtered, err := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, false, true)
	if err != nil {
		return nil, xerrors.Errorf("filter: %w", err)
	}
	if len(filtered) == 0 {
		return nil, withType(
			xerrors.Errorf("filter result is empty: number=%v", appId, roomNumber),
			ErrNoWatchableRoom)
	}

	return rs.watch(ctx, appId, filtered[0].Id, clientInfo, macKey, filtered[0].HostId)
}
