package lobby

import (
	"bytes"

	"wsnet2/binary"
	"wsnet2/log"
)

type PropQuery struct {
	Key string
	Op  string
	Val []byte
}

func (q *PropQuery) match(val []byte) bool {
	ret := bytes.Compare(val, q.Val)
	switch q.Op {
	case "=":
		return ret == 0
	case "!":
		return ret != 0
	case "<":
		return ret < 0
	case "<=":
		return ret <= 0
	case ">":
		return ret > 0
	case ">=":
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
