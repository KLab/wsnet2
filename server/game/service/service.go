package service

import (
	"context"

	"github.com/jmoiron/sqlx"

	"wsnet2/config"
	"wsnet2/game"
	"wsnet2/pb"
)

type GameService struct {
	conf  *config.GameConf
	repos map[pb.AppId]*game.Repository
}

func New(db *sqlx.DB, conf *config.GameConf) (*GameService, error) {
	repos, err := game.NewRepos(db, conf)
	if err != nil {
		return nil, err
	}
	return &GameService{
		conf:  conf,
		repos: repos,
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
	}
	return err
}
