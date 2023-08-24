package game

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/shiguredo/websocket"
	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/metrics"
)

const (
	writeTimeout = 3 * time.Second

	waitCloseTimeout = 3 * time.Second
)

// Peer : websocketの接続
//
// CloseCodeが次の場合はクライアントは再接続を試行しない
//   - (1000) CloseNormalClosure (C#: WebsocketCloseStatus.NormalClosure)
//   - (1001) CloseGoingAway (C#: WebsocketCloseStatus.EndpointUnavailable)
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
	conn.SetCloseHandler(func(code int, text string) error { return nil }) // CloseMessageの返送はこちらで制御する
	err := cli.AttachPeer(p, lastEvSeq)
	if err != nil {
		p.closeWithMessage(websocket.CloseGoingAway, err.Error())
		return nil, xerrors.Errorf("AttachPeer (%v, peer=%p): %w", cli.Id, p, err)
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
// websocketハンドラのgoroutineからcli.AttachPeer経由で呼ばれる.
func (p *Peer) SendReady(lastMsgSeq int) error {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return xerrors.New("peer closed")
	}
	p.client.logger.Infof("peer ready (%v, peer=%p): lastMsg=%v", p.client.Id, p, lastMsgSeq)
	ev := binary.NewEvPeerReady(lastMsgSeq)
	return writeMessage(p.conn, websocket.BinaryMessage, ev.Marshal())
}

// SendSystemEvent : SystemEventを送信する.
// 送信失敗時はPeerを閉じて再接続できるようにする.
// 個別のgoroutineで呼ばれるのでerrorは返さない. see: (*Client).SendSystemEvent()
func (p *Peer) SendSystemEvent(ev *binary.SystemEvent) {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return
	}
	metrics.MessageSent.Add(1)
	err := writeMessage(p.conn, websocket.BinaryMessage, ev.Marshal())
	if err != nil {
		p.client.logger.Warnf("peer send %v (%v, peer=%p): %+v", ev.Type(), p.client.Id, p, err)
		p.sendCloseAndCloseConn(websocket.CloseInternalServerErr, err.Error())
	}
}

// SendEvents : evbufに蓄積されてるイベントを送信
// 送信失敗時はPeerを閉じて再接続できるようにする. errorは返さない.
// 再接続しても復帰不能な場合はerrorを返す（Client.EventLoopを止める）.
func (p *Peer) SendEvents(evbuf *common.RingBuf[*binary.RegularEvent]) error {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	if p.closed {
		return nil
	}

	evs, err := evbuf.Read(p.evSeqNum)
	if err != nil {
		// evSeqNumが古すぎるため. 復帰不能.
		// 頻発するようならevbufのサイズ(ClientConf.EventBufSize)を拡張したほうがよいかも
		p.client.logger.Errorf("peer evbuf.Read (%v, %p): %+v", p.client.Id, p, err)
		p.sendCloseAndCloseConn(
			websocket.CloseGoingAway, // client: EndpointUnavailable
			err.Error())
		return err
	}

	seqNum := p.evSeqNum
	for _, ev := range evs {
		seqNum++
		buf := ev.Marshal(seqNum)
		err := writeMessage(p.conn, websocket.BinaryMessage, buf)
		if err != nil {
			// 新しいpeerで復帰できるかもしれない
			p.client.logger.Warnf("peer send %v (%v, %p): %+v", ev.Type(), p.client.Id, p, err)
			p.sendCloseAndCloseConn(websocket.CloseInternalServerErr, err.Error())
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

// closeWithMessage : CloseMessageを送信して一定時間後にWebsocketを閉じる
// 内部で p.muWrite をロックする
func (p *Peer) closeWithMessage(code int, msg string) {
	p.muWrite.Lock()
	defer p.muWrite.Unlock()
	p.sendCloseAndCloseConn(code, msg)
}

// sendCloseAndCloseconn : CloseMessageを送信して一定時間後にWebsocketを閉じる
// 呼び出す前に p.muWrite をロックすること
func (p *Peer) sendCloseAndCloseConn(code int, msg string) {
	if p.closed {
		return
	}
	p.closed = true
	writeMessage(p.conn, websocket.CloseMessage, formatCloseMessage(code, msg))
	// wait close message from client
	time.AfterFunc(waitCloseTimeout, func() { p.conn.Close() })
}

func (p *Peer) MsgLoop(ctx context.Context) {
loop:
	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			if !p.closed {
				if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					// CloseMessage送信前のEOFもAbnorormalClosureになる
					// アプリkillでも起こるのでErrorにはしない
					p.client.logger.Warnf("peer close error (%v, %p): %v", p.client.Id, p, err)
					p.closeWithMessage(websocket.CloseUnsupportedData, "unexpected close message")
				} else if websocket.IsUnexpectedCloseError(err) {
					// その他のCloseMessageはクライアントがおかしいのでErrorとして報告
					p.client.logger.Errorf("peer close error (%v, %p): %v", p.client.Id, p, err)
					p.closeWithMessage(websocket.CloseUnsupportedData, "unexpected close message")
				} else {
					// 不慮の切断など。モバイルではよく起きるし再接続もあるのでWarn
					p.client.logger.Warnf("peer read error (%v, %p): %T %v", p.client.Id, p, err, err)
					if !errors.Is(err, net.ErrClosed) {
						// 切断していないならclose messageを送ってから切断
						p.closeWithMessage(websocket.CloseInternalServerErr, err.Error())
					} else {
						p.closed = true
						p.conn.Close()
					}
				}
			} else if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				// clientから送るのはNormalClosureのみなのでそれ以外は警告（切断後なのでErrorにはしなくていい）
				p.client.logger.Warnf("peer read error (%v, %p): %T %v", p.client.Id, p, err, err)
			} else {
				// 正常切断
				p.client.logger.Infof("peer closed (%v, %p): %v", p.client.Id, p, err)
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

func writeMessage(conn *websocket.Conn, messageType int, data []byte) error {
	metrics.MessageSent.Add(1)
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return conn.WriteMessage(messageType, data)
}

func formatCloseMessage(closeCode int, text string) []byte {
	if len(text) > 123 {
		text = text[:123]
	}
	return websocket.FormatCloseMessage(closeCode, text)
}
