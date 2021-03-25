package game

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"

	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	// RoomMsgChSize : Msgチャネルのバッファサイズ
	RoomMsgChSize = 10
)

type Room struct {
	*pb.RoomInfo
	repo *Repository

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

	lastMsg binary.Dict // map[clientID]unixtime_millisec

	logger *zap.SugaredLogger
}

func NewRoom(ctx context.Context, repo *Repository, info *pb.RoomInfo, masterInfo *pb.ClientInfo, deadlineSec uint32, conf *config.GameConf, loglevel log.Level) (*Room, *JoinedInfo, ErrorWithCode) {
	pubProps, iProps, err := common.InitProps(info.PublicProps)
	if err != nil {
		return nil, nil, WithCode(xerrors.Errorf("PublicProps unmarshal error: %w", err), codes.InvalidArgument)
	}
	info.PublicProps = iProps
	privProps, iProps, err := common.InitProps(info.PrivateProps)
	if err != nil {
		return nil, nil, WithCode(xerrors.Errorf("PrivateProps unmarshal error: %w", err), codes.InvalidArgument)
	}
	info.PrivateProps = iProps

	r := &Room{
		RoomInfo: info,
		repo:     repo,
		deadline: time.Duration(deadlineSec) * time.Second,

		publicProps:  pubProps,
		privateProps: privProps,

		msgCh: make(chan Msg, RoomMsgChSize),
		done:  make(chan struct{}),

		players:     make(map[ClientID]*Client),
		masterOrder: []ClientID{},
		watchers:    make(map[ClientID]*Client),
		lastMsg:     make(binary.Dict),

		logger: log.Get(loglevel).With(zap.String("room", info.Id)).Sugar(),
	}

	go r.MsgLoop()

	jch := make(chan *JoinedInfo, 1)
	ech := make(chan ErrorWithCode, 1)

	select {
	case <-ctx.Done():
		return nil, nil, WithCode(
			xerrors.Errorf("NewRoom write msg timeout or context done: room=%v client=%v", r.Id, masterInfo.Id),
			codes.DeadlineExceeded)
	case r.msgCh <- &MsgCreate{masterInfo, jch, ech}:
	}

	r.logger.Debugf("NewRoom: info={%v}, pubProp:%v, privProp:%v", r.RoomInfo, r.publicProps, r.privateProps)

	select {
	case <-ctx.Done():
		return nil, nil, WithCode(
			xerrors.Errorf("NewRoom msgCreate timeout or context done: room=%v client=%v", r.Id, masterInfo.Id),
			codes.DeadlineExceeded)
	case ewc := <-ech:
		return nil, nil, WithCode(
			xerrors.Errorf("NewRoom msgCreate: %w", ewc), ewc.Code())
	case joined := <-jch:
		return r, joined, nil
	}
}

func (r *Room) ID() RoomID {
	return RoomID(r.Id)
}

// MsgLoop goroutine dispatch messages.
func (r *Room) MsgLoop() {
	r.logger.Debug("Room.MsgLoop() start.")
Loop:
	for {
		select {
		case <-r.Done():
			r.logger.Info("Room closed.")
			break Loop
		case msg := <-r.msgCh:
			r.logger.Debugf("Room msg: %T %v", msg, msg)
			r.updateLastMsg(msg.SenderID())
			if err := r.dispatch(msg); err != nil {
				r.logger.Errorf("Room msg error: %v", err)
			}
		}
	}
	r.repo.RemoveRoom(r)

	r.drainMsg()
	r.logger.Debug("Room.MsgLoop() finish")
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
			r.logger.Debugf("Discard msg: %T %v", msg, msg)
		case <-ch:
			return
		}
	}
}

// Done returns a channel which cloased when room is done.
func (r *Room) Done() <-chan struct{} {
	return r.done
}

func (r *Room) writeLastMsg(cid ClientID) {
	millisec := uint64(time.Now().UnixNano()) / 1000000
	r.lastMsg[string(cid)] = binary.MarshalULong(millisec)
}

func (r *Room) removeLastMsg(cid ClientID) {
	delete(r.lastMsg, string(cid))
}

/// UpdateLastMsg : PlayerがMsgを受信したとき更新する.
/// 既に登録されているPlayerのみ書き込み (watcherを含めないため)
func (r *Room) updateLastMsg(cid ClientID) {
	id := string(cid)
	if _, ok := r.lastMsg[id]; ok {
		r.writeLastMsg(cid)
	}
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
		r.logger.Debugf("Player may be aleady left:  client=%v", cid)
		return
	}

	r.logger.Infof("Player removed: client=%v %v", cid, err)
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
		r.logger.Infof("Master switched: master:%v->%v", r.master.Id, r.masterOrder[0])
		r.master = r.players[r.masterOrder[0]]
	}

	r.RoomInfo.Players = uint32(len(r.players))
	r.repo.updateRoomInfo(r)

	r.broadcast(binary.NewEvLeft(string(cid), r.master.Id))

	r.removeLastMsg(cid)
}

func (r *Room) removeWatcher(c *Client, err error) {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	cid := c.ID()

	if _, ok := r.watchers[cid]; !ok {
		r.logger.Debugf("Watcher may be aleady left: client=%v", cid)
		return
	}

	r.logger.Infof("Watcher removed: client=%v %v", cid, err)
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
	case *MsgPing:
		return r.msgPing(m)
	case *MsgNodeCount:
		return r.msgNodeCount(m)
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
	case *MsgSwitchMaster:
		return r.msgSwitchMaster(m)
	case *MsgKick:
		return r.msgKick(m)
	case *MsgClientError:
		return r.msgClientError(m)
	default:
		return xerrors.Errorf("unknown msg type: %T %v", m, m)
	}
}

// sendTo : 特定クライアントに送信.
// muClients のロックを取得してから呼び出す.
// 送信できない場合続行不能なので退室させる.
func (r *Room) sendTo(c *Client, ev *binary.RegularEvent) error {
	err := c.Send(ev)
	if err != nil {
		// removeClient locks muClients so that must be called another goroutine.
		go r.removeClient(c, err)
	}
	return err
}

// broadcast : 全員に送信.
// muClients のロックを取得してから呼び出すこと
func (r *Room) broadcast(ev *binary.RegularEvent) {
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
		msg.Err <- err
		return xerrors.Errorf("NewPlayer error. room=%v, client=%v: %w", r.ID(), msg.Info.Id, err)
	}
	r.master = master
	r.players[master.ID()] = master
	r.masterOrder = append(r.masterOrder, master.ID())

	rinfo := r.RoomInfo.Clone()
	cinfo := r.master.ClientInfo.Clone()
	players := []*pb.ClientInfo{cinfo}
	msg.Joined <- &JoinedInfo{rinfo, players, master, master.ID(), r.deadline}
	r.broadcast(binary.NewEvJoined(cinfo))

	r.writeLastMsg(master.ID())

	return nil
}

func (r *Room) msgJoin(msg *MsgJoin) error {
	if !r.Joinable {
		err := xerrors.Errorf("Room is not joinable. room=%v, client=%v", r.ID(), msg.Info.Id)
		msg.Err <- WithCode(err, codes.FailedPrecondition)
		return err
	}

	r.muClients.Lock()
	defer r.muClients.Unlock()

	if _, ok := r.players[msg.SenderID()]; ok {
		err := xerrors.Errorf("Player already exists. room=%v, client=%v", r.ID(), msg.SenderID())
		msg.Err <- WithCode(err, codes.AlreadyExists)
		return err
	}
	// hub経由で観戦している場合は考慮しない
	if _, ok := r.watchers[msg.SenderID()]; ok {
		err := xerrors.Errorf("Player already exists as a watcher. room=%v, client=%v", r.ID(), msg.SenderID())
		msg.Err <- WithCode(err, codes.AlreadyExists)
		return err
	}

	if r.MaxPlayers == uint32(len(r.players)) {
		err := xerrors.Errorf("Room full. room=%v max=%v, client=%v", r.ID(), r.MaxPlayers, msg.Info.Id)
		msg.Err <- WithCode(err, codes.ResourceExhausted)
		return err
	}

	client, err := NewPlayer(msg.Info, r)
	if err != nil {
		err = WithCode(
			xerrors.Errorf("NewPlayer error. room=%v, client=%v: %w", r.ID(), msg.Info.Id, err),
			err.Code())
		msg.Err <- err
		return err
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
	msg.Joined <- &JoinedInfo{rinfo, players, client, r.master.ID(), r.deadline}
	r.broadcast(binary.NewEvJoined(cinfo))

	r.writeLastMsg(client.ID())

	return nil
}

func (r *Room) msgWatch(msg *MsgWatch) error {
	if !r.Watchable {
		err := xerrors.Errorf("Room is not watchable. room=%v, client=%v", r.ID(), msg.Info.Id)
		msg.Err <- WithCode(err, codes.FailedPrecondition)
		return err
	}

	r.muClients.Lock()
	defer r.muClients.Unlock()

	if _, ok := r.players[msg.SenderID()]; ok {
		err := xerrors.Errorf("Watcher already exists as a player. room=%v, client=%v", r.ID(), msg.SenderID())
		msg.Err <- WithCode(err, codes.AlreadyExists)
		return err
	}
	// hub経由で観戦している場合は考慮しない
	if _, ok := r.watchers[msg.SenderID()]; ok {
		err := xerrors.Errorf("Watcher already exists. room=%v, client=%v", r.ID(), msg.SenderID())
		msg.Err <- WithCode(err, codes.AlreadyExists)
		return err
	}

	client, err := NewWatcher(msg.Info, r)
	if err != nil {
		err = WithCode(
			xerrors.Errorf("NewWatcher error. room=%v, client=%v: %w", r.ID(), msg.Info.Id, err),
			err.Code())
		msg.Err <- err
		return err
	}
	r.watchers[client.ID()] = client
	r.RoomInfo.Watchers += client.nodeCount
	r.repo.updateRoomInfo(r)

	rinfo := r.RoomInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(r.players))
	for _, c := range r.players {
		players = append(players, c.ClientInfo.Clone())
	}

	msg.Joined <- &JoinedInfo{rinfo, players, client, r.master.ID(), r.deadline}
	return nil
}

func (r *Room) msgPing(msg *MsgPing) error {
	ev := binary.NewEvPong(msg.Timestamp, r.RoomInfo.Watchers, r.lastMsg)
	return msg.Sender.SendSystemEvent(ev)
}

func (r *Room) msgNodeCount(msg *MsgNodeCount) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	c := msg.Sender
	if c.nodeCount == msg.Count {
		return nil
	}
	r.RoomInfo.Watchers = (r.RoomInfo.Watchers - c.nodeCount) + msg.Count
	r.logger.Debugf("Update NodeCount: %v before=%v after=%v (total=%v)", c.ID(), c.nodeCount, msg.Count, r.RoomInfo.Watchers)
	c.nodeCount = msg.Count
	r.repo.updateRoomInfo(r)
	return nil
}

func (r *Room) msgLeave(msg *MsgLeave) error {
	r.removeClient(msg.Sender, nil)
	return nil
}

func (r *Room) msgRoomProp(msg *MsgRoomProp) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	if msg.Sender != r.master {
		// 送信元にエラー通知
		r.sendTo(msg.Sender, binary.NewEvPermissionDenied(msg))
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
		r.logger.Debugf("Room update PublicProps: %v", r.publicProps)
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
		r.logger.Debugf("Room update PrivateProps: %v", r.privateProps)
	}

	r.repo.updateRoomInfo(r)

	if msg.ClientDeadline != 0 {
		deadline := time.Duration(msg.ClientDeadline) * time.Second
		if deadline != r.deadline {
			r.logger.Debugf("Room notify new deadline: %v", deadline)
			r.deadline = deadline
			for _, c := range r.players {
				c.newDeadline <- deadline
			}
		}
	}

	r.sendTo(msg.Sender, binary.NewEvSucceeded(msg))
	r.broadcast(binary.NewEvRoomProp(msg.Sender.Id, msg.MsgRoomPropPayload))
	return nil
}

func (r *Room) msgClientProp(msg *MsgClientProp) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	if !msg.Sender.isPlayer {
		// 送信元にエラー通知
		r.sendTo(msg.Sender, binary.NewEvPermissionDenied(msg))
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

	r.sendTo(msg.Sender, binary.NewEvSucceeded(msg))
	r.broadcast(binary.NewEvClientProp(msg.Sender.Id, msg.Payload()))
	return nil
}

func (r *Room) msgTargets(msg *MsgTargets) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	ev := binary.NewEvMessage(msg.Sender.Id, msg.Data)

	absent := make([]string, 0, len(r.players))

	for _, t := range msg.Targets {
		c, ok := r.players[ClientID(t)]
		if !ok {
			r.logger.Infof("target %s is absent", t)
			absent = append(absent, t)
			continue
		}
		_ = r.sendTo(c, ev)
	}

	// 居なかった人を通知
	if len(absent) > 0 {
		r.sendTo(msg.Sender, binary.NewEvTargetNotFound(msg, absent))
	}

	return nil
}

func (r *Room) msgToMaster(msg *MsgToMaster) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	_ = r.sendTo(r.master, binary.NewEvMessage(msg.Sender.Id, msg.Data))
	return nil
}

func (r *Room) msgBroadcast(msg *MsgBroadcast) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	r.broadcast(binary.NewEvMessage(msg.Sender.Id, msg.Data))
	return nil
}

func (r *Room) msgSwitchMaster(msg *MsgSwitchMaster) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	if msg.Sender != r.master {
		// 送信元にエラー通知
		r.sendTo(msg.Sender, binary.NewEvPermissionDenied(msg))
		return xerrors.Errorf("MsgSwitchMaster: sender %q is not master %q", msg.Sender.Id, r.master.Id)
	}

	target, found := r.players[msg.Target]
	if !found {
		// 対象が居ないことを通知
		r.sendTo(msg.Sender, binary.NewEvTargetNotFound(msg, []string{string(msg.Target)}))
		return xerrors.Errorf("MsgSwitchMaster: player not found: room=%v, target=%v", r.Id, msg.Target)
	}

	r.master = target

	r.logger.Debugf("Master switched: master:%v->%v", msg.Sender.ID(), r.master.Id)

	r.sendTo(msg.Sender, binary.NewEvSucceeded(msg))
	r.broadcast(binary.NewEvMasterSwitched(msg.Sender.Id, r.master.Id))
	return nil
}

func (r *Room) msgKick(msg *MsgKick) error {
	r.muClients.RLock()

	if msg.Sender != r.master {
		// 送信元にエラー通知
		r.sendTo(msg.Sender, binary.NewEvPermissionDenied(msg))
		r.muClients.RUnlock()
		return xerrors.Errorf("MsgKick: sender %q is not master %q", msg.Sender.Id, r.master.Id)
	}

	target, found := r.players[msg.Target]
	if !found {
		// 対象が居ないことを通知
		r.sendTo(msg.Sender, binary.NewEvTargetNotFound(msg, []string{string(msg.Target)}))
		r.muClients.RUnlock()
		return xerrors.Errorf("MsgKick: player not found: room=%v, target=%v", r.Id, msg.Target)
	}

	r.sendTo(msg.Sender, binary.NewEvSucceeded(msg))

	// removeClientがmuClientsのロックを取るため呼び出し前にUnlockしておく
	r.muClients.RUnlock()

	r.removeClient(target, xerrors.Errorf("client kicked: room=%v client=%v", r.ID(), target.Id))
	return nil
}

func (r *Room) msgClientError(msg *MsgClientError) error {
	r.removeClient(msg.Sender, msg.Err)
	return nil
}

// IRoom実装

func (r *Room) Deadline() time.Duration {
	return r.deadline
}

func (r *Room) WaitGroup() *sync.WaitGroup {
	return &r.wgClient
}

func (r *Room) Logger() *zap.SugaredLogger {
	return r.logger
}

func (r *Room) SendMessage(msg Msg) {
	r.msgCh <- msg
}

func (r *Room) Repo() IRepo {
	return r.repo
}
