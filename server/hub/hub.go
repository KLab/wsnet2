package hub

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"wsnet2/binary"
	"wsnet2/game"
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

func (h *Hub) Start() {
	h.logger.Debug("hub start")
	defer h.logger.Debug("hub end")

	//TODO: 実装
	time.Sleep(time.Minute)
}
