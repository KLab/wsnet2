package binary

import (
	"math"
	"reflect"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMarshalNull(t *testing.T) {
	exp := []byte{byte(TypeNull)}

	b := MarshalNull()
	if diff := cmp.Diff(b, exp); diff != "" {
		t.Fatalf("MarshalNull: (-got +want)\n%s", diff)
	}
	r, l, e := Unmarshal(b)
	if e != nil {
		t.Fatalf("Unmarshal error: %v", e)
	}
	if r != nil || l != 1 {
		t.Fatalf("Unmarshal = %v (len=%v) wants %v (len=%v)", r, l, nil, 1)
	}
}

func TestMarshalBool(t *testing.T) {
	tests := []struct {
		val bool
		buf []byte
	}{
		{true, []byte{byte(TypeTrue)}},
		{false, []byte{byte(TypeFalse)}},
	}
	for _, test := range tests {
		b := MarshalBool(test.val)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalBool:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		v, ok := r.(bool)
		if !ok {
			t.Fatalf("Unmarshal type mismatch: %T wants %T", r, v)
		}
		if v != test.val || l != len(test.buf) {
			t.Fatalf("Unmarshal = %v (len=%v) wants %v (len=%v)", v, l, test.val, len(test.buf))
		}
	}
}

func TestMarshalInteger(t *testing.T) {
	tests := []struct {
		marshal func(int) []byte
		in      int
		out     int
		buf     []byte
	}{
		{MarshalByte, -1, 0, []byte{byte(TypeByte), 0x00}},
		{MarshalByte, 0, 0, []byte{byte(TypeByte), 0x00}},
		{MarshalByte, 255, 255, []byte{byte(TypeByte), 0xff}},
		{MarshalByte, 256, 255, []byte{byte(TypeByte), 0xff}},

		{MarshalSByte, -129, -128, []byte{byte(TypeSByte), 0x00}},
		{MarshalSByte, -128, -128, []byte{byte(TypeSByte), 0x00}},
		{MarshalSByte, 0, 0, []byte{byte(TypeSByte), 0x80}},
		{MarshalSByte, 127, 127, []byte{byte(TypeSByte), 0xff}},
		{MarshalSByte, 128, 127, []byte{byte(TypeSByte), 0xff}},

		{MarshalUShort, -1, 0, []byte{byte(TypeUShort), 0x00, 0x00}},
		{MarshalUShort, 0, 0, []byte{byte(TypeUShort), 0x00, 0x00}},
		{MarshalUShort, 0x0102, 0x0102, []byte{byte(TypeUShort), 0x01, 0x02}},
		{MarshalUShort, 0xffff, 0xffff, []byte{byte(TypeUShort), 0xff, 0xff}},
		{MarshalUShort, 0x10000, 0xffff, []byte{byte(TypeUShort), 0xff, 0xff}},

		{MarshalShort, -0x8001, -0x8000, []byte{byte(TypeShort), 0x00, 0x00}},
		{MarshalShort, -0x8000, -0x8000, []byte{byte(TypeShort), 0x00, 0x00}},
		{MarshalShort, 0, 0, []byte{byte(TypeShort), 0x80, 0x00}},
		{MarshalShort, 1, 1, []byte{byte(TypeShort), 0x80, 0x01}},
		{MarshalShort, 0x7fff, 0x7fff, []byte{byte(TypeShort), 0xff, 0xff}},
		{MarshalShort, 0x8000, 0x7fff, []byte{byte(TypeShort), 0xff, 0xff}},

		{MarshalUInt, -1, 0, []byte{byte(TypeUInt), 0x00, 0x00, 0x00, 0x00}},
		{MarshalUInt, 0, 0, []byte{byte(TypeUInt), 0x00, 0x00, 0x00, 0x00}},
		{MarshalUInt, 0x01020304, 0x01020304, []byte{byte(TypeUInt), 0x01, 0x02, 0x03, 0x04}},
		{MarshalUInt, 0xffffffff, 0xffffffff, []byte{byte(TypeUInt), 0xff, 0xff, 0xff, 0xff}},
		{MarshalUInt, 0x100000000, 0xffffffff, []byte{byte(TypeUInt), 0xff, 0xff, 0xff, 0xff}},

		{MarshalInt, -0x80000001, -0x80000000, []byte{byte(TypeInt), 0x00, 0x00, 0x00, 0x00}},
		{MarshalInt, -0x80000000, -0x80000000, []byte{byte(TypeInt), 0x00, 0x00, 0x00, 0x00}},
		{MarshalInt, 0x00000000, 0x00000000, []byte{byte(TypeInt), 0x80, 0x00, 0x00, 0x00}},
		{MarshalInt, 0x01020304, 0x01020304, []byte{byte(TypeInt), 0x81, 0x02, 0x03, 0x04}},
		{MarshalInt, 0x7fffffff, 0x7fffffff, []byte{byte(TypeInt), 0xff, 0xff, 0xff, 0xff}},
		{MarshalInt, 0x80000000, 0x7fffffff, []byte{byte(TypeInt), 0xff, 0xff, 0xff, 0xff}},

		{MarshalLong, math.MinInt64, math.MinInt64,
			[]byte{byte(TypeLong), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{MarshalLong, -1, -1,
			[]byte{byte(TypeLong), 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{MarshalLong, 0, 0,
			[]byte{byte(TypeLong), 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{MarshalLong, 1, 1,
			[]byte{byte(TypeLong), 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{MarshalLong, math.MaxInt64, math.MaxInt64,
			[]byte{byte(TypeLong), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{MarshalLong, 0x0102030405060708, 0x0102030405060708,
			[]byte{byte(TypeLong), 0x81, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}},
	}
	for _, test := range tests {
		b := test.marshal(test.in)
		if !reflect.DeepEqual(b, test.buf) {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%x):\n%#v\n%#v", fname, test.in, b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%x): Unmarshal error: %v", fname, test.in, e)
		}
		if r != test.out || l != len(test.buf) {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%x): Unmarshal = %v (len=%v) wants %v (len=%v)",
				fname, test.in, r, l, test.out, len(test.buf))
		}
	}
}

func TestMarshalUInt64(t *testing.T) {
	tests := []struct {
		val uint64
		buf []byte
	}{
		{0, []byte{byte(TypeULong), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}
	for _, test := range tests {
		b := MarshalULong(test.val)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalULong:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if r != test.val || l != len(test.buf) {
			t.Fatalf("Unmarshal = %v (len=%v) wants %v (len=%v)", r, l, test.val, len(test.buf))
		}
	}
}
