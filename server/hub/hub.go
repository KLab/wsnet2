package hub

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/game"
	"wsnet2/pb"
)

type Hub struct {
	ID       RoomID
	repo     *Repository
	clientId string

	deadline time.Duration

	publicProps  binary.Dict
	privateProps binary.Dict

	msgCh    chan game.Msg
	done     chan struct{}
	wgClient sync.WaitGroup

	muClients sync.RWMutex
	watchers  map[ClientID]*game.Client

	lastMsg binary.Dict // map[clientID]unixtime_millisec

	logger *zap.SugaredLogger
}

func (h *Hub) connectGame() error {
	var room pb.RoomInfo
	err := h.repo.db.Get(&room, "SELECT * FROM room WHERE id = ?", h.ID)
	if err != nil {
		return xerrors.Errorf("connectGame: Failed to get room: %w", err)
	}

	return xerrors.New("not implemented")
}

func (h *Hub) Start() {
	h.logger.Debug("hub start")
	defer h.logger.Debug("hub end")

	if err := h.connectGame(); err != nil {
		h.logger.Error("Failed to connect game server")
	}

	//TODO: 実装
	time.Sleep(time.Minute)
}
