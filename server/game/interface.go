// clientとRoomを切り離すための抽象化レイヤー
package game

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type RoomID string

type IRoom interface {
	ID() RoomID
	Repo() IRepo

	Deadline() time.Duration
	WaitGroup() *sync.WaitGroup
	Logger() *zap.SugaredLogger

	// Done returns a channel which cloased when room is done.
	Done() <-chan struct{}

	// Timeout : client側でtimeout検知したとき. Client.MsgLoopから呼ばれる
	Timeout(c *Client)

	SendMessage(msg Msg)
}

type IRepo interface {
	RemoveClient(c *Client)
}
