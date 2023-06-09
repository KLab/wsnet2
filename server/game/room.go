package game

import (
	"context"
	"sync"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"

	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/config"
	"wsnet2/log"
	"wsnet2/metrics"
	"wsnet2/pb"
)

const (
	// RoomMsgChSize : Msgチャネルのバッファサイズ
	RoomMsgChSize = 10
)

type Room struct {
	*pb.RoomInfo
	repo *Repository

	conf *config.GameConf

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

	logger log.Logger

	chRoomInfo   chan struct{}
	mRoomInfo    sync.Mutex // used by updateRoomInfo
	lastRoomInfo *pb.RoomInfo
}

func NewRoom(ctx context.Context, repo *Repository, info *pb.RoomInfo, masterInfo *pb.ClientInfo, macKey string, deadlineSec uint32, conf *config.GameConf, logger log.Logger) (*Room, *JoinedInfo, ErrorWithCode) {
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
		conf:     conf,
		deadline: time.Duration(deadlineSec) * time.Second,

		publicProps:  pubProps,
		privateProps: privProps,

		msgCh: make(chan Msg, RoomMsgChSize),
		done:  make(chan struct{}),

		players:     make(map[ClientID]*Client),
		masterOrder: []ClientID{},
		watchers:    make(map[ClientID]*Client),
		lastMsg:     make(binary.Dict),

		logger: logger,

		chRoomInfo:   make(chan struct{}, 1),
		lastRoomInfo: info.Clone(),
	}

	go r.MsgLoop()
	go r.roomInfoUpdater()

	jch := make(chan *JoinedInfo, 1)
	ech := make(chan ErrorWithCode, 1)

	select {
	case <-ctx.Done():
		return nil, nil, WithCode(
			xerrors.Errorf("write msg timeout or context done: room=%v client=%v", r.Id, masterInfo.Id),
			codes.DeadlineExceeded)
	case r.msgCh <- &MsgCreate{masterInfo, macKey, jch, ech}:
	}

	select {
	case <-ctx.Done():
		return nil, nil, WithCode(
			xerrors.Errorf("msgCreate timeout or context done: room=%v client=%v", r.Id, masterInfo.Id),
			codes.DeadlineExceeded)
	case ewc := <-ech:
		return nil, nil, WithCode(
			xerrors.Errorf("msgCreate: %w", ewc), ewc.Code())
	case joined := <-jch:
		return r, joined, nil
	}
}

func (r *Room) ID() RoomID {
	return RoomID(r.Id)
}

func (r *Room) ClientConf() *config.ClientConf {
	return &r.conf.ClientConf
}

// MsgLoop goroutine dispatch messages.
func (r *Room) MsgLoop() {
	metrics.Rooms.Add(1)
	defer metrics.Rooms.Add(-1)
Loop:
	for {
		select {
		case <-r.Done():
			r.logger.Infof("room closed: %v", r.Id)
			break Loop
		case msg := <-r.msgCh:
			r.updateLastMsg(msg.SenderID())
			if err := r.dispatch(msg); err != nil {
				r.logger.Errorf("dispatch %T: %+v", msg, err)
			}
		}
	}
	r.repo.RemoveRoom(r)
	r.drainMsg()
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
			r.logger.Debugf("discard msg: %T %v", msg, msg)
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

// UpdateLastMsg : PlayerがMsgを受信したとき更新する.
// 既に登録されているPlayerのみ書き込み (watcherを含めないため)
func (r *Room) updateLastMsg(cid ClientID) {
	id := string(cid)
	if _, ok := r.lastMsg[id]; ok {
		r.writeLastMsg(cid)
	}
}

// removeClient :  Player/Watcherを退室させる.
// muClients のロックを取得してから呼び出す.
func (r *Room) removeClient(c *Client, cause string) {
	if c.isPlayer {
		r.removePlayer(c, cause)
	} else {
		r.removeWatcher(c, cause)
	}
}

func (r *Room) removePlayer(c *Client, cause string) {
	cid := c.ID()

	if r.players[cid] != c {
		c.logger.Infof("player may be aleady removed: %v, %p", cid, c)
		return
	}

	delete(r.players, cid)

	for i, id := range r.masterOrder {
		if id == cid {
			r.masterOrder = append(r.masterOrder[:i], r.masterOrder[i+1:]...)
			break
		}
	}

	r.repo.PlayerLog(c, PlayerLogLeave)

	c.logger.Infof("player left: %v: %v", cid, cause)
	c.Removed(cause)

	if len(r.players) == 0 {
		close(r.done)
		return
	}

	if r.master.ID() == cid {
		r.master = r.players[r.masterOrder[0]]
		r.logger.Infof("master switched: %v -> %v", cid, r.master.ID())
	}

	r.RoomInfo.Players = uint32(len(r.players))
	r.updateRoomInfo()

	r.broadcast(binary.NewEvLeft(string(cid), r.master.Id, cause))

	r.removeLastMsg(cid)
}

func (r *Room) roomInfoUpdater() {
	for {
		select {
		case <-r.done:
			return
		case <-r.chRoomInfo:
			for {
				// mRoomInfo.Lock() はすぐにロック取れるので、先にDB接続を確保する
				t1 := time.Now()
				conn, err := r.repo.db.Connx(context.Background())
				if err != nil {
					r.logger.Errorf("roomInfoUpdater: conn: %+v", err)
					time.Sleep(time.Second)
					continue
				}
				if d := time.Since(t1); d > time.Second {
					r.logger.Infof("roomInfoUpdater: took %v to get a db conn", d)
				}

				r.mRoomInfo.Lock()
				ri := r.lastRoomInfo
				select {
				case <-r.chRoomInfo:
				default:
				}
				r.mRoomInfo.Unlock()

				r.repo.updateRoomInfo(ri, conn, r.logger)
				conn.Close()
				break
			}
		}
	}
}

func (r *Room) updateRoomInfo() {
	r.mRoomInfo.Lock()
	defer r.mRoomInfo.Unlock()
	r.lastRoomInfo = r.RoomInfo.Clone()

	select {
	case r.chRoomInfo <- struct{}{}:
	default:
	}
}

func (r *Room) removeWatcher(c *Client, cause string) {
	cid := c.ID()

	if r.watchers[cid] != c {
		r.logger.Debugf("Watcher may be aleady left: %v, p", cid, c)
		return
	}

	delete(r.watchers, cid)
	c.logger.Infof("watcher left: %v: %v", cid, cause)

	r.RoomInfo.Watchers -= c.nodeCount
	r.updateRoomInfo()
	c.Removed(cause)
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
	case *MsgAdminKick:
		return r.msgAdminKick(m)
	case *MsgGetRoomInfo:
		return r.msgGetRoomInfo(m)
	case *MsgClientError:
		return r.msgClientError(m)
	case *MsgClientTimeout:
		return r.msgClientTimeout(m)
	default:
		return xerrors.Errorf("unknown msg type (%T): %v", m, m)
	}
}

// sendTo : 特定クライアントに送信.
// muClients のロックを取得してから呼び出す.
// 送信できない場合続行不能なので退室させる.
func (r *Room) sendTo(c *Client, ev *binary.RegularEvent) error {
	err := c.Send(ev)
	if err != nil {
		// players/watchersのループ内で呼ばれているため、removeClientは別goroutineで呼ぶ
		go func() {
			r.muClients.Lock()
			r.removeClient(c, err.Error())
			r.muClients.Unlock()
		}()
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

	master, err := NewPlayer(msg.Info, msg.MACKey, r)
	if err != nil {
		msg.Err <- err
		return xerrors.Errorf("NewPlayer room=%v, client=%v: %w", r.Id, msg.Info.Id, err)
	}
	master.logger.Infof("new player: %v", master.Id)

	r.master = master
	r.players[master.ID()] = master
	r.masterOrder = append(r.masterOrder, master.ID())
	r.repo.PlayerLog(master, PlayerLogCreate)

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

	// Timeout前の再入室はclientを差し替え、EvJoinedではなくEvRejoinedを通知
	oldp, rejoin := r.players[msg.SenderID()]
	// 観戦しながらの入室は不許可（ただしhub経由で観戦している場合は考慮しない）
	if _, ok := r.watchers[msg.SenderID()]; ok {
		err := xerrors.Errorf("Player already exists as a watcher. room=%v, client=%v", r.ID(), msg.SenderID())
		msg.Err <- WithCode(err, codes.AlreadyExists)
		return err
	}

	if !rejoin && r.MaxPlayers <= uint32(len(r.players)) {
		err := xerrors.Errorf("Room full. room=%v max=%v, client=%v", r.ID(), r.MaxPlayers, msg.Info.Id)
		msg.Err <- WithCode(err, codes.ResourceExhausted)
		return err
	}

	client, err := NewPlayer(msg.Info, msg.MACKey, r)
	if err != nil {
		err = WithCode(
			xerrors.Errorf("NewPlayer room=%v, client=%v: %w", r.ID(), msg.Info.Id, err),
			err.Code())
		msg.Err <- err
		return err
	}
	r.players[client.ID()] = client
	if rejoin {
		oldp.Removed("client rejoined as a new client")
		if r.master == oldp {
			r.master = client
		}
		r.repo.PlayerLog(client, PlayerLogRejoin)
		client.logger.Infof("rejoin player: %v", client.Id)
	} else {
		r.masterOrder = append(r.masterOrder, client.ID())
		r.repo.PlayerLog(client, PlayerLogJoin)
		r.RoomInfo.Players = uint32(len(r.players))
		r.updateRoomInfo()
		client.logger.Infof("new player: %v", client.Id)
	}

	rinfo := r.RoomInfo.Clone()
	cinfo := client.ClientInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(r.players))
	for _, c := range r.players {
		players = append(players, c.ClientInfo.Clone())
	}
	msg.Joined <- &JoinedInfo{rinfo, players, client, r.master.ID(), r.deadline}
	if rejoin {
		r.broadcast(binary.NewEvRejoined(cinfo))
	} else {
		r.broadcast(binary.NewEvJoined(cinfo))
	}

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

	// Playerとして参加中に観戦は不許可
	if _, ok := r.players[msg.SenderID()]; ok {
		err := xerrors.Errorf("Watcher already exists as a player. room=%v, client=%v", r.ID(), msg.SenderID())
		msg.Err <- WithCode(err, codes.AlreadyExists)
		return err
	}

	client, err := NewWatcher(msg.Info, msg.MACKey, r)
	if err != nil {
		err = WithCode(
			xerrors.Errorf("NewWatcher error. room=%v, client=%v: %w", r.ID(), msg.Info.Id, err),
			err.Code())
		msg.Err <- err
		return err
	}
	oldc, rejoin := r.watchers[client.ID()]
	r.watchers[client.ID()] = client
	if rejoin {
		oldc.Removed("client rejoined as a new client")
		r.RoomInfo.Watchers -= oldc.nodeCount
		client.logger.Infof("rejoin watcher: %v", client.Id)
	} else {
		client.logger.Infof("new watcher: %v", client.Id)
	}
	r.RoomInfo.Watchers += client.nodeCount
	r.updateRoomInfo()

	rinfo := r.RoomInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(r.players))
	for _, c := range r.players {
		players = append(players, c.ClientInfo.Clone())
	}

	msg.Joined <- &JoinedInfo{rinfo, players, client, r.master.ID(), r.deadline}
	return nil
}

func (r *Room) msgPing(msg *MsgPing) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	if msg.Sender.isPlayer {
		if r.players[msg.SenderID()] != msg.Sender {
			return nil
		}
	} else {
		if r.watchers[msg.SenderID()] != msg.Sender {
			return nil
		}
	}
	msg.Sender.logger.Debugf("ping %v: %v", msg.Sender.Id, msg.Timestamp)
	ev := binary.NewEvPong(msg.Timestamp, r.RoomInfo.Watchers, r.lastMsg)
	msg.Sender.SendSystemEvent(ev)
	return nil
}

func (r *Room) msgNodeCount(msg *MsgNodeCount) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	c := msg.Sender
	if r.watchers[c.ID()] != c {
		return nil
	}
	if c.nodeCount == msg.Count {
		return nil
	}
	r.RoomInfo.Watchers = (r.RoomInfo.Watchers - c.nodeCount) + msg.Count
	c.logger.Debugf("nodeCount %v: %v -> %v (total=%v)", c.Id, c.nodeCount, msg.Count, r.RoomInfo.Watchers)
	c.nodeCount = msg.Count
	r.updateRoomInfo()
	return nil
}

func (r *Room) msgLeave(msg *MsgLeave) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	r.removeClient(msg.Sender, msg.Message)
	return nil
}

func (r *Room) msgRoomProp(msg *MsgRoomProp) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	if msg.Sender != r.master {
		// 送信元にエラー通知
		r.sendTo(msg.Sender, binary.NewEvPermissionDenied(msg))
		return xerrors.Errorf("sender %q is not master %q", msg.Sender.Id, r.master.Id)
	}

	msg.Sender.logger.Debugf("update room props: v=%v j=%v w=%v group=%v maxp=%v deadline=%v public=%v private=%v",
		msg.Visible, msg.Joinable, msg.Watchable, msg.SearchGroup, msg.MaxPlayer, msg.ClientDeadline, msg.PublicProps, msg.PrivateProps)

	outputlog := r.RoomInfo.Visible != msg.Visible ||
		r.RoomInfo.Joinable != msg.Joinable ||
		r.RoomInfo.Watchable != msg.Watchable ||
		r.RoomInfo.SearchGroup != msg.SearchGroup ||
		r.RoomInfo.MaxPlayers != msg.MaxPlayer

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
	}

	r.updateRoomInfo()

	if msg.ClientDeadline != 0 {
		deadline := time.Duration(msg.ClientDeadline) * time.Second
		if deadline != r.deadline {
			r.deadline = deadline
			for _, c := range r.players {
				c.newDeadline <- deadline
			}
			outputlog = true
		}
	}

	if outputlog {
		msg.Sender.logger.Infof("room props: v=%v, j=%v, w=%v, group=%v, maxp=%v, deadline=%v",
			r.Visible, r.Joinable, r.Watchable, r.SearchGroup, r.MaxPlayers, r.deadline)
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
		return xerrors.Errorf("sender %q is not player", msg.Sender.Id)
	}
	if r.players[msg.Sender.ID()] != msg.Sender {
		return nil
	}

	msg.Sender.logger.Debugf("update client prop: %v", msg.Props)

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
	}

	r.sendTo(msg.Sender, binary.NewEvSucceeded(msg))
	r.broadcast(binary.NewEvClientProp(msg.Sender.Id, msg.Payload()))
	return nil
}

func (r *Room) msgTargets(msg *MsgTargets) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	if msg.Sender.isPlayer {
		if r.players[msg.SenderID()] != msg.Sender {
			return nil
		}
	} else {
		if r.watchers[msg.SenderID()] != msg.Sender {
			return nil
		}
	}

	msg.Sender.logger.Debugf("message to targets: %v, %v", msg.Targets, msg.Data)

	ev := binary.NewEvMessage(msg.Sender.Id, msg.Data)

	absent := make([]string, 0, len(r.players))

	for _, t := range msg.Targets {
		c, ok := r.players[ClientID(t)]
		if !ok {
			msg.Sender.logger.Infof("target %s is absent", t)
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
	if msg.Sender.isPlayer {
		if r.players[msg.SenderID()] != msg.Sender {
			return nil
		}
	} else {
		if r.watchers[msg.SenderID()] != msg.Sender {
			return nil
		}
	}

	msg.Sender.logger.Debugf("message to master: %v", msg.Data)

	_ = r.sendTo(r.master, binary.NewEvMessage(msg.Sender.Id, msg.Data))
	return nil
}

func (r *Room) msgBroadcast(msg *MsgBroadcast) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	if msg.Sender.isPlayer {
		if r.players[msg.SenderID()] != msg.Sender {
			return nil
		}
	} else {
		if r.watchers[msg.SenderID()] != msg.Sender {
			return nil
		}
	}

	msg.Sender.logger.Debugf("message to all: %v", msg.Data)

	r.broadcast(binary.NewEvMessage(msg.Sender.Id, msg.Data))
	return nil
}

func (r *Room) msgSwitchMaster(msg *MsgSwitchMaster) error {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	if msg.Sender != r.master {
		// 送信元にエラー通知
		r.sendTo(msg.Sender, binary.NewEvPermissionDenied(msg))
		return xerrors.Errorf("sender %q is not master %q", msg.Sender.Id, r.master.Id)
	}

	target, found := r.players[msg.Target]
	if !found {
		// 対象が居ないことを通知
		r.sendTo(msg.Sender, binary.NewEvTargetNotFound(msg, []string{string(msg.Target)}))
		return xerrors.Errorf("player not found: %v", msg.Target)
	}

	r.master = target

	msg.Sender.logger.Infof("master switched: %v -> %v", msg.Sender.ID(), r.master.Id)

	r.sendTo(msg.Sender, binary.NewEvSucceeded(msg))
	r.broadcast(binary.NewEvMasterSwitched(msg.Sender.Id, r.master.Id))
	return nil
}

func (r *Room) msgKick(msg *MsgKick) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()

	if msg.Sender != r.master {
		// 送信元にエラー通知
		r.sendTo(msg.Sender, binary.NewEvPermissionDenied(msg))
		return xerrors.Errorf("sender %q is not master %q", msg.Sender.Id, r.master.Id)
	}

	target, found := r.players[msg.Target]
	if !found {
		// 対象が居ないことを通知
		r.sendTo(msg.Sender, binary.NewEvTargetNotFound(msg, []string{string(msg.Target)}))
		return xerrors.Errorf("player not found: %v", msg.Target)
	}

	r.sendTo(msg.Sender, binary.NewEvSucceeded(msg))

	r.removeClient(target, msg.Message)
	return nil
}

func (r *Room) msgAdminKick(msg *MsgAdminKick) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()
	target, ok := r.players[msg.Target]
	if !ok {
		msg.Res <- xerrors.Errorf("player not found: target=%v", msg.Target)
		return nil
	}

	r.removeClient(target, "kicked by admin")
	msg.Res <- nil
	return nil
}

func (r *Room) msgGetRoomInfo(msg *MsgGetRoomInfo) error {
	ri := r.RoomInfo.Clone()

	r.muClients.RLock()
	defer r.muClients.RUnlock()
	cis := make([]*pb.ClientInfo, 0, len(r.masterOrder))
	for _, id := range r.masterOrder {
		cis = append(cis, r.players[id].ClientInfo.Clone())
	}
	lmt := make(map[string]uint64)
	for p, d := range r.lastMsg {
		t, _, err := binary.UnmarshalAs(d, binary.TypeULong)
		if err != nil {
			return xerrors.Errorf("Unmarshal LastMsg[%s]: %w", p, err)
		}
		lmt[p] = t.(uint64)
	}

	msg.Res <- &pb.GetRoomInfoRes{
		RoomInfo:     ri,
		ClientInfos:  cis,
		MasterId:     r.master.Id,
		LastMsgTimes: lmt,
	}

	return nil
}

func (r *Room) msgClientError(msg *MsgClientError) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()
	r.removeClient(msg.Sender, msg.ErrMsg)
	return nil
}

func (r *Room) msgClientTimeout(msg *MsgClientTimeout) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()
	r.removeClient(msg.Sender, "timeout")
	return nil
}

// IRoom実装

func (r *Room) Deadline() time.Duration {
	return r.deadline
}

func (r *Room) WaitGroup() *sync.WaitGroup {
	return &r.wgClient
}

func (r *Room) Logger() log.Logger {
	return r.logger
}

func (r *Room) SendMessage(msg Msg) {
	select {
	case <-r.done:
	case r.msgCh <- msg:
	}
}

func (r *Room) Repo() IRepo {
	return r.repo
}
