package lobby

import (
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/log"
)

type gameServerID uint32

type gameServer struct {
	Id            uint32
	Hostname      string
	PublicName    string `db:"public_name"`
	GRPCPort      int    `db:"grpc_port"`
	WebSocketPort int    `db:"ws_port"`
}

type gameCache struct {
	sync.Mutex
	db     *sqlx.DB
	expire time.Duration
	valid  time.Duration

	servers     map[gameServerID]*gameServer
	order       []gameServerID
	lastUpdated time.Time
}

func newGameCache(db *sqlx.DB, expire time.Duration, valid time.Duration) *gameCache {
	return &gameCache{
		db:      db,
		expire:  expire,
		valid:   valid,
		servers: make(map[gameServerID]*gameServer),
		order:   []gameServerID{},
	}
}

func (c *gameCache) updateInner() error {
	query := "SELECT id, hostname, public_name, grpc_port, ws_port FROM game_server WHERE status=1 AND heartbeat >= ?"

	var servers []gameServer
	err := c.db.Select(&servers, query, time.Now().Add(-c.valid).Unix())
	if err != nil {
		return err
	}

	log.Debugf("Now alive game servers: %+v", servers)

	c.servers = make(map[gameServerID]*gameServer, len(servers))
	c.order = make([]gameServerID, len(servers))
	for i, s := range servers {
		c.servers[gameServerID(s.Id)] = &servers[i]
		c.order[i] = gameServerID(s.Id)
	}
	c.lastUpdated = time.Now()
	return nil
}

func (c *gameCache) update() error {
	if c.lastUpdated.Add(c.expire).Before(time.Now()) {
		return c.updateInner()
	}
	return nil
}

func (c *gameCache) Get(id uint32) (*gameServer, error) {
	c.Lock()
	defer c.Unlock()
	if err := c.update(); err != nil {
		return nil, err
	}

	if len(c.servers) == 0 {
		return nil, xerrors.New("no available game server")
	}
	game := c.servers[gameServerID(id)]
	if game == nil {
		return nil, xerrors.Errorf("game server not found (id=%v)", id)
	}
	return game, nil
}

func (c *gameCache) Rand() (*gameServer, error) {
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

func (c *gameCache) All() ([]*gameServer, error) {
	c.Lock()
	defer c.Unlock()
	if err := c.update(); err != nil {
		return nil, err
	}

	res := make([]*gameServer, 0, len(c.servers))
	for _, gs := range c.servers {
		res = append(res, gs)
	}
	return res, nil
}
