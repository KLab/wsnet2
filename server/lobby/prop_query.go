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
	OpContain
	OpNotContain
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
	if q.Op == OpContain || q.Op == OpNotContain {
		return q.contain(val)
	}

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

func (q *PropQuery) containBool(val []byte) bool {
	qv, _, e := binary.UnmarshalAs(q.Val, binary.TypeTrue, binary.TypeFalse)
	if e != nil {
		return q.Op == OpNotContain
	}
	qval := qv.(bool)

	list, _, e := binary.UnmarshalAs(val, binary.TypeBools)
	if e != nil {
		return q.Op == OpNotContain
	}

	for _, v := range list.([]bool) {
		if v == qval {
			return q.Op == OpContain
		}
	}

	return q.Op == OpNotContain
}

func (q *PropQuery) containNum(val []byte, listtype, valtype binary.Type) bool {
	qv, _, e := binary.UnmarshalAs(q.Val, valtype)
	if e != nil {
		return q.Op == OpNotContain
	}
	qval := qv.(int)

	list, _, e := binary.UnmarshalAs(val, listtype)
	if e != nil {
		return q.Op == OpNotContain
	}

	for _, v := range list.([]int) {
		if v == qval {
			return q.Op == OpContain
		}
	}

	return q.Op == OpNotContain
}

func (q *PropQuery) containChar(val []byte) bool {
	qv, _, e := binary.UnmarshalAs(q.Val, binary.TypeChar)
	if e != nil {
		return q.Op == OpNotContain
	}
	qval := qv.(rune)

	list, _, e := binary.UnmarshalAs(val, binary.TypeChars)
	if e != nil {
		return q.Op == OpNotContain
	}

	for _, v := range list.([]rune) {
		if v == qval {
			return q.Op == OpContain
		}
	}

	return q.Op == OpNotContain
}

func (q *PropQuery) containULong(val []byte) bool {
	qv, _, e := binary.UnmarshalAs(q.Val, binary.TypeULong)
	if e != nil {
		return q.Op == OpNotContain
	}
	qval := qv.(uint64)

	list, _, e := binary.UnmarshalAs(val, binary.TypeULongs)
	if e != nil {
		return q.Op == OpNotContain
	}

	for _, v := range list.([]uint64) {
		if v == qval {
			return q.Op == OpContain
		}
	}

	return q.Op == OpNotContain
}

func (q *PropQuery) containFloat(val []byte) bool {
	qv, _, e := binary.UnmarshalAs(q.Val, binary.TypeFloat)
	if e != nil {
		return q.Op == OpNotContain
	}
	qval := qv.(float32)

	list, _, e := binary.UnmarshalAs(val, binary.TypeFloats)
	if e != nil {
		return q.Op == OpNotContain
	}

	for _, v := range list.([]float32) {
		if v == qval {
			return q.Op == OpContain
		}
	}

	return q.Op == OpNotContain
}

func (q *PropQuery) containDouble(val []byte) bool {
	qv, _, e := binary.UnmarshalAs(q.Val, binary.TypeDouble)
	if e != nil {
		return q.Op == OpNotContain
	}
	qval := qv.(float64)

	list, _, e := binary.UnmarshalAs(val, binary.TypeDoubles)
	if e != nil {
		return q.Op == OpNotContain
	}

	for _, v := range list.([]float64) {
		if v == qval {
			return q.Op == OpContain
		}
	}

	return q.Op == OpNotContain
}

func (q *PropQuery) contain(val []byte) bool {
	switch binary.Type(val[0]) {
	case binary.TypeNull:
		return q.Op == OpNotContain
	case binary.TypeList:
		l, _, e := binary.UnmarshalAs(val, binary.TypeList)
		if e != nil {
			return q.Op == OpNotContain
		}
		for _, v := range l.(binary.List) {
			if bytes.Compare(v, q.Val) == 0 {
				return q.Op == OpContain
			}
		}
		return q.Op == OpNotContain
	case binary.TypeBools:
		return q.containBool(val)
	case binary.TypeSBytes:
		return q.containNum(val, binary.TypeSBytes, binary.TypeSByte)
	case binary.TypeBytes:
		return q.containNum(val, binary.TypeBytes, binary.TypeByte)
	case binary.TypeChars:
		return q.containChar(val)
	case binary.TypeShorts:
		return q.containNum(val, binary.TypeShorts, binary.TypeShort)
	case binary.TypeUShorts:
		return q.containNum(val, binary.TypeUShorts, binary.TypeUShort)
	case binary.TypeInts:
		return q.containNum(val, binary.TypeInts, binary.TypeInt)
	case binary.TypeUInts:
		return q.containNum(val, binary.TypeUInts, binary.TypeUInt)
	case binary.TypeLongs:
		return q.containNum(val, binary.TypeLongs, binary.TypeLong)
	case binary.TypeULongs:
		return q.containULong(val)
	case binary.TypeFloats:
		return q.containFloat(val)
	case binary.TypeDoubles:
		return q.containDouble(val)
	}

	log.Errorf("PropQuery.contain: property is not a list: %v", binary.Type(val[0]))
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
