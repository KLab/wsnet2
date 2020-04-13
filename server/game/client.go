package game

import (
	"sync"
	"time"

	"wsnet2/pb"
)

const (
	ClientEventBufSize = 64 // todo: 設定化
)

type ClientID string

type Client struct {
	*pb.ClientInfo

	room    *Room
	removed chan struct{}
	done    chan struct{}

	evbuf *EvBuf

	mu   sync.Mutex
	peer *Peer
}

func NewClient(info *pb.ClientInfo, room *Room) *Client {
	return &Client{
		ClientInfo: info,

		room:    room,
		removed: make(chan struct{}),
		done:    make(chan struct{}),
		evbuf:   NewEvBuf(ClientEventBufSize),
	}
}

func (c *Client) ID() ClientID {
	return ClientID(c.Id)
}

// MsgLoop goroutine.
func (c *Client) MsgLoop(deadline time.Duration) {
	t := time.NewTimer(deadline)
loop:
	for {
		select {
		case <-t.C:
			c.room.Timeout(c)
			break loop
		case <-c.room.Done():
			if !t.Stop() {
				<-t.C
			}
			break loop
		case <-c.removed:
			if !t.Stop() {
				<-t.C
			}
			break loop
		// todo: deadline変更をchanで受け取る
		case m := <-c.peer.MsgCh:
			if !t.Stop() {
				<-t.C
			}
			c.room.msgCh <- m
			t.Reset(deadline)
		}
	}

	close(c.done)
	for {
		select {
		case <-c.peer.MsgCh:
		default:
			return
		}
	}
}

// RoomのMsgLoopから呼ばれる
func (c *Client) Removed() {
	close(c.removed)
}

func (c *Client) Send(e Event) error {
	return c.evbuf.Write(e)
}
