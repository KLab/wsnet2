// clientとRoomを切り離すための抽象化レイヤー
package game

import (
	"sync"
	"time"
	"wsnet2/config"

	"go.uber.org/zap"
)

type RoomID string

type IRoom interface {
	ID() RoomID
	Repo() IRepo

	ClientConf() *config.ClientConf

	Deadline() time.Duration
	WaitGroup() *sync.WaitGroup
	Logger() *zap.SugaredLogger

	// Done returns a channel which cloased when room is done.
	Done() <-chan struct{}

	SendMessage(msg Msg)
}

type IRepo interface {
	RemoveClient(c *Client)
}
