package binary

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNull(t *testing.T) {
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

func TestBool(t *testing.T) {
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

func TestInteger(t *testing.T) {
	tests := []struct {
		marshal func(int) []byte
		in      int
		out     int
		buf     []byte
	}{
		{MarshalByte, 0, 0, []byte{byte(TypeByte), 0x00}},
		{MarshalByte, 255, 255, []byte{byte(TypeByte), 0xff}},
		{MarshalByte, -1, 0, []byte{byte(TypeByte), 0x00}},
		{MarshalByte, 256, 255, []byte{byte(TypeByte), 0xff}},

		{MarshalSByte, -129, -128, []byte{byte(TypeSByte), 0x00}},
		{MarshalSByte, -128, -128, []byte{byte(TypeSByte), 0x00}},
		{MarshalSByte, 0, 0, []byte{byte(TypeSByte), 0x80}},
		{MarshalSByte, 127, 127, []byte{byte(TypeSByte), 0xff}},
		{MarshalSByte, 128, 127, []byte{byte(TypeSByte), 0xff}},
	}
	for _, test := range tests {
		b := test.marshal(test.in)
		if !reflect.DeepEqual(b, test.buf) {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%v):\n%#v\n%#v", fname, test.in, b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%v): Unmarshal error: %v", fname, test.in, e)
		}
		if r != test.out || l != len(test.buf) {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%v): Unmarshal = %v (len=%v) wants %v (len=%v)",
				fname, test.in, r, l, test.out, len(test.buf))
		}
	}
}
