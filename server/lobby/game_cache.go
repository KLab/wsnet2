package lobby

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"

	"wsnet2/log"
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
}

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
	valid  int64

	servers     map[GameServerID]*GameServer
	order       []GameServerID
	lastUpdated time.Time
}

func NewGameCache(db *sqlx.DB, expire time.Duration, valid int64) *GameCache {
	return &GameCache{
		db:      db,
		expire:  expire,
		valid:   valid,
		servers: make(map[GameServerID]*GameServer),
		order:   []GameServerID{},
	}
}

func (c *GameCache) update() error {
	query := "SELECT id, hostname, public_name, grpc_port, ws_port FROM game_server WHERE status=1 AND heartbeat >= ?"

	var servers []GameServer
	err := c.db.Select(&servers, query, time.Now().Unix()-c.valid)
	if err != nil {
		return err
	}

	log.Debugf("Now alive game servers: %+v", servers)

	c.servers = make(map[GameServerID]*GameServer, len(servers))
	for _, s := range servers {
		c.servers[GameServerID(s.Id)] = &s
		c.order = append(c.order, GameServerID(s.Id))
	}
	c.lastUpdated = time.Now()
	return nil
}

func (c *GameCache) Get(id uint32) (*GameServer, error) {
	c.Lock()
	defer c.Unlock()
	if c.lastUpdated.Add(c.expire).Before(time.Now()) {
		err := c.update()
		if err != nil {
			return nil, err
		}
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
	if c.lastUpdated.Add(c.expire).Before(time.Now()) {
		err := c.update()
		if err != nil {
			return nil, err
		}
	}
	if len(c.order) == 0 {
		return nil, xerrors.New("no available game server")
	}
	id := c.order[rand.Intn(len(c.order))]
	return c.servers[id], nil
}
