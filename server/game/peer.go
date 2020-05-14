package game

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"golang.org/x/xerrors"

	"wsnet2/binary"
)

type Peer struct {
	client *Client
	conn   *websocket.Conn
	msgCh  chan binary.Msg

	done     chan struct{}
	detached chan struct{}

	muWrite sync.Mutex
	closed  bool

	evSeqNum int
}

func NewPeer(ctx context.Context, cli *Client, conn *websocket.Conn, lastEvSeq int) (*Peer, error) {
	p := &Peer{
		client: cli,
		conn:   conn,
		msgCh:  make(chan binary.Msg),

		done:     make(chan struct{}),
		detached: make(chan struct{}),

		evSeqNum: lastEvSeq,
	}
	err := cli.AttachPeer(p, lastEvSeq)
	if err != nil {
		p.closeWithMessage(websocket.CloseInternalServerErr, err.Error())
		return nil, err
	}
	go p.MsgLoop(ctx)
	return p, nil
}

func (p *Peer) MsgCh() <-chan binary.Msg {
	return p.msgCh
}

func (p *Peer) Done() <-chan struct{} {
	return p.done
}

func (p *Peer) LastEventSeq() int {
	return p.evSeqNum
}

func (p *Peer) SendReady(lastMsgSeq int) error {
	//	p.conn.WriteMessage()
	return nil
}

// SendEvent : Eventをwebsocketで送信.
// Client.EventLoopから呼ばれる.
// error時のdetach処理はClient側で行う
func (p *Peer) SendEvent(evs []*binary.Event) error {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return xerrors.Errorf("peer closed")
	}

	last := p.evSeqNum
	for _, ev := range evs {
		last++
		err := p.conn.WriteMessage(websocket.BinaryMessage, ev.Serialize(last))
		if err != nil {
			return err
		}
	}

	p.evSeqNum = last
	return nil
}

func (p *Peer) Close(msg string) {
	if p == nil {
		return
	}
	p.closeWithMessage(websocket.CloseNormalClosure, msg)
}

// Detached from Client (called by Client)
func (p *Peer) Detached() {
	if p == nil {
		return
	}
	close(p.detached)
}

// CloseWithClientError : クライアントエラーによってwebsocketを切断する.
// Clientのgoroutineから呼ばれる.
func (p *Peer) CloseWithClientError(err error) {
	p.closeWithMessage(websocket.CloseInternalServerErr, err.Error())
}

func (p *Peer) closeWithMessage(code int, msg string) {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		p.client.room.logger.Debugf("peer already closed: client=%v peer=%p %v", p.client.Id, p, msg)
		return
	}
	p.client.room.logger.Debugf("peer close: client=%v peer=%p %v", p.client.Id, p, msg)
	p.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, msg))
	p.conn.Close()
	p.closed = true
}

func (p *Peer) MsgLoop(ctx context.Context) {
	p.client.room.logger.Debugf("Peer.MsgLoop start: client=%v peer=%p", p.client.Id, p)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-p.detached:
			break loop
		case <-p.client.done:
			break loop
		default:
		}

		_, data, err := p.conn.ReadMessage()
		if err != nil {
			logger := p.client.room.logger
			if websocket.IsCloseError(err) {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logger.Infof("Peer closed: client=%v peer=%p %v", p.client.Id, p, err)
				} else {
					logger.Errorf("peer close error: client=%v peer=%p %v", p.client.Id, p, err)
				}
			} else {
				logger.Errorf("Peer read error: client=%v peer=%p %v", p.client.Id, p, err)
				p.closeWithMessage(websocket.CloseInternalServerErr, err.Error())
			}
			break loop
		}

		msg, err := binary.UnmarshalMsg(data)
		if err != nil {
			p.client.room.logger.Errorf("DecodeMsg error: client=%v peer=%p %v", p.client.Id, p, err)
			p.closeWithMessage(websocket.CloseInvalidFramePayloadData, err.Error())
			break loop
		}

		p.msgCh <- msg
	}

	p.client.DetachPeer(p)
	close(p.msgCh)
	close(p.done)
	p.client.room.logger.Debugf("Peer.MsgLoop finish: client=%v peer=%p", p.client.Id, p)
}
