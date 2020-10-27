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

	deadline time.Duration

	publicProps  binary.Dict
	privateProps binary.Dict

	msgCh    chan game.Msg
	ready    chan struct{}
	done     chan struct{}
	wgClient sync.WaitGroup

	muClients sync.RWMutex
	players   map[ClientID]*Player
	watchers  map[ClientID]*game.Client
	master    ClientID

	lastMsg binary.Dict // map[clientID]unixtime_millisec

	logger *zap.SugaredLogger
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

func (h *Hub) connectGame() (*websocket.Conn, error) {
	var room pb.RoomInfo
	err := h.repo.db.Get(&room, "SELECT * FROM room WHERE id = ?", h.id)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to get room: %w", err)
	}

	gs, err := h.repo.gameCache.Get(room.HostId)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to get game server: %w", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", gs.Hostname, gs.GRPCPort)
	conn, err := h.repo.grpcPool.Get(grpcAddr)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to dial to game server: %w", err)
	}
	defer conn.Close()

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

	pubProps, iProps, err := common.InitProps(res.RoomInfo.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("PublicProps unmarshal error: %w", err)
	}
	res.RoomInfo.PublicProps = iProps
	privProps, iProps, err := common.InitProps(res.RoomInfo.PrivateProps)
	if err != nil {
		return nil, xerrors.Errorf("PrivateProps unmarshal error: %w", err)
	}
	res.RoomInfo.PrivateProps = iProps

	h.RoomInfo = res.RoomInfo
	h.publicProps = pubProps
	h.privateProps = privProps

	h.players = make(map[ClientID]*Player)
	for _, c := range res.Players {
		props, iProps, err := common.InitProps(c.Props)
		if err != nil {
			return nil, xerrors.Errorf("PublicProps unmarshal error: %w", err)
		}
		c.Props = iProps
		h.players[ClientID(c.Id)] = &Player{
			ClientInfo: c,
			props:      props,
		}
	}

	close(h.ready)

	// Hub -> Game は Hostname で接続する
	url := strings.Replace(res.Url, gs.PublicName, gs.Hostname, 1)
	h.logger.Debugf("Dial Game: %v\n", url)
	ws, err := h.dialGame(url, res.AuthKey, 0)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to dial game server: %w", err)
	}

	return ws, nil
}

func (h *Hub) pinger(conn *websocket.Conn) {
	// FIXME: 送信間隔はdeadlineから算出する
	//        deadlineの更新にも対応できるように
	t := time.NewTicker(time.Second * 2)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			msg := binary.NewMsgPing(time.Now())
			if err := conn.WriteMessage(websocket.BinaryMessage, msg.Marshal()); err != nil {
				return
			}
		case <-h.Done():
			return
		}
	}
}

func (h *Hub) Start() {
	h.logger.Debug("hub start")
	defer h.logger.Debug("hub end")
	defer close(h.done)

	go func() {
		select {
		case <-h.ready:
			h.MsgLoop()
		case <-h.Done():
			h.repo.RemoveHub(h)
			h.drainMsg()
		}
	}()

	ws, err := h.connectGame()
	if err != nil {
		h.logger.Errorf("Failed to connect game server: %v\n", err)
		return
	}

	go h.pinger(ws)

	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			h.logger.Errorf("ReadMessage error: %v\n", err)
			return
		}

		switch ty := binary.EvType(b[0]); ty {
		default:
			h.logger.Debugf("ReadMessage: %v, %v\n", ty, b)
		}
	}
}

// MsgLoop goroutine dispatch messages.
func (h *Hub) MsgLoop() {
	h.logger.Debug("Hub.MsgLoop() start.")
Loop:
	for {
		select {
		case <-h.Done():
			h.logger.Info("Hub closed.")
			break Loop
		case msg := <-h.msgCh:
			h.logger.Debugf("Hub msg: %T %v", msg, msg)
			h.updateLastMsg(msg.SenderID())
			if err := h.dispatch(msg); err != nil {
				h.logger.Errorf("Hub msg error: %v", err)
			}
		}
	}
	h.repo.RemoveHub(h)

	h.drainMsg()
	h.logger.Debug("Hub.MsgLoop() finish")
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
