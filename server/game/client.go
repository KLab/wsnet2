package game

import (
	"sync"
	"time"

	"wsnet2/pb"
)

const (
	ClientEventBufSize = 64

	// 部屋が終了した後で再接続が来た時もバッファに残ったデータを送信できるので一定時間残す
	ClientWaitAfterClose = time.Second * 30
)

type ClientID string

type Client struct {
	*pb.ClientInfo
	room *Room

	removed     chan struct{}
	done        chan struct{}
	newDeadline chan time.Duration

	evbuf *EvBuf

	mu       sync.RWMutex
	peer     *Peer
	waitPeer chan struct{}
	newPeer  chan *Peer
}

func NewClient(info *pb.ClientInfo, room *Room) *Client {
	c := &Client{
		ClientInfo: info,
		room:       room,

		removed:     make(chan struct{}),
		done:        make(chan struct{}),
		newDeadline: make(chan time.Duration),

		waitPeer: make(chan struct{}),
		newPeer:  make(chan *Peer),

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
	c.room.logger.Debugf("Client.MsgLoop start: client=%v", c.Id)
	var peerMsgCh <-chan Msg
	t := time.NewTimer(deadline)
loop:
	for {
		select {
		case <-t.C:
			c.room.logger.Debugf("client timeout: client=%v", c.Id)
			c.room.Timeout(c)
			break loop

		case <-c.room.Done():
			c.room.logger.Debugf("room done: client=%v", c.Id)
			if !t.Stop() {
				<-t.C
			}
			break loop

		case <-c.removed:
			c.room.logger.Debugf("client removed: client=%v", c.Id)
			if !t.Stop() {
				<-t.C
			}
			break loop

		case deadline = <-c.newDeadline:
			c.room.logger.Debugf("new deadline: client=%v, deadline=%v", c.Id, deadline)
			if !t.Stop() {
				<-t.C
			}
			t.Reset(deadline)

		case peer := <-c.newPeer:
			go c.drainMsg(peerMsgCh)
			if peer == nil {
				c.room.logger.Debugf("peer detached: client=%v, peer=%p", c.Id, peer)
				peerMsgCh = nil
				continue
			}
			c.room.logger.Debugf("assign new peer: client=%v, peer=%p", c.Id, peer)
			if !t.Stop() {
				<-t.C
			}
			peerMsgCh = peer.MsgCh()
			t.Reset(deadline)

		case m := <-peerMsgCh:
			c.room.logger.Debugf("peer message: client=%v %v", c.Id, m)
			if !t.Stop() {
				<-t.C
			}
			c.room.msgCh <- m
			t.Reset(deadline)
		}
	}
	c.room.logger.Debugf("Client.MsgLoop close: client=%v", c.Id)
	close(c.done)

	go func() {
		time.Sleep(ClientWaitAfterClose)
		c.room.repo.RemoveClient(c)
	}()

	c.drainMsg(peerMsgCh)
	c.room.logger.Debugf("Client.MsgLoop finish: client=%v", c.Id)
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

// RoomのMsgLoopから呼ばれる
func (c *Client) Send(e Event) error {
	return c.evbuf.Write(e)
}

// attachPeer: peerを紐付ける
//  peerのgoroutineから呼ばれる
func (c *Client) attachPeer(p *Peer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// TODO: seqnumをpeerに通知(送信までする)

	if c.peer == nil {
		close(c.waitPeer)
	}
	c.peer = p
	c.newPeer <- p
}

// detachPeer: peerを切り離す
// peer側で切断やエラーを検知したときにpeerのgoroutineから呼ばれる.
func (c *Client) detachPeer(p *Peer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.peer != p {
		return // すでにdetach済み
	}
	c.peer = nil
	c.newPeer <- nil
	c.waitPeer = make(chan struct{})
}

func (c *Client) getWritePeer() (*Peer, <-chan struct{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.peer, c.waitPeer
}

func (c *Client) EventLoop() {
loop:
	for {
		// dataが来るまで待つ
		select {
		case <-c.done:
			break loop
		case <-c.evbuf.HasData():
			c.room.logger.Debugf("client EventLoop: data available. client=%v", c.Id)
		}

		peer, wait := c.getWritePeer()
		if peer == nil {
			c.room.logger.Debugf("client EventLoop: wait peer. client=%v", c.Id)
			select {
			case <-c.done:
				break loop
			case <-wait:
				c.room.logger.Debugf("client EventLoop: peer available. client=%v", c.Id)
				continue
			}
		}

		evs, last := c.evbuf.Read()
		c.room.logger.Debugf("client EventLoop: send event %v - %v, client=%v", last-len(evs)+1, last, c.Id)
		peer.SendEvent(c.evbuf.Read())
	}

	c.room.logger.Debugf("client EventLoop finish: client=%v", c.Id)
}
