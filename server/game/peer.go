package game

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/shiguredo/websocket"
	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/metrics"
)

// Peer : websocketの接続
//
// CloseCodeが次の場合はクライアントは再接続を試行しない
//  - (1000) CloseNormalClosure (C#: WebsocketCloseStatus.NormalClosure)
//  - (1001) CloseGoingAway (C#: WebsocketCloseStatus.EndpointUnavailable)
//
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
		p.closeWithMessage(websocket.CloseGoingAway, err.Error())
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

// SendReady : EvPeerReadyを送信する.
// clientのAttachPeerから呼ばれる.
func (p *Peer) SendReady(lastMsgSeq int) error {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return xerrors.New("peer closed")
	}
	ev := binary.NewEvPeerReady(lastMsgSeq)
	metrics.MessageSent.Add(1)
	return p.conn.WriteMessage(websocket.BinaryMessage, ev.Marshal())
}

// SendSystemEvent : SystemEventを送信する.
// 送信失敗時はPeerを閉じて再接続できるようにする.
func (p *Peer) SendSystemEvent(ev *binary.SystemEvent) error {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return nil
	}
	metrics.MessageSent.Add(1)
	err := p.conn.WriteMessage(websocket.BinaryMessage, ev.Marshal())
	if err != nil {
		p.client.logger.Errorf("peer SendSystemEvent write (%v, %p): %+v", p.client.Id, p, err)
		metrics.MessageSent.Add(1)
		p.conn.WriteMessage(websocket.CloseMessage,
			formatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
		p.conn.Close()
		p.closed = true
	}
	return err
}

// SendEvents : evbufに蓄積されてるイベントを送信
// 送信失敗時はPeerを閉じて再接続できるようにする. errorは返さない.
// 再接続しても復帰不能な場合はerrorを返す（Client.EventLoopを止める）.
func (p *Peer) SendEvents(evbuf *EvBuf) error {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return nil
	}

	evs, err := evbuf.Read(p.evSeqNum)
	if err != nil {
		// evSeqNumが古すぎるため. 復帰不能.
		p.client.logger.Errorf("peer evbuf.Read (%v, %p): %+v", p.client.Id, p, err)
		metrics.MessageSent.Add(1)
		p.conn.WriteMessage(websocket.CloseMessage,
			formatCloseMessage(websocket.CloseGoingAway, err.Error()))
		p.conn.Close()
		p.closed = true
		return err
	}

	seqNum := p.evSeqNum
	for _, ev := range evs {
		seqNum++
		buf := ev.Marshal(seqNum)
		metrics.MessageSent.Add(1)
		err := p.conn.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			// 新しいpeerで復帰できるかもしれない
			p.client.logger.Errorf("peer WriteMessage (%v, %p): %+v", p.client.Id, p, err)
			metrics.MessageSent.Add(1)
			p.conn.WriteMessage(websocket.CloseMessage,
				formatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
			p.conn.Close()
			p.closed = true
			return nil
		}
	}
	p.evSeqNum = seqNum
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
		return
	}
	metrics.MessageSent.Add(1)
	p.conn.WriteMessage(websocket.CloseMessage, formatCloseMessage(code, msg))
	p.conn.Close()
	p.closed = true
}

func (p *Peer) MsgLoop(ctx context.Context) {
loop:
	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				p.client.logger.Infof("peer closed (%v, %p): %+v", p.client.Id, p, err)
			} else if websocket.IsUnexpectedCloseError(err) {
				p.client.logger.Errorf("peer close error (%v, %p): %+v", p.client.Id, p, err)
			} else {
				if !errors.Is(err, net.ErrClosed) {
					p.client.logger.Errorf("peer read error (%v, %p): %T %+v", p.client.Id, p, err, err)
					p.closeWithMessage(websocket.CloseInternalServerErr, err.Error())
				}
			}
			break loop
		}
		metrics.MessageRecv.Add(1)

		msg, err := binary.UnmarshalMsg(p.client.hmac, data)
		if err != nil {
			p.client.logger.Errorf("peer UnmarshalMsg (%v, %p): %+v", p.client.Id, p, err)
			p.closeWithMessage(websocket.CloseInvalidFramePayloadData, err.Error())
			break loop
		}

		select {
		case <-ctx.Done():
			break loop
		case <-p.detached:
			break loop
		case <-p.client.done:
			break loop
		case p.msgCh <- msg:
		}
	}

	p.client.DetachPeer(p)
	close(p.msgCh)
	close(p.done)
}

func formatCloseMessage(closeCode int, text string) []byte {
	if len(text) > 123 {
		text = text[:123]
	}
	return websocket.FormatCloseMessage(closeCode, text)
}
