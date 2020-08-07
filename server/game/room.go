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
	// roomKeyLen : Roomキーの文字列長
	roomKeyLen = 16
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

	muClients   sync.RWMutex
	players     map[ClientID]*Client
	master      *Client
	masterOrder []ClientID
	watchers    map[ClientID]*Client

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

func NewRoom(repo *Repository, info *pb.RoomInfo, masterInfo *pb.ClientInfo, deadlineSec uint32, conf *config.GameConf, loglevel log.Level) (*Room, <-chan JoinedInfo, error) {
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
		key:      RandomHex(roomKeyLen),
		deadline: time.Duration(deadlineSec) * time.Second,

		publicProps:  pubProps,
		privateProps: privProps,

		msgCh: make(chan Msg, RoomMsgChSize),
		done:  make(chan struct{}),

		players:     make(map[ClientID]*Client),
		masterOrder: []ClientID{},
		watchers:    make(map[ClientID]*Client),

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
				r.logger.Errorf("Room msg error: %v", err)
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
	if c.isPlayer {
		r.removePlayer(c, err)
	} else {
		r.removeWatcher(c, err)
	}
}

func (r *Room) removePlayer(c *Client, err error) {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	cid := c.ID()

	if _, ok := r.players[cid]; !ok {
		r.logger.Debugf("Player may be aleady left: room=%v, client=%v", r.Id, cid)
		return
	}

	r.logger.Infof("Player removed: room=%v, client=%v %v", r.Id, cid, err)
	delete(r.players, cid)

	for i, id := range r.masterOrder {
		if id == cid {
			r.masterOrder = append(r.masterOrder[:i], r.masterOrder[i+1:]...)
			break
		}
	}

	c.Removed(err)

	if len(r.players) == 0 {
		close(r.done)
		return
	}

	if r.master.ID() == cid {
		r.logger.Infof("Master switched: room=%v master:%v->%v", r.Id, r.master.Id, r.masterOrder[0])
		r.master = r.players[r.masterOrder[0]]
	}

	r.RoomInfo.Players = uint32(len(r.players))
	r.repo.updateRoomInfo(r)

	r.broadcast(binary.NewEvLeft(string(cid), r.master.Id))
}

func (r *Room) removeWatcher(c *Client, err error) {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	cid := c.ID()

	if _, ok := r.watchers[cid]; !ok {
		r.logger.Debugf("Watcher may be aleady left: room=%v, client=%v", r.Id, cid)
		return
	}

	r.logger.Infof("Watcher removed: room=%v, client=%v %v", r.Id, cid, err)
	delete(r.watchers, cid)

	r.RoomInfo.Watchers -= c.nodeCount
	r.repo.updateRoomInfo(r)

	c.Removed(err)
}

func (r *Room) dispatch(msg Msg) error {
	switch m := msg.(type) {
	case *MsgCreate:
		return r.msgCreate(m)
	case *MsgJoin:
		return r.msgJoin(m)
	case *MsgWatch:
		return r.msgWatch(m)
	case *MsgLeave:
		return r.msgLeave(m)
	case *MsgRoomProp:
		return r.msgRoomProp(m)
	case *MsgClientProp:
		return r.msgClientProp(m)
	case *MsgTargets:
		return r.msgTargets(m)
	case *MsgToMaster:
		return r.msgToMaster(m)
	case *MsgBroadcast:
		return r.msgBroadcast(m)
	case *MsgClientError:
		return r.msgClientError(m)
	default:
		return xerrors.Errorf("unknown msg type: %T %v", m, m)
	}
}

// sendTo : 特定クライアントに送信.
// muClients のロックを取得してから呼び出す.
// 送信できない場合続行不能なので退室させる.
func (r *Room) sendTo(c *Client, ev *binary.Event) error {
	err := c.Send(ev)
	if err != nil {
		// removeClient locks muClients so that must be called another goroutine.
		go r.removeClient(c, err)
	}
	return err
}

// broadcast : 全員に送信.
// muClients のロックを取得してから呼び出すこと
func (r *Room) broadcast(ev *binary.Event) {
	for _, c := range r.players {
		_ = r.sendTo(c, ev)
	}
	for _, c := range r.watchers {
		_ = r.sendTo(c, ev)
	}
}

func (r *Room) msgCreate(msg *MsgCreate) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	master, err := NewPlayer(msg.Info, r)
	if err != nil {
		close(msg.Joined)
		return xerrors.Errorf("NewPlayer error. room=%v, client=%v, err=%w", r.ID(), msg.Info.Id, err)
	}
	r.master = master
	r.players[master.ID()] = master
	r.masterOrder = append(r.masterOrder, master.ID())

	rinfo := r.RoomInfo.Clone()
	cinfo := r.master.ClientInfo.Clone()
	players := []*pb.ClientInfo{cinfo}
	msg.Joined <- JoinedInfo{rinfo, players, master, master.ID(), r.deadline}
	r.broadcast(binary.NewEvJoined(cinfo))
	return nil
}

func (r *Room) msgJoin(msg *MsgJoin) error {
	if !r.Joinable {
		close(msg.Joined)
		return xerrors.Errorf("Room is not joinable. room=%v, client=%v", r.ID(), msg.Info.Id)
	}

	r.muClients.Lock()
	defer r.muClients.Unlock()

	if r.MaxPlayers == uint32(len(r.players)) {
		close(msg.Joined)
		return xerrors.Errorf("Room full. room=%v max=%v, client=%v", r.ID(), r.MaxPlayers, msg.Info.Id)
	}

	client, err := NewPlayer(msg.Info, r)
	if err != nil {
		close(msg.Joined)
		return xerrors.Errorf("NewPlayer error. room=%v, client=%v, err=%w", r.ID(), msg.Info.Id, err)
	}
	r.players[client.ID()] = client
	r.masterOrder = append(r.masterOrder, client.ID())
	r.RoomInfo.Players = uint32(len(r.players))
	r.repo.updateRoomInfo(r)

	rinfo := r.RoomInfo.Clone()
	cinfo := client.ClientInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(r.players))
	for _, c := range r.players {
		players = append(players, c.ClientInfo.Clone())
	}
	msg.Joined <- JoinedInfo{rinfo, players, client, r.master.ID(), r.deadline}
	r.broadcast(binary.NewEvJoined(cinfo))
	return nil
}

func (r *Room) msgWatch(msg *MsgWatch) error {
	if !r.Watchable {
		close(msg.Joined)
		return xerrors.Errorf("Room is not watchable. room=%v, client=%v", r.ID(), msg.Info.Id)
	}

	r.muClients.Lock()
	defer r.muClients.Unlock()

	client, err := NewWatcher(msg.Info, r)
	if err != nil {
		close(msg.Joined)
		return xerrors.Errorf("NewWatcher error. room=%v, client=%v, err=%w", r.ID(), msg.Info.Id, err)
	}
	r.watchers[client.ID()] = client
	r.RoomInfo.Watchers += client.nodeCount
	r.repo.updateRoomInfo(r)

	rinfo := r.RoomInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(r.players))
	for _, c := range r.players {
		players = append(players, c.ClientInfo.Clone())
	}

	msg.Joined <- JoinedInfo{rinfo, players, client, r.master.ID(), r.deadline}
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

	r.RoomInfo.Visible = msg.Visible
	r.RoomInfo.Joinable = msg.Joinable
	r.RoomInfo.Watchable = msg.Watchable
	r.RoomInfo.SearchGroup = msg.SearchGroup
	r.RoomInfo.MaxPlayers = msg.MaxPlayer

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

	r.muClients.RLock()
	defer r.muClients.RUnlock()

	if msg.ClientDeadline != 0 {
		deadline := time.Duration(msg.ClientDeadline) * time.Second
		if deadline != r.deadline {
			r.logger.Debugf("Room notify new deadline: room=%v %v", r.Id, deadline)
			r.deadline = deadline
			for _, c := range r.players {
				c.newDeadline <- deadline
			}
		}
	}

	r.broadcast(binary.NewEvRoomProp(msg.Sender.Id, msg.MsgRoomPropPayload))
	return nil
}

func (r *Room) msgClientProp(msg *MsgClientProp) error {
	if !msg.Sender.isPlayer {
		return xerrors.Errorf("MsgClientProp: sender %q is not player", msg.Sender.Id)
	}

	r.logger.Debugf("MsgClientProp: client=%v, props=%v", msg.Sender.Id, msg.Props)
	if len(msg.Props) > 0 {
		c := msg.Sender
		for k, v := range msg.Props {
			if _, ok := c.props[k]; ok && len(v) == 0 {
				delete(c.props, k)
			} else {
				c.props[k] = v
			}
		}
		c.ClientInfo.Props = binary.MarshalDict(c.props)
		r.logger.Debugf("Client update Props: client=%v %v", c.Id, c.props)
	}

	r.muClients.RLock()
	defer r.muClients.RUnlock()
	r.broadcast(binary.NewEvClientProp(msg.Sender.Id, msg.Payload()))
	return nil
}

func (r *Room) msgTargets(msg *MsgTargets) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	ev := binary.NewEvMessage(msg.Sender.Id, msg.Data)

	// todo: 居なかった人を通知したほうがいいかも？

	for _, t := range msg.Targets {
		c, ok := r.players[ClientID(t)]
		if !ok {
			r.logger.Infof("target %s is absent", t)
			continue
		}
		_ = r.sendTo(c, ev)
	}

	return nil
}

func (r *Room) msgToMaster(msg *MsgToMaster) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	// todo: 送信できなかったら通知したい
	_ = r.sendTo(r.master, binary.NewEvMessage(msg.Sender.Id, msg.Data))
	return nil
}

func (r *Room) msgBroadcast(msg *MsgBroadcast) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	r.broadcast(binary.NewEvMessage(msg.Sender.Id, msg.Data))
	return nil
}

func (r *Room) msgClientError(msg *MsgClientError) error {
	r.removeClient(msg.Sender, msg.Err)
	return nil
}
