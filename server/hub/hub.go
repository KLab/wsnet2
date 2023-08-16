package hub

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/config"
	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/pb"
)

type Hub struct {
	repo     *Repository
	hubPK    int64
	roomId   RoomID
	appId    AppID
	clientId string

	room *client.Room
	conn *client.Connection

	msgCh chan game.Msg
	done  <-chan struct{}

	watchers map[ClientID]*game.Client
	wgClient sync.WaitGroup

	// game に通知した直近の nodeCount
	lastNodeCount    uint32
	nodeCount        atomic.Uint32
	nodeCountUpdated chan struct{}

	logger log.Logger
}

var _ game.IRoom = &Hub{}

func NewHub(repo *Repository, pk int64, appid AppID, roomid RoomID, grpc *grpc.ClientConn, wsHost string, logger log.Logger) (*Hub, error) {
	// hub->game 接続に使うclientId. このhubを作成するトリガーになったclientIdは使わない
	// roomIdもhostIdもユニークなので hostId:roomId はユニークになるはず。
	clientid := fmt.Sprintf("hub:%d:%s", repo.hostId, roomid)
	clinfo := &pb.ClientInfo{
		Id:    clientid,
		IsHub: true,
	}

	ctx := context.Background() // hubの寿命はリクエストなどに紐付かない

	lg := logger.WithOptions(zap.AddCallerSkip(1))
	room, conn, err := client.WatchDirect(
		ctx, grpc, wsHost, appid, string(roomid), clinfo,
		func(err error) { lg.Warnf("%v: %v", clientid, err) })
	if err != nil {
		return nil, xerrors.Errorf("client.WatchDirect: %w", err)
	}

	done := make(chan struct{})
	go func() {
		msg, err := conn.Wait(ctx)
		close(done)
		if err != nil {
			logger.Errorf("connection closed with error: room=%v %v, %+v", roomid, msg, err)
		} else {
			logger.Infof("connection closed: room=%v %v", roomid, msg)
		}
	}()

	hub := &Hub{
		repo:     repo,
		hubPK:    pk,
		roomId:   roomid,
		clientId: clientid,
		room:     room,
		conn:     conn,
		msgCh:    make(chan game.Msg, game.RoomMsgChSize),
		done:     done,
		watchers: make(map[ClientID]*game.Client),

		nodeCountUpdated: make(chan struct{}, 1),

		logger: logger,
	}

	go hub.ProcessLoop()
	go hub.nodeCountUpdater()

	return hub, nil
}

func (h *Hub) ID() RoomID {
	return h.roomId
}

func (h *Hub) ClientConf() *config.ClientConf {
	return &h.repo.conf.ClientConf
}

func (h *Hub) Repo() game.IRepo {
	return h.repo
}

func (h *Hub) Deadline() time.Duration {
	return time.Duration(h.room.ClientDeadline) * time.Second
}

func (h *Hub) WaitGroup() *sync.WaitGroup {
	return &h.wgClient
}

func (h *Hub) Logger() log.Logger {
	return h.logger
}

func (h *Hub) Done() <-chan struct{} {
	return h.done
}

func (h *Hub) SendMessage(msg game.Msg) {
	select {
	case <-h.done:
	case h.msgCh <- msg:
	}
}

func (h *Hub) removeWatcher(cid game.ClientID, cause string) {
	c, ok := h.watchers[cid]
	if !ok {
		h.logger.Debugf("Watcher may be aleady removed: %v, %p", cid, c)
		return
	}

	h.logger.Infof("Watcher removed: client=%v %v", cid, cause)
	delete(h.watchers, cid)
	h.storeNodeCount()

	c.Removed(cause)
}

func (h *Hub) storeNodeCount() {
	count := uint32(0)
	for _, c := range h.watchers {
		count += c.NodeCount()
	}
	h.nodeCount.Store(count)
	select {
	case h.nodeCountUpdated <- struct{}{}:
	default:
	}
}

func (h *Hub) nodeCountUpdater() {
	// interval以上の間隔をあけ、updateされたら更新する
	interval := time.Duration(h.repo.conf.NodeCountInterval)
	for {
		select {
		case <-h.Done():
			return
		case <-h.nodeCountUpdated:
		}

		count := h.nodeCount.Load()
		if count == h.lastNodeCount {
			continue
		}

		h.repo.updateHubWatchers(h, int(count))
		if err := h.conn.SendSystemMsg(binary.NewMsgNodeCount(count)); err != nil {
			h.logger.Infof("send nodecount: %v", err)

			// retry after interval
			select {
			case h.nodeCountUpdated <- struct{}{}:
			default:
			}
		} else {
			h.lastNodeCount = count
		}

		select {
		case <-h.Done():
			return
		case <-time.After(interval):
		}
	}
}

// ProcessLoop goroutine dispatch messages and events.
func (h *Hub) ProcessLoop() {
Loop:
	for {
		select {
		case msg := <-h.msgCh:
			h.dispatchMsg(msg)
		case ev, ok := <-h.conn.Events():
			if !ok {
				h.logger.Debugf("connection events closed")
				break Loop
			}
			if err := h.room.Update(ev); err != nil {
				h.logger.Errorf("room update: %+v", err)
			}
			if binary.IsRegularEvent(ev) {
				h.logger.Debugf("broadcast: %v", ev.Type())
				h.broadcast(ev.(*binary.RegularEvent))
			}
		}
	}
	h.drainMsg()
	h.logger.Debug("Hub.ProcessLoop() finish")
}

// drainMsg drain msgCh until all clients closed.
// clientのgoroutineがmsgChに書き込むところで停止するのを防ぐ
func (h *Hub) drainMsg() {
	ch := make(chan struct{})
	go func() {
		h.wgClient.Wait()
		ch <- struct{}{}
	}()

	for {
		select {
		case msg := <-h.msgCh:
			h.logger.Debugf("Discard msg: %T %v", msg, msg)
		case <-ch:
			return
		}
	}
}

func (h *Hub) dispatchMsg(msg game.Msg) {
	switch m := msg.(type) {
	case *game.MsgWatch:
		h.msgWatch(m)
	case *game.MsgLeave:
		h.msgLeave(m)
	case *game.MsgPing:
		h.msgPing(m)
	case *game.MsgClientError:
		h.msgClientError(m)
	case *game.MsgClientTimeout:
		h.msgClientTimeout(m)

	// clientから来たメッセージをgameに伝える.
	case *game.MsgTargets:
		m.Sender.Logger().Debugf("message to targets: %v, %v", m.Targets, m.Data)
		h.proxyMessage(m.RegularMsg)
	case *game.MsgToMaster:
		m.Sender.Logger().Debugf("message to master: %v", m.Data)
		h.proxyMessage(m.RegularMsg)
	case *game.MsgBroadcast:
		m.Sender.Logger().Debugf("message to all: %v", m.Data)
		h.proxyMessage(m.RegularMsg)

	default:
		h.logger.Errorf("unknown msg type: %T %v", m, m)
	}
}

// broadcast : 全員に送信.
func (h *Hub) broadcast(ev *binary.RegularEvent) {
	errs := map[game.ClientID]string{}
	for _, c := range h.watchers {
		err := c.Send(ev)
		if err != nil {
			errs[c.ID()] = err.Error()
		}
	}
	for id, msg := range errs {
		h.removeWatcher(id, msg)
	}
}

func (h *Hub) msgWatch(msg *game.MsgWatch) {
	if !h.room.Watchable {
		err := xerrors.Errorf("Room is not watchable. room=%v, client=%v", h.ID(), msg.Info.Id)
		h.logger.Info(err.Error())
		msg.Err <- game.WithCode(err, codes.FailedPrecondition)
		return
	}

	// Playerとして参加中の観戦は不許可
	if _, ok := h.room.Players[string(msg.SenderID())]; ok {
		err := xerrors.Errorf("Watcher already exists as a player. room=%v, client=%v", h.ID(), msg.SenderID())
		h.logger.Warn(err.Error())
		msg.Err <- game.WithCode(err, codes.AlreadyExists)
		return
	}

	client, err := game.NewWatcher(msg.Info, msg.MACKey, h)
	if err != nil {
		err = game.WithCode(
			xerrors.Errorf("NewWatcher error. room=%v, client=%v: %w", h.ID(), msg.Info.Id, err),
			err.Code())
		h.logger.Warn(err.Error())
		msg.Err <- err
		return
	}
	oldc, rejoin := h.watchers[client.ID()]
	h.watchers[client.ID()] = client
	if rejoin {
		oldc.Removed("client rejoined as a new client")
		client.Logger().Infof("rejoin watcher: %v", client.Id)
	} else {
		client.Logger().Infof("new watcher: %v", client.Id)
	}
	h.storeNodeCount()

	rinfo := &pb.RoomInfo{
		Id:           h.room.Id,
		AppId:        h.appId,
		HostId:       h.repo.hostId,
		Visible:      h.room.Visible,
		Joinable:     h.room.Joinable,
		Watchable:    h.room.Watchable,
		Number:       &pb.RoomNumber{Number: *h.room.Number},
		SearchGroup:  h.room.SearchGroup,
		MaxPlayers:   h.room.MaxPlayers,
		Players:      uint32(len(h.room.Players)),
		Watchers:     h.room.Watchers,
		PublicProps:  binary.MarshalDict(h.room.PublicProps),
		PrivateProps: binary.MarshalDict(h.room.PrivateProps),
	}
	rinfo.SetCreated(h.room.Created)

	players := make([]*pb.ClientInfo, 0, len(h.room.Players))
	for _, p := range h.room.Players {
		players = append(players, &pb.ClientInfo{
			Id:    p.Id,
			Props: binary.MarshalDict(p.Props),
		})
	}

	msg.Joined <- &game.JoinedInfo{
		Room:     rinfo,
		Players:  players,
		Client:   client,
		MasterId: game.ClientID(h.room.Master.Id),
		Deadline: h.Deadline(),
	}
}

func (h *Hub) msgLeave(msg *game.MsgLeave) {
	h.removeWatcher(msg.Sender.ID(), msg.Message)
}

func (h *Hub) msgPing(msg *game.MsgPing) {
	if h.watchers[msg.SenderID()] != msg.Sender {
		return
	}
	msg.Sender.Logger().Debugf("ping %v: %v", msg.Sender.Id, msg.Timestamp)
	ev := binary.NewEvPong(msg.Timestamp, h.room.Watchers, h.room.LastMsgTimes)
	msg.Sender.SendSystemEvent(ev)
}

func (h *Hub) msgClientError(msg *game.MsgClientError) {
	h.removeWatcher(msg.Sender.ID(), msg.ErrMsg)
}

func (h *Hub) msgClientTimeout(msg *game.MsgClientTimeout) {
	h.removeWatcher(msg.Sender.ID(), "timeout")
}

// clientから受け取った RegularMsg を gameサーバーに転送する
func (h *Hub) proxyMessage(msg binary.RegularMsg) {
	err := h.conn.Send(msg.Type(), msg.Payload())
	if err != nil {
		h.logger.Errorf("send message: %+v", err)
	}
}
