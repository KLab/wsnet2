package lobby

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"wsnet2/binary"
	"wsnet2/log"
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
	props       []binary.Dict
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

func (q *roomCacheQuery) do() ([]pb.RoomInfo, []binary.Dict, error) {
	q.Lock()
	defer q.Unlock()

	now := time.Now()

	if q.lastUpdated.Add(q.expire).After(now) {
		return q.result, q.props, q.lastError
	}

	rooms := []pb.RoomInfo{}
	err := q.db.Select(&rooms, q.query, q.args...)
	if err != nil {
		q.result = nil
		q.lastError = err
		return nil, nil, err
	}

	props := []binary.Dict{}
	for _, r := range rooms {
		um, err := unmarshalProps(r.PublicProps)
		if err != nil {
			log.Errorf("props unmarshal error: %v", err)
			props = append(props, binary.Dict{})
			continue
		}
		props = append(props, um)
	}

	q.result = rooms
	q.props = props
	q.lastError = nil
	q.lastUpdated = time.Now()

	return q.result, q.props, q.lastError
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

func (c *RoomCache) GetRooms(appId string, searchGroup uint32) ([]pb.RoomInfo, []binary.Dict, error) {
	c.Lock()
	q := c.queries[appId][searchGroup]
	if q == nil {
		if c.queries[appId] == nil {
			c.queries[appId] = make(map[uint32]*roomCacheQuery)
		}
		q = newRoomCacheQuery(c.db, c.expire, "SELECT * FROM room WHERE app_id = ? AND search_group = ? AND visible = 1 LIMIT 1000", appId, searchGroup)
		c.queries[appId][searchGroup] = q
	}
	c.Unlock()

	return q.do()
}
