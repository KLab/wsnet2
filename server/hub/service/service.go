package service

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"wsnet2/config"
	"wsnet2/hub"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	registerQuery = "" +
		"INSERT INTO `hub_server` (`hostname`, `public_name`, `grpc_port`, `ws_port`, `status`) VALUES (:hostname, :public_name, :grpc_port, :ws_port, :status) " +
		"ON DUPLICATE KEY UPDATE `public_name`=:public_name, `grpc_port`=:grpc_port, `ws_port`=:ws_port, `status`=:status, id=last_insert_id(id)"
	heartbeatQuery = "" +
		"UPDATE `hub_server` SET `status`=:status, heartbeat=:now WHERE `id`=:hostid"

	HostStatusStarting = 0
	HostStatusRunning  = 1
)

type HubService struct {
	pb.UnimplementedGameServer // Create, Join の空実装

	HostId int64

	conf *config.GameConf
	repo *hub.Repository

	db          *sqlx.DB
	preparation sync.WaitGroup

	wsURLFormat string
}

func New(db *sqlx.DB, conf *config.GameConf) (*HubService, error) {
	hostId, err := registerHost(db, conf)
	if err != nil {
		return nil, err
	}

	repo, err := hub.NewRepository(db, conf, uint32(hostId))
	if err != nil {
		return nil, err
	}

	return &HubService{HostId: hostId, conf: conf, repo: repo, db: db}, nil
}

func registerHost(db *sqlx.DB, conf *config.GameConf) (int64, error) {
	bind := map[string]interface{}{
		"hostname":    conf.Hostname,
		"public_name": conf.PublicName,
		"grpc_port":   conf.GRPCPort,
		"ws_port":     conf.WebsocketPort,
		"status":      HostStatusRunning,
	}
	res, err := sqlx.NamedExec(db, registerQuery, bind)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *HubService) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error
	select {
	case <-ctx.Done():
	case err = <-s.heartbeat(ctx):
	case err = <-s.servePprof(ctx):
	case err = <-s.serveGRPC(ctx):
	case err = <-s.serveWebSocket(ctx):
	}
	return err
}

// heartbeat :
// TODO: 一時停止する仕組みが必要
func (s *HubService) heartbeat(ctx context.Context) <-chan error {
	wait := make(chan struct{})
	go func() {
		s.preparation.Wait()
		close(wait)
	}()

	errCh := make(chan error)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-wait:
		}

		log.Debugf("heartbeat start")
		t := time.NewTicker(time.Duration(s.conf.HeartBeatInterval))
		bind := map[string]interface{}{
			"hostid": s.HostId,
			"status": HostStatusRunning,
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}

			bind["now"] = time.Now().Unix()
			if _, err := sqlx.NamedExec(s.db, heartbeatQuery, bind); err != nil {
				errCh <- err
				return
			}
		}
	}()

	return errCh
}