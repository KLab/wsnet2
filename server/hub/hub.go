package hub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"wsnet2/binary"
	"wsnet2/game"
	"wsnet2/pb"
)

type Hub struct {
	ID       RoomID
	repo     *Repository
	appId    AppID
	clientId string

	deadline time.Duration

	publicProps  binary.Dict
	privateProps binary.Dict

	msgCh    chan game.Msg
	done     chan struct{}
	wgClient sync.WaitGroup

	muClients sync.RWMutex
	watchers  map[ClientID]*game.Client

	lastMsg binary.Dict // map[clientID]unixtime_millisec

	logger *zap.SugaredLogger
}

func (h *Hub) connectGame() error {
	var room pb.RoomInfo
	err := h.repo.db.Get(&room, "SELECT * FROM room WHERE id = ?", h.ID)
	if err != nil {
		return xerrors.Errorf("connectGame: Failed to get room: %w", err)
	}

	gs, err := h.repo.gameCache.Get(room.HostId)
	if err != nil {
		return xerrors.Errorf("connectGame: Failed to get game server: %w", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", gs.Hostname, gs.GRPCPort)
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		return xerrors.Errorf("connectGame: Failed to dial to game server: %w", err)
	}
	defer conn.Close()

	client := pb.NewGameClient(conn)
	req := &pb.JoinRoomReq{
		AppId:  h.appId,
		RoomId: string(h.ID),
		ClientInfo: &pb.ClientInfo{
			Id: h.clientId,
		},
	}

	res, err := client.Watch(context.TODO(), req)
	if err != nil {
		return xerrors.Errorf("connectGame: Failed to 'Watch' request to game server: %w", err)
	}

	h.logger.Info("Joined room: %v", res)

	return nil
}

func (h *Hub) Start() {
	h.logger.Debug("hub start")
	defer h.logger.Debug("hub end")

	if err := h.connectGame(); err != nil {
		h.logger.Error("Failed to connect game server")
	}

	//TODO: 実装
	time.Sleep(time.Minute)
}
