package lobby

import (
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/log"
)

type hubServerID gameServerID
type hubServer gameServer

type hubCache struct {
	sync.Mutex
	db     *sqlx.DB
	expire time.Duration
	valid  time.Duration

	servers     map[hubServerID]*hubServer
	order       []hubServerID
	lastUpdated time.Time
}

func newHubCache(db *sqlx.DB, expire time.Duration, valid time.Duration) *hubCache {
	return &hubCache{
		db:      db,
		expire:  expire,
		valid:   valid,
		servers: make(map[hubServerID]*hubServer),
		order:   []hubServerID{},
	}
}

func (c *hubCache) updateInner() error {
	query := "SELECT id, hostname, public_name, grpc_port, ws_port FROM hub_server WHERE status=1 AND heartbeat >= ?"

	var servers []hubServer
	err := c.db.Select(&servers, query, time.Now().Add(-c.valid).Unix())
	if err != nil {
		return err
	}

	log.Debugf("Now alive hub servers: %+v", servers)

	c.servers = make(map[hubServerID]*hubServer, len(servers))
	c.order = make([]hubServerID, len(servers))
	for i, s := range servers {
		c.servers[hubServerID(s.Id)] = &servers[i]
		c.order[i] = hubServerID(s.Id)
	}
	c.lastUpdated = time.Now()
	return nil
}

func (c *hubCache) update() error {
	if c.lastUpdated.Add(c.expire).Before(time.Now()) {
		return c.updateInner()
	}
	return nil
}

func (c *hubCache) Get(id uint32) (*hubServer, error) {
	c.Lock()
	defer c.Unlock()
	if err := c.update(); err != nil {
		return nil, err
	}

	if len(c.servers) == 0 {
		return nil, xerrors.New("no available hub server")
	}
	hub := c.servers[hubServerID(id)]
	if hub == nil {
		return nil, xerrors.Errorf("hub server not found (id=%v)", id)
	}
	return hub, nil
}

func (c *hubCache) Rand() (*hubServer, error) {
	c.Lock()
	defer c.Unlock()
	if err := c.update(); err != nil {
		return nil, err
	}

	if len(c.order) == 0 {
		return nil, xerrors.New("no available hub server")
	}
	id := c.order[rand.Intn(len(c.order))]
	return c.servers[id], nil
}
