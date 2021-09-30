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
	"google.golang.org/grpc/codes"

	"wsnet2/auth"
	"wsnet2/binary"
	"wsnet2/common"
	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/metrics"
	"wsnet2/pb"
)

type Player struct {
	*pb.ClientInfo
	props binary.Dict
}

type Hub struct {
	*pb.RoomInfo
	hubPK    int64 // dbのautoincrementで採番されるsurrogate key. DB更新にだけ使う。
	roomId   RoomID
	repo     *Repository
	appId    AppID
	clientId string

	deadline    time.Duration
	newDeadline chan time.Duration

	publicProps  binary.Dict
	privateProps binary.Dict

	msgCh          chan game.Msg
	evCh           chan binary.Event // TODO: drainとかちゃんと考える
	ready          chan struct{}
	done           chan struct{}
	normallyClosed bool
	wgClient       sync.WaitGroup

	muClients sync.RWMutex
	players   map[ClientID]*Player
	watchers  map[ClientID]*game.Client
	master    ClientID

	lastMsg binary.Dict // map[clientID]unixtime_millisec

	// game に通知した直近の nodeCount
	lastNodeCount int

	// hub -> game の seq.
	seq int

	// hub -> game に使う conn
	gameConn *websocket.Conn
	muWrite  sync.Mutex

	logger *zap.SugaredLogger
}

func NewHub(repo *Repository, appId AppID, roomId RoomID) *Hub {
	// hub->game 接続に使うclientId. このhubを作成するトリガーになったclientIdは使わない
	// roomIdもhostIdもユニークなので hostId:roomId はユニークになるはず。
	clientId := fmt.Sprintf("hub:%d:%s", repo.hostId, roomId)

	logger := log.Get(log.CurrentLevel()).With(
		zap.String("type", "hub"),
		zap.String("room", string(roomId)),
		zap.String("clientId", clientId),
	).Sugar()

	hub := &Hub{
		roomId:   RoomID(roomId),
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
	return h.roomId
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
	h.msgCh <- msg
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

	// pong.Watchersで上書きされるのでなくてもいいかも？
	h.RoomInfo.Watchers -= c.NodeCount()

	c.Removed(err)
}

func (h *Hub) dialGame(url, authKey string, seq int) error {
	hdr := http.Header{}
	hdr.Add("Wsnet2-App", h.appId)
	hdr.Add("Wsnet2-User", h.clientId)
	hdr.Add("Wsnet2-LastEventSeq", strconv.Itoa(seq))

	authdata, err := auth.GenerateAuthData(authKey, h.clientId, time.Now())
	if err != nil {
		return xerrors.Errorf("dialGame: generate authdata error: %w\n", err)
	}
	hdr.Add("Authorization", "Bearer "+authdata)

	ws, res, err := h.repo.ws.Dial(url, hdr)
	if err != nil {
		return xerrors.Errorf("dialGame: dial error: %v, %w\n", res, err)
	}

	h.logger.Infof("dialGame: response: %v\n", res)
	h.gameConn = ws
	return nil
}

func (h *Hub) requestWatch(addr string) (*pb.JoinedRoomRes, error) {
	conn, err := h.repo.grpcPool.Get(addr)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to dial to game server: %w", err)
	}

	client := pb.NewGameClient(conn)
	req := &pb.JoinRoomReq{
		AppId:  h.appId,
		RoomId: string(h.roomId),
		ClientInfo: &pb.ClientInfo{
			Id:    h.clientId,
			IsHub: true,
		},
	}

	res, err := client.Watch(context.TODO(), req)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to 'Watch' request to game server: %w", err)
	}

	h.logger.Info("Joined room: %v", res)

	return res, nil
}

func (h *Hub) WriteMessage(messageType int, data []byte) error {
	h.muWrite.Lock()
	defer h.muWrite.Unlock()
	metrics.MessageSent.Add(1)
	return h.gameConn.WriteMessage(messageType, data)
}

func calcPingInterval(deadline time.Duration) time.Duration {
	return deadline / 3
}

func (h *Hub) pinger() {
	deadline := h.deadline
	t := time.NewTicker(calcPingInterval(deadline))
	defer t.Stop()
	for {
		select {
		case <-t.C:
			msg := binary.NewMsgPing(time.Now())
			if err := h.WriteMessage(websocket.BinaryMessage, msg.Marshal()); err != nil {
				h.logger.Errorf("pinger: WrteMessage error: %v\n", err)
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

const updateInterval = 1 // TODO: config化

func (h *Hub) nodeCountUpdater() {
	t := time.NewTicker(time.Duration(updateInterval) * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			h.muClients.RLock()
			nodeCount := len(h.watchers)
			h.muClients.RUnlock()
			if nodeCount != h.lastNodeCount {
				msg := binary.NewMsgNodeCount(uint32(nodeCount))
				metrics.MessageSent.Add(1)
				if err := h.WriteMessage(websocket.BinaryMessage, msg.Marshal()); err != nil {
					h.logger.Errorf("nodeCountUpdater: WrteMessage error: %v\n", err)
					return
				}
				h.lastNodeCount = nodeCount
				_, err := h.repo.db.Exec("UPDATE `hub` SET `watchers`=? WHERE id=?", nodeCount, h.hubPK)
				if err != nil {
					h.logger.Errorf("failed to update hub.watchers: %v", err)
				}
			}
		case <-h.Done():
			return
		}
	}
}

// DBのhubテーブルにレコードを追加する
func (h *Hub) registerDB() error {
	res, err := h.repo.db.Exec("INSERT INTO hub (`host_id`, `room_id`, `watchers`, `created`) VALUES (?,?,?,?)",
		h.repo.hostId, string(h.roomId), 0, time.Now().UTC())
	if err != nil {
		h.logger.Errorf(`Failed to "insert into hub": %v`, err)
		return err
	}
	h.hubPK, _ = res.LastInsertId()
	return nil
}

// エラーの処理のしようがないからerrorは返さない
func (h *Hub) unregisterDB() {
	_, err := h.repo.db.Exec("DELETE FROM hub WHERE id=?", h.hubPK)
	if err != nil {
		h.logger.Errorf(`Failed to "delete from hub where id=%v": %v`, h.hubPK, err)
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
	h.master = game.ClientID(res.MasterId)
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
	err := h.repo.db.Get(&room, "SELECT * FROM room WHERE id = ?", h.roomId)
	if err != nil {
		return nil, xerrors.Errorf("Failed to get room: %w", err)
	}
	return h.repo.gameCache.Get(room.HostId)
}

func (h *Hub) Start() {
	h.logger.Debug("hub start")
	metrics.Hubs.Add(1)
	defer func() {
		close(h.done)
		metrics.Hubs.Add(-1)
		h.logger.Debug("hub end")
	}()

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

	err = h.registerDB()
	if err != nil {
		return
	}
	defer h.unregisterDB()

	go h.ProcessLoop()

	// Hub -> Game は Hostname で接続する
	url := strings.Replace(res.Url, gs.PublicName, gs.Hostname, 1)
	h.logger.Debugf("Dial Game: %v\n", url)

	err = h.dialGame(url, res.AuthKey, 0)
	if err != nil {
		h.logger.Errorf("Failed to dial game server: %v\n", err)
		return
	}

	go h.pinger()
	go h.nodeCountUpdater()

	// TODO: どこで h.gameConn を Close するか考える
	for {
		_, b, err := h.gameConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				h.logger.Infof("Game closed: %v", err)
				h.normallyClosed = true
			} else if websocket.IsUnexpectedCloseError(err) {
				h.logger.Errorf("Game close error: %v", err)
			} else {
				h.logger.Errorf("Game read error: %T %v", err, err)
			}
			return
		}
		metrics.MessageRecv.Add(1)
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

func (h *Hub) dispatch(msg game.Msg) error {
	switch m := msg.(type) {
	case *game.MsgWatch:
		return h.msgWatch(m)
	case *game.MsgPing:
		return h.msgPing(m)
	case *game.MsgClientError:
		return h.msgClientError(m)

	// clientから来たメッセージをgameに伝える.
	case *game.MsgTargets:
		return h.proxyMessage(m.RegularMsg)
	case *game.MsgToMaster:
		return h.proxyMessage(m.RegularMsg)
	case *game.MsgBroadcast:
		return h.proxyMessage(m.RegularMsg)

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
		err := xerrors.Errorf("Room is not watchable. room=%v, client=%v", h.ID(), msg.Info.Id)
		msg.Err <- game.WithCode(err, codes.FailedPrecondition)
		return err
	}

	h.muClients.Lock()
	defer h.muClients.Unlock()

	if _, ok := h.players[msg.SenderID()]; ok {
		err := xerrors.Errorf("Watcher already exists as a player. room=%v, client=%v", h.ID(), msg.SenderID())
		msg.Err <- game.WithCode(err, codes.AlreadyExists)
		return err
	}
	// 他のhub経由で観戦している場合は考慮しない（Gameでの直接観戦も同様）
	if _, ok := h.watchers[msg.SenderID()]; ok {
		err := xerrors.Errorf("Watcher already exists. room=%v, client=%v", h.ID(), msg.SenderID())
		msg.Err <- game.WithCode(err, codes.AlreadyExists)
		return err
	}

	client, err := game.NewWatcher(msg.Info, h)
	if err != nil {
		err = game.WithCode(
			xerrors.Errorf("NewWatcher error. room=%v, client=%v: %w", h.ID(), msg.Info.Id, err),
			err.Code())
		msg.Err <- err
		return err
	}
	h.watchers[client.ID()] = client
	// pong.Watchersで上書きされるのでなくてもいいかも？
	h.RoomInfo.Watchers += client.NodeCount()

	rinfo := h.RoomInfo.Clone()
	players := make([]*pb.ClientInfo, 0, len(h.players))
	for _, c := range h.players {
		players = append(players, c.ClientInfo.Clone())
	}

	msg.Joined <- &game.JoinedInfo{
		Room:     rinfo,
		Players:  players,
		Client:   client,
		MasterId: h.master,
		Deadline: h.deadline,
	}
	return nil
}

func (h *Hub) msgPing(msg *game.MsgPing) error {
	ev := binary.NewEvPong(msg.Timestamp, h.RoomInfo.Watchers, h.lastMsg)
	return msg.Sender.SendSystemEvent(ev)
}

func (h *Hub) msgClientError(msg *game.MsgClientError) error {
	h.removeClient(msg.Sender, msg.Err)
	return nil
}

// clientから受け取った RegularMsg を gameサーバーに転送する
func (h *Hub) proxyMessage(msg binary.RegularMsg) error {
	// client->hubとhub->gameでseq が異なるからmsgの使いまわしができない。
	// アロケーションもったいないけど頻度は多くないだろうから気にしない。
	h.seq++ // TODO: EvTypePeerReady がくる前に proxyMessageが呼ばれないか確認する。
	packet := binary.BuildRegularMsgFrame(msg.Type(), int(h.seq), msg.Payload())
	metrics.MessageSent.Add(1)
	return h.WriteMessage(websocket.BinaryMessage, packet)
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
	case binary.EvTypeClientProp:
		return h.evClientProp(ev)
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
		return xerrors.Errorf("Unmarshal EvPeerReady payload error: %w", err)
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
	return nil
}

func (h *Hub) evPermissionDenied(ev binary.Event) error {
	h.logger.Errorf("evPermissionDenied: payload=% x", ev.Payload())
	return nil
}

func (h *Hub) evTargetNotFound(ev binary.Event) error {
	h.logger.Errorf("evTargetNotFound: payload=% x", ev.Payload())
	return nil
}
