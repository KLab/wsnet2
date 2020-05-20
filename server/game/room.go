package game

import (
	"sync"
	"time"

	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	// RoomMsgChSize : Msgチャネルのバッファサイズ
	RoomMsgChSize = 10
)

type RoomID string

type Room struct {
	*pb.RoomInfo
	repo *Repository

	deadline time.Duration

	publicProps  binary.Dict
	privateProps binary.Dict

	msgCh    chan Msg
	done     chan struct{}
	wgClient sync.WaitGroup

	muClients sync.RWMutex
	clients   map[ClientID]*Client
	master    *Client
	order     []ClientID
	// todo: photonのactorNrみたいに連番ふったほうがよいかも。gameObjectのphotonView.IDみたいなのを作るときに必要

	// todo: pongにclientたちの最終送信時刻を入れたい
	// muLastMsg sync.RWMutex
	// lastMsg   map[ClientID]int // unixtime

	logger log.Logger
}

func initProps(props []byte) (binary.Dict, []byte, error) {
	if len(props) == 0 {
		props = binary.MarshalDict(nil)
	}
	d, t, _, err := binary.Unmarshal(props)
	if err != nil {
		return nil, nil, err
	}
	if t != binary.TypeDict {
		return nil, nil, xerrors.Errorf("type is not Dict: %v", t)
	}
	return d.(binary.Dict), props, nil
}

func NewRoom(repo *Repository, info *pb.RoomInfo, masterInfo *pb.ClientInfo, conf *config.GameConf) (*Room, *Client, <-chan JoinedInfo, error) {
	pubProps, iProps, err := initProps(info.PublicProps)
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("PublicProps unmarshal error: %w", err)
	}
	info.PublicProps = iProps
	privProps, iProps, err := initProps(info.PrivateProps)
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("PrivateProps unmarshal error: %w", err)
	}
	info.PrivateProps = iProps

	r := &Room{
		RoomInfo: info,
		repo:     repo,
		deadline: time.Duration(info.ClientDeadline) * time.Second,

		publicProps:  pubProps,
		privateProps: privProps,

		msgCh: make(chan Msg, RoomMsgChSize),
		done:  make(chan struct{}),

		clients: make(map[ClientID]*Client),
		order:   []ClientID{},

		logger: log.Get(log.CurrentLevel()),
	}

	r.wgClient.Add(1)
	master := NewClient(masterInfo, r)
	r.master = master
	r.clients[ClientID(master.Id)] = master
	r.order = append(r.order, master.ID())

	go r.MsgLoop()

	ch := make(chan JoinedInfo)
	r.msgCh <- &MsgCreate{ch}

	return r, master, ch, nil
}

func (r *Room) ID() RoomID {
	return RoomID(r.Id)
}

// MsgLoop goroutine dispatch messages.
func (r *Room) MsgLoop() {
	r.logger.Debugf("Room.MsgLoop() start: room=%v", r.Id)
Loop:
	for {
		select {
		case <-r.Done():
			r.logger.Infof("Room closed: room=%v", r.Id)
			break Loop
		case msg := <-r.msgCh:
			r.logger.Debugf("Room msg: room=%v, %T %v", r.Id, msg, msg)
			r.dispatch(msg)
		}
	}
	r.repo.RemoveRoom(r)

	r.drainMsg()
	r.logger.Debugf("Room.MsgLoop() finish: room=%v", r.Id)
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
			r.logger.Debugf("Discard msg: room=%v %T %v", r.Id, msg, msg)
		case <-ch:
			return
		}
	}
}

// Done returns a channel which cloased when room is done.
func (r *Room) Done() <-chan struct{} {
	return r.done
}

// Timeout : client側でtimeout検知したとき. Client.MsgLoopから呼ばれる
func (r *Room) Timeout(c *Client) {
	r.removeClient(c, xerrors.Errorf("client timeout: %v", c.Id))
}

func (r *Room) removeClient(c *Client, err error) {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	cid := ClientID(c.Id)

	if _, ok := r.clients[cid]; !ok {
		r.logger.Debugf("Client may be aleady leaved: room=%v, client=%v", r.Id, cid)
		return
	}

	r.logger.Infof("Client removed: room=%v, client=%v %v", r.Id, cid, err)
	delete(r.clients, cid)
	// todo: orderの書き換え

	c.Removed(err)

	if len(r.clients) == 0 {
		close(r.done)
		return
	}

	r.broadcast(binary.NewEvLeave(string(cid)))
}

func (r *Room) dispatch(msg Msg) error {
	switch m := msg.(type) {
	case *MsgCreate:
		return r.msgCreate(m)
	case *MsgJoin:
		return r.msgJoin(m)
	case *MsgLeave:
		return r.msgLeave(m)
	case *MsgBroadcast:
		return r.msgBroadcast(m)
	case *MsgClientError:
		return r.msgClientError(m)
	default:
		return xerrors.Errorf("unknown msg type: %T %v", m, m)
	}
}

func (r *Room) broadcast(ev *binary.Event) {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	for _, c := range r.clients {
		if err := c.Send(ev); err != nil {
			// removeClient locks muClients so that must be called another goroutine.
			go r.removeClient(c, err)
		}
	}
}

func (r *Room) msgCreate(msg *MsgCreate) error {
	rinfo := r.RoomInfo.Clone()
	cinfo := r.master.ClientInfo.Clone()
	msg.Joined <- JoinedInfo{rinfo, cinfo}
	r.broadcast(binary.NewEvJoined(cinfo))
	return nil
}

func (r *Room) msgJoin(msg *MsgJoin) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	if r.MaxPlayers == uint32(len(r.clients)) {
		close(msg.Joined)
		return xerrors.Errorf("Room full. room=%v max=%v, client=%v", r.ID(), r.MaxPlayers, msg.Info.Id)
	}

	r.wgClient.Add(1)
	c := NewClient(msg.Info, r)
	r.clients[ClientID(c.Id)] = c

	rinfo := r.RoomInfo.Clone()
	cinfo := c.ClientInfo.Clone()
	msg.Joined <- JoinedInfo{rinfo, cinfo}
	r.broadcast(binary.NewEvJoined(cinfo))
	return nil
}

func (r *Room) msgLeave(msg *MsgLeave) error {
	r.removeClient(msg.Sender, nil)
	return nil
}

func (r *Room) msgBroadcast(msg *MsgBroadcast) error {
	r.broadcast(binary.NewEvMessage(msg.Sender.Id, msg.Payload))
	return nil
}

func (r *Room) msgClientError(msg *MsgClientError) error {
	r.removeClient(msg.Sender, msg.Err)
	return nil
}
