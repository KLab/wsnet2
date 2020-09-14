package lobby

import (
	"bytes"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmihailenco/msgpack/v4"

	"wsnet2/binary"
)

func TestPropQueryMatchBool(t *testing.T) {
	props := binary.Dict{
		"true":  binary.MarshalBool(true),
		"false": binary.MarshalBool(false),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"true", OpEqual, binary.MarshalBool(true)}, true},
		{PropQuery{"true", OpEqual, binary.MarshalBool(false)}, false},

		{PropQuery{"true", OpNot, binary.MarshalBool(true)}, false},
		{PropQuery{"true", OpNot, binary.MarshalBool(false)}, true},

		{PropQuery{"false", OpEqual, binary.MarshalBool(true)}, false},
		{PropQuery{"false", OpEqual, binary.MarshalBool(false)}, true},

		{PropQuery{"false", OpNot, binary.MarshalBool(true)}, true},
		{PropQuery{"false", OpNot, binary.MarshalBool(false)}, false},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchByte(t *testing.T) {
	props := binary.Dict{
		"0":        binary.MarshalByte(0),
		"MaxUint8": binary.MarshalByte(math.MaxUint8),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"0", OpEqual, binary.MarshalByte(0)}, true},
		{PropQuery{"0", OpEqual, binary.MarshalByte(1)}, false},

		{PropQuery{"0", OpNot, binary.MarshalByte(0)}, false},
		{PropQuery{"0", OpNot, binary.MarshalByte(1)}, true},

		{PropQuery{"0", OpLessThan, binary.MarshalByte(0)}, false},
		{PropQuery{"0", OpLessThan, binary.MarshalByte(1)}, true},

		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalByte(0)}, true},
		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalByte(1)}, true},

		{PropQuery{"0", OpGreaterThan, binary.MarshalByte(0)}, false},
		{PropQuery{"0", OpGreaterThan, binary.MarshalByte(1)}, false},

		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalByte(0)}, true},
		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalByte(1)}, false},

		{PropQuery{"MaxUint8", OpEqual, binary.MarshalByte(math.MaxUint8 - 1)}, false},
		{PropQuery{"MaxUint8", OpEqual, binary.MarshalByte(math.MaxUint8)}, true},

		{PropQuery{"MaxUint8", OpNot, binary.MarshalByte(math.MaxUint8 - 1)}, true},
		{PropQuery{"MaxUint8", OpNot, binary.MarshalByte(math.MaxUint8)}, false},

		{PropQuery{"MaxUint8", OpLessThan, binary.MarshalByte(math.MaxUint8 - 1)}, false},
		{PropQuery{"MaxUint8", OpLessThan, binary.MarshalByte(math.MaxUint8)}, false},

		{PropQuery{"MaxUint8", OpLessThanOrEqual, binary.MarshalByte(math.MaxUint8 - 1)}, false},
		{PropQuery{"MaxUint8", OpLessThanOrEqual, binary.MarshalByte(math.MaxUint8)}, true},

		{PropQuery{"MaxUint8", OpGreaterThan, binary.MarshalByte(math.MaxUint8 - 1)}, true},
		{PropQuery{"MaxUint8", OpGreaterThan, binary.MarshalByte(math.MaxUint8)}, false},

		{PropQuery{"MaxUint8", OpGreaterThanOrEqual, binary.MarshalByte(math.MaxUint8 - 1)}, true},
		{PropQuery{"MaxUint8", OpGreaterThanOrEqual, binary.MarshalByte(math.MaxUint8)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchSByte(t *testing.T) {
	props := binary.Dict{
		"MinInt8": binary.MarshalSByte(math.MinInt8),
		"MaxInt8": binary.MarshalSByte(math.MaxInt8),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"MinInt8", OpEqual, binary.MarshalSByte(math.MinInt8)}, true},
		{PropQuery{"MinInt8", OpEqual, binary.MarshalSByte(math.MinInt8 + 1)}, false},

		{PropQuery{"MinInt8", OpNot, binary.MarshalSByte(math.MinInt8)}, false},
		{PropQuery{"MinInt8", OpNot, binary.MarshalSByte(math.MinInt8 + 1)}, true},

		{PropQuery{"MinInt8", OpLessThan, binary.MarshalSByte(math.MinInt8)}, false},
		{PropQuery{"MinInt8", OpLessThan, binary.MarshalSByte(math.MinInt8 + 1)}, true},

		{PropQuery{"MinInt8", OpLessThanOrEqual, binary.MarshalSByte(math.MinInt8)}, true},
		{PropQuery{"MinInt8", OpLessThanOrEqual, binary.MarshalSByte(math.MinInt8 + 1)}, true},

		{PropQuery{"MinInt8", OpGreaterThan, binary.MarshalSByte(math.MinInt8)}, false},
		{PropQuery{"MinInt8", OpGreaterThan, binary.MarshalSByte(math.MinInt8 + 1)}, false},

		{PropQuery{"MinInt8", OpGreaterThanOrEqual, binary.MarshalSByte(math.MinInt8)}, true},
		{PropQuery{"MinInt8", OpGreaterThanOrEqual, binary.MarshalSByte(math.MinInt8 + 1)}, false},

		{PropQuery{"MaxInt8", OpEqual, binary.MarshalSByte(math.MaxInt8 - 1)}, false},
		{PropQuery{"MaxInt8", OpEqual, binary.MarshalSByte(math.MaxInt8)}, true},

		{PropQuery{"MaxInt8", OpNot, binary.MarshalSByte(math.MaxInt8 - 1)}, true},
		{PropQuery{"MaxInt8", OpNot, binary.MarshalSByte(math.MaxInt8)}, false},

		{PropQuery{"MaxInt8", OpLessThan, binary.MarshalSByte(math.MaxInt8 - 1)}, false},
		{PropQuery{"MaxInt8", OpLessThan, binary.MarshalSByte(math.MaxInt8)}, false},

		{PropQuery{"MaxInt8", OpLessThanOrEqual, binary.MarshalSByte(math.MaxInt8 - 1)}, false},
		{PropQuery{"MaxInt8", OpLessThanOrEqual, binary.MarshalSByte(math.MaxInt8)}, true},

		{PropQuery{"MaxInt8", OpGreaterThan, binary.MarshalSByte(math.MaxInt8 - 1)}, true},
		{PropQuery{"MaxInt8", OpGreaterThan, binary.MarshalSByte(math.MaxInt8)}, false},

		{PropQuery{"MaxInt8", OpGreaterThanOrEqual, binary.MarshalSByte(math.MaxInt8 - 1)}, true},
		{PropQuery{"MaxInt8", OpGreaterThanOrEqual, binary.MarshalSByte(math.MaxInt8)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchUShort(t *testing.T) {
	props := binary.Dict{
		"0":         binary.MarshalUShort(0),
		"MaxUint16": binary.MarshalUShort(math.MaxUint16),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"0", OpEqual, binary.MarshalUShort(0)}, true},
		{PropQuery{"0", OpEqual, binary.MarshalUShort(1)}, false},

		{PropQuery{"0", OpNot, binary.MarshalUShort(0)}, false},
		{PropQuery{"0", OpNot, binary.MarshalUShort(1)}, true},

		{PropQuery{"0", OpLessThan, binary.MarshalUShort(0)}, false},
		{PropQuery{"0", OpLessThan, binary.MarshalUShort(1)}, true},

		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalUShort(0)}, true},
		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalUShort(1)}, true},

		{PropQuery{"0", OpGreaterThan, binary.MarshalUShort(0)}, false},
		{PropQuery{"0", OpGreaterThan, binary.MarshalUShort(1)}, false},

		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalUShort(0)}, true},
		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalUShort(1)}, false},

		{PropQuery{"MaxUint16", OpEqual, binary.MarshalUShort(math.MaxUint16 - 1)}, false},
		{PropQuery{"MaxUint16", OpEqual, binary.MarshalUShort(math.MaxUint16)}, true},

		{PropQuery{"MaxUint16", OpNot, binary.MarshalUShort(math.MaxUint16 - 1)}, true},
		{PropQuery{"MaxUint16", OpNot, binary.MarshalUShort(math.MaxUint16)}, false},

		{PropQuery{"MaxUint16", OpLessThan, binary.MarshalUShort(math.MaxUint16 - 1)}, false},
		{PropQuery{"MaxUint16", OpLessThan, binary.MarshalUShort(math.MaxUint16)}, false},

		{PropQuery{"MaxUint16", OpLessThanOrEqual, binary.MarshalUShort(math.MaxUint16 - 1)}, false},
		{PropQuery{"MaxUint16", OpLessThanOrEqual, binary.MarshalUShort(math.MaxUint16)}, true},

		{PropQuery{"MaxUint16", OpGreaterThan, binary.MarshalUShort(math.MaxUint16 - 1)}, true},
		{PropQuery{"MaxUint16", OpGreaterThan, binary.MarshalUShort(math.MaxUint16)}, false},

		{PropQuery{"MaxUint16", OpGreaterThanOrEqual, binary.MarshalUShort(math.MaxUint16 - 1)}, true},
		{PropQuery{"MaxUint16", OpGreaterThanOrEqual, binary.MarshalUShort(math.MaxUint16)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchShort(t *testing.T) {
	props := binary.Dict{
		"MinInt16": binary.MarshalShort(math.MinInt16),
		"MaxInt16": binary.MarshalShort(math.MaxInt16),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"MinInt16", OpEqual, binary.MarshalShort(math.MinInt16)}, true},
		{PropQuery{"MinInt16", OpEqual, binary.MarshalShort(math.MinInt16 + 1)}, false},

		{PropQuery{"MinInt16", OpNot, binary.MarshalShort(math.MinInt16)}, false},
		{PropQuery{"MinInt16", OpNot, binary.MarshalShort(math.MinInt16 + 1)}, true},

		{PropQuery{"MinInt16", OpLessThan, binary.MarshalShort(math.MinInt16)}, false},
		{PropQuery{"MinInt16", OpLessThan, binary.MarshalShort(math.MinInt16 + 1)}, true},

		{PropQuery{"MinInt16", OpLessThanOrEqual, binary.MarshalShort(math.MinInt16)}, true},
		{PropQuery{"MinInt16", OpLessThanOrEqual, binary.MarshalShort(math.MinInt16 + 1)}, true},

		{PropQuery{"MinInt16", OpGreaterThan, binary.MarshalShort(math.MinInt16)}, false},
		{PropQuery{"MinInt16", OpGreaterThan, binary.MarshalShort(math.MinInt16 + 1)}, false},

		{PropQuery{"MinInt16", OpGreaterThanOrEqual, binary.MarshalShort(math.MinInt16)}, true},
		{PropQuery{"MinInt16", OpGreaterThanOrEqual, binary.MarshalShort(math.MinInt16 + 1)}, false},

		{PropQuery{"MaxInt16", OpEqual, binary.MarshalShort(math.MaxInt16 - 1)}, false},
		{PropQuery{"MaxInt16", OpEqual, binary.MarshalShort(math.MaxInt16)}, true},

		{PropQuery{"MaxInt16", OpNot, binary.MarshalShort(math.MaxInt16 - 1)}, true},
		{PropQuery{"MaxInt16", OpNot, binary.MarshalShort(math.MaxInt16)}, false},

		{PropQuery{"MaxInt16", OpLessThan, binary.MarshalShort(math.MaxInt16 - 1)}, false},
		{PropQuery{"MaxInt16", OpLessThan, binary.MarshalShort(math.MaxInt16)}, false},

		{PropQuery{"MaxInt16", OpLessThanOrEqual, binary.MarshalShort(math.MaxInt16 - 1)}, false},
		{PropQuery{"MaxInt16", OpLessThanOrEqual, binary.MarshalShort(math.MaxInt16)}, true},

		{PropQuery{"MaxInt16", OpGreaterThan, binary.MarshalShort(math.MaxInt16 - 1)}, true},
		{PropQuery{"MaxInt16", OpGreaterThan, binary.MarshalShort(math.MaxInt16)}, false},

		{PropQuery{"MaxInt16", OpGreaterThanOrEqual, binary.MarshalShort(math.MaxInt16 - 1)}, true},
		{PropQuery{"MaxInt16", OpGreaterThanOrEqual, binary.MarshalShort(math.MaxInt16)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchUInt(t *testing.T) {
	props := binary.Dict{
		"0":         binary.MarshalUInt(0),
		"MaxUint32": binary.MarshalUInt(math.MaxUint32),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"0", OpEqual, binary.MarshalUInt(0)}, true},
		{PropQuery{"0", OpEqual, binary.MarshalUInt(1)}, false},

		{PropQuery{"0", OpNot, binary.MarshalUInt(0)}, false},
		{PropQuery{"0", OpNot, binary.MarshalUInt(1)}, true},

		{PropQuery{"0", OpLessThan, binary.MarshalUInt(0)}, false},
		{PropQuery{"0", OpLessThan, binary.MarshalUInt(1)}, true},

		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalUInt(0)}, true},
		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalUInt(1)}, true},

		{PropQuery{"0", OpGreaterThan, binary.MarshalUInt(0)}, false},
		{PropQuery{"0", OpGreaterThan, binary.MarshalUInt(1)}, false},

		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalUInt(0)}, true},
		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalUInt(1)}, false},

		{PropQuery{"MaxUint32", OpEqual, binary.MarshalUInt(math.MaxUint32 - 1)}, false},
		{PropQuery{"MaxUint32", OpEqual, binary.MarshalUInt(math.MaxUint32)}, true},

		{PropQuery{"MaxUint32", OpNot, binary.MarshalUInt(math.MaxUint32 - 1)}, true},
		{PropQuery{"MaxUint32", OpNot, binary.MarshalUInt(math.MaxUint32)}, false},

		{PropQuery{"MaxUint32", OpLessThan, binary.MarshalUInt(math.MaxUint32 - 1)}, false},
		{PropQuery{"MaxUint32", OpLessThan, binary.MarshalUInt(math.MaxUint32)}, false},

		{PropQuery{"MaxUint32", OpLessThanOrEqual, binary.MarshalUInt(math.MaxUint32 - 1)}, false},
		{PropQuery{"MaxUint32", OpLessThanOrEqual, binary.MarshalUInt(math.MaxUint32)}, true},

		{PropQuery{"MaxUint32", OpGreaterThan, binary.MarshalUInt(math.MaxUint32 - 1)}, true},
		{PropQuery{"MaxUint32", OpGreaterThan, binary.MarshalUInt(math.MaxUint32)}, false},

		{PropQuery{"MaxUint32", OpGreaterThanOrEqual, binary.MarshalUInt(math.MaxUint32 - 1)}, true},
		{PropQuery{"MaxUint32", OpGreaterThanOrEqual, binary.MarshalUInt(math.MaxUint32)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchInt(t *testing.T) {
	props := binary.Dict{
		"MinInt32": binary.MarshalInt(math.MinInt32),
		"MaxInt32": binary.MarshalInt(math.MaxInt32),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"MinInt32", OpEqual, binary.MarshalInt(math.MinInt32)}, true},
		{PropQuery{"MinInt32", OpEqual, binary.MarshalInt(math.MinInt32 + 1)}, false},

		{PropQuery{"MinInt32", OpNot, binary.MarshalInt(math.MinInt32)}, false},
		{PropQuery{"MinInt32", OpNot, binary.MarshalInt(math.MinInt32 + 1)}, true},

		{PropQuery{"MinInt32", OpLessThan, binary.MarshalInt(math.MinInt32)}, false},
		{PropQuery{"MinInt32", OpLessThan, binary.MarshalInt(math.MinInt32 + 1)}, true},

		{PropQuery{"MinInt32", OpLessThanOrEqual, binary.MarshalInt(math.MinInt32)}, true},
		{PropQuery{"MinInt32", OpLessThanOrEqual, binary.MarshalInt(math.MinInt32 + 1)}, true},

		{PropQuery{"MinInt32", OpGreaterThan, binary.MarshalInt(math.MinInt32)}, false},
		{PropQuery{"MinInt32", OpGreaterThan, binary.MarshalInt(math.MinInt32 + 1)}, false},

		{PropQuery{"MinInt32", OpGreaterThanOrEqual, binary.MarshalInt(math.MinInt32)}, true},
		{PropQuery{"MinInt32", OpGreaterThanOrEqual, binary.MarshalInt(math.MinInt32 + 1)}, false},

		{PropQuery{"MaxInt32", OpEqual, binary.MarshalInt(math.MaxInt32 - 1)}, false},
		{PropQuery{"MaxInt32", OpEqual, binary.MarshalInt(math.MaxInt32)}, true},

		{PropQuery{"MaxInt32", OpNot, binary.MarshalInt(math.MaxInt32 - 1)}, true},
		{PropQuery{"MaxInt32", OpNot, binary.MarshalInt(math.MaxInt32)}, false},

		{PropQuery{"MaxInt32", OpLessThan, binary.MarshalInt(math.MaxInt32 - 1)}, false},
		{PropQuery{"MaxInt32", OpLessThan, binary.MarshalInt(math.MaxInt32)}, false},

		{PropQuery{"MaxInt32", OpLessThanOrEqual, binary.MarshalInt(math.MaxInt32 - 1)}, false},
		{PropQuery{"MaxInt32", OpLessThanOrEqual, binary.MarshalInt(math.MaxInt32)}, true},

		{PropQuery{"MaxInt32", OpGreaterThan, binary.MarshalInt(math.MaxInt32 - 1)}, true},
		{PropQuery{"MaxInt32", OpGreaterThan, binary.MarshalInt(math.MaxInt32)}, false},

		{PropQuery{"MaxInt32", OpGreaterThanOrEqual, binary.MarshalInt(math.MaxInt32 - 1)}, true},
		{PropQuery{"MaxInt32", OpGreaterThanOrEqual, binary.MarshalInt(math.MaxInt32)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchULong(t *testing.T) {
	props := binary.Dict{
		"0":         binary.MarshalULong(0),
		"MaxUint64": binary.MarshalULong(math.MaxUint64),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"0", OpEqual, binary.MarshalULong(0)}, true},
		{PropQuery{"0", OpEqual, binary.MarshalULong(1)}, false},

		{PropQuery{"0", OpNot, binary.MarshalULong(0)}, false},
		{PropQuery{"0", OpNot, binary.MarshalULong(1)}, true},

		{PropQuery{"0", OpLessThan, binary.MarshalULong(0)}, false},
		{PropQuery{"0", OpLessThan, binary.MarshalULong(1)}, true},

		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalULong(0)}, true},
		{PropQuery{"0", OpLessThanOrEqual, binary.MarshalULong(1)}, true},

		{PropQuery{"0", OpGreaterThan, binary.MarshalULong(0)}, false},
		{PropQuery{"0", OpGreaterThan, binary.MarshalULong(1)}, false},

		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalULong(0)}, true},
		{PropQuery{"0", OpGreaterThanOrEqual, binary.MarshalULong(1)}, false},

		{PropQuery{"MaxUint64", OpEqual, binary.MarshalULong(math.MaxUint64 - 1)}, false},
		{PropQuery{"MaxUint64", OpEqual, binary.MarshalULong(math.MaxUint64)}, true},

		{PropQuery{"MaxUint64", OpNot, binary.MarshalULong(math.MaxUint64 - 1)}, true},
		{PropQuery{"MaxUint64", OpNot, binary.MarshalULong(math.MaxUint64)}, false},

		{PropQuery{"MaxUint64", OpLessThan, binary.MarshalULong(math.MaxUint64 - 1)}, false},
		{PropQuery{"MaxUint64", OpLessThan, binary.MarshalULong(math.MaxUint64)}, false},

		{PropQuery{"MaxUint64", OpLessThanOrEqual, binary.MarshalULong(math.MaxUint64 - 1)}, false},
		{PropQuery{"MaxUint64", OpLessThanOrEqual, binary.MarshalULong(math.MaxUint64)}, true},

		{PropQuery{"MaxUint64", OpGreaterThan, binary.MarshalULong(math.MaxUint64 - 1)}, true},
		{PropQuery{"MaxUint64", OpGreaterThan, binary.MarshalULong(math.MaxUint64)}, false},

		{PropQuery{"MaxUint64", OpGreaterThanOrEqual, binary.MarshalULong(math.MaxUint64 - 1)}, true},
		{PropQuery{"MaxUint64", OpGreaterThanOrEqual, binary.MarshalULong(math.MaxUint64)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchLong(t *testing.T) {
	props := binary.Dict{
		"MinInt64": binary.MarshalLong(math.MinInt64),
		"MaxInt64": binary.MarshalLong(math.MaxInt64),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"MinInt64", OpEqual, binary.MarshalLong(math.MinInt64)}, true},
		{PropQuery{"MinInt64", OpEqual, binary.MarshalLong(math.MinInt64 + 1)}, false},

		{PropQuery{"MinInt64", OpNot, binary.MarshalLong(math.MinInt64)}, false},
		{PropQuery{"MinInt64", OpNot, binary.MarshalLong(math.MinInt64 + 1)}, true},

		{PropQuery{"MinInt64", OpLessThan, binary.MarshalLong(math.MinInt64)}, false},
		{PropQuery{"MinInt64", OpLessThan, binary.MarshalLong(math.MinInt64 + 1)}, true},

		{PropQuery{"MinInt64", OpLessThanOrEqual, binary.MarshalLong(math.MinInt64)}, true},
		{PropQuery{"MinInt64", OpLessThanOrEqual, binary.MarshalLong(math.MinInt64 + 1)}, true},

		{PropQuery{"MinInt64", OpGreaterThan, binary.MarshalLong(math.MinInt64)}, false},
		{PropQuery{"MinInt64", OpGreaterThan, binary.MarshalLong(math.MinInt64 + 1)}, false},

		{PropQuery{"MinInt64", OpGreaterThanOrEqual, binary.MarshalLong(math.MinInt64)}, true},
		{PropQuery{"MinInt64", OpGreaterThanOrEqual, binary.MarshalLong(math.MinInt64 + 1)}, false},

		{PropQuery{"MaxInt64", OpEqual, binary.MarshalLong(math.MaxInt64 - 1)}, false},
		{PropQuery{"MaxInt64", OpEqual, binary.MarshalLong(math.MaxInt64)}, true},

		{PropQuery{"MaxInt64", OpNot, binary.MarshalLong(math.MaxInt64 - 1)}, true},
		{PropQuery{"MaxInt64", OpNot, binary.MarshalLong(math.MaxInt64)}, false},

		{PropQuery{"MaxInt64", OpLessThan, binary.MarshalLong(math.MaxInt64 - 1)}, false},
		{PropQuery{"MaxInt64", OpLessThan, binary.MarshalLong(math.MaxInt64)}, false},

		{PropQuery{"MaxInt64", OpLessThanOrEqual, binary.MarshalLong(math.MaxInt64 - 1)}, false},
		{PropQuery{"MaxInt64", OpLessThanOrEqual, binary.MarshalLong(math.MaxInt64)}, true},

		{PropQuery{"MaxInt64", OpGreaterThan, binary.MarshalLong(math.MaxInt64 - 1)}, true},
		{PropQuery{"MaxInt64", OpGreaterThan, binary.MarshalLong(math.MaxInt64)}, false},

		{PropQuery{"MaxInt64", OpGreaterThanOrEqual, binary.MarshalLong(math.MaxInt64 - 1)}, true},
		{PropQuery{"MaxInt64", OpGreaterThanOrEqual, binary.MarshalLong(math.MaxInt64)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchStr8(t *testing.T) {
	props := binary.Dict{
		"":      binary.MarshalStr8(""),
		"abc":   binary.MarshalStr8("abc"),
		"あいうえお": binary.MarshalStr8("あいうえお"),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"", OpEqual, binary.MarshalStr8("")}, true},
		{PropQuery{"", OpNot, binary.MarshalStr8("")}, false},
		{PropQuery{"abc", OpEqual, binary.MarshalStr8("abc")}, true},
		{PropQuery{"abc", OpNot, binary.MarshalStr8("abc")}, false},
		{PropQuery{"あいうえお", OpEqual, binary.MarshalStr8("あいうえお")}, true},
		{PropQuery{"あいうえお", OpNot, binary.MarshalStr8("あいうえお")}, false},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueryMatchStr16(t *testing.T) {
	props := binary.Dict{
		"":      binary.MarshalStr16(""),
		"abc":   binary.MarshalStr16("abc"),
		"あいうえお": binary.MarshalStr16("あいうえお"),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"", OpEqual, binary.MarshalStr16("")}, true},
		{PropQuery{"", OpNot, binary.MarshalStr16("")}, false},
		{PropQuery{"abc", OpEqual, binary.MarshalStr16("abc")}, true},
		{PropQuery{"abc", OpNot, binary.MarshalStr16("abc")}, false},
		{PropQuery{"あいうえお", OpEqual, binary.MarshalStr16("あいうえお")}, true},
		{PropQuery{"あいうえお", OpNot, binary.MarshalStr16("あいうえお")}, false},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}

func TestPropQueriesMatch(t *testing.T) {
	props := binary.Dict{
		"0":   binary.MarshalInt(0),
		"abc": binary.MarshalStr16("abc"),
	}
	tests := []struct {
		queries  PropQueries
		expected bool
	}{
		{PropQueries{{"0", OpEqual, binary.MarshalInt(0)}, {"abc", OpEqual, binary.MarshalStr16("abc")}}, true},
		{PropQueries{{"0", OpEqual, binary.MarshalInt(1)}, {"abc", OpEqual, binary.MarshalStr16("abc")}}, false},
		{PropQueries{{"0", OpEqual, binary.MarshalInt(0)}, {"abc", OpEqual, binary.MarshalStr16("def")}}, false},
		{PropQueries{{"0", OpNot, binary.MarshalInt(1)}, {"abc", OpNot, binary.MarshalStr16("def")}}, true},
	}
	for _, test := range tests {
		if actual := test.queries.match(props); actual != test.expected {
			t.Fatalf("mismatch: %v %v, actual=%v, expected=%v", props, test, actual, test.expected)
		}
	}
}

func TestPropQueryMsgpack(t *testing.T) {
	// mapでもarrayでもデコードできることを確認したかった（できた）
	body, err := msgpack.Marshal(
		[]interface{}{
			[]interface{}{
				[]interface{}{"key1", byte(OpEqual), []byte{byte(binary.TypeTrue)}},
			},
			[]interface{}{
				// []interface{}{"key2", byte(OpNot), []byte{byte(binary.TypeByte), 0}},
				map[string]interface{}{"Key": "key2", "Op": byte(OpNot), "Val": []byte{byte(binary.TypeByte), 0}},
			},
		})
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var actual [][]PropQuery
	err = msgpack.NewDecoder(bytes.NewReader(body)).UseJSONTag(true).Decode(&actual)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	expect := [][]PropQuery{
		{{"key1", OpEqual, []byte{byte(binary.TypeTrue)}}},
		{{"key2", OpNot, []byte{byte(binary.TypeByte), 0}}},
	}

	if diff := cmp.Diff(actual, expect); diff != "" {
		t.Fatalf("Unmarshal (-got +want)\n%s", diff)
	}
}
