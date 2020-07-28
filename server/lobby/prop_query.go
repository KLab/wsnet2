package lobby

import (
	"bytes"

	"golang.org/x/xerrors"

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

func unmarshalProps(props []byte) (binary.Dict, error) {
	um, _, err := binary.Unmarshal(props)
	if err != nil {
		return nil, err
	}
	dict, ok := um.(binary.Dict)
	if !ok {
		return nil, xerrors.Errorf("type is not Dict: %v", binary.Type(props[0]))
	}
	return dict, nil
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
