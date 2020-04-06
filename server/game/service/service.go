package service

import (
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

func (s *GameService) Serve() error {
	var err error
	select{
	case err = <-s.grpcServe():
	}
	return err
}
