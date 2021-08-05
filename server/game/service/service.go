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

			select {
			case <-s.shutdownChan:
				bind["status"] = HostStatusClosing
			default:
			}

			if _, err := sqlx.NamedExec(s.db, heartbeatQuery, bind); err != nil {
				errCh <- err
				return
			}

			if bind["status"] == HostStatusClosing {
				n := s.numRooms()
				if n == 0 {
					log.Infof("the host is closing and no running room found, end the heartbeat")
					errCh <- nil
					return
				} else {
					log.Infof("the host is closing and waiting for %v rooms to close", n)
				}
			}
		}
	}()

	return errCh
}

func (s *GameService) Shutdown(ctx context.Context) error {
	log.Infof("GameService %v is gracefully shutdowning", s.HostId)

	select {
	case <-s.shutdownChan:
	default:
		close(s.shutdownChan)
	}

	// Immediately execute a heartbeat query in order not to leak status updates
	bind := map[string]interface{}{
		"now":    time.Now().Unix(),
		"hostid": s.HostId,
		"status": HostStatusClosing,
	}
	if _, err := sqlx.NamedExec(s.db, heartbeatQuery, bind); err != nil {
		return err
	}

	// Wait for all the rooms to close
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		if s.numRooms() == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
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
