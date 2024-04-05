package lobby

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sort"
	"sync"
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
	gameCache *gameCache
	hubCache  *hubCache
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
		gameCache: newGameCache(db, time.Second*1, time.Duration(conf.ValidHeartBeat)),
		hubCache:  newHubCache(db, time.Second*1, time.Duration(conf.ValidHeartBeat)),
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

	client, err := rs.newGameClient(game.Hostname, game.GRPCPort)
	if err != nil {
		return nil, xerrors.Errorf("newGameClient: %w", err)
	}

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

func filter(rooms []*pb.RoomInfo, props []binary.Dict, queries []PropQueries, limit int, checkJoinable, checkWatchable bool, logger log.Logger) []*pb.RoomInfo {
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
				match := q.match(props[i], logger)
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
	return filtered
}

func (rs *RoomService) join(ctx context.Context, appId, roomId string, clientInfo *pb.ClientInfo, macKey string, hostId uint32) (*pb.JoinedRoomRes, error) {
	game, err := rs.gameCache.Get(hostId)
	if err != nil {
		return nil, xerrors.Errorf("get game server(%v): %w", hostId, err)
	}

	client, err := rs.newGameClient(game.Hostname, game.GRPCPort)
	if err != nil {
		return nil, xerrors.Errorf("newGameClient: %w", err)
	}

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

func (rs *RoomService) JoinById(ctx context.Context, appId, roomId string, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string, logger log.Logger) (*pb.JoinedRoomRes, error) {
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

	filtered := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, true, false, logger)
	if len(filtered) == 0 {
		return nil, withType(
			xerrors.Errorf("filter result is empty: room=%v", roomId),
			ErrNoJoinableRoom)
	}

	return rs.join(ctx, appId, filtered[0].Id, clientInfo, macKey, filtered[0].HostId)
}

func (rs *RoomService) JoinByNumber(ctx context.Context, appId string, roomNumber int32, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string, logger log.Logger) (*pb.JoinedRoomRes, error) {
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

	filtered := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, true, false, logger)
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
	filtered := filter(rooms, props, queries, 1000, true, false, logger)

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

func (rs *RoomService) Search(ctx context.Context, appId string, searchGroup uint32, queries []PropQueries, limit int, joinable, watchable bool, logger log.Logger) ([]*pb.RoomInfo, error) {
	rooms, props, err := rs.roomCache.GetRooms(ctx, appId, searchGroup)
	if err != nil {
		return nil, xerrors.Errorf("get rooms (group=%v): %w", searchGroup, err)
	}

	return filter(rooms, props, queries, limit, joinable, watchable, logger), nil
}

func (rs *RoomService) SearchByIds(ctx context.Context, appId string, roomIds []string, queries []PropQueries, logger log.Logger) ([]*pb.RoomInfo, error) {
	if len(roomIds) == 0 {
		return []*pb.RoomInfo{}, nil
	}

	sql, params, err := sqlx.In("SELECT * FROM room WHERE app_id = ? AND id IN (?)", appId, roomIds)
	if err != nil {
		return nil, xerrors.Errorf("sqlx.In: %w", err)
	}

	return rs.searchBySQL(ctx, sql, params, queries, logger)
}

func (rs *RoomService) SearchByNumbers(ctx context.Context, appId string, roomNumbers []int32, queries []PropQueries, logger log.Logger) ([]*pb.RoomInfo, error) {
	if len(roomNumbers) == 0 {
		return []*pb.RoomInfo{}, nil
	}

	sql, params, err := sqlx.In("SELECT * FROM room WHERE app_id = ? AND number IN (?)", appId, roomNumbers)
	if err != nil {
		return nil, xerrors.Errorf("sqlx.In: %w", err)
	}

	return rs.searchBySQL(ctx, sql, params, queries, logger)
}

func (rs *RoomService) SearchCurrentRooms(ctx context.Context, appId, clientId string, queries []PropQueries, logger log.Logger) ([]*pb.RoomInfo, error) {
	allGameServers, err := rs.gameCache.All()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var roomIds []string

	wg.Add(len(allGameServers))
	for _, svr := range allGameServers {
		go func(svr *gameServer) {
			defer wg.Done()

			client, err := rs.newGameClient(svr.Hostname, svr.GRPCPort)
			if err != nil {
				logger.Errorf("SearchCurrentRooms: newGameClient: %+v", err)
				return
			}

			req := &pb.CurrentRoomsReq{
				AppId:    appId,
				ClientId: clientId,
			}
			res, err := client.CurrentRooms(ctx, req)
			if err != nil {
				logger.Errorf("SearchCurrentRooms: app=%q client=%q host=%q err=%+v", appId, clientId, svr.Hostname, err)
				return
			}
			if len(res.RoomIds) > 0 {
				mu.Lock()
				roomIds = append(roomIds, res.RoomIds...)
				mu.Unlock()
			}
		}(svr)
	}

	wg.Wait()

	if len(roomIds) == 0 {
		return []*pb.RoomInfo{}, nil
	}

	sql, params, err := sqlx.In("SELECT * FROM room WHERE app_id = ? AND id IN (?)", appId, roomIds)
	if err != nil {
		return nil, xerrors.Errorf("sqlx.In: %w", err)
	}

	rooms, err := rs.searchBySQL(ctx, sql, params, queries, logger)
	if err != nil {
		return nil, xerrors.Errorf("searchBySQL: %w", err)
	}

	sort.Slice(rooms, func(i, j int) bool {
		ti := rooms[i].Created.Timestamp.AsTime()
		tj := rooms[j].Created.Timestamp.AsTime()
		return ti.Before(tj)
	})

	return rooms, nil
}

func (rs *RoomService) searchBySQL(ctx context.Context, sql string, params []any, queries []PropQueries, logger log.Logger) ([]*pb.RoomInfo, error) {
	var rooms []*pb.RoomInfo
	err := rs.db.SelectContext(ctx, &rooms, sql, params...)
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
	return filter(rooms, props, queries, len(rooms), false, false, logger), nil
}

func (rs *RoomService) watch(ctx context.Context, room *pb.RoomInfo, clientInfo *pb.ClientInfo, macKey string) (*pb.JoinedRoomRes, error) {
	var hubIDs []uint32
	err := rs.db.Select(&hubIDs, "SELECT `host_id` FROM `hub` WHERE `room_id`=? AND `watchers`<?", room.Id, rs.conf.HubMaxWatchers)
	if err != nil {
		return nil, xerrors.Errorf("select hub: %w", err)
	}

	var hub *hubServer
	if len(hubIDs) > 0 {
		n := rand.IntN(len(hubIDs))
		hub, err = rs.hubCache.Get(hubIDs[n])
	} else {
		hub, err = rs.hubCache.Rand()
	}
	if err != nil {
		return nil, xerrors.Errorf("get hub server: %w", err)
	}

	client, err := rs.newGameClient(hub.Hostname, hub.GRPCPort)
	if err != nil {
		return nil, xerrors.Errorf("newGameClient: %w", err)
	}

	game, err := rs.gameCache.Get(room.HostId)
	if err != nil {
		return nil, xerrors.Errorf("get game server: %w", err)
	}

	req := &pb.JoinRoomReq{
		AppId:      room.AppId,
		RoomId:     room.Id,
		ClientInfo: clientInfo,
		MacKey:     macKey,
		GrpcHost:   fmt.Sprintf("%s:%d", game.Hostname, game.GRPCPort),
		WsHost:     fmt.Sprintf("%s:%d", game.Hostname, game.WebSocketPort),
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

func (rs *RoomService) WatchById(ctx context.Context, appId, roomId string, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string, logger log.Logger) (*pb.JoinedRoomRes, error) {
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

	filtered := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, false, true, logger)
	if len(filtered) == 0 {
		return nil, withType(
			xerrors.Errorf("filter result is empty: room=%v", roomId),
			ErrNoWatchableRoom)
	}

	return rs.watch(ctx, filtered[0], clientInfo, macKey)
}

func (rs *RoomService) WatchByNumber(ctx context.Context, appId string, roomNumber int32, queries []PropQueries, clientInfo *pb.ClientInfo, macKey string, logger log.Logger) (*pb.JoinedRoomRes, error) {
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

	filtered := filter([]*pb.RoomInfo{&room}, []binary.Dict{props}, queries, 1, false, true, logger)
	if len(filtered) == 0 {
		return nil, withType(
			xerrors.Errorf("filter result is empty: number=%v", roomNumber),
			ErrNoWatchableRoom)
	}

	return rs.watch(ctx, filtered[0], clientInfo, macKey)
}

func (rs *RoomService) AdminKick(ctx context.Context, appId, targetID string, logger log.Logger) error {
	if _, found := rs.apps[appId]; !found {
		return xerrors.Errorf("Unknown appId: %v", appId)
	}

	go rs.adminKick(appId, targetID, logger)
	return nil
}

func (rs *RoomService) adminKick(appID, targetID string, logger log.Logger) {
	allGameServers, err := rs.gameCache.All()
	if err != nil {
		logger.Errorf("adminKick: get all game servers: %+v", err)
		return
	}

	for _, game := range allGameServers {
		client, err := rs.newGameClient(game.Hostname, game.GRPCPort)
		if err != nil {
			logger.Errorf("adminKick: newGameClient: %+v", err)
			continue
		}

		req := &pb.KickReq{
			AppId:    appID,
			RoomId:   "",
			ClientId: targetID,
		}
		_, err = client.Kick(context.Background(), req)
		if err != nil {
			logger.Errorf("adminKick: app=%q target=%q host=%q err=%+v", appID, targetID, game.Hostname, err)
			continue
		}
	}

}

func (rs *RoomService) newGameClient(host string, port int) (pb.GameClient, error) {
	grpcAddr := fmt.Sprintf("%s:%d", host, port)
	conn, err := rs.grpcPool.Get(grpcAddr)
	if err != nil {
		return nil, xerrors.Errorf("grpcPool.Get(%v): %w", grpcAddr, err)
	}

	client := pb.NewGameClient(conn)

	return client, nil
}
