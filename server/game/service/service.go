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
	registerSQL = "" +
		"INSERT INTO `host` (`hostname`, `public_name`, status) VALUES (:hostname, :public_name, :status) " +
		"ON DUPLICATE KEY UPDATE `public_name`=:public_name, `status`=:status, id=last_insert_id(id)"
	heartbeatSQL = "" +
		"UPDATE `host` SET `status`=:status, heartbeat=:now WHERE `id`=:hostid"

	HostStatusStarting = 0
	HostStatusRunning  = 1
)

type GameService struct {
	HostId int64

	conf  *config.GameConf
	repos map[pb.AppId]*game.Repository

	db          *sqlx.DB
	preparation sync.WaitGroup
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
		"status":      HostStatusRunning,
	}
	res, err := sqlx.NamedExec(db, registerSQL, bind)
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
		t := time.NewTicker(time.Duration(s.conf.HeartBeatInterval) * time.Second)
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

			bind["now"] = time.Now()
			if _, err := sqlx.NamedExec(s.db, heartbeatSQL, bind); err != nil {
				errCh <- err
				return
			}
		}
	}()

	return errCh
}
