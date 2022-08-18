package common

import (
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/log"
)

type GameServerID uint32

type GameServer struct {
	Id            uint32
	Hostname      string
	PublicName    string `db:"public_name"`
	GRPCPort      int    `db:"grpc_port"`
	WebSocketPort int    `db:"ws_port"`
}

type GameCache struct {
	sync.Mutex
	db     *sqlx.DB
	expire time.Duration
	valid  time.Duration

	servers     map[GameServerID]*GameServer
	order       []GameServerID
	lastUpdated time.Time
}

func NewGameCache(db *sqlx.DB, expire time.Duration, valid time.Duration) *GameCache {
	return &GameCache{
		db:      db,
		expire:  expire,
		valid:   valid,
		servers: make(map[GameServerID]*GameServer),
		order:   []GameServerID{},
	}
}

func (c *GameCache) updateInner() error {
	query := "SELECT id, hostname, public_name, grpc_port, ws_port FROM game_server WHERE status=1 AND heartbeat >= ?"

	var servers []GameServer
	err := c.db.Select(&servers, query, time.Now().Add(-c.valid).Unix())
	if err != nil {
		return err
	}

	log.Debugf("Now alive game servers: %+v", servers)

	c.servers = make(map[GameServerID]*GameServer, len(servers))
	c.order = make([]GameServerID, len(servers))
	for i, s := range servers {
		c.servers[GameServerID(s.Id)] = &servers[i]
		c.order[i] = GameServerID(s.Id)
	}
	c.lastUpdated = time.Now()
	return nil
}

func (c *GameCache) update() error {
	if c.lastUpdated.Add(c.expire).Before(time.Now()) {
		return c.updateInner()
	}
	return nil
}

func (c *GameCache) Get(id uint32) (*GameServer, error) {
	c.Lock()
	defer c.Unlock()
	if err := c.update(); err != nil {
		return nil, err
	}

	if len(c.servers) == 0 {
		return nil, xerrors.New("no available game server")
	}
	game := c.servers[GameServerID(id)]
	if game == nil {
		return nil, xerrors.Errorf("game server not found (id=%v)", id)
	}
	return game, nil
}

func (c *GameCache) Rand() (*GameServer, error) {
	c.Lock()
	defer c.Unlock()
	if err := c.update(); err != nil {
		return nil, err
	}

	if len(c.order) == 0 {
		return nil, xerrors.New("no available game server")
	}
	id := c.order[rand.Intn(len(c.order))]
	return c.servers[id], nil
}

func (c *GameCache) All() ([]*GameServer, error) {
	c.Lock()
	defer c.Unlock()
	if err := c.update(); err != nil {
		return nil, err
	}

	res := make([]*GameServer, 0, len(c.servers))
	for _, gs := range c.servers {
		res = append(res, gs)
	}
	return res, nil
}
