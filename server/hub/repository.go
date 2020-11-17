package hub

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"wsnet2/common"
	"wsnet2/config"
	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/pb"
)

type AppID = pb.AppId
type RoomID = game.RoomID
type ClientID = game.ClientID

type Repository struct {
	hostId uint32

	conf     *config.GameConf
	db       *sqlx.DB
	grpcPool *common.GrpcPool

	gameCache *common.GameCache
	ws        *websocket.Dialer

	mu      sync.RWMutex
	hubs    map[RoomID]*Hub
	clients map[ClientID]map[RoomID]*game.Client
}

func NewRepository(db *sqlx.DB, conf *config.GameConf, hostId uint32) (*Repository, error) {
	repo := &Repository{
		hostId:   hostId,
		conf:     conf,
		db:       db,
		grpcPool: common.NewGrpcPool(grpc.WithInsecure()),

		gameCache: common.NewGameCache(db, time.Second*1, time.Duration(time.Second*5)), /* TODO: 第三引数はconfigから持ってくる（ValidHeartBeat） */
		ws: &websocket.Dialer{
			Subprotocols:    []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},

		hubs:    make(map[RoomID]*Hub),
		clients: make(map[ClientID]map[RoomID]*game.Client),
	}
	return repo, nil
}

func (r *Repository) GetOrCreateHub(appId AppID, roomId RoomID) (*Hub, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hub, ok := r.hubs[roomId]
	if ok {
		return hub, nil
	}

	hub = NewHub(r, appId, roomId)
	r.hubs[roomId] = hub

	go func() {
		hub.Start()
		r.RemoveHub(hub)
	}()

	return hub, nil
}

func (r *Repository) RemoveHub(hub *Hub) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rid := hub.ID()
	delete(r.hubs, rid)
	log.Debugf("hub removed from repository: room=%v", rid)
}

func (r *Repository) WatchRoom(ctx context.Context, appId AppID, roomId RoomID, client *pb.ClientInfo) (*pb.JoinedRoomRes, game.ErrorWithCode) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	hub, err := r.GetOrCreateHub(appId, roomId)
	if err != nil {
		return nil, game.WithCode(xerrors.Errorf("WatchRoom: %w", err), codes.NotFound)
	}

	jch := make(chan *game.JoinedInfo, 1)
	errch := make(chan game.ErrorWithCode, 1)
	msg := &game.MsgWatch{
		Info:   client,
		Joined: jch,
		Err:    errch,
	}
	select {
	case <-hub.Done():
		return nil, game.WithCode(
			xerrors.Errorf("WatchRoom hub closed: room=%v client=%v", roomId, client.Id),
			codes.Internal)
	case <-ctx.Done():
		return nil, game.WithCode(
			xerrors.Errorf("WatchRoom write msg timeout or context done: room=%v client=%v", roomId, client.Id),
			codes.DeadlineExceeded)
	case hub.msgCh <- msg:
	}

	var joined *game.JoinedInfo
	select {
	case <-hub.Done():
		return nil, game.WithCode(
			xerrors.Errorf("WatchRoom hub closed: room=%v client=%v", roomId, client.Id),
			codes.Internal)
	case <-ctx.Done():
		return nil, game.WithCode(
			xerrors.Errorf("WatchRoom timeout or context done: room=%v client=%v", roomId, client.Id),
			codes.DeadlineExceeded)
	case ewc := <-errch:
		return nil, game.WithCode(xerrors.Errorf("WatchRoom: %w", ewc), ewc.Code())
	case joined = <-jch:
	}

	cli := joined.Client

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clients[cli.ID()]; !ok {
		r.clients[cli.ID()] = make(map[RoomID]*game.Client)
	}
	r.clients[cli.ID()][roomId] = cli

	return &pb.JoinedRoomRes{
		RoomInfo: joined.Room,
		Players:  joined.Players,
		AuthKey:  cli.AuthKey(),
		MasterId: string(joined.MasterId),
		Deadline: uint32(joined.Deadline / time.Second),
	}, nil
}

func (r *Repository) RemoveClient(cli *game.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cid := cli.ID()
	rid := cli.RoomID()
	if cmap, ok := r.clients[cid]; ok {
		// IDが同じでも別クライアントの場合には削除しない
		if c, ok := cmap[rid]; ok && c != cli {
			log.Debugf("cannot remove client from repository (already replaced new client): room=%v, client=%v", rid, cid)
			return
		}
		delete(cmap, rid)
		if len(cmap) == 0 {
			delete(r.clients, cid)
		}
	}
	log.Debugf("client removed from repository: room=%v, client=%v", rid, cid)
}

func (r *Repository) GetClient(roomId, userId string) (*game.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cli, ok := r.clients[ClientID(userId)][RoomID(roomId)]
	if !ok {
		return nil, xerrors.Errorf("client not found: room=%v, client=%v", roomId, userId)
	}
	return cli, nil
}
