package game

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"golang.org/x/xerrors"
)

type Peer struct {
	client *Client
	conn   *websocket.Conn
	msgCh  chan Msg

	done     chan struct{}
	detached chan struct{}

	muWrite sync.Mutex
	closed  bool

	msgSeqNum int
	evSeqNum  int
}

func NewPeer(ctx context.Context, cli *Client, conn *websocket.Conn) (*Peer, error) {
	p := &Peer{
		client: cli,
		conn:   conn,
		msgCh:  make(chan Msg),

		done:     make(chan struct{}),
		detached: make(chan struct{}),
	}
	err := cli.AttachPeer(p)
	if err != nil {
		p.closeWithMessage(websocket.CloseInternalServerErr, err.Error())
		return nil, err
	}
	go p.MsgLoop(ctx)
	return p, nil
}

func (p *Peer) MsgCh() <-chan Msg {
	return p.msgCh
}

func (p *Peer) Done() <-chan struct{} {
	return p.done
}

func (p *Peer) LastEventSeq() int {
	return p.evSeqNum
}

// SendEvent : Eventをwebsocketで送信.
// Client.EventLoopから呼ばれる.
// error時のdetach処理はClient側で行う
func (p *Peer) SendEvent(evs []Event, last int) error {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return xerrors.Errorf("peer closed")
	}

	for _, ev := range evs {
		err := p.conn.WriteMessage(websocket.BinaryMessage, ev.Encode())
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

func (p *Peer) ClientError(err error) {
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

		seq, msg, err := DecodeMsg(p.client, data)
		if err != nil {
			p.client.room.logger.Errorf("DecodeMsg error: client=%v peer=%p %v", p.client.Id, p, err)
			p.closeWithMessage(websocket.CloseInvalidFramePayloadData, err.Error())
			break loop
		}

		if seq != p.msgSeqNum+1 {
			err := xerrors.Errorf("sequence num skipped %d to %d", p.msgSeqNum, seq)
			p.client.room.logger.Errorf("Peer MsgLoop error: client=%v peer=%p %v", p.client.Id, p, err)
			p.closeWithMessage(websocket.CloseInvalidFramePayloadData, err.Error())
			break loop
		}

		p.msgSeqNum = seq
		p.msgCh <- msg
	}

	p.client.DetachPeer(p)
	close(p.done)
	p.client.room.logger.Debugf("Peer.MsgLoop finish: client=%v peer=%p", p.client.Id, p)
}
