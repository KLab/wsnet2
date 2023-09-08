package client

import "wsnet2/lobby"

type Query []lobby.PropQueries

func NewQuery() *Query {
	q := Query([]lobby.PropQueries{[]lobby.PropQuery{}})
	return &q
}

func (q *Query) Equal(key string, val []byte) *Query {
	q.and(key, lobby.OpEqual, val)
	return q
}

func (q *Query) Not(key string, val []byte) *Query {
	q.and(key, lobby.OpNot, val)
	return q
}

func (q *Query) LessThan(key string, val []byte) *Query {
	q.and(key, lobby.OpLessThan, val)
	return q
}
func (q *Query) LessEqual(key string, val []byte) *Query {
	q.and(key, lobby.OpLessThanOrEqual, val)
	return q
}

func (q *Query) GreaterThan(key string, val []byte) *Query {
	q.and(key, lobby.OpGreaterThan, val)
	return q
}

func (q *Query) GreaterEqual(key string, val []byte) *Query {
	q.and(key, lobby.OpGreaterThanOrEqual, val)
	return q
}

func (q *Query) and(key string, op lobby.OpType, val []byte) {
	pq := lobby.PropQuery{Key: key, Op: op, Val: val}
	for i := range *q {
		(*q)[i] = append((*q)[i], pq)
	}
}
