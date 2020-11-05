package hub

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"wsnet2/auth"
	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/pb"
)

type Player struct {
	*pb.ClientInfo
	props binary.Dict
}

type Hub struct {
	*pb.RoomInfo
	id       RoomID
	repo     *Repository
	appId    AppID
	clientId string

	deadline    time.Duration
	newDeadline chan time.Duration

	publicProps  binary.Dict
	privateProps binary.Dict

	msgCh    chan game.Msg
	evCh     chan binary.Event // TODO: drainとかちゃんと考える
	ready    chan struct{}
	done     chan struct{}
	wgClient sync.WaitGroup

	muClients sync.RWMutex
	players   map[ClientID]*Player
	watchers  map[ClientID]*game.Client
	master    ClientID

	lastMsg binary.Dict // map[clientID]unixtime_millisec

	seq int

	logger *zap.SugaredLogger
}

func NewHub(repo *Repository, appId AppID, roomId RoomID) *Hub {
	// hub->game 接続に使うclientId. このhubを作成するトリガーになったclientIdは使わない
	// roomIdもhostIdもユニークなので hostId:roomId はユニークになるはず。
	clientId := fmt.Sprintf("hub:%d:%s", repo.hostId, roomId)

	// todo: log.CurrentLevel()
	logger := log.Get(log.DEBUG).With(
		zap.String("type", "hub"),
		zap.String("room", string(roomId)),
		zap.String("clientId", clientId),
	).Sugar()

	hub := &Hub{
		id:       RoomID(roomId),
		repo:     repo,
		appId:    appId,
		clientId: clientId,

		newDeadline: make(chan time.Duration, 1),

		publicProps:  make(binary.Dict),
		privateProps: make(binary.Dict),

		msgCh: make(chan game.Msg, game.RoomMsgChSize),
		evCh:  make(chan binary.Event, 1), // FIXME: 値をちゃんと考える
		ready: make(chan struct{}),
		done:  make(chan struct{}),

		players:  make(map[ClientID]*Player),
		watchers: make(map[ClientID]*game.Client),

		lastMsg: make(binary.Dict),

		logger: logger,
		// todo: hubをもっと埋める
	}

	return hub
}

func (h *Hub) ID() RoomID {
	return h.id
}

func (h *Hub) Repo() game.IRepo {
	return h.repo
}

func (h *Hub) Deadline() time.Duration {
	return h.deadline
}

func (h *Hub) WaitGroup() *sync.WaitGroup {
	return &h.wgClient
}

func (h *Hub) Logger() *zap.SugaredLogger {
	return h.logger
}

func (h *Hub) Done() <-chan struct{} {
	return h.done
}

func (h *Hub) Timeout(c *game.Client) {
	// TODO: 実装
}

func (h *Hub) SendMessage(msg game.Msg) {
	// TODO: 実装
}

func (h *Hub) writeLastMsg(cid ClientID) {
	millisec := uint64(time.Now().UnixNano()) / 1000000
	h.lastMsg[string(cid)] = binary.MarshalULong(millisec)
}

/// UpdateLastMsg : PlayerがMsgを受信したとき更新する.
/// 既に登録されているPlayerのみ書き込み (watcherを含めないため)
func (h *Hub) updateLastMsg(cid ClientID) {
	id := string(cid)
	if _, ok := h.lastMsg[id]; ok {
		h.writeLastMsg(cid)
	}
}

func (h *Hub) removeClient(c *game.Client, err error) {
	h.removeWatcher(c, err)
}

func (h *Hub) removeWatcher(c *game.Client, err error) {
	h.muClients.Lock()
	defer h.muClients.Unlock()

	cid := c.ID()

	if _, ok := h.watchers[cid]; !ok {
		h.logger.Debugf("Watcher may be aleady left: client=%v", cid)
		return
	}

	h.logger.Infof("Watcher removed: client=%v %v", cid, err)
	delete(h.watchers, cid)

	h.RoomInfo.Watchers -= c.NodeCount()

	c.Removed(err)
}

func (h *Hub) dialGame(url, authKey string, seq int) (*websocket.Conn, error) {
	hdr := http.Header{}
	hdr.Add("Wsnet2-App", h.appId)
	hdr.Add("Wsnet2-User", h.clientId)
	hdr.Add("Wsnet2-LastEventSeq", strconv.Itoa(seq))

	authdata, err := auth.GenerateAuthData(authKey, h.clientId, time.Now())
	if err != nil {
		return nil, xerrors.Errorf("dialGame: generate authdata error: %w\n", err)
	}
	hdr.Add("Authorization", "Bearer "+authdata)

	ws, res, err := h.repo.ws.Dial(url, hdr)
	if err != nil {
		return nil, xerrors.Errorf("dialGame: dial error: %v, %w\n", res, err)
	}

	h.logger.Infof("dialGame: response: %v\n", res)

	return ws, nil
}

func (h *Hub) requestWatch(addr string) (*pb.JoinedRoomRes, error) {
	conn, err := h.repo.grpcPool.Get(addr)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to dial to game server: %w", err)
	}

	client := pb.NewGameClient(conn)
	req := &pb.JoinRoomReq{
		AppId:  h.appId,
		RoomId: string(h.id),
		ClientInfo: &pb.ClientInfo{
			Id: h.clientId,
		},
	}

	res, err := client.Watch(context.TODO(), req)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to 'Watch' request to game server: %w", err)
	}

	h.logger.Info("Joined room: %v", res)

	return res, nil
}

func calcPingInterval(deadline time.Duration) time.Duration {
	return deadline / 3
}

func (h *Hub) pinger(conn *websocket.Conn) {
	deadline := h.deadline
	t := time.NewTicker(calcPingInterval(deadline))
	defer t.Stop()
	for {
		select {
		case <-t.C:
			msg := binary.NewMsgPing(time.Now())
			if err := conn.WriteMessage(websocket.BinaryMessage, msg.Marshal()); err != nil {
				return
			}
		case newDeadline := <-h.newDeadline:
			h.logger.Debugf("pinger: update deadline: %v to %v\n", deadline, newDeadline)
			t.Reset(calcPingInterval(newDeadline))
		case <-h.Done():
			return
		}
	}
}

func (h *Hub) copyInitialValues(res *pb.JoinedRoomRes) error {
	pubProps, iProps, err := common.InitProps(res.RoomInfo.PublicProps)
	if err != nil {
		return xerrors.Errorf("PublicProps unmarshal error: %w", err)
	}
	res.RoomInfo.PublicProps = iProps
	privProps, iProps, err := common.InitProps(res.RoomInfo.PrivateProps)
	if err != nil {
		return xerrors.Errorf("PrivateProps unmarshal error: %w", err)
	}
	res.RoomInfo.PrivateProps = iProps

	h.RoomInfo = res.RoomInfo
	h.deadline = time.Duration(res.Deadline) * time.Second
	h.publicProps = pubProps
	h.privateProps = privProps

	h.players = make(map[ClientID]*Player)
	for _, c := range res.Players {
		props, iProps, err := common.InitProps(c.Props)
		if err != nil {
			return xerrors.Errorf("PublicProps unmarshal error: %w", err)
		}
		c.Props = iProps
		h.players[ClientID(c.Id)] = &Player{
			ClientInfo: c,
			props:      props,
		}
	}
	return nil
}

func (h *Hub) getGameServer() (*common.GameServer, error) {
	var room pb.RoomInfo
	err := h.repo.db.Get(&room, "SELECT * FROM room WHERE id = ?", h.id)
	if err != nil {
		return nil, xerrors.Errorf("Failed to get room: %w", err)
	}
	return h.repo.gameCache.Get(room.HostId)
}

func (h *Hub) Start() {
	h.logger.Debug("hub start")
	defer h.logger.Debug("hub end")
	defer close(h.done)

	go h.ProcessLoop()

	gs, err := h.getGameServer()
	if err != nil {
		h.logger.Errorf("Failed to get game server: %v\n", err)
		return
	}

	res, err := h.requestWatch(fmt.Sprintf("%s:%d", gs.Hostname, gs.GRPCPort))
	if err != nil {
		h.logger.Errorf("Failed to Watch request: %v\n", err)
		return
	}

	if err := h.copyInitialValues(res); err != nil {
		h.logger.Errorf("Failed to copy initial values: %v\n", err)
		return
	}

	// msgChの受け入れ準備完了
	h.ready <- struct{}{}

	// Hub -> Game は Hostname で接続する
	url := strings.Replace(res.Url, gs.PublicName, gs.Hostname, 1)
	h.logger.Debugf("Dial Game: %v\n", url)

	ws, err := h.dialGame(url, res.AuthKey, 0)
	if err != nil {
		h.logger.Errorf("Failed to dial game server: %v\n", err)
		return
	}

	go h.pinger(ws)

	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			h.logger.Errorf("ReadMessage error: %v\n", err)
			return
		}
		ev, _, err := binary.UnmarshalEvent(b)
		if err != nil {
			h.logger.Errorf("UnmarshalEvent error: %v\n", err)
			return
		}
		h.evCh <- ev
	}
}

// ProcessLoop goroutine dispatch messages and events.
func (h *Hub) ProcessLoop() {
	h.logger.Debug("Hub.ProcessLoop() start")
	defer func() {
		h.repo.RemoveHub(h)
		h.drainMsg()
		h.logger.Debug("Hub.ProcessLoop() finish")
	}()

	select {
	case <-h.ready:
		h.logger.Info("Hub ready")
	case <-h.Done():
		h.logger.Info("Hub closed before ready")
		return
	}

Loop:
	for {
		select {
		case <-h.Done():
			h.logger.Info("Hub closed")
			break Loop
		case msg := <-h.msgCh:
			h.logger.Debugf("Hub msg: %T %v", msg, msg)
			h.updateLastMsg(msg.SenderID())
			if err := h.dispatch(msg); err != nil {
				h.logger.Errorf("Hub msg error: %v", err)
			}
		case ev := <-h.evCh:
			h.logger.Debugf("Hub event: %T %v", ev, ev)
			if err := h.dispatchEvent(ev); err != nil {
				h.logger.Errorf("Hub event error: %v", err)
			}
		}
	}
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

func (h *Hub) dispatch(msg game.Msg) error {
	switch m := msg.(type) {
	case *game.MsgWatch:
		return h.msgWatch(m)
	case *game.MsgClientError:
		return h.msgClientError(m)
	default:
		return xerrors.Errorf("unknown msg type: %T %v", m, m)
	}
}

// sendTo : 特定クライアントに送信.
// muClients のロックを取得してから呼び出す.
// 送信できない場合続行不能なので退室させる.
func (h *Hub) sendTo(c *game.Client, ev *binary.RegularEvent) error {
	err := c.Send(ev)
	if err != nil {
		// removeClient locks muClients so that must be called another goroutine.
		go h.removeClient(c, err)
	}
	return err
}

// broadcast : 全員に送信.
// muClients のロックを取得してから呼び出すこと
func (h *Hub) broadcast(ev *binary.RegularEvent) {
	for _, c := range h.watchers {
		_ = h.sendTo(c, ev)
	}
}

func (h *Hub) msgWatch(msg *game.MsgWatch) error {
	if !h.Watchable {
		close(msg.Joined)
		return xerrors.Errorf("Room is not watchable. room=%v, client=%v", h.ID(), msg.Info.Id)
	}

	h.muClients.Lock()
	defer h.muClients.Unlock()

	if _, ok := h.players[msg.SenderID()]; ok {
		close(msg.Joined)
		return xerrors.Errorf("Player already exists. room=%v, client=%v", h.ID(), msg.SenderID())
	}
	// 他のhub経由で観戦している場合は考慮しない（Gameでの直接観戦も同様）
	if _, ok := h.watchers[msg.SenderID()]; ok {
		close(msg.Joined)
		return xerrors.Errorf("Player already exists as a watcher. room=%v, client=%v", h.ID(), msg.SenderID())
	}

	client, err := game.NewWatcher(msg.Info, h)
	if err != nil {
		close(msg.Joined)
		return xerrors.Errorf("NewWatcher error. room=%v, client=%v, err=%w", h.ID(), msg.Info.Id, err)
	}
	h.watchers[client.ID()] = client
	h.RoomInfo.Watchers += client.NodeCount()

	rinfo := h.RoomInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(h.players))
	for _, c := range h.players {
		players = append(players, c.ClientInfo.Clone())
	}

	msg.Joined <- game.JoinedInfo{
		Room:     rinfo,
		Players:  players,
		Client:   client,
		MasterId: h.master,
		Deadline: h.deadline,
	}
	return nil
}

func (h *Hub) msgClientError(msg *game.MsgClientError) error {
	h.removeClient(msg.Sender, msg.Err)
	return nil
}

func (h *Hub) dispatchEvent(ev binary.Event) error {
	switch ev.Type() {
	case binary.EvTypePong:
		return h.evPong(ev)
	case binary.EvTypePeerReady:
		return h.evPeerReady(ev)
	case binary.EvTypeJoined:
		return h.evJoined(ev)
	case binary.EvTypeLeft:
		return h.evLeft(ev)
	case binary.EvTypeRoomProp:
		return h.evRoomProp(ev)
	case binary.EvTypeMasterSwitched:
		return h.evMasterSwitched(ev)
	case binary.EvTypeMessage:
		return h.evMessage(ev)
	case binary.EvTypeSucceeded:
		return h.evSucceeded(ev)
	case binary.EvTypePermissionDenied:
		return h.evPermissionDenied(ev)
	case binary.EvTypeTargetNotFound:
		return h.evTargetNotFound(ev)
	default:
		return xerrors.Errorf("unknown event type: %T %v", ev, ev)
	}
}

func (h *Hub) evPong(ev binary.Event) error {
	pong, err := binary.UnmarshalEvPongPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Unmarshal EvPong payload error: %w", err)
	}
	h.RoomInfo.Watchers = pong.Watchers
	h.lastMsg = pong.LastMsg
	return nil
}

func (h *Hub) evPeerReady(ev binary.Event) error {
	seq, err := binary.UnmarshalEvPeerReadyPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Unmarshal EvPong payload error: %w", err)
	}
	h.seq = seq
	return nil
}

func (h *Hub) evJoined(ev binary.Event) error {
	ci, err := binary.UnmarshalEvJoinedPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Unmarshal EvJoined payload error: %w", err)
	}
	props, iProps, err := common.InitProps(ci.Props)
	if err != nil {
		return xerrors.Errorf("PublicProps unmarshal error: %w", err)
	}
	ci.Props = iProps

	h.muClients.Lock()
	defer h.muClients.Unlock()

	h.players[ClientID(ci.Id)] = &Player{
		ClientInfo: ci,
		props:      props,
	}

	h.broadcast(ev.(*binary.RegularEvent))
	return nil
}

func (h *Hub) evLeft(ev binary.Event) error {
	left, err := binary.UnmarshalEvLeftPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Unmarshal EvLeft payload error: %w", err)
	}

	h.muClients.Lock()
	defer h.muClients.Unlock()

	delete(h.players, game.ClientID(left.ClientId))
	h.master = game.ClientID(left.MasterId)

	h.broadcast(ev.(*binary.RegularEvent))
	return nil
}

func (h *Hub) evRoomProp(ev binary.Event) error {
	rpp, err := binary.UnmarshalEvRoomPropPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Unmarshal EvRoomProp payload error: %w", err)
	}

	h.RoomInfo.Visible = rpp.Visible
	h.RoomInfo.Joinable = rpp.Joinable
	h.RoomInfo.Watchable = rpp.Watchable
	h.RoomInfo.SearchGroup = rpp.SearchGroup
	h.RoomInfo.MaxPlayers = rpp.MaxPlayer

	if len(rpp.PublicProps) > 0 {
		for k, v := range rpp.PublicProps {
			if _, ok := h.publicProps[k]; ok && len(v) == 0 {
				delete(h.publicProps, k)
			} else {
				h.publicProps[k] = v
			}
		}
		h.RoomInfo.PublicProps = binary.MarshalDict(h.publicProps)
		h.logger.Debugf("Hub update PublicProps: %v", h.publicProps)
	}

	if len(rpp.PrivateProps) > 0 {
		for k, v := range rpp.PrivateProps {
			if _, ok := h.privateProps[k]; ok && len(v) == 0 {
				delete(h.privateProps, k)
			} else {
				h.privateProps[k] = v
			}
		}
		h.RoomInfo.PrivateProps = binary.MarshalDict(h.privateProps)
		h.logger.Debugf("Hub update PrivateProps: %v", h.privateProps)
	}

	if rpp.ClientDeadline != 0 {
		deadline := time.Duration(rpp.ClientDeadline) * time.Second
		if deadline != h.deadline {
			h.logger.Debugf("Hub notify new deadline: %v", deadline)
			h.deadline = deadline
			h.newDeadline <- deadline
		}
	}

	h.muClients.Lock()
	defer h.muClients.Unlock()

	h.broadcast(ev.(*binary.RegularEvent))
	return nil
}

func (h *Hub) evClientProp(ev binary.Event) error {
	cpp, err := binary.UnmarshalEvClientPropPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Unmarshal EvClientProp payload error: %w", err)
	}

	h.muClients.Lock()
	defer h.muClients.Unlock()

	h.logger.Debugf("EvClientProp: client=%v, props=%v", cpp.Id, cpp.Props)
	if len(cpp.Props) > 0 {
		c, found := h.players[ClientID(cpp.Id)]
		if !found {
			return xerrors.Errorf("player not found: client=%v", cpp.Id)
		}
		for k, v := range cpp.Props {
			if _, ok := c.props[k]; ok && len(v) == 0 {
				delete(c.props, k)
			} else {
				c.props[k] = v
			}
		}
		c.ClientInfo.Props = binary.MarshalDict(c.props)
		h.logger.Debugf("Client update Props: client=%v %v", c.Id, c.props)
	}

	h.broadcast(ev.(*binary.RegularEvent))
	return nil
}

func (h *Hub) evMasterSwitched(ev binary.Event) error {
	master, err := binary.UnmarshalEvMasterSwitchedPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Unmarshal EvMasterSwitched payload error: %w", err)
	}

	h.muClients.Lock()
	h.master = ClientID(master)
	h.broadcast(ev.(*binary.RegularEvent))
	h.muClients.Unlock()
	return nil
}

func (h *Hub) evMessage(ev binary.Event) error {
	h.muClients.Lock()
	h.broadcast(ev.(*binary.RegularEvent))
	h.muClients.Unlock()
	return nil
}

func (h *Hub) evSucceeded(ev binary.Event) error {
	// TODO: 実装
	return nil
}

func (h *Hub) evPermissionDenied(ev binary.Event) error {
	// TODO: 実装
	return nil
}

func (h *Hub) evTargetNotFound(ev binary.Event) error {
	// TODO: 実装
	return nil
}
