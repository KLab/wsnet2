package game

import (
	"sync"
	"time"

	"wsnet2/log"
	"wsnet2/pb"
)

const (
	ClientEventBufSize = 64
)

type ClientID string

type Client struct {
	*pb.ClientInfo
	room *Room

	removed     chan struct{}
	done        chan struct{}
	newDeadline chan time.Duration
	newPeer     chan *Peer

	evbuf *EvBuf

	mu   sync.Mutex
}

func NewClient(info *pb.ClientInfo, room *Room) *Client {
	c := &Client{
		ClientInfo: info,
		room:       room,

		removed:     make(chan struct{}),
		done:        make(chan struct{}),
		newDeadline: make(chan time.Duration),
		newPeer:     make(chan *Peer),

		evbuf: NewEvBuf(ClientEventBufSize),
	}

	go c.MsgLoop(room.deadline)

	return c
}

func (c *Client) ID() ClientID {
	return ClientID(c.Id)
}

// MsgLoop goroutine.
func (c *Client) MsgLoop(deadline time.Duration) {
	log.Debugf("Client.MsgLoop start: client=%v", c.Id)
	var peerMsgCh <-chan Msg
	t := time.NewTimer(deadline)
loop:
	for {
		select {
		case <-t.C:
			log.Debugf("client timeout: client=%v", c.Id)
			c.room.Timeout(c)
			break loop

		case <-c.room.Done():
			if !t.Stop() {
				<-t.C
			}
			break loop

		case <-c.removed:
			log.Debugf("client removed: client=%v", c.Id)
			if !t.Stop() {
				<-t.C
			}
			break loop

		case deadline = <-c.newDeadline:
			log.Debugf("new deadline: client=%v, deadline=%v", c.Id, deadline)
			if !t.Stop() {
				<-t.C
			}
			t.Reset(deadline)

		case peer := <-c.newPeer:
			log.Debugf("assign new peer: client=%v, peer=%v", c.Id, peer)
			if !t.Stop() {
				<-t.C
			}
			go c.drainMsg(peerMsgCh)
			peerMsgCh = peer.MsgCh
			t.Reset(deadline)

		case m := <-peerMsgCh:
			log.Debugf("peer message: client=%v %v", c.Id, m)
			if !t.Stop() {
				<-t.C
			}
			c.room.msgCh <- m
			t.Reset(deadline)
		}
	}
	log.Debugf("Client.MsgLoop close: client=%v", c.Id)
	close(c.done)

	c.drainMsg(peerMsgCh)
	log.Debugf("Client.MsgLoop finish: client=%v", c.Id)
}

func (c *Client) drainMsg(msgCh <-chan Msg) {
	if msgCh == nil {
		return
	}
	for {
		if _, ok := <-msgCh; !ok {
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
