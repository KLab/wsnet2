package lobby

import (
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/log"
)

type HubServerID GameServerID
type HubServer GameServer

type HubCache struct {
	sync.Mutex
	db     *sqlx.DB
	expire time.Duration
	valid  time.Duration

	servers     map[HubServerID]*HubServer
	order       []HubServerID
	lastUpdated time.Time
}

func NewHubCache(db *sqlx.DB, expire time.Duration, valid time.Duration) *HubCache {
	return &HubCache{
		db:      db,
		expire:  expire,
		valid:   valid,
		servers: make(map[HubServerID]*HubServer),
		order:   []HubServerID{},
	}
}

func (c *HubCache) update() error {
	query := "SELECT id, hostname, public_name, grpc_port, ws_port FROM hub_server WHERE status=1 AND heartbeat >= ?"

	var servers []HubServer
	err := c.db.Select(&servers, query, time.Now().Add(-c.valid).Unix())
	if err != nil {
		return err
	}

	log.Debugf("Now alive hub servers: %+v", servers)

	c.servers = make(map[HubServerID]*HubServer, len(servers))
	c.order = make([]HubServerID, len(servers))
	for i, s := range servers {
		c.servers[HubServerID(s.Id)] = &servers[i]
		c.order[i] = HubServerID(s.Id)
	}
	c.lastUpdated = time.Now()
	return nil
}

func (c *HubCache) Get(id uint32) (*HubServer, error) {
	c.Lock()
	defer c.Unlock()
	if c.lastUpdated.Add(c.expire).Before(time.Now()) {
		err := c.update()
		if err != nil {
			return nil, err
		}
	}
	if len(c.servers) == 0 {
		return nil, xerrors.New("no available hub server")
	}
	hub := c.servers[HubServerID(id)]
	if hub == nil {
		return nil, xerrors.Errorf("hub server not found (id=%v)", id)
	}
	return hub, nil
}

func (c *HubCache) Rand() (*HubServer, error) {
	c.Lock()
	defer c.Unlock()
	if c.lastUpdated.Add(c.expire).Before(time.Now()) {
		err := c.update()
		if err != nil {
			return nil, err
		}
	}
	if len(c.order) == 0 {
		return nil, xerrors.New("no available hub server")
	}
	id := c.order[rand.Intn(len(c.order))]
	return c.servers[id], nil
}
