package cachedquery

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type Scanner func(rows *sqlx.Rows) (interface{}, error)

type Query struct {
	sync.Mutex
	expire      time.Duration
	db          *sqlx.DB
	query       string
	args        []interface{}
	scanner     Scanner
	lastUpdated time.Time
	result      interface{}
	lastError   error
}

func New(db *sqlx.DB, expire time.Duration, scanner Scanner, sql string, args ...interface{}) *Query {
	return &Query{
		expire:  expire,
		db:      db,
		query:   sql,
		args:    args,
		scanner: scanner,
	}
}

func (q *Query) Query() (interface{}, error) {
	now := time.Now()
	q.Lock()
	defer q.Unlock()

	if q.lastUpdated.Add(q.expire).After(now) {
		return q.result, q.lastError
	}

	rows, err := q.db.Queryx(q.query, q.args...)
	if err != nil {
		q.result = nil
		q.lastError = err
		return nil, err
	}

	q.result, q.lastError = q.scanner(rows)
	q.lastUpdated = time.Now()
	rows.Close()
	return q.result, q.lastError
}
