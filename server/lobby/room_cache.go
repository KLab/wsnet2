package lobby

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"wsnet2/pb"
)

type roomCacheQuery struct {
	sync.Mutex
	db     *sqlx.DB
	expire time.Duration
	query  string
	args   []interface{}

	lastUpdated time.Time
	result      []pb.RoomInfo
	lastError   error
}

func newRoomCacheQuery(db *sqlx.DB, expire time.Duration, sql string, args ...interface{}) *roomCacheQuery {
	return &roomCacheQuery{
		db:     db,
		expire: expire,
		query:  sql,
		args:   args,
	}
}

func (q *roomCacheQuery) do() ([]pb.RoomInfo, error) {
	q.Lock()
	defer q.Unlock()

	now := time.Now()

	if q.lastUpdated.Add(q.expire).After(now) {
		return q.result, q.lastError
	}

	rooms := []pb.RoomInfo{}
	err := q.db.Select(&rooms, q.query, q.args...)
	if err != nil {
		q.result = nil
		q.lastError = err
		return nil, err
	}

	q.result = rooms
	q.lastError = nil
	q.lastUpdated = time.Now()

	return q.result, q.lastError
}

type RoomCache struct {
	sync.Mutex
	db      *sqlx.DB
	expire  time.Duration
	queries map[string]map[uint32]*roomCacheQuery
}

func NewRoomCache(db *sqlx.DB, expire time.Duration) *RoomCache {
	return &RoomCache{
		db:      db,
		expire:  expire,
		queries: make(map[string]map[uint32]*roomCacheQuery),
	}
}

func (c *RoomCache) GetRooms(appId string, searchGroup uint32) ([]pb.RoomInfo, error) {
	c.Lock()
	q := c.queries[appId][searchGroup]
	if q == nil {
		if c.queries[appId] == nil {
			c.queries[appId] = make(map[uint32]*roomCacheQuery)
		}
		q = newRoomCacheQuery(c.db, c.expire, "SELECT * FROM room WHERE app_id = ? AND search_group = ? LIMIT 1000", appId, searchGroup)
		c.queries[appId][searchGroup] = q
	}
	c.Unlock()

	return q.do()
}
