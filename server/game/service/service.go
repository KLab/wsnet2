package service

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"wsnet2/config"
	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	registerQuery = "" +
		"INSERT INTO `game_server` (`hostname`, `public_name`, `grpc_port`, `ws_port`, `status`) VALUES (:hostname, :public_name, :grpc_port, :ws_port, :status) " +
		"ON DUPLICATE KEY UPDATE `public_name`=:public_name, `grpc_port`=:grpc_port, `ws_port`=:ws_port, `status`=:status, id=last_insert_id(id)"
	heartbeatQuery = "" +
		"UPDATE `game_server` SET `status`=:status, heartbeat=:now WHERE `id`=:hostid"

	HostStatusStarting = 0
	HostStatusRunning  = 1
	HostStatusClosing  = 2
)

type GameService struct {
	pb.UnimplementedGameServer

	HostId int64

	conf  *config.GameConf
	repos map[pb.AppId]*game.Repository

	db          *sqlx.DB
	preparation sync.WaitGroup

	wsURLFormat string

	shutdownChan chan struct{}
	done         chan error
}

func New(db *sqlx.DB, conf *config.GameConf) (*GameService, error) {
	hostId, err := registerHost(db, conf)
	if err != nil {
		return nil, err
	}
	repos, err := game.NewRepos(db, conf, uint32(hostId))
	if err != nil {
		return nil, err
	}
	return &GameService{
		HostId: hostId,
		conf:   conf,
		repos:  repos,
		db:     db,

		shutdownChan: make(chan struct{}),
		done:         make(chan error),
	}, nil
}

func (s *GameService) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error
	select {
	case <-ctx.Done():
	case err = <-s.serveGRPC(ctx):
	case err = <-s.serveWebSocket(ctx):
	case err = <-s.servePprof(ctx):
	case err = <-s.heartbeat(ctx):
	case err = <-s.done:
	}
	return err
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

func (s *GameService) shutdownRequested() bool {
	select {
	case <-s.shutdownChan:
		return true
	default:
		return false
	}
}

// heartbeat :
// TODO: 一時停止する仕組みが必要
func (s *GameService) heartbeat(ctx context.Context) <-chan error {
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

			if s.shutdownRequested() {
				bind["status"] = HostStatusClosing
				log.Infof("the host is shutting down and waiting for %v rooms to be closed", s.numRooms())
			}

			if _, err := sqlx.NamedExec(s.db, heartbeatQuery, bind); err != nil {
				errCh <- err
				return
			}
		}
	}()

	return errCh
}

// Shutdown requests the termination of the GameService and waits for the serving rooms to be closed.
func (s *GameService) Shutdown(ctx context.Context) {
	log.Infof("GameService %v is gracefully shutting down", s.HostId)

	select {
	case <-s.shutdownChan:
		// Shutdown is already requested
		return
	default:
		close(s.shutdownChan)
	}

	// Immediately execute a heartbeat query in order not to miss the status update
	bind := map[string]interface{}{
		"now":    time.Now().Unix(),
		"hostid": s.HostId,
		"status": HostStatusClosing,
	}
	if _, err := sqlx.NamedExec(s.db, heartbeatQuery, bind); err != nil {
		s.done <- err
		return
	}

	// Wait for all the rooms to be closed
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		if s.numRooms() == 0 {
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

func (s *GameService) numRooms() int {
	numRooms := 0
	for _, repo := range s.repos {
		numRooms += repo.GetRoomCount()
	}
	return numRooms
}
