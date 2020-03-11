package game

import (
	"fmt"
	"sync"
	"time"

	"wsnet2/log"
)

const (
	// RoomMsgChSize : Msgチャネルのバッファサイズ
	RoomMsgChSize = 10

	// RoomDefaultClientDeadline : クライアント切断判定の無通信時間の初期値
	RoomDefaultClientDeadline = 30 * time.Second
)

type RoomID string

type Room struct {
	ID        RoomID
	maxClient int
	deadline  time.Duration

	msgCh    chan Msg
	done     chan struct{}
	wgClient sync.WaitGroup

	muClients sync.RWMutex
	clients   map[ClientID]*Client
	leaved    map[ClientID]error
	master    *Client
	order     []ClientID

	logger log.Logger
}

func NewRoom(id RoomID, maxClient int, master *Client) *Room {
	r := &Room{
		ID:        id,
		maxClient: maxClient,
		deadline:  RoomDefaultClientDeadline,

		msgCh: make(chan Msg, RoomMsgChSize),
		done:  make(chan struct{}),

		clients: make(map[ClientID]*Client),
		leaved:  make(map[ClientID]error),
		master:  master,
		order:   []ClientID{master.ID},

		logger: log.Get(log.CurrentLevel()),
	}

	r.clients[master.ID] = master
	r.wgClient.Add(1)

	return r
}

// MsgLoop goroutine dispatch messages.
func (r *Room) MsgLoop() {
	r.logger.Debugf("Room.MsgLoop() start: room=%v", r.ID)
Loop:
	for {
		select {
		case <-r.Done():
			r.logger.Infof("Room closed: room=%v", r.ID)
			break Loop
		case msg := <-r.msgCh:
			r.logger.Debugf("Room msg: room=%v, %T %v", r.ID, msg, msg)
			r.dispatch(msg)
		}
	}

	r.drainMsg()
	r.logger.Debugf("Room.MsgLoop() finish: room=%v", r.ID)
}

// drainMsg drain msgCh until all clients closed.
// clientのgoroutineがmsgChに書き込むところで停止するのを防ぐ
func (r *Room) drainMsg() {
	ch := make(chan struct{})
	go func() {
		r.wgClient.Wait()
		ch <- struct{}{}
	}()

	for {
		select {
		case msg := <-r.msgCh:
			r.logger.Debugf("Discard msg: room=%v %T %v", r.ID, msg, msg)
		case <-ch:
			return
		}
	}
}

// Done returns a channel which cloased when room is done.
func (r *Room) Done() <-chan struct{} {
	return r.done
}

// Post a mssage to room
func (r *Room) Post(m Msg) {
	r.msgCh <- m
}

func (r *Room) Timeout(c *Client) {
	r.removeClient(c, fmt.Errorf("client timeout: %v", c.ID))
}

func (r *Room) removeClient(c *Client, err error) {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	cid := c.ID

	if _, ok := r.clients[cid]; !ok {
		r.logger.Debugf("Client may be aleady leaved: room=%v, client=%v", r.ID, cid)
		return
	}

	r.logger.Infof("Client removed: room=%v, client=%v", r.ID, cid)
	delete(r.clients, cid)
	r.leaved[cid] = err
	c.Removed()

	if len(r.clients) == 0 {
		close(r.done)
		return
	}

	r.Post(MsgLeave{c.ID})
}

func (r *Room) dispatch(msg Msg) error {
	switch m := msg.(type) {
	case MsgJoin:
		return r.msgJoin(m)
	default:
		return fmt.Errorf("unknown msg type: %T %v", m, m)
	}
}

func (r *Room) broadcast(ev Event) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	for _, c := range r.clients {
		if err := c.Send(ev); err != nil {
			// removeClient locks muClients so that must be called another goroutine.
			go r.removeClient(c, err)
		}
	}
	return nil
}

func (r *Room) msgJoin(msg MsgJoin) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	if r.maxClient == len(r.clients) {
		close(msg.Res)
		return fmt.Errorf("Room full. max=%v, client=%v", r.maxClient, msg.ID)
	}

	c := NewClient(msg.ID, r)
	r.clients[c.ID] = c

	msg.Res <- JoinResponse{
		Client: c,
	}
	r.broadcast(EvJoined{c})
	return nil
}
