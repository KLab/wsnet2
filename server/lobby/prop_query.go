package lobby

import (
	"bytes"

	"wsnet2/binary"
	"wsnet2/log"
)

//go:generate stringer -type=OpType
type OpType byte

const (
	OpEqual OpType = iota
	OpNot
	OpLessThan
	OpLessThanOrEqual
	OpGreaterThan
	OpGreaterThanOrEqual
)

type PropQuery struct {
	Key string
	Op  OpType
	Val []byte
}

func (q *PropQuery) match(val []byte) bool {
	ret := bytes.Compare(val, q.Val)
	switch q.Op {
	case OpEqual:
		return ret == 0
	case OpNot:
		return ret != 0
	case OpLessThan:
		return ret < 0
	case OpLessThanOrEqual:
		return ret <= 0
	case OpGreaterThan:
		return ret > 0
	case OpGreaterThanOrEqual:
		return ret >= 0
	}
	log.Errorf("unsupported operator: %v", q.Op)
	return false
}

type PropQueries []PropQuery

func (pqs *PropQueries) match(props binary.Dict) bool {
	for _, q := range *pqs {
		if !q.match(props[q.Key]) {
			return false
		}
	}
	return true
}
