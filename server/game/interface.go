// clientとRoomを切り離すための抽象化レイヤー
package game

import (
	"sync"
	"time"
	"wsnet2/config"
	"wsnet2/log"
)

type RoomID string

type IRoom interface {
	ID() RoomID
	Repo() IRepo

	ClientConf() *config.ClientConf

	Deadline() time.Duration
	WaitGroup() *sync.WaitGroup
	Logger() log.Logger

	// Done returns a channel which cloased when room is done.
	Done() <-chan struct{}

	SendMessage(msg Msg)
}

type IRepo interface {
	RemoveClient(c *Client)
	PlayerLog(c *Client, msg PlayerLogMsg)
}
