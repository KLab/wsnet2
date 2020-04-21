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
	waitPeer chan *Peer
	newPeer  chan *Peer

	evErr chan error
}

func NewClient(info *pb.ClientInfo, room *Room) *Client {
	c := &Client{
		ClientInfo: info,
		room:       room,

		removed:     make(chan struct{}),
		done:        make(chan struct{}),
		newDeadline: make(chan time.Duration),

		evbuf: NewEvBuf(ClientEventBufSize),

		waitPeer: make(chan *Peer, 1),
		newPeer:  make(chan *Peer, 1),

		evErr: make(chan error),
	}

	go c.MsgLoop(room.deadline)
	go c.EventLoop()

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
			c.room.logger.Infof("client timeout: client=%v", c.Id)
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
				c.room.logger.Infof("Peer detached: client=%v, peer=%p", c.Id, peer)
				peerMsgCh = nil
				continue
			}
			c.room.logger.Infof("New peer attached: client=%v, peer=%p", c.Id, peer)
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
		case err := <-c.evErr:
			c.room.logger.Debugf("error from EventLoop: client=%v %v", c.Id, err)
			c.room.msgCh <- MsgClientError{
				Client: c,
				Err:    err,
			}
			break loop
		}
	}
	c.room.logger.Debugf("Client.MsgLoop close: client=%v", c.Id)
	close(c.done)

	go func() {
		time.Sleep(ClientWaitAfterClose)
		c.room.repo.RemoveClient(c)
	}()

	c.drainMsg(peerMsgCh)
	c.room.wgClient.Done()
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
	c.peer.Close("removed from room")
}

// RoomのMsgLoopから呼ばれる
func (c *Client) Send(e Event) error {
	c.room.logger.Debugf("client.send: client=%v %T", c.Id, e)
	return c.evbuf.Write(e)
}

// attachPeer: peerを紐付ける
//  peerのgoroutineから呼ばれる
func (c *Client) AttachPeer(p *Peer) error {
	c.room.logger.Debugf("attach peer: client=%v, peer=%p", c.Id, p)
	c.mu.Lock()
	defer c.mu.Unlock()
	// TODO: seqnumをpeerに通知(送信までする)

	// 未読Eventを再送. client終了後でも送信する.
	evs, last, err := c.evbuf.Read(p.LastEventSeq())
	if err != nil {
		c.room.logger.Debugf("attach error: client%v, peer=%p %v", c.Id, p, err)
		return err
	}
	if err := p.SendEvent(evs, last); err != nil {
		return err
	}

	if c.peer == nil {
		c.waitPeer <- p
	} else {
		c.peer.Detached()
	}
	c.peer = p
	c.newPeer <- p
	return nil
}

// detachPeer: peerを切り離す
// peer側で切断やエラーを検知したときにpeerのgoroutineから呼ばれる.
func (c *Client) DetachPeer(p *Peer) {
	c.room.logger.Debugf("detach peer: client=%v, peer=%p", c.Id, p)
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.peer != p {
		return // すでにdetach済み
	}
	c.peer.Detached()
	c.peer = nil
	c.newPeer <- nil
	c.waitPeer = make(chan *Peer, 1)
}

func (c *Client) getWritePeer() (*Peer, <-chan *Peer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.peer, c.waitPeer
}

func (c *Client) EventLoop() {
	c.room.logger.Debugf("client.EventLoop start: client=%v", c.Id)
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
			// peerがattachされるまで待つ
			c.room.logger.Debugf("client EventLoop: wait peer. client=%v", c.Id)
			select {
			case <-c.done:
				break loop
			case peer = <-wait:
				c.room.logger.Debugf("client EventLoop: peer available. client=%v, peer=%p", c.Id, peer)
			}
		}

		evs, last, err := c.evbuf.Read(peer.LastEventSeq())
		if err != nil {
			// 端末側の持っているLastEventSeqが古すぎる. 基本的に復帰不能
			c.room.logger.Errorf("evbuf.Read error: client=%v, peer=%p: %v", c.Id, peer, err)
			peer.ClientError(err)
			c.evErr <- err
			break loop
		}

		if err := peer.SendEvent(evs, last); err != nil {
			// 送信失敗は新しいpeerなら復帰できるかもしれない.
			c.room.logger.Infof("peer SendEvent error, detach: client=%v, peer=%p: %v", c.Id, peer, err)
			c.DetachPeer(peer)
		}
	}

	c.room.logger.Debugf("client EventLoop finish: client=%v", c.Id)
}
