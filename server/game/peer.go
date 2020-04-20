package game

import (
	"context"

	"github.com/gorilla/websocket"
	"golang.org/x/xerrors"
)

type Peer struct {
	client *Client
	conn   *websocket.Conn
	msgCh  chan Msg

	done     chan struct{}
	detached chan struct{}

	msgSeqNum int
	evSeqNum  int
}

func NewPeer(ctx context.Context, cli *Client, conn *websocket.Conn) *Peer {
	p := &Peer{
		client: cli,
		conn:   conn,
		msgCh:  make(chan Msg),

		done:     make(chan struct{}),
		detached: make(chan struct{}),
	}
	go p.MsgLoop(ctx)
	cli.AttachPeer(p)
	return p
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
	for _, ev := range evs {
		err := p.conn.WriteMessage(websocket.BinaryMessage, ev.Encode())
		if err != nil {
			return err
		}
	}

	p.evSeqNum = last
	return nil
}

func (p *Peer) Close() {
	if p == nil {
		return
	}
	p.client.room.logger.Debugf("peer close: client=%v peer=%p", p.client.Id, p)
	p.writeCloseMessage(websocket.CloseNormalClosure, "")
	p.conn.Close()
}

// Detached from Client (called by Client)
func (p *Peer) Detached() {
	if p == nil {
		return
	}
	close(p.detached)
}

func (p *Peer) ClientError(err error) {
	if p == nil {
		return
	}
	p.writeCloseMessage(websocket.CloseInternalServerErr, err.Error())
	p.conn.Close()
}

func (p *Peer) writeCloseMessage(code int, msg string) {
	p.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, msg))
}

func (p *Peer) MsgLoop(ctx context.Context) {
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
			if websocket.IsCloseError(err) &&
				!websocket.IsUnexpectedCloseError(
					err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logger.Infof("Peer closed: client=%v peer=%p %v", p.client.Id, p, err)
			} else {
				logger.Errorf("Peer read error: client=%v peer=%p %v", p.client.Id, p, err)
			}
			break loop
		}

		seq, msg, err := DecodeMsg(p.client, data)
		if err != nil {
			p.client.room.logger.Errorf("DecodeMsg error: client=%v peer=%p %v", p.client.Id, p, err)
			p.writeCloseMessage(websocket.CloseInvalidFramePayloadData, err.Error())
			break loop
		}

		if seq != p.msgSeqNum+1 {
			err := xerrors.Errorf("sequence num skipped %d to %d", p.msgSeqNum, seq)
			p.client.room.logger.Errorf("Peer MsgLoop error: client=%v peer=%p %v", p.client.Id, p, err)
			p.writeCloseMessage(websocket.CloseInvalidFramePayloadData, err.Error())
			break loop
		}

		p.msgSeqNum = seq
		p.msgCh <- msg
	}

	p.client.DetachPeer(p)
	close(p.done)
	p.conn.Close()
}
