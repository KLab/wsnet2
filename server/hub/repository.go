package hub

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"wsnet2/common"
	"wsnet2/config"
	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/metrics"
	"wsnet2/pb"
)

type AppID = pb.AppId
type RoomID = game.RoomID
type ClientID = game.ClientID

type Repository struct {
	hostId uint32

	conf     *config.HubConf
	db       *sqlx.DB
	grpcPool *common.GrpcPool

	muhubs sync.RWMutex
	hubs   map[RoomID]*Hub

	muclients sync.RWMutex
	clients   map[ClientID]map[RoomID]*game.Client
}

func NewRepository(db *sqlx.DB, conf *config.HubConf, hostId uint32) (*Repository, error) {
	// レコードが残っていると再起動したとき元いた部屋に入れないので削除しておく
	if _, err := db.Exec("DELETE FROM hub WHERE `host_id` = ?", hostId); err != nil {
		return nil, xerrors.Errorf("delete rooms: %w", err)
	}

	repo := &Repository{
		hostId:   hostId,
		conf:     conf,
		db:       db,
		grpcPool: common.NewGrpcPool(grpc.WithTransportCredentials(insecure.NewCredentials())),

		hubs:    make(map[RoomID]*Hub),
		clients: make(map[ClientID]map[RoomID]*game.Client),
	}
	return repo, nil
}

func (r *Repository) insertHub(ctx context.Context, tx sqlx.ExecerContext, roomId RoomID) (int64, error) {
	res, err := tx.ExecContext(ctx,
		"INSERT INTO `hub` (`host_id`, `room_id`, `watchers`, `created`) VALUES (?,?,?,?)",
		r.hostId, string(roomId), 0, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) deleteHub(hub *Hub) {
	_, err := r.db.Exec("DELETE FROM `hub` WHERE `id` = ?", hub.hubPK)
	if err != nil {
		hub.logger.Errorf("delete from db: %v", err)
	}
}

func (r *Repository) updateHubWatchers(hub *Hub, watchers int) {
	_, err := r.db.Exec("UPDATE `hub` SET `watchers`= ? WHERE `id` = ?", watchers, hub.hubPK)
	if err != nil {
		hub.logger.Errorf("update hub.watchers: %v", err)
	}
}

func (r *Repository) getOrCreateHub(ctx context.Context, appId AppID, roomId RoomID, grpcHost, wsHost string) (_ *Hub, err error) {
	r.muhubs.Lock()
	defer r.muhubs.Unlock()
	hub, ok := r.hubs[roomId]
	if !ok {
		logger := log.Get(log.CurrentLevel()).With(log.KeyApp, appId, log.KeyRoom, roomId)
		logger.Infof("create new hub: app=%v room=%v", appId, roomId)

		grpc, err := r.grpcPool.Get(grpcHost)
		if err != nil {
			return nil, xerrors.Errorf("grpcPool get: %w", err)
		}

		tx, err := r.db.Begin()
		if err != nil {
			return nil, xerrors.Errorf("db.Begin: %w", err)
		}
		pk, err := r.insertHub(ctx, tx, roomId)
		if err != nil {
			tx.Rollback()
			return nil, xerrors.Errorf("insert into hub: %w")
		}

		hub, err = NewHub(r, pk, appId, roomId, grpc, wsHost, logger)
		if err != nil {
			tx.Rollback()
			return nil, xerrors.Errorf("new hub: %w", err)
		}

		err = tx.Commit()
		if err != nil {
			return nil, xerrors.Errorf("commit: %w", err)
		}

		r.hubs[roomId] = hub
		metrics.Hubs.Add(1)

		go func() {
			<-hub.Done()
			delete(r.hubs, roomId)
			r.deleteHub(hub)
			logger.Infof("hub removed: room=%v", roomId)
			metrics.Hubs.Add(-1)
		}()
	}

	return hub, nil
}

func (r *Repository) WatchRoom(ctx context.Context, appId AppID, roomId RoomID, client *pb.ClientInfo, grpcHost, wsHost, macKey string) (*pb.JoinedRoomRes, game.ErrorWithCode) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	r.muclients.RLock()
	clients := len(r.clients)
	r.muclients.RUnlock()
	if clients >= r.conf.MaxClients {
		return nil, game.WithCode(
			xerrors.Errorf("reached to the max_clients"), codes.ResourceExhausted)
	}

	hub, err := r.getOrCreateHub(ctx, appId, roomId, grpcHost, wsHost)
	if err != nil {
		return nil, game.WithCode(xerrors.Errorf("getOrCreateHub: %w", err), codes.NotFound)
	}

	jch := make(chan *game.JoinedInfo, 1)
	errch := make(chan game.ErrorWithCode, 1)
	msg := &game.MsgWatch{
		Info:   client,
		MACKey: macKey,
		Joined: jch,
		Err:    errch,
	}
	select {
	case <-hub.Done():
		return nil, game.WithCode(
			xerrors.Errorf("hub closed: room=%v client=%v", roomId, client.Id),
			codes.NotFound)
	case <-ctx.Done():
		return nil, game.WithCode(
			xerrors.Errorf("send MsgWatch timeout or context done: room=%v client=%v", roomId, client.Id),
			codes.DeadlineExceeded)
	case hub.msgCh <- msg:
	}

	var joined *game.JoinedInfo
	select {
	case <-hub.Done():
		return nil, game.WithCode(
			xerrors.Errorf("hub closed: room=%v client=%v", roomId, client.Id),
			codes.NotFound)
	case <-ctx.Done():
		return nil, game.WithCode(
			xerrors.Errorf("waiting JoinedInfo timeout or context done: room=%v client=%v", roomId, client.Id),
			codes.DeadlineExceeded)
	case ewc := <-errch:
		return nil, game.WithCode(xerrors.Errorf("WatchRoom: %w", ewc), ewc.Code())
	case joined = <-jch:
	}

	cli := joined.Client

	r.muclients.Lock()
	defer r.muclients.Unlock()
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
	r.muclients.Lock()
	defer r.muclients.Unlock()
	cid := cli.ID()
	rid := cli.RoomID()
	if cmap, ok := r.clients[cid]; ok {
		// IDが同じでも別クライアントの場合には削除しない
		if c, ok := cmap[rid]; ok && c != cli {
			c.Logger().Debugf("cannot remove client from repository (already replaced new client): room=%v, client=%v", rid, cid)
			return
		}
		delete(cmap, rid)
		if len(cmap) == 0 {
			delete(r.clients, cid)
		}
	}
	cli.Logger().Debugf("client removed from repository: room=%v, client=%v", rid, cid)
}

func (r *Repository) GetClient(roomId, userId string) (*game.Client, error) {
	r.muclients.RLock()
	defer r.muclients.RUnlock()
	cli, ok := r.clients[ClientID(userId)][RoomID(roomId)]
	if !ok {
		return nil, xerrors.Errorf("client not found: room=%v, client=%v", roomId, userId)
	}
	return cli, nil
}

func (r *Repository) GetHubCount() int {
	r.muhubs.RLock()
	defer r.muhubs.RUnlock()
	return len(r.hubs)
}

func (r *Repository) PlayerLog(c *game.Client, msg game.PlayerLogMsg) {}
