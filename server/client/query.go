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

func (q *Query) and(key string, op lobby.OpType, val []byte) {
	pq := lobby.PropQuery{Key: key, Op: op, Val: val}
	for i := range *q {
		(*q)[i] = append((*q)[i], pq)
	}
}
