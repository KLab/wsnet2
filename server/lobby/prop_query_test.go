package lobby

import (
	"math"
	"testing"

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
		{PropQuery{"true", "=", binary.MarshalBool(true)}, true},
		{PropQuery{"true", "=", binary.MarshalBool(false)}, false},

		{PropQuery{"true", "!", binary.MarshalBool(true)}, false},
		{PropQuery{"true", "!", binary.MarshalBool(false)}, true},

		{PropQuery{"false", "=", binary.MarshalBool(true)}, false},
		{PropQuery{"false", "=", binary.MarshalBool(false)}, true},

		{PropQuery{"false", "!", binary.MarshalBool(true)}, true},
		{PropQuery{"false", "!", binary.MarshalBool(false)}, false},
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
		{PropQuery{"0", "=", binary.MarshalByte(0)}, true},
		{PropQuery{"0", "=", binary.MarshalByte(1)}, false},

		{PropQuery{"0", "!", binary.MarshalByte(0)}, false},
		{PropQuery{"0", "!", binary.MarshalByte(1)}, true},

		{PropQuery{"0", "<", binary.MarshalByte(0)}, false},
		{PropQuery{"0", "<", binary.MarshalByte(1)}, true},

		{PropQuery{"0", "<=", binary.MarshalByte(0)}, true},
		{PropQuery{"0", "<=", binary.MarshalByte(1)}, true},

		{PropQuery{"0", ">", binary.MarshalByte(0)}, false},
		{PropQuery{"0", ">", binary.MarshalByte(1)}, false},

		{PropQuery{"0", ">=", binary.MarshalByte(0)}, true},
		{PropQuery{"0", ">=", binary.MarshalByte(1)}, false},

		{PropQuery{"MaxUint8", "=", binary.MarshalByte(math.MaxUint8 - 1)}, false},
		{PropQuery{"MaxUint8", "=", binary.MarshalByte(math.MaxUint8)}, true},

		{PropQuery{"MaxUint8", "!", binary.MarshalByte(math.MaxUint8 - 1)}, true},
		{PropQuery{"MaxUint8", "!", binary.MarshalByte(math.MaxUint8)}, false},

		{PropQuery{"MaxUint8", "<", binary.MarshalByte(math.MaxUint8 - 1)}, false},
		{PropQuery{"MaxUint8", "<", binary.MarshalByte(math.MaxUint8)}, false},

		{PropQuery{"MaxUint8", "<=", binary.MarshalByte(math.MaxUint8 - 1)}, false},
		{PropQuery{"MaxUint8", "<=", binary.MarshalByte(math.MaxUint8)}, true},

		{PropQuery{"MaxUint8", ">", binary.MarshalByte(math.MaxUint8 - 1)}, true},
		{PropQuery{"MaxUint8", ">", binary.MarshalByte(math.MaxUint8)}, false},

		{PropQuery{"MaxUint8", ">=", binary.MarshalByte(math.MaxUint8 - 1)}, true},
		{PropQuery{"MaxUint8", ">=", binary.MarshalByte(math.MaxUint8)}, true},
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
		{PropQuery{"MinInt8", "=", binary.MarshalSByte(math.MinInt8)}, true},
		{PropQuery{"MinInt8", "=", binary.MarshalSByte(math.MinInt8 + 1)}, false},

		{PropQuery{"MinInt8", "!", binary.MarshalSByte(math.MinInt8)}, false},
		{PropQuery{"MinInt8", "!", binary.MarshalSByte(math.MinInt8 + 1)}, true},

		{PropQuery{"MinInt8", "<", binary.MarshalSByte(math.MinInt8)}, false},
		{PropQuery{"MinInt8", "<", binary.MarshalSByte(math.MinInt8 + 1)}, true},

		{PropQuery{"MinInt8", "<=", binary.MarshalSByte(math.MinInt8)}, true},
		{PropQuery{"MinInt8", "<=", binary.MarshalSByte(math.MinInt8 + 1)}, true},

		{PropQuery{"MinInt8", ">", binary.MarshalSByte(math.MinInt8)}, false},
		{PropQuery{"MinInt8", ">", binary.MarshalSByte(math.MinInt8 + 1)}, false},

		{PropQuery{"MinInt8", ">=", binary.MarshalSByte(math.MinInt8)}, true},
		{PropQuery{"MinInt8", ">=", binary.MarshalSByte(math.MinInt8 + 1)}, false},

		{PropQuery{"MaxInt8", "=", binary.MarshalSByte(math.MaxInt8 - 1)}, false},
		{PropQuery{"MaxInt8", "=", binary.MarshalSByte(math.MaxInt8)}, true},

		{PropQuery{"MaxInt8", "!", binary.MarshalSByte(math.MaxInt8 - 1)}, true},
		{PropQuery{"MaxInt8", "!", binary.MarshalSByte(math.MaxInt8)}, false},

		{PropQuery{"MaxInt8", "<", binary.MarshalSByte(math.MaxInt8 - 1)}, false},
		{PropQuery{"MaxInt8", "<", binary.MarshalSByte(math.MaxInt8)}, false},

		{PropQuery{"MaxInt8", "<=", binary.MarshalSByte(math.MaxInt8 - 1)}, false},
		{PropQuery{"MaxInt8", "<=", binary.MarshalSByte(math.MaxInt8)}, true},

		{PropQuery{"MaxInt8", ">", binary.MarshalSByte(math.MaxInt8 - 1)}, true},
		{PropQuery{"MaxInt8", ">", binary.MarshalSByte(math.MaxInt8)}, false},

		{PropQuery{"MaxInt8", ">=", binary.MarshalSByte(math.MaxInt8 - 1)}, true},
		{PropQuery{"MaxInt8", ">=", binary.MarshalSByte(math.MaxInt8)}, true},
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
		{PropQuery{"0", "=", binary.MarshalUShort(0)}, true},
		{PropQuery{"0", "=", binary.MarshalUShort(1)}, false},

		{PropQuery{"0", "!", binary.MarshalUShort(0)}, false},
		{PropQuery{"0", "!", binary.MarshalUShort(1)}, true},

		{PropQuery{"0", "<", binary.MarshalUShort(0)}, false},
		{PropQuery{"0", "<", binary.MarshalUShort(1)}, true},

		{PropQuery{"0", "<=", binary.MarshalUShort(0)}, true},
		{PropQuery{"0", "<=", binary.MarshalUShort(1)}, true},

		{PropQuery{"0", ">", binary.MarshalUShort(0)}, false},
		{PropQuery{"0", ">", binary.MarshalUShort(1)}, false},

		{PropQuery{"0", ">=", binary.MarshalUShort(0)}, true},
		{PropQuery{"0", ">=", binary.MarshalUShort(1)}, false},

		{PropQuery{"MaxUint16", "=", binary.MarshalUShort(math.MaxUint16 - 1)}, false},
		{PropQuery{"MaxUint16", "=", binary.MarshalUShort(math.MaxUint16)}, true},

		{PropQuery{"MaxUint16", "!", binary.MarshalUShort(math.MaxUint16 - 1)}, true},
		{PropQuery{"MaxUint16", "!", binary.MarshalUShort(math.MaxUint16)}, false},

		{PropQuery{"MaxUint16", "<", binary.MarshalUShort(math.MaxUint16 - 1)}, false},
		{PropQuery{"MaxUint16", "<", binary.MarshalUShort(math.MaxUint16)}, false},

		{PropQuery{"MaxUint16", "<=", binary.MarshalUShort(math.MaxUint16 - 1)}, false},
		{PropQuery{"MaxUint16", "<=", binary.MarshalUShort(math.MaxUint16)}, true},

		{PropQuery{"MaxUint16", ">", binary.MarshalUShort(math.MaxUint16 - 1)}, true},
		{PropQuery{"MaxUint16", ">", binary.MarshalUShort(math.MaxUint16)}, false},

		{PropQuery{"MaxUint16", ">=", binary.MarshalUShort(math.MaxUint16 - 1)}, true},
		{PropQuery{"MaxUint16", ">=", binary.MarshalUShort(math.MaxUint16)}, true},
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
		{PropQuery{"MinInt16", "=", binary.MarshalShort(math.MinInt16)}, true},
		{PropQuery{"MinInt16", "=", binary.MarshalShort(math.MinInt16 + 1)}, false},

		{PropQuery{"MinInt16", "!", binary.MarshalShort(math.MinInt16)}, false},
		{PropQuery{"MinInt16", "!", binary.MarshalShort(math.MinInt16 + 1)}, true},

		{PropQuery{"MinInt16", "<", binary.MarshalShort(math.MinInt16)}, false},
		{PropQuery{"MinInt16", "<", binary.MarshalShort(math.MinInt16 + 1)}, true},

		{PropQuery{"MinInt16", "<=", binary.MarshalShort(math.MinInt16)}, true},
		{PropQuery{"MinInt16", "<=", binary.MarshalShort(math.MinInt16 + 1)}, true},

		{PropQuery{"MinInt16", ">", binary.MarshalShort(math.MinInt16)}, false},
		{PropQuery{"MinInt16", ">", binary.MarshalShort(math.MinInt16 + 1)}, false},

		{PropQuery{"MinInt16", ">=", binary.MarshalShort(math.MinInt16)}, true},
		{PropQuery{"MinInt16", ">=", binary.MarshalShort(math.MinInt16 + 1)}, false},

		{PropQuery{"MaxInt16", "=", binary.MarshalShort(math.MaxInt16 - 1)}, false},
		{PropQuery{"MaxInt16", "=", binary.MarshalShort(math.MaxInt16)}, true},

		{PropQuery{"MaxInt16", "!", binary.MarshalShort(math.MaxInt16 - 1)}, true},
		{PropQuery{"MaxInt16", "!", binary.MarshalShort(math.MaxInt16)}, false},

		{PropQuery{"MaxInt16", "<", binary.MarshalShort(math.MaxInt16 - 1)}, false},
		{PropQuery{"MaxInt16", "<", binary.MarshalShort(math.MaxInt16)}, false},

		{PropQuery{"MaxInt16", "<=", binary.MarshalShort(math.MaxInt16 - 1)}, false},
		{PropQuery{"MaxInt16", "<=", binary.MarshalShort(math.MaxInt16)}, true},

		{PropQuery{"MaxInt16", ">", binary.MarshalShort(math.MaxInt16 - 1)}, true},
		{PropQuery{"MaxInt16", ">", binary.MarshalShort(math.MaxInt16)}, false},

		{PropQuery{"MaxInt16", ">=", binary.MarshalShort(math.MaxInt16 - 1)}, true},
		{PropQuery{"MaxInt16", ">=", binary.MarshalShort(math.MaxInt16)}, true},
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
		{PropQuery{"0", "=", binary.MarshalUInt(0)}, true},
		{PropQuery{"0", "=", binary.MarshalUInt(1)}, false},

		{PropQuery{"0", "!", binary.MarshalUInt(0)}, false},
		{PropQuery{"0", "!", binary.MarshalUInt(1)}, true},

		{PropQuery{"0", "<", binary.MarshalUInt(0)}, false},
		{PropQuery{"0", "<", binary.MarshalUInt(1)}, true},

		{PropQuery{"0", "<=", binary.MarshalUInt(0)}, true},
		{PropQuery{"0", "<=", binary.MarshalUInt(1)}, true},

		{PropQuery{"0", ">", binary.MarshalUInt(0)}, false},
		{PropQuery{"0", ">", binary.MarshalUInt(1)}, false},

		{PropQuery{"0", ">=", binary.MarshalUInt(0)}, true},
		{PropQuery{"0", ">=", binary.MarshalUInt(1)}, false},

		{PropQuery{"MaxUint32", "=", binary.MarshalUInt(math.MaxUint32 - 1)}, false},
		{PropQuery{"MaxUint32", "=", binary.MarshalUInt(math.MaxUint32)}, true},

		{PropQuery{"MaxUint32", "!", binary.MarshalUInt(math.MaxUint32 - 1)}, true},
		{PropQuery{"MaxUint32", "!", binary.MarshalUInt(math.MaxUint32)}, false},

		{PropQuery{"MaxUint32", "<", binary.MarshalUInt(math.MaxUint32 - 1)}, false},
		{PropQuery{"MaxUint32", "<", binary.MarshalUInt(math.MaxUint32)}, false},

		{PropQuery{"MaxUint32", "<=", binary.MarshalUInt(math.MaxUint32 - 1)}, false},
		{PropQuery{"MaxUint32", "<=", binary.MarshalUInt(math.MaxUint32)}, true},

		{PropQuery{"MaxUint32", ">", binary.MarshalUInt(math.MaxUint32 - 1)}, true},
		{PropQuery{"MaxUint32", ">", binary.MarshalUInt(math.MaxUint32)}, false},

		{PropQuery{"MaxUint32", ">=", binary.MarshalUInt(math.MaxUint32 - 1)}, true},
		{PropQuery{"MaxUint32", ">=", binary.MarshalUInt(math.MaxUint32)}, true},
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
		{PropQuery{"MinInt32", "=", binary.MarshalInt(math.MinInt32)}, true},
		{PropQuery{"MinInt32", "=", binary.MarshalInt(math.MinInt32 + 1)}, false},

		{PropQuery{"MinInt32", "!", binary.MarshalInt(math.MinInt32)}, false},
		{PropQuery{"MinInt32", "!", binary.MarshalInt(math.MinInt32 + 1)}, true},

		{PropQuery{"MinInt32", "<", binary.MarshalInt(math.MinInt32)}, false},
		{PropQuery{"MinInt32", "<", binary.MarshalInt(math.MinInt32 + 1)}, true},

		{PropQuery{"MinInt32", "<=", binary.MarshalInt(math.MinInt32)}, true},
		{PropQuery{"MinInt32", "<=", binary.MarshalInt(math.MinInt32 + 1)}, true},

		{PropQuery{"MinInt32", ">", binary.MarshalInt(math.MinInt32)}, false},
		{PropQuery{"MinInt32", ">", binary.MarshalInt(math.MinInt32 + 1)}, false},

		{PropQuery{"MinInt32", ">=", binary.MarshalInt(math.MinInt32)}, true},
		{PropQuery{"MinInt32", ">=", binary.MarshalInt(math.MinInt32 + 1)}, false},

		{PropQuery{"MaxInt32", "=", binary.MarshalInt(math.MaxInt32 - 1)}, false},
		{PropQuery{"MaxInt32", "=", binary.MarshalInt(math.MaxInt32)}, true},

		{PropQuery{"MaxInt32", "!", binary.MarshalInt(math.MaxInt32 - 1)}, true},
		{PropQuery{"MaxInt32", "!", binary.MarshalInt(math.MaxInt32)}, false},

		{PropQuery{"MaxInt32", "<", binary.MarshalInt(math.MaxInt32 - 1)}, false},
		{PropQuery{"MaxInt32", "<", binary.MarshalInt(math.MaxInt32)}, false},

		{PropQuery{"MaxInt32", "<=", binary.MarshalInt(math.MaxInt32 - 1)}, false},
		{PropQuery{"MaxInt32", "<=", binary.MarshalInt(math.MaxInt32)}, true},

		{PropQuery{"MaxInt32", ">", binary.MarshalInt(math.MaxInt32 - 1)}, true},
		{PropQuery{"MaxInt32", ">", binary.MarshalInt(math.MaxInt32)}, false},

		{PropQuery{"MaxInt32", ">=", binary.MarshalInt(math.MaxInt32 - 1)}, true},
		{PropQuery{"MaxInt32", ">=", binary.MarshalInt(math.MaxInt32)}, true},
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
		{PropQuery{"0", "=", binary.MarshalULong(0)}, true},
		{PropQuery{"0", "=", binary.MarshalULong(1)}, false},

		{PropQuery{"0", "!", binary.MarshalULong(0)}, false},
		{PropQuery{"0", "!", binary.MarshalULong(1)}, true},

		{PropQuery{"0", "<", binary.MarshalULong(0)}, false},
		{PropQuery{"0", "<", binary.MarshalULong(1)}, true},

		{PropQuery{"0", "<=", binary.MarshalULong(0)}, true},
		{PropQuery{"0", "<=", binary.MarshalULong(1)}, true},

		{PropQuery{"0", ">", binary.MarshalULong(0)}, false},
		{PropQuery{"0", ">", binary.MarshalULong(1)}, false},

		{PropQuery{"0", ">=", binary.MarshalULong(0)}, true},
		{PropQuery{"0", ">=", binary.MarshalULong(1)}, false},

		{PropQuery{"MaxUint64", "=", binary.MarshalULong(math.MaxUint64 - 1)}, false},
		{PropQuery{"MaxUint64", "=", binary.MarshalULong(math.MaxUint64)}, true},

		{PropQuery{"MaxUint64", "!", binary.MarshalULong(math.MaxUint64 - 1)}, true},
		{PropQuery{"MaxUint64", "!", binary.MarshalULong(math.MaxUint64)}, false},

		{PropQuery{"MaxUint64", "<", binary.MarshalULong(math.MaxUint64 - 1)}, false},
		{PropQuery{"MaxUint64", "<", binary.MarshalULong(math.MaxUint64)}, false},

		{PropQuery{"MaxUint64", "<=", binary.MarshalULong(math.MaxUint64 - 1)}, false},
		{PropQuery{"MaxUint64", "<=", binary.MarshalULong(math.MaxUint64)}, true},

		{PropQuery{"MaxUint64", ">", binary.MarshalULong(math.MaxUint64 - 1)}, true},
		{PropQuery{"MaxUint64", ">", binary.MarshalULong(math.MaxUint64)}, false},

		{PropQuery{"MaxUint64", ">=", binary.MarshalULong(math.MaxUint64 - 1)}, true},
		{PropQuery{"MaxUint64", ">=", binary.MarshalULong(math.MaxUint64)}, true},
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
		{PropQuery{"MinInt64", "=", binary.MarshalLong(math.MinInt64)}, true},
		{PropQuery{"MinInt64", "=", binary.MarshalLong(math.MinInt64 + 1)}, false},

		{PropQuery{"MinInt64", "!", binary.MarshalLong(math.MinInt64)}, false},
		{PropQuery{"MinInt64", "!", binary.MarshalLong(math.MinInt64 + 1)}, true},

		{PropQuery{"MinInt64", "<", binary.MarshalLong(math.MinInt64)}, false},
		{PropQuery{"MinInt64", "<", binary.MarshalLong(math.MinInt64 + 1)}, true},

		{PropQuery{"MinInt64", "<=", binary.MarshalLong(math.MinInt64)}, true},
		{PropQuery{"MinInt64", "<=", binary.MarshalLong(math.MinInt64 + 1)}, true},

		{PropQuery{"MinInt64", ">", binary.MarshalLong(math.MinInt64)}, false},
		{PropQuery{"MinInt64", ">", binary.MarshalLong(math.MinInt64 + 1)}, false},

		{PropQuery{"MinInt64", ">=", binary.MarshalLong(math.MinInt64)}, true},
		{PropQuery{"MinInt64", ">=", binary.MarshalLong(math.MinInt64 + 1)}, false},

		{PropQuery{"MaxInt64", "=", binary.MarshalLong(math.MaxInt64 - 1)}, false},
		{PropQuery{"MaxInt64", "=", binary.MarshalLong(math.MaxInt64)}, true},

		{PropQuery{"MaxInt64", "!", binary.MarshalLong(math.MaxInt64 - 1)}, true},
		{PropQuery{"MaxInt64", "!", binary.MarshalLong(math.MaxInt64)}, false},

		{PropQuery{"MaxInt64", "<", binary.MarshalLong(math.MaxInt64 - 1)}, false},
		{PropQuery{"MaxInt64", "<", binary.MarshalLong(math.MaxInt64)}, false},

		{PropQuery{"MaxInt64", "<=", binary.MarshalLong(math.MaxInt64 - 1)}, false},
		{PropQuery{"MaxInt64", "<=", binary.MarshalLong(math.MaxInt64)}, true},

		{PropQuery{"MaxInt64", ">", binary.MarshalLong(math.MaxInt64 - 1)}, true},
		{PropQuery{"MaxInt64", ">", binary.MarshalLong(math.MaxInt64)}, false},

		{PropQuery{"MaxInt64", ">=", binary.MarshalLong(math.MaxInt64 - 1)}, true},
		{PropQuery{"MaxInt64", ">=", binary.MarshalLong(math.MaxInt64)}, true},
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
		{PropQuery{"", "=", binary.MarshalStr8("")}, true},
		{PropQuery{"", "!", binary.MarshalStr8("")}, false},
		{PropQuery{"abc", "=", binary.MarshalStr8("abc")}, true},
		{PropQuery{"abc", "!", binary.MarshalStr8("abc")}, false},
		{PropQuery{"あいうえお", "=", binary.MarshalStr8("あいうえお")}, true},
		{PropQuery{"あいうえお", "!", binary.MarshalStr8("あいうえお")}, false},
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
		{PropQuery{"", "=", binary.MarshalStr16("")}, true},
		{PropQuery{"", "!", binary.MarshalStr16("")}, false},
		{PropQuery{"abc", "=", binary.MarshalStr16("abc")}, true},
		{PropQuery{"abc", "!", binary.MarshalStr16("abc")}, false},
		{PropQuery{"あいうえお", "=", binary.MarshalStr16("あいうえお")}, true},
		{PropQuery{"あいうえお", "!", binary.MarshalStr16("あいうえお")}, false},
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
		{PropQueries{{"0", "=", binary.MarshalInt(0)}, {"abc", "=", binary.MarshalStr16("abc")}}, true},
		{PropQueries{{"0", "=", binary.MarshalInt(1)}, {"abc", "=", binary.MarshalStr16("abc")}}, false},
		{PropQueries{{"0", "=", binary.MarshalInt(0)}, {"abc", "=", binary.MarshalStr16("def")}}, false},
		{PropQueries{{"0", "!", binary.MarshalInt(1)}, {"abc", "!", binary.MarshalStr16("def")}}, true},
	}
	for _, test := range tests {
		if actual := test.queries.match(props); actual != test.expected {
			t.Fatalf("mismatch: %v %v, actual=%v, expected=%v", props, test, actual, test.expected)
		}
	}
}
