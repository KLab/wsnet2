package hub

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"wsnet2/config"
	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/pb"
)

type RoomID = game.RoomID
type ClientID = game.ClientID

type Repository struct {
	hostId uint32

	app  pb.App
	conf *config.GameConf
	db   *sqlx.DB

	mu      sync.RWMutex
	hubs    map[RoomID]*Hub
	clients map[ClientID]map[RoomID]*game.Client
}

func NewRepository(db *sqlx.DB, conf *config.GameConf, hostId uint32) (*Repository, error) {
	repo := &Repository{
		hostId: hostId,
		conf:   conf,
		db:     db,

		hubs:    make(map[RoomID]*Hub),
		clients: make(map[ClientID]map[RoomID]*game.Client),
	}
	return repo, nil
}

func (r *Repository) GetOrCreateHub(roomId RoomID) (*Hub, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hub, ok := r.hubs[roomId]
	if ok {
		return hub, nil
	}

	// hub->game 接続に使うclientId. このhubを作成するトリガーになったclientIdは使わない
	// roomIdもhostIdもユニークなので hostId:roomId はユニークになるはず。
	clientId := fmt.Sprintf("hub:%s:%s", r.hostId, roomId)

	// todo: log.CurrentLevel()
	logger := log.Get(log.DEBUG).With(
		zap.String("type", "hub"),
		zap.String("room", string(roomId)),
		zap.String("clientId", clientId),
	).Sugar()

	hub = &Hub{
		ID:       RoomID(roomId),
		repo:     r,
		clientId: clientId,

		logger: logger,
		// todo: hubをもっと埋める
	}

	go hub.Start()

	return hub, nil
}
