package game

import (
	"sync"
	"time"
)

const (
	ClientEventBufSize = 64
)

type ClientID string

type Client struct {
	ID      ClientID
	room    *Room
	removed chan struct{}
	done    chan struct{}

	evbuf *EvBuf

	muPeer  sync.Mutex
	peer    *Peer
	peerMsg chan Msg
}

func NewClient(id ClientID, room *Room) *Client {
	return &Client{
		ID:      id,
		room:    room,
		removed: make(chan struct{}),
		done:    make(chan struct{}),
		evbuf:   NewEvBuf(ClientEventBufSize),
	}
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
		case m := <-c.peerMsg:
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
		case <-c.peerMsg:
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
