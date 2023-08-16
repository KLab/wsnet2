package service

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"wsnet2/common"
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
)

type HubService struct {
	pb.UnimplementedGameServer // Create, Join の空実装

	HostId int64

	conf *config.HubConf
	repo *hub.Repository

	db          *sqlx.DB
	preparation sync.WaitGroup

	wsURLFormat string

	shutdownChan chan struct{}
	done         chan error
}

func New(db *sqlx.DB, conf *config.HubConf) (*HubService, error) {
	hostId, err := registerHost(db, conf)
	if err != nil {
		return nil, err
	}

	repo, err := hub.NewRepository(db, conf, uint32(hostId))
	if err != nil {
		return nil, err
	}

	return &HubService{
		HostId:       hostId,
		conf:         conf,
		repo:         repo,
		db:           db,
		preparation:  sync.WaitGroup{},
		shutdownChan: make(chan struct{}),
		done:         make(chan error),
	}, nil
}

func registerHost(db *sqlx.DB, conf *config.HubConf) (int64, error) {
	bind := map[string]interface{}{
		"hostname":    conf.Hostname,
		"public_name": conf.PublicName,
		"grpc_port":   conf.GRPCPort,
		"ws_port":     conf.WebsocketPort,
		"status":      common.HostStatusRunning,
	}
	res, err := sqlx.NamedExec(db, registerQuery, bind)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *HubService) shutdownRequested() bool {
	select {
	case <-s.shutdownChan:
		return true
	default:
		return false
	}
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
	case err = <-s.done:
	}
	return err
}

// heartbeat :
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
			"status": common.HostStatusRunning,
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}

			bind["now"] = time.Now().Unix()
			if s.shutdownRequested() {
				bind["status"] = common.HostStatusClosing
			}
			if _, err := sqlx.NamedExec(s.db, heartbeatQuery, bind); err != nil {
				errCh <- err
				return
			}
		}
	}()

	return errCh
}

// Shutdown requests the termination of the HubService and waits for the serving hubs to be closed.
func (s *HubService) Shutdown(ctx context.Context) {
	log.Infof("HubService %v is gracefully shutting down", s.HostId)

	if s.shutdownRequested() {
		return
	}
	close(s.shutdownChan)
	defer close(s.done)

	// Immediately execute a heartbeat query in order not to miss the status update
	bind := map[string]interface{}{
		"now":    time.Now().Unix(),
		"hostid": s.HostId,
		"status": common.HostStatusClosing,
	}
	if _, err := sqlx.NamedExec(s.db, heartbeatQuery, bind); err != nil {
		s.done <- err
		return
	}

	// Wait for all the hubs to be closed
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		if s.repo.GetHubCount() == 0 {
			log.Infof("graceful shutdown completed")
			s.done <- nil
			return
		}

		select {
		case <-ctx.Done():
			s.done <- ctx.Err()
			return
		case <-ticker.C:
		}
	}
}
