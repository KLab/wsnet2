package binary_test

import (
	"testing"

	"wsnet2/binary"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalRecursive(t *testing.T) {

	tests := []struct {
		data  []byte
		wants interface{}
	}{
		{
			[]byte{byte(binary.TypeInt), 0x80, 0x00, 0x00, 0x01},
			int(1),
		},
		{
			[]byte{byte(binary.TypeUInt), 0x01, 0x02, 0x03, 0x04, byte(binary.TypeStr8), 0x03, 'a', 'b', 'c'},
			[]interface{}{int(0x1020304), "abc"},
		},
		{
			[]byte{
				byte(binary.TypeList), 2,
				0, 1, byte(binary.TypeNull),
				0, 5, byte(binary.TypeBools), 0, 11, 0xa5, 0x60,
			},
			[]interface{}{nil, []bool{true, false, true, false, false, true, false, true, false, true, true}},
		},
		{
			[]byte{byte(binary.TypeObj), 0, 0, 0},
			binary.RawObj{0, []interface{}{}},
		},
		{
			[]byte{byte(binary.TypeObj), 1, 0, 1, byte(binary.TypeTrue)},
			binary.RawObj{1, []interface{}{true}},
		},
		{
			[]byte{
				byte(binary.TypeObj), 2, 0, 28,
				byte(binary.TypeFalse),
				byte(binary.TypeUShort), 0x80, 0x01,
				byte(binary.TypeDict), 2,
				3, 'a', 'b', 'c', 0, 1, byte(binary.TypeTrue),
				3, 'd', 'e', 'f', 0, 9, byte(binary.TypeULong), 0x80, 0x70, 0x60, 0x50, 0x40, 0x30, 0x20, 0x10,
				byte(binary.TypeObj), 1, 0, 1, byte(binary.TypeFalse),
			},
			[]interface{}{
				binary.RawObj{2, []interface{}{
					false,
					0x8001,
					map[string]interface{}{"abc": true, "def": uint64(0x8070605040302010)},
				}},
				binary.RawObj{1, []interface{}{false}},
			},
		},
	}

	for i, test := range tests {
		out, err := binary.UnmarshalRecursive(test.data)
		if err != nil {
			t.Fatalf("%v: %v", i, err)
		}

		if diff := cmp.Diff(out, test.wants); diff != "" {
			t.Fatalf("%v: UnmarshalRecursive: (-got +want)\n%s", i, diff)
		}
	}
}
