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

	key string

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
	um, _, err := binary.Unmarshal(props)
	if err != nil {
		return nil, nil, err
	}
	dict, ok := um.(binary.Dict)
	if !ok {
		return nil, nil, xerrors.Errorf("type is not Dict: %v", binary.Type(props[0]))
	}
	return dict, props, nil
}

func NewRoom(repo *Repository, info *pb.RoomInfo, masterInfo *pb.ClientInfo, conf *config.GameConf, loglevel log.Level) (*Room, <-chan JoinedInfo, error) {
	pubProps, iProps, err := initProps(info.PublicProps)
	if err != nil {
		return nil, nil, xerrors.Errorf("PublicProps unmarshal error: %w", err)
	}
	info.PublicProps = iProps
	privProps, iProps, err := initProps(info.PrivateProps)
	if err != nil {
		return nil, nil, xerrors.Errorf("PrivateProps unmarshal error: %w", err)
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

		logger: log.Get(loglevel),
	}

	go r.MsgLoop()

	ch := make(chan JoinedInfo)
	r.msgCh <- &MsgCreate{masterInfo, ch}

	r.logger.Debugf("NewRoom: info={%v}, pubProp:%v, privProp:%v", r.RoomInfo, r.publicProps, r.privateProps)

	return r, ch, nil
}

func (r *Room) ID() RoomID {
	return RoomID(r.Id)
}

func (r *Room) Key() string {
	return r.key
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
			if err := r.dispatch(msg); err != nil {
				r.logger.Infof("Room msg error: %v", err)
			}
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
	case *MsgRoomProp:
		return r.msgRoomProp(m)
	case *MsgBroadcast:
		return r.msgBroadcast(m)
	case *MsgClientError:
		return r.msgClientError(m)
	default:
		return xerrors.Errorf("unknown msg type: %T %v", m, m)
	}
}

// muClients のロックを取得してから呼び出すこと
func (r *Room) broadcast(ev *binary.Event) {
	for _, c := range r.clients {
		if err := c.Send(ev); err != nil {
			// removeClient locks muClients so that must be called another goroutine.
			go r.removeClient(c, err)
		}
	}
}

func (r *Room) notifyDeadline(deadline time.Duration) {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	for _, c := range r.clients {
		c.newDeadline <- deadline
	}
}

func (r *Room) msgCreate(msg *MsgCreate) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	r.wgClient.Add(1)
	master := NewClient(msg.Info, r)
	r.master = master
	r.clients[ClientID(master.Id)] = master
	r.order = append(r.order, master.ID())

	rinfo := r.RoomInfo.Clone()
	cinfo := r.master.ClientInfo.Clone()
	players := []*pb.ClientInfo{cinfo}
	msg.Joined <- JoinedInfo{rinfo, players, master}
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
	client := NewClient(msg.Info, r)
	r.clients[ClientID(client.Id)] = client
	r.RoomInfo.Players = uint32(len(r.clients))
	r.repo.updateRoomInfo(r)

	rinfo := r.RoomInfo.Clone()
	cinfo := client.ClientInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(r.clients))
	for _, c := range r.clients {
		players = append(players, c.ClientInfo.Clone())
	}
	msg.Joined <- JoinedInfo{rinfo, players, client}
	r.broadcast(binary.NewEvJoined(cinfo))
	return nil
}

func (r *Room) msgLeave(msg *MsgLeave) error {
	r.removeClient(msg.Sender, nil)
	return nil
}

func (r *Room) msgRoomProp(msg *MsgRoomProp) error {
	if msg.Sender != r.master {
		return xerrors.Errorf("MsgRoomProp: sender %q is not master %q", msg.Sender.Id, r.master.Id)
	}
	r.logger.Debugf("Room MsgRoomProps: %v", msg.MsgRoomPropPayload)

	deadlineUpdated := r.ClientDeadline != msg.ClientDeadline
	r.RoomInfo.Visible = msg.Visible
	r.RoomInfo.Joinable = msg.Joinable
	r.RoomInfo.Watchable = msg.Watchable
	r.RoomInfo.SearchGroup = msg.SearchGroup
	r.RoomInfo.MaxPlayers = msg.MaxPlayer
	r.RoomInfo.ClientDeadline = msg.ClientDeadline

	if len(msg.PublicProps) > 0 {
		for k, v := range msg.PublicProps {
			if _, ok := r.publicProps[k]; ok && len(v) == 0 {
				delete(r.publicProps, k)
			} else {
				r.publicProps[k] = v
			}
		}
		r.RoomInfo.PublicProps = binary.MarshalDict(r.publicProps)
		r.logger.Debugf("Room update PublicProps: room=%v %v", r.Id, r.publicProps)
	}

	if len(msg.PrivateProps) > 0 {
		for k, v := range msg.PrivateProps {
			if _, ok := r.privateProps[k]; ok && len(v) == 0 {
				delete(r.privateProps, k)
			} else {
				r.privateProps[k] = v
			}
		}
		r.RoomInfo.PrivateProps = binary.MarshalDict(r.privateProps)
		r.logger.Debugf("Room update PrivateProps: room=%v %v", r.Id, r.privateProps)
	}

	r.repo.updateRoomInfo(r)

	if deadlineUpdated {
		r.deadline = time.Duration(msg.ClientDeadline) * time.Second
		r.logger.Debugf("Room notify new deadline: room=%v %v", r.Id, r.deadline)
		r.notifyDeadline(r.deadline)
	}

	r.muClients.RLock()
	defer r.muClients.RUnlock()
	r.broadcast(binary.NewEvRoomProp(msg.Sender.Id, msg.MsgRoomPropPayload))
	return nil
}

func (r *Room) msgBroadcast(msg *MsgBroadcast) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	r.broadcast(binary.NewEvMessage(msg.Sender.Id, msg.Payload))
	return nil
}

func (r *Room) msgClientError(msg *MsgClientError) error {
	r.removeClient(msg.Sender, msg.Err)
	return nil
}

func (r *Room) getClient(id ClientID) (*Client, error) {
	r.muClients.Lock()
	defer r.muClients.Unlock()
	cli, ok := r.clients[id]
	if !ok {
		return nil, xerrors.Errorf("client not found: client=%v", id)
	}
	return cli, nil
}
