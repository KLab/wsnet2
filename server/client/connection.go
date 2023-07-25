package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"hash"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shiguredo/websocket"
	"golang.org/x/xerrors"

	"wsnet2/auth"
	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/pb"
)

const reconnectInterval = 3 * time.Second

var dialer = &websocket.Dialer{
	Subprotocols:    []string{"wsnet2"},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type msgerr struct {
	msg string
	err error
}

type marshaledMsg struct {
	seq   int
	frame []byte
}

type unrecoverableError struct {
	error
}

func (err *unrecoverableError) Unwrap() error {
	return err.error
}

func unrecoverable(err error) unrecoverableError {
	return unrecoverableError{err}
}

// Connection : Roomへの接続
type Connection struct {
	appid  string
	userid string
	url    string
	bearer string

	deadline atomic.Uint32

	mumsg  sync.Mutex
	msgseq int
	msgbuf *common.RingBuf[marshaledMsg]
	hmac   hash.Hash

	lastev int
	evch   chan binary.Event

	sysmsg chan binary.Msg

	done chan msgerr
}

// Send : Msg (RegularMsg) を送信（バッファに書き込み、自動再送対象）
func (r *Connection) Send(typ binary.MsgType, payload []byte) error {
	r.mumsg.Lock()
	defer r.mumsg.Unlock()
	next := r.msgseq + 1
	err := r.msgbuf.Write(marshaledMsg{
		next,
		binary.BuildRegularMsgFrame(typ, next, payload, r.hmac),
	})
	if err != nil {
		return xerrors.Errorf("write to msgbuf: %w", err)
	}
	r.msgseq++
	return nil
}

// SendSystemMsg : SystemMsg (NonRegularMsg) を送信
func (r *Connection) SendSystemMsg(msg binary.Msg) error {
	if _, ok := msg.(binary.RegularMsg); ok {
		return xerrors.Errorf("not a system msg: %T %v", msg, msg)
	}
	select {
	case r.sysmsg <- msg:
		return nil
	default:
		return xerrors.Errorf("system msg sender is not ready")
	}
}

// Events : Eventが流れてくるチャネル
func (c *Connection) Events() <-chan binary.Event {
	return c.evch
}

// Wait : 接続終了(退室)を待つ
func (c *Connection) Wait(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "context done", ctx.Err()
	case d := <-c.done:
		return d.msg, d.err
	}
}

// newConn allocates and starts new connection
func newConn(ctx context.Context, accinfo *AccessInfo, joined *pb.JoinedRoomRes, warn func(error)) (*Connection, error) {

	bearer, err := auth.GenerateAuthData(joined.AuthKey, accinfo.UserId, time.Now())
	if err != nil {
		return nil, xerrors.Errorf("bearer: %w", err)
	}

	mac := hmac.New(sha1.New, []byte(accinfo.MACKey))

	conn := &Connection{
		appid:  accinfo.AppId,
		userid: accinfo.UserId,
		url:    joined.Url,
		bearer: "Bearer " + bearer,

		msgbuf: common.NewRingBuf[marshaledMsg](32),
		hmac:   mac,

		evch:   make(chan binary.Event, 32),
		sysmsg: make(chan binary.Msg),
		done:   make(chan msgerr, 1),
	}

	conn.deadline.Store(joined.Deadline)

	if warn == nil {
		warn = func(error) {}
	}

	go func() {
		msg, err := conn.connect(ctx, warn)
		conn.done <- msgerr{msg, err}
		close(conn.evch)
	}()

	return conn, nil
}

func (conn *Connection) connect(ctx context.Context, warn func(error)) (string, error) {
	var retrylimit *time.Timer
	var lasterr error

	for {
		if retrylimit == nil {
			retrylimit = time.NewTimer(time.Duration(conn.deadline.Load()) * time.Second)
		}
		select {
		case <-ctx.Done():
			return "context done", ctx.Err()
		case <-retrylimit.C:
			return "retry limit", lasterr
		default:
		}

		interval := time.NewTimer(reconnectInterval)

		hdr := http.Header{}
		hdr.Add("Wsnet2-App", conn.appid)
		hdr.Add("Wsnet2-User", conn.userid)
		hdr.Add("Wsnet2-LastEventSeq", strconv.Itoa(conn.lastev))
		hdr.Add("Authorization", conn.bearer)

		ws, res, err := dialer.DialContext(ctx, conn.url, hdr)
		if err != nil {
			if res != nil && res.StatusCode >= 400 && res.StatusCode < 500 {
				return "websocket dial failed", xerrors.Errorf("dial: %w", err)
			}
			warn(err)
			lasterr = err
			select {
			case <-ctx.Done():
				return "context done", ctx.Err()
			case <-interval.C:
				continue
			}
		}

		conctx, cancel := context.WithCancel(ctx)
		done := make(chan error, 4)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			done <- conn.receiver(conctx, ws, func(lastmsgseq int) {
				retrylimit = nil
				var mu sync.Mutex
				wg.Add(3)
				go func() {
					done <- conn.pinger(conctx, ws, &mu)
					wg.Done()
				}()
				go func() {
					done <- conn.sender(conctx, ws, &mu, lastmsgseq)
					wg.Done()
				}()
				go func() {
					done <- conn.systemSender(conctx, ws, &mu)
					wg.Done()
				}()
			})
			wg.Done()
		}()

		err = <-done
		cancel()
		wg.Wait()

		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return err.(*websocket.CloseError).Text, nil
		}
		if ue := unrecoverable(nil); errors.As(err, &ue) {
			return "give up on reconnection", ue.Unwrap()
		}

		warn(err)
		lasterr = err

		select {
		case <-ctx.Done():
			return "context done", ctx.Err()
		case <-interval.C:
		}
	}
}

func (conn *Connection) receiver(ctx context.Context, ws *websocket.Conn, startsender func(int)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ws.SetReadDeadline(time.Now().Add(time.Duration(conn.deadline.Load()) * time.Second))
		_, data, err := ws.ReadMessage()
		if err != nil {
			return err // websocket.IsCloseError()がwrapを考慮してくれないのでこのまま返す
		}

		ev, seq, err := binary.UnmarshalEvent(data)
		if err != nil {
			return xerrors.Errorf("receiver unmarshal: %w", err)
		}

		lastev := conn.lastev
		if _, ok := ev.(*binary.RegularEvent); ok {
			lastev++
			if seq != lastev {
				return xerrors.Errorf("invalid event sequence num: %v wants %v", seq, lastev)
			}
		}

		switch ev.Type() {
		case binary.EvTypePeerReady:
			msgseq, err := binary.UnmarshalEvPeerReadyPayload(ev.Payload())
			if err != nil {
				return xerrors.Errorf("unmarshal peer-ready payload %v: %w", ev.Type(), err)
			}
			startsender(msgseq)

		case binary.EvTypeRoomProp:
			deadline, err := binary.GetRoomPropClientDeadline(ev.Payload())
			if err != nil {
				return xerrors.Errorf("get client deadline: %w", err)
			}
			if deadline != 0 {
				conn.deadline.Store(deadline)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			select {
			case <-ctx.Done():
				return ctx.Err()
			case conn.evch <- ev:
				conn.lastev = lastev
			}
		}
	}
}

func (conn *Connection) pinger(ctx context.Context, ws *websocket.Conn, mu *sync.Mutex) error {
	for {
		conn.mumsg.Lock()
		msg := binary.NewMsgPing(time.Now()).Marshal(conn.hmac)
		conn.mumsg.Unlock()

		mu.Lock()
		ws.SetWriteDeadline(time.Now().Add(time.Second))
		err := ws.WriteMessage(websocket.BinaryMessage, msg)
		mu.Unlock()
		if err != nil {
			return xerrors.Errorf("pinger: %w", err)
		}

		t := time.Duration(conn.deadline.Load()) * time.Second / 3
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTimer(t).C:
		}
	}
}

func (conn *Connection) sender(ctx context.Context, ws *websocket.Conn, mu *sync.Mutex, lastseq int) error {
	for {
		msgs, err := conn.msgbuf.Read(lastseq)
		if err != nil {
			return unrecoverable(xerrors.Errorf("sender read: %w", err))
		}

		for _, msg := range msgs {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			mu.Lock()
			ws.SetWriteDeadline(time.Now().Add(time.Second))
			err := ws.WriteMessage(websocket.BinaryMessage, msg.frame)
			mu.Unlock()
			if err != nil {
				return xerrors.Errorf("sender write(%v): %w", msg.seq, err)
			}
			lastseq = msg.seq
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-conn.msgbuf.HasData():
		}
	}
}

func (conn *Connection) systemSender(ctx context.Context, ws *websocket.Conn, mu *sync.Mutex) error {
	// 送信中の投げ込みも受け付けるようcap=1のチャネルを挟む
	mc := make(chan binary.Msg, 1)
	// systemSenderが動き始めてからsysmsgへの書き込みを受け付ける (see: conn.SendSystemMsg())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case mc <- <-conn.sysmsg:
			}
		}
	}()

	for {
		var msg binary.Msg
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg = <-mc:
		}

		conn.mumsg.Lock()
		frame := msg.Marshal(conn.hmac)
		conn.mumsg.Unlock()

		mu.Lock()
		ws.SetWriteDeadline(time.Now().Add(time.Second))
		err := ws.WriteMessage(websocket.BinaryMessage, frame)
		mu.Unlock()
		if err != nil {
			return xerrors.Errorf("systemSender write: %w", err)
		}
	}
}
