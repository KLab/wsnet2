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
		{0x0102030405060708,
			[]byte{byte(TypeULong), 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}},
		{math.MaxUint64,
			[]byte{byte(TypeULong), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
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

func TestMarshalStr8(t *testing.T) {
	s := "0123456789abcdef0123456789abcdef" // len=32
	s = s + s + s + s + s + s + s + s       // len=256

	tests := []struct {
		val string
		buf []byte
	}{
		{"", []byte{byte(TypeStr8), 0}},
		{"abc", []byte{byte(TypeStr8), 3, 'a', 'b', 'c'}},
		{"あいうえお", append([]byte{byte(TypeStr8), 3 * 5}, []byte("あいうえお")...)},
		{s, append([]byte{byte(TypeStr8), 255}, []byte(s[:255])...)},
	}
	for _, test := range tests {
		b := MarshalStr8(test.val)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalStr8:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		exp := []byte(test.val)
		if len(exp) > 255 {
			exp = exp[:255]
		}
		if r != string(exp) || l != len(test.buf) {
			t.Fatalf("Unmarshal = %v (len=%v) wants %v (len=%v)", r, l, string(exp), len(test.buf))
		}
	}
}

func TestMarshalStr16(t *testing.T) {
	s := "0123456789abcdef0123456789abcdef" // len=32
	s256 := s + s + s + s + s + s + s + s   // len=256
	s65536 := s256
	for len(s65536) < 65536 {
		s65536 += s65536
	}

	tests := []struct {
		val string
		buf []byte
	}{
		{"", []byte{byte(TypeStr16), 0, 0}},
		{"abc", []byte{byte(TypeStr16), 0, 3, 'a', 'b', 'c'}},
		{"あいうえお", append([]byte{byte(TypeStr16), 0, 3 * 5}, []byte("あいうえお")...)},
		{s256, append([]byte{byte(TypeStr16), 0x01, 0x00}, []byte(s256)...)},
		{s65536, append([]byte{byte(TypeStr16), 0xff, 0xff}, []byte(s65536[:65535])...)},
	}
	for _, test := range tests {
		b := MarshalStr16(test.val)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalStr16:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		exp := []byte(test.val)
		if len(exp) > math.MaxUint16 {
			exp = exp[:math.MaxUint16]
		}
		if r != string(exp) || l != len(test.buf) {
			t.Fatalf("Unmarshal = %v (len=%v) wants %v (len=%v)", r, l, string(exp), len(test.buf))
		}
	}
}

func TestMarshalObj(t *testing.T) {
	obj := &Obj{1, []byte{1, 2, 3, 4, 5}}
	buf := []byte{byte(TypeObj), 1, 0, 5, 1, 2, 3, 4, 5}

	b := MarshalObj(obj)
	if !reflect.DeepEqual(b, buf) {
		t.Fatalf("MarshalObj:\n%#v\n%#v", b, buf)
	}
	r, l, e := Unmarshal(b)
	if e != nil {
		t.Fatalf("Unmarshal error: %v", e)
	}
	if diff := cmp.Diff(r, obj); diff != "" {
		t.Fatalf("Unmarshal (-got +want)\n%s", diff)
	}
	if l != len(buf) {
		t.Fatalf("Unmarshal length = %v, wants %v", l, len(buf))
	}
}

func TestMarshalList(t *testing.T) {
	list := List{
		[]byte{byte(TypeStr8), 3, 'a', 'b', 'c'},
		[]byte{byte(TypeNull)},
		[]byte{byte(TypeByte), 1},
	}
	buf := []byte{byte(TypeList), 3,
		0, 5, byte(TypeStr8), 3, 'a', 'b', 'c',
		0, 1, byte(TypeNull),
		0, 2, byte(TypeByte), 1,
	}

	b := MarshalList(list)
	if !reflect.DeepEqual(b, buf) {
		t.Fatalf("MarshalList:\n%#v\n%#v", b, buf)
	}
	r, l, e := Unmarshal(b)
	if e != nil {
		t.Fatalf("Unmarshal error: %v", e)
	}
	if diff := cmp.Diff(r, list); diff != "" {
		t.Fatalf("Unmarshal (-got +want)\n%s", diff)
	}
	if l != len(buf) {
		t.Fatalf("Unmarshal length = %v, wants %v", l, len(buf))
	}
}

func TestMarshalDict(t *testing.T) {
	dict := Dict{
		"abc":   []byte{byte(TypeStr8), 3, 'a', 'b', 'c'},
		"null":  []byte{byte(TypeNull)},
		"byte1": []byte{byte(TypeByte), 1},
	}
	buf := []byte{byte(TypeDict), 3,
		3, 'a', 'b', 'c', 0, 5, byte(TypeStr8), 3, 'a', 'b', 'c',
		4, 'n', 'u', 'l', 'l', 0, 1, byte(TypeNull),
		5, 'b', 'y', 't', 'e', '1', 0, 2, byte(TypeByte), 1,
	}

	b := MarshalDict(dict)
	if !reflect.DeepEqual(b, buf) {
		t.Fatalf("MarshalDict:\n%#v\n%#v", b, buf)
	}
	r, l, e := Unmarshal(b)
	if e != nil {
		t.Fatalf("Unmarshal error: %v", e)
	}
	if diff := cmp.Diff(r, dict); diff != "" {
		t.Fatalf("Unmarshal (-got +want)\n%s", diff)
	}
	if l != len(buf) {
		t.Fatalf("Unmarshal length = %v, wants %v", l, len(buf))
	}

}
