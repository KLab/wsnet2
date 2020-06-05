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

type GameQuery struct {
	db    *sqlx.DB
	valid int64

	muServers   sync.Mutex
	servers     map[GameServerID]*GameServer
	order       []GameServerID
	lastUpdated time.Time
}

func NewGameQuery(db *sqlx.DB, valid int64) *GameQuery {
	return &GameQuery{
		db:      db,
		valid:   valid,
		servers: make(map[GameServerID]*GameServer),
		order:   []GameServerID{},
	}
}

func (q *GameQuery) update() error {
	query := "SELECT id, hostname, public_name, grpc_port, ws_port FROM game_server WHERE status=1 AND heartbeat >= ?"

	var servers []GameServer
	err := q.db.Select(&servers, query, time.Now().Unix()-q.valid)
	if err != nil {
		return err
	}

	log.Debugf("Now alive game servers: %+v", servers)

	q.servers = make(map[GameServerID]*GameServer, len(servers))
	for _, s := range servers {
		q.servers[GameServerID(s.Id)] = &s
		q.order = append(q.order, GameServerID(s.Id))
	}
	q.lastUpdated = time.Now()
	return nil
}

func (q *GameQuery) Get(id uint32) (*GameServer, error) {
	q.muServers.Lock()
	defer q.muServers.Unlock()
	if q.lastUpdated.Add(time.Second * 1).Before(time.Now()) {
		err := q.update()
		if err != nil {
			return nil, err
		}
	}
	if len(q.servers) == 0 {
		return nil, xerrors.New("no available game server")
	}
	game := q.servers[GameServerID(id)]
	if game == nil {
		return nil, xerrors.Errorf("game server not found (id=%v)", id)
	}
	return game, nil
}

func (q *GameQuery) Rand() (*GameServer, error) {
	q.muServers.Lock()
	defer q.muServers.Unlock()
	if q.lastUpdated.Add(time.Second * 1).Before(time.Now()) {
		err := q.update()
		if err != nil {
			return nil, err
		}
	}
	if len(q.order) == 0 {
		return nil, xerrors.New("no available game server")
	}
	id := q.order[rand.Intn(len(q.order))]
	return q.servers[id], nil
}
