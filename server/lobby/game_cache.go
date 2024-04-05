package lobby

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/common"
	"wsnet2/log"
)

type hostInfo struct {
	Id            uint32
	Hostname      string
	PublicName    string `db:"public_name"`
	GRPCPort      int    `db:"grpc_port"`
	WebSocketPort int    `db:"ws_port"`
}

type gameServer struct {
	hostInfo
	Status int32
}

type gameCache struct {
	sync.Mutex
	db     *sqlx.DB
	expire time.Duration
	valid  time.Duration

	servers     map[uint32]*gameServer
	order       []uint32
	lastUpdated time.Time
}

func newGameCache(db *sqlx.DB, expire time.Duration, valid time.Duration) *gameCache {
	return &gameCache{
		db:      db,
		expire:  expire,
		valid:   valid,
		servers: make(map[uint32]*gameServer),
		order:   []uint32{},
	}
}

func (c *gameCache) updateInner() error {
	// 再入室のために、graceful shutdown中のサーバー(status == closing == 2)の情報も取得する.
	query := ("SELECT id, hostname, public_name, grpc_port, ws_port, status\n" +
		"FROM game_server WHERE status IN (1, 2) AND heartbeat >= ?")

	var servers []gameServer
	err := c.db.Select(&servers, query, time.Now().Add(-c.valid).Unix())
	if err != nil {
		return xerrors.Errorf("selecting game servers: %w", err)
	}

	log.Debugf("Now alive game servers: %+v", servers)

	c.servers = make(map[uint32]*gameServer, len(servers))
	c.order = make([]uint32, 0, len(servers))
	for i := range servers {
		s := &servers[i]
		c.servers[s.Id] = s
		// Rand() がgraceful shutdown中のサーバーを返さないために、
		// status=running のサーバーのみ order に追加する.
		if s.Status == common.HostStatusRunning {
			c.order = append(c.order, s.Id)
		}
	}
	c.lastUpdated = time.Now()
	return nil
}

func (c *gameCache) update() error {
	if time.Since(c.lastUpdated) > c.expire {
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
	game := c.servers[id]
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
	id := c.order[rand.IntN(len(c.order))]
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
