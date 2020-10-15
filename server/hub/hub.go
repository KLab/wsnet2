package hub

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"wsnet2/auth"
	"wsnet2/binary"
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

	wgConnect sync.WaitGroup

	msgCh    chan game.Msg
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
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
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

	pubProps, iProps, err := game.InitProps(res.RoomInfo.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("PublicProps unmarshal error: %w", err)
	}
	res.RoomInfo.PublicProps = iProps
	privProps, iProps, err := game.InitProps(res.RoomInfo.PrivateProps)
	if err != nil {
		return nil, xerrors.Errorf("PrivateProps unmarshal error: %w", err)
	}
	res.RoomInfo.PrivateProps = iProps

	h.RoomInfo = res.RoomInfo
	h.publicProps = pubProps
	h.privateProps = privProps

	h.players = make(map[ClientID]*Player)
	for _, c := range res.Players {
		props, iProps, err := game.InitProps(c.Props)
		if err != nil {
			return nil, xerrors.Errorf("PublicProps unmarshal error: %w", err)
		}
		c.Props = iProps
		h.players[ClientID(c.Id)] = &Player{
			ClientInfo: c,
			props:      props,
		}
	}

	h.wgConnect.Done()

	h.logger.Debugf("URL: %v\n", res.Url)
	h.logger.Debugf("AuthKey: %v\n", res.AuthKey)
	ws, err := h.dialGame(res.Url, res.AuthKey, 0)
	if err != nil {
		return nil, xerrors.Errorf("connectGame: Failed to dial game server: %w", err)
	}

	return ws, nil
}

func (h *Hub) Start() {
	h.logger.Debug("hub start")
	defer h.logger.Debug("hub end")

	ws, err := h.connectGame()
	if err != nil {
		h.logger.Errorf("Failed to connect game server: %v\n", err)
		return
	}

	h.logger.Infof("Established")

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

func (h *Hub) Watch(info *pb.ClientInfo) <-chan game.JoinedInfo {
	ch := make(chan game.JoinedInfo)
	go func() {
		h.wgConnect.Wait()
		client, err := game.NewWatcher(info, h)
		if err != nil {
			return
		}
		h.muClients.Lock()
		h.watchers[client.ID()] = client
		rinfo := h.RoomInfo.Clone()
		players := make([]*pb.ClientInfo, 0, len(h.players))
		for _, c := range h.players {
			players = append(players, c.ClientInfo.Clone())
		}
		h.muClients.Unlock()
		ch <- game.JoinedInfo{rinfo, players, client, h.master, h.deadline}
	}()
	return ch
}
