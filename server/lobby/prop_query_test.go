package lobby

import (
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
		{PropQuery{"true", "!", binary.MarshalBool(true)}, false},
		{PropQuery{"true", "=", binary.MarshalBool(false)}, false},
		{PropQuery{"true", "!", binary.MarshalBool(false)}, true},
		{PropQuery{"false", "=", binary.MarshalBool(true)}, false},
		{PropQuery{"false", "!", binary.MarshalBool(true)}, true},
		{PropQuery{"false", "=", binary.MarshalBool(false)}, true},
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
		"0":  binary.MarshalByte(0),
		"255": binary.MarshalByte(255),
	}
	tests := []struct {
		query    PropQuery
		expected bool
	}{
		{PropQuery{"0", "=", binary.MarshalByte(-1)}, true},
		{PropQuery{"0", "=", binary.MarshalByte(0)}, true},
		{PropQuery{"0", "=", binary.MarshalByte(1)}, false},

		{PropQuery{"0", "!", binary.MarshalByte(-1)}, false},
		{PropQuery{"0", "!", binary.MarshalByte(0)}, false},
		{PropQuery{"0", "!", binary.MarshalByte(1)}, true},

		{PropQuery{"0", "<", binary.MarshalByte(-1)}, false},
		{PropQuery{"0", "<", binary.MarshalByte(0)}, false},
		{PropQuery{"0", "<", binary.MarshalByte(1)}, true},

		{PropQuery{"0", "<=", binary.MarshalByte(-1)}, true},
		{PropQuery{"0", "<=", binary.MarshalByte(0)}, true},
		{PropQuery{"0", "<=", binary.MarshalByte(1)}, true},

		{PropQuery{"0", ">", binary.MarshalByte(-1)}, false},
		{PropQuery{"0", ">", binary.MarshalByte(0)}, false},
		{PropQuery{"0", ">", binary.MarshalByte(1)}, false},

		{PropQuery{"0", ">=", binary.MarshalByte(-1)}, true},
		{PropQuery{"0", ">=", binary.MarshalByte(0)}, true},
		{PropQuery{"0", ">=", binary.MarshalByte(1)}, false},

		{PropQuery{"255", "=", binary.MarshalByte(254)}, false},
		{PropQuery{"255", "=", binary.MarshalByte(255)}, true},
		{PropQuery{"255", "=", binary.MarshalByte(256)}, true},

		{PropQuery{"255", "!", binary.MarshalByte(254)}, true},
		{PropQuery{"255", "!", binary.MarshalByte(255)}, false},
		{PropQuery{"255", "!", binary.MarshalByte(256)}, false},

		{PropQuery{"255", "<", binary.MarshalByte(254)}, false},
		{PropQuery{"255", "<", binary.MarshalByte(255)}, false},
		{PropQuery{"255", "<", binary.MarshalByte(256)}, false},

		{PropQuery{"255", "<=", binary.MarshalByte(254)}, false},
		{PropQuery{"255", "<=", binary.MarshalByte(255)}, true},
		{PropQuery{"255", "<=", binary.MarshalByte(256)}, true},

		{PropQuery{"255", ">", binary.MarshalByte(254)}, true},
		{PropQuery{"255", ">", binary.MarshalByte(255)}, false},
		{PropQuery{"255", ">", binary.MarshalByte(256)}, false},

		{PropQuery{"255", ">=", binary.MarshalByte(254)}, true},
		{PropQuery{"255", ">=", binary.MarshalByte(255)}, true},
		{PropQuery{"255", ">=", binary.MarshalByte(256)}, true},
	}
	for _, test := range tests {
		if actual := test.query.match(props[test.query.Key]); actual != test.expected {
			t.Fatalf("mismatch: %v %v %v, actual=%v, expected=%v", props[test.query.Key], test.query.Op, test.query.Val, actual, test.expected)
		}
	}
}
