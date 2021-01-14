package game

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"

	"wsnet2/auth"
	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/pb"
)

const (
	ClientEventBufSize = 128

	// 部屋が終了した後で再接続が来た時もバッファに残ったデータを送信できるので一定時間残す
	ClientWaitAfterClose = time.Second * 30

	ClientAuthKeyLen = 32

	// ClientAuthDataDeadline : クライアントのAuthDataの有効時間
	// client側で接続時に毎回生成するので短くて良い
	ClientAuthDataDeadline = time.Second * 10
)

type ClientID string

type Client struct {
	*pb.ClientInfo
	room IRoom

	isPlayer  bool
	nodeCount uint32

	props binary.Dict

	removed     chan struct{}
	removeCause string
	done        chan struct{}
	newDeadline chan time.Duration

	evbuf *EvBuf

	mu        sync.RWMutex
	msgSeqNum int
	peer      *Peer
	waitPeer  chan *Peer
	newPeer   chan *Peer

	authKey string

	evErr chan error
}

func NewPlayer(info *pb.ClientInfo, room IRoom) (*Client, ErrorWithCode) {
	return newClient(info, room, true)
}

func NewWatcher(info *pb.ClientInfo, room IRoom) (*Client, ErrorWithCode) {
	return newClient(info, room, false)
}

func newClient(info *pb.ClientInfo, room IRoom, isPlayer bool) (*Client, ErrorWithCode) {
	props, iProps, err := common.InitProps(info.Props)
	if err != nil {
		return nil, WithCode(
			xerrors.Errorf("Props unmarshal error: %w", err),
			codes.InvalidArgument)
	}
	info.Props = iProps
	c := &Client{
		ClientInfo: info,
		room:       room,
		isPlayer:   isPlayer,
		nodeCount:  1,

		props: props,

		removed:     make(chan struct{}),
		done:        make(chan struct{}),
		newDeadline: make(chan time.Duration, 1),

		evbuf: NewEvBuf(ClientEventBufSize),

		waitPeer: make(chan *Peer, 1),
		newPeer:  make(chan *Peer, 1),

		authKey: RandomHex(ClientAuthKeyLen),

		evErr: make(chan error),
	}

	room.WaitGroup().Add(1)

	go c.MsgLoop(room.Deadline())
	go c.EventLoop()

	return c, nil
}

func (c *Client) ID() ClientID {
	return ClientID(c.Id)
}

func (c *Client) RoomID() RoomID {
	return c.room.ID()
}

func (c *Client) AuthKey() string {
	return c.authKey
}

func (c *Client) NodeCount() uint32 {
	return c.nodeCount
}

func (c *Client) ValidAuthData(authData string) error {
	return auth.ValidAuthData(authData, c.authKey, c.Id, time.Now().Add(-ClientAuthDataDeadline))
}

// MsgLoop goroutine.
func (c *Client) MsgLoop(deadline time.Duration) {
	c.room.Logger().Debugf("Client.MsgLoop start: client=%v", c.Id)
	var peerMsgCh <-chan binary.Msg
	var curPeer *Peer
	t := time.NewTimer(deadline)
loop:
	for {
		select {
		case <-t.C:
			c.room.Logger().Infof("client timeout: client=%v", c.Id)
			c.room.Timeout(c)
			break loop

		case <-c.room.Done():
			c.room.Logger().Debugf("room done: client=%v", c.Id)
			curPeer.Close("room closed")
			if !t.Stop() {
				<-t.C
			}
			break loop

		case <-c.removed:
			c.room.Logger().Debugf("client removed: client=%v", c.Id)
			if !t.Stop() {
				<-t.C
			}
			break loop

		case newDeadline := <-c.newDeadline:
			c.room.Logger().Debugf("new deadline: client=%v, deadline=%v", c.Id, newDeadline)
			if !t.Stop() {
				<-t.C
			}
			// 突然短くされてもclientが把握できないので
			// 変更直後だけ旧deadline分の猶予をもたせる.
			t.Reset(deadline + newDeadline)
			deadline = newDeadline

		case peer := <-c.newPeer:
			go c.drainMsg(peerMsgCh)
			if peer == nil {
				c.room.Logger().Infof("Peer detached: client=%v, peer=%p", c.Id, peer)
				peerMsgCh = nil
				curPeer = nil
				continue
			}
			c.room.Logger().Infof("New peer attached: client=%v, peer=%p", c.Id, peer)
			peerMsgCh = peer.MsgCh()
			curPeer = peer
			// つなげて切るだけのクライアントをタイムアウトさせるため、t.Resetしない

		case m, ok := <-peerMsgCh:
			if !ok {
				// peer側でchをcloseした.
				c.room.Logger().Errorf("peerMsgCh closed: client=%v, peer=%p", c.Id, curPeer)
				// DetachPeerは呼ばれているはず
				peerMsgCh = nil
				curPeer = nil
				continue
			}
			c.room.Logger().Debugf("peer message: client=%v %v %v", c.Id, m.Type(), m)
			msg, err := ConstructMsg(c, m)
			if err != nil {
				// おかしなデータを送ってくるクライアントは遮断する
				c.room.SendMessage(
					&MsgClientError{
						Sender: c,
						Err:    err,
					})
				break loop
			}
			if regmsg, ok := m.(binary.RegularMsg); ok {
				seq := regmsg.SequenceNum()

				c.mu.Lock()
				cSeq := c.msgSeqNum
				valid := seq == cSeq+1
				if valid {
					c.msgSeqNum = seq
				}
				c.mu.Unlock()

				if !valid {
					// 再接続時の再送に期待して切断
					err := xerrors.Errorf("invalid sequence num: %d to %d", cSeq, seq)
					c.room.Logger().Errorf("msg error: client=%v %s", err)
					c.DetachAndClosePeer(curPeer, err)
					continue
				}
			}
			if !t.Stop() {
				<-t.C
			}
			c.room.SendMessage(msg)
			t.Reset(deadline)

		case err := <-c.evErr:
			c.room.Logger().Debugf("error from EventLoop: client=%v %v", c.Id, err)
			c.room.SendMessage(
				&MsgClientError{
					Sender: c,
					Err:    err,
				})
			break loop
		}
	}
	c.room.Logger().Debugf("Client.MsgLoop close: client=%v", c.Id)
	close(c.done)

	go func() {
		time.Sleep(ClientWaitAfterClose)
		c.room.Repo().RemoveClient(c)
	}()

	c.drainMsg(peerMsgCh)
	c.room.WaitGroup().Done()
	c.room.Logger().Debugf("Client.MsgLoop finish: client=%v", c.Id)
}

func (c *Client) drainMsg(msgCh <-chan binary.Msg) {
	if msgCh == nil {
		return
	}
	for range msgCh {
	}
}

// RoomのMsgLoopから呼ばれる
func (c *Client) Removed(cause error) {
	close(c.removed)
	c.removeCause = "client leave"
	if cause != nil {
		c.removeCause = fmt.Sprintf("removed from room: %v", cause)
	}

	c.mu.RLock()
	p := c.peer
	c.mu.RUnlock()
	if p != nil {
		p.Close(c.removeCause)
	}
}

// RoomのMsgLoopから呼ばれる
func (c *Client) Send(e *binary.RegularEvent) error {
	c.room.Logger().Debugf("client.send: client=%v %v", c.Id, e.Type())
	return c.evbuf.Write(e)
}

// RoomのMsgLoopから呼ばれる.
func (c *Client) SendSystemEvent(e *binary.SystemEvent) error {
	c.room.Logger().Debugf("client.SystemSend: client=%v %v", c.Id, e.Type)
	p := c.peer
	if p == nil {
		return xerrors.Errorf("client.SendSystemEvent: no peer attached")
	}

	// SystemEventは送信順序を問わない. 多少遅れても構わない.
	// roomのmsgloopを止めないために新しいgoroutineで送信する.
	go p.SendSystemEvent(e)

	return nil
}

// attachPeer: peerを紐付ける
//  peerのgoroutineから呼ばれる
func (c *Client) AttachPeer(p *Peer, lastEvSeq int) error {
	c.room.Logger().Debugf("attach peer: client=%v peer=%p", c.Id, p)
	c.mu.Lock()
	defer c.mu.Unlock()

	// 未読Eventを再送. client終了後でも送信する.
	if err := p.SendEvents(c.evbuf); err != nil {
		return err
	}

	select {
	case <-c.removed:
		c.room.Logger().Debugf("client has been removed: client=%v peer=%p %s", c.Id, p, c.removeCause)
		return xerrors.Errorf("client has been removed: %s", c.removeCause)
	default:
	}

	// msgSeqNumの後のメッセージから送信してもらう(再送含む)
	if err := p.SendReady(c.msgSeqNum); err != nil {
		c.room.Logger().Debugf("attach error: client=%v peer=%p %v", c.Id, p, err)
		return err
	}

	if c.peer == nil {
		c.waitPeer <- p
	} else {
		c.peer.Close("new peer attached")
	}
	c.peer = p
	c.newPeer <- p
	return nil
}

// DetachPeer : peerを切り離す.
// Peer.MsgLoopで切断やエラーを検知したときに呼ばれる.
// websocketの切断は呼び出し側で行う
func (c *Client) DetachPeer(p *Peer) {
	c.room.Logger().Debugf("detach peer: client=%v, peer=%p", c.Id, p)
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.peer != p {
		return // すでにdetach済み
	}
	c.peer.Detached()
	go c.drainMsg(c.peer.MsgCh())
	c.peer = nil
	c.newPeer <- nil
	c.waitPeer = make(chan *Peer, 1)
}

// DetachAndClosePeer : peerを切り離して、peerのwebsocketをcloseする.
// Client側のエラーによりpeerを切断する場合はこっち
func (c *Client) DetachAndClosePeer(p *Peer, err error) {
	c.room.Logger().Debugf("detach and close peer: client=%v, peer=%p", c.Id, p)
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.peer != p {
		return // すでにdetach済み. closeもされているはず.
	}

	p.CloseWithClientError(err)
	c.peer.Detached()
	go c.drainMsg(c.peer.MsgCh())
	c.peer = nil
	c.newPeer <- nil
	c.waitPeer = make(chan *Peer, 1)
}

func (c *Client) getWritePeer() (*Peer, <-chan *Peer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.peer, c.waitPeer
}

// EventLoop : EvBufにEventが入ってきたらPeerに送信してもらう
func (c *Client) EventLoop() {
	c.room.Logger().Debugf("client.EventLoop start: client=%v", c.Id)
loop:
	for {
		// dataが来るまで待つ
		select {
		case <-c.done:
			break loop
		case <-c.evbuf.HasData():
			c.room.Logger().Debugf("client.EventLoop: evbuf has data. client=%v", c.Id)
		}

		peer, wait := c.getWritePeer()
		if peer == nil {
			// peerがattachされるまで待つ
			c.room.Logger().Debugf("client.EventLoop: wait peer. client=%v", c.Id)
			select {
			case <-c.done:
				break loop
			case peer = <-wait:
				c.room.Logger().Debugf("client.EventLoop: peer available. client=%v, peer=%p", c.Id, peer)
			}
		}

		if err := peer.SendEvents(c.evbuf); err != nil {
			// 再接続でも復帰不能なので終わる.
			c.room.Logger().Errorf("clinet.EventLoop: send event error: client=%v peer=%p %v", c.Id, peer, err)
			c.evErr <- err
			break loop
		}
	}

	c.room.Logger().Debugf("client.EventLoop finish: client=%v", c.Id)
}
