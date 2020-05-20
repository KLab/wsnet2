package service

import (
	"context"

	"github.com/jmoiron/sqlx"

	"wsnet2/config"
)

type LobbyService struct {
	conf  *config.LobbyConf
}

func New(db *sqlx.DB, conf *config.LobbyConf) (*LobbyService, error) {
	return &LobbyService{
		conf: conf,
	}, nil
}

func (s *LobbyService) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error
	select {
	case <-ctx.Done():
	case err = <-s.serveAPI(ctx):
	case err = <-s.servePprof(ctx):
	}
	return err
}
