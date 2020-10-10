package hub

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"wsnet2/config"
	"wsnet2/game"
	"wsnet2/lobby"
	"wsnet2/log"
	"wsnet2/pb"
)

type AppID = pb.AppId
type RoomID = game.RoomID
type ClientID = game.ClientID

type Repository struct {
	hostId uint32

	app  pb.App /* 1つのrepositoryで異なるappのroomを扱うなら不要？ */
	conf *config.GameConf
	db   *sqlx.DB

	gameCache *lobby.GameCache
	ws        *websocket.Dialer

	mu      sync.RWMutex
	hubs    map[RoomID]*Hub
	clients map[ClientID]map[RoomID]*game.Client
}

func NewRepository(db *sqlx.DB, conf *config.GameConf, hostId uint32) (*Repository, error) {
	repo := &Repository{
		hostId: hostId,
		conf:   conf,
		db:     db,

		gameCache: lobby.NewGameCache(db, time.Second*1, time.Duration(time.Second*5)), /* TODO: 第三引数はconfigから持ってくる（ValidHeartBeat） */
		ws: &websocket.Dialer{
			Subprotocols:    []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},

		hubs:    make(map[RoomID]*Hub),
		clients: make(map[ClientID]map[RoomID]*game.Client),
	}
	return repo, nil
}

func (r *Repository) GetOrCreateHub(appId AppID, roomId RoomID) (*Hub, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hub, ok := r.hubs[roomId]
	if ok {
		return hub, nil
	}

	// hub->game 接続に使うclientId. このhubを作成するトリガーになったclientIdは使わない
	// roomIdもhostIdもユニークなので hostId:roomId はユニークになるはず。
	clientId := fmt.Sprintf("hub:%d:%s", r.hostId, roomId)

	// todo: log.CurrentLevel()
	logger := log.Get(log.DEBUG).With(
		zap.String("type", "hub"),
		zap.String("room", string(roomId)),
		zap.String("clientId", clientId),
	).Sugar()

	hub = &Hub{
		ID:       RoomID(roomId),
		repo:     r,
		appId:    appId,
		clientId: clientId,

		logger: logger,
		// todo: hubをもっと埋める
	}

	go hub.Start()

	return hub, nil
}
