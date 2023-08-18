package game

import (
	"crypto/hmac"
	"crypto/sha1"
	"hash"
	"sync"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"

	"wsnet2/auth"
	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/log"
	"wsnet2/pb"
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

	evbuf *common.RingBuf[*binary.RegularEvent]

	mu           sync.RWMutex
	msgSeqNum    int
	peer         *Peer
	waitPeer     chan *Peer
	renewPeer    chan struct{}
	connectCount int
	received     bool

	authKey string
	hmac    hash.Hash

	logger log.Logger

	evErr chan error
}

func NewPlayer(info *pb.ClientInfo, macKey string, room IRoom) (*Client, ErrorWithCode) {
	return newClient(info, macKey, room, true)
}

func NewWatcher(info *pb.ClientInfo, macKey string, room IRoom) (*Client, ErrorWithCode) {
	return newClient(info, macKey, room, false)
}

func newClient(info *pb.ClientInfo, macKey string, room IRoom, isPlayer bool) (*Client, ErrorWithCode) {
	props, iProps, err := common.InitProps(info.Props)
	if err != nil {
		return nil, WithCode(
			xerrors.Errorf("InitProps: %w", err),
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

		evbuf: common.NewRingBuf[*binary.RegularEvent](room.ClientConf().EventBufSize),

		waitPeer:  make(chan *Peer, 1),
		renewPeer: make(chan struct{}, 1),

		authKey: RandomHex(room.ClientConf().AuthKeyLen),
		hmac:    hmac.New(sha1.New, []byte(macKey)),

		logger: room.Logger().With(log.KeyClient, info.Id),

		evErr: make(chan error),
	}
	if info.IsHub {
		c.nodeCount = 0
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

func (c *Client) Logger() log.Logger {
	return c.logger
}

func (c *Client) ValidAuthData(authData string) error {
	// clientのtimestampは信用できないのでhashだけ検証
	_, err := auth.ValidAuthDataHash(authData, c.authKey, c.Id)
	return err
}

// MsgLoop goroutine.
func (c *Client) MsgLoop(deadline time.Duration) {
	var peerMsgCh <-chan binary.Msg
	var curPeer *Peer
	t := time.NewTimer(deadline)
loop:
	for {
		select {
		case <-t.C:
			if !c.received {
				// lobbyに繋がるがgameに繋げなかったり繋いでもmsg送ってこないのは何かある
				c.logger.Errorf("client timeout: %v connectCount=%v no msg received", c.Id, c.connectCount)
			} else {
				c.logger.Infof("client timeout: %v connectCount=%v", c.Id, c.connectCount)
			}
			c.room.SendMessage(&MsgClientTimeout{Sender: c})
			break loop

		case <-c.room.Done():
			c.logger.Debugf("client room done: %v", c.Id)
			curPeer.Close("room closed")
			if !t.Stop() {
				<-t.C
			}
			break loop

		case <-c.removed:
			c.logger.Debugf("client removed: %v", c.Id)
			if !t.Stop() {
				<-t.C
			}
			break loop

		case newDeadline := <-c.newDeadline:
			if !t.Stop() {
				<-t.C
			}
			// 突然短くされてもclientが把握できないので
			// 変更直後だけ旧deadline分の猶予をもたせる.
			t.Reset(deadline + newDeadline)
			deadline = newDeadline

		case <-c.renewPeer:
			go c.drainMsg(peerMsgCh)
			c.mu.Lock()
			if c.peer == nil {
				peerMsgCh = nil
				curPeer = nil
				if c.isPlayer {
					c.room.Repo().PlayerLog(c, PlayerLogDetach)
				}
			} else {
				c.connectCount++
				c.logger.Infof("new peer attached: %v peer=%p", c.Id, c.peer)
				peerMsgCh = c.peer.MsgCh()
				curPeer = c.peer
				if c.isPlayer {
					c.room.Repo().PlayerLog(c, PlayerLogAttach)
				}
				// つなげて切るだけのクライアントをタイムアウトさせるため、t.Resetしない
			}
			c.mu.Unlock()

		case m, ok := <-peerMsgCh:
			if !ok {
				// peer側でchをcloseした.
				c.logger.Debugf("peerMsgCh closed: %v peer=%p", c.Id, curPeer)
				// DetachPeerは呼ばれているはず
				peerMsgCh = nil
				curPeer = nil
				continue
			}
			msg, err := ConstructMsg(c, m)
			if err != nil {
				// おかしなデータを送ってくるクライアントは遮断する
				c.logger.Errorf("client invalid msg: %v %+v", c.Id, err)
				c.room.SendMessage(
					&MsgClientError{
						Sender: c,
						ErrMsg: err.Error(),
					})
				break loop
			}
			c.received = true
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
					err := xerrors.Errorf("invalid sequence num: %d, wants %d", seq, cSeq+1)
					c.logger.Warnf("client msg: %v %+v", c.Id, err)
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
			c.room.SendMessage(
				&MsgClientError{
					Sender: c,
					ErrMsg: err.Error(),
				})
			break loop
		}
	}
	c.logger.Debugf("Client.MsgLoop closed: %v", c.Id)
	close(c.done)

	go func() {
		time.Sleep(time.Duration(c.room.ClientConf().WaitAfterClose))
		c.room.Repo().RemoveClient(c)
	}()

	c.drainMsg(peerMsgCh)
	c.room.WaitGroup().Done()
}

func (c *Client) drainMsg(msgCh <-chan binary.Msg) {
	if msgCh == nil {
		return
	}
	for range msgCh {
	}
}

// RoomのMsgLoopから呼ばれる
func (c *Client) Removed(cause string) {
	close(c.removed)
	c.removeCause = cause

	c.mu.RLock()
	p := c.peer
	c.mu.RUnlock()
	if p != nil {
		go p.Close(c.removeCause)
	}
}

// RoomのMsgLoopから呼ばれる
func (c *Client) Send(e *binary.RegularEvent) error {
	return c.evbuf.Write(e)
}

// RoomのMsgLoopから呼ばれる.
func (c *Client) SendSystemEvent(e *binary.SystemEvent) {
	c.mu.RLock()
	p := c.peer
	c.mu.RUnlock()
	if p == nil {
		return
	}

	// SystemEventは送信順序を問わない. 多少遅れても構わない.
	// roomのmsgloopを止めないために新しいgoroutineで送信する.
	go p.SendSystemEvent(e)
}

func (c *Client) sendRenewPeer() {
	select {
	case c.renewPeer <- struct{}{}:
	default:
	}
}

// attachPeer: peerを紐付ける
// peerのgoroutineから呼ばれる
func (c *Client) AttachPeer(p *Peer, lastEvSeq int) error {
	c.logger.Debugf("attach peer: %v peer=%p", c.Id, p)
	c.mu.Lock()
	defer c.mu.Unlock()

	// 未読Eventを再送. client終了後でも送信する.
	if err := p.SendEvents(c.evbuf); err != nil {
		return xerrors.Errorf("SendEvents: %w", err)
	}

	select {
	case <-c.done:
		return xerrors.Errorf("client has been done")
	case <-c.removed:
		return xerrors.Errorf("client has been removed")
	default:
	}

	// msgSeqNumの後のメッセージから送信してもらう(再送含む)
	if err := p.SendReady(c.msgSeqNum); err != nil {
		return xerrors.Errorf("SendReady: %w", err)
	}

	if c.peer == nil {
		c.waitPeer <- p
	} else {
		c.peer.Close("new peer attached")
	}
	c.peer = p
	c.sendRenewPeer()
	return nil
}

// DetachPeer : peerを切り離す.
// Peer.MsgLoopで切断やエラーを検知したときに呼ばれる.
// websocketの切断は呼び出し側で行う
func (c *Client) DetachPeer(p *Peer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.peer != p {
		return // すでにdetach済み
	}
	c.logger.Infof("detach peer: %v peer=%p", c.Id, p)
	c.peer.Detached()
	go c.drainMsg(c.peer.MsgCh())

	select {
	case <-c.done:
		c.logger.Debugf("detach peer: client already closed: %v", c.Id)
		return
	default:
	}

	c.peer = nil
	c.waitPeer = make(chan *Peer, 1)
	c.sendRenewPeer()
}

// DetachAndClosePeer : peerを切り離して、peerのwebsocketをcloseする.
// Client側のエラーによりpeerを切断する場合はこっち
func (c *Client) DetachAndClosePeer(p *Peer, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.peer != p {
		c.logger.Debugf("detach+close peer: peer aleady detached: %v", c.Id)
		return // すでにdetach済み. closeもされているはず.
	}
	c.logger.Infof("detach+close peer: %v", c.Id)
	p.CloseWithClientError(err)
	c.peer.Detached()
	go c.drainMsg(c.peer.MsgCh())

	select {
	case <-c.done:
		c.logger.Debugf("detach+close peer: client already closed: %v", c.Id)
		return
	default:
	}

	c.peer = nil
	c.waitPeer = make(chan *Peer, 1)
	c.sendRenewPeer()
}

func (c *Client) getWritePeer() (*Peer, <-chan *Peer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.peer, c.waitPeer
}

// EventLoop : evbufにEventが入ってきたらPeerに送信してもらう
func (c *Client) EventLoop() {
loop:
	for {
		// dataが来るまで待つ
		select {
		case <-c.done:
			break loop
		case <-c.evbuf.HasData():
		}

		peer, wait := c.getWritePeer()
		if peer == nil {
			// peerがattachされるまで待つ
			c.logger.Debugf("client.EventLoop: waiting peer: %v", c.Id)
			select {
			case <-c.done:
				break loop
			case peer = <-wait:
			}
		}

		if err := peer.SendEvents(c.evbuf); err != nil {
			// 再接続でも復帰不能なので終わる.
			c.evErr <- xerrors.Errorf("send event: %w", err)
			break loop
		}
	}
}
