package service

import (
	"context"

	"github.com/jmoiron/sqlx"

	"wsnet2/config"
	"wsnet2/lobby"
)

type LobbyService struct {
	conf        *config.LobbyConf
	roomService *lobby.RoomService
}

func New(db *sqlx.DB, conf *config.LobbyConf) (*LobbyService, error) {
	roomService := lobby.NewRoomService(db)
	return &LobbyService{
		conf:        conf,
		roomService: roomService,
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
