package binary

import (
	"bytes"
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

func TestMarshalChar(t *testing.T) {
	tests := []struct {
		val rune
		buf []byte
	}{
		{0, []byte{byte(TypeChar), 0x00, 0x00}},
		{'j', []byte{byte(TypeChar), 0x00, 0x6a}},
		{'あ', []byte{byte(TypeChar), 0x30, 0x42}},
	}
	for _, test := range tests {
		b := MarshalChar(test.val)
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

func TestMarshalLong(t *testing.T) {
	tests := []struct {
		in  int64
		out int64
		buf []byte
	}{
		{math.MinInt64, math.MinInt64,
			[]byte{byte(TypeLong), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{-1, -1,
			[]byte{byte(TypeLong), 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{0, 0,
			[]byte{byte(TypeLong), 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{1, 1,
			[]byte{byte(TypeLong), 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{math.MaxInt64, math.MaxInt64,
			[]byte{byte(TypeLong), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{0x0102030405060708, 0x0102030405060708,
			[]byte{byte(TypeLong), 0x81, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}},
	}
	for _, test := range tests {
		b := MarshalLong(test.in)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalLong(%x):\n%#v\n%#v", test.in, b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("MarshalLong(%x): Unmarshal error: %v", test.in, e)
		}
		if r != test.out || l != len(test.buf) {
			t.Fatalf("MarshalLong(%x): Unmarshal = %v (len=%v) wants %v (len=%v)",
				test.in, r, l, test.out, len(test.buf))
		}
	}
}

func TestMarshalULong(t *testing.T) {
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

func TestMarshalFloat(t *testing.T) {
	var cmpf float32
	var cmpb []byte

	tests := []struct {
		val float32
		buf []byte
	}{ // sorted by val.
		{ // -Inf (sign=1, exp=ff, frac=000000) 0xff800000
			float32(math.Inf(-1)),
			[]byte{byte(TypeFloat), 0x00, 0x7f, 0xff, 0xff},
		},
		{ // min (sign=1, exp=fe, frac=7fffff) 0xff7fffff
			-math.MaxFloat32,
			[]byte{byte(TypeFloat), 0x00, 0x80, 0x00, 0x00},
		},
		{ // negative normal border (sign=1, exp=1, frac=000000) 0x80800000
			math.Float32frombits(0x80800000),
			[]byte{byte(TypeFloat), 0x7f, 0x7f, 0xff, 0xff},
		},
		{ // subnormal border (sign=1, exp=0, frac=7fffff) 0x807fffff
			math.Float32frombits(0x807fffff),
			[]byte{byte(TypeFloat), 0x7f, 0x80, 0x00, 0x00},
		},
		{ // subnormal border (sign=1, exp=0, frac=000001) 0x80000001
			math.Float32frombits(0x80000001),
			[]byte{byte(TypeFloat), 0x7f, 0xff, 0xff, 0xfe},
		},
		{ // -0 (sign=1, exp=0, frac=000000) 0x80000000
			float32(math.Copysign(0, -1)),
			[]byte{byte(TypeFloat), 0x7f, 0xff, 0xff, 0xff},
		},
		{ // +0 (sign=0, exp=0, frac=000000) 0x00000000
			0,
			[]byte{byte(TypeFloat), 0x80, 0x00, 0x00, 0x00},
		},
		{ // subnormal border (sign=0, exp=0, frac=000001) 0x00000001
			math.Float32frombits(0x00000001),
			[]byte{byte(TypeFloat), 0x80, 0x00, 0x00, 0x01},
		},
		{ // subnormal border (sign=0, exp=0, frac=7fffff 0x007fffff
			math.Float32frombits(0x007fffff),
			[]byte{byte(TypeFloat), 0x80, 0x7f, 0xff, 0xff},
		},
		{ // normal border (sign=0, exp=1, frac=000000) 0x00800000
			math.Float32frombits(0x00800000),
			[]byte{byte(TypeFloat), 0x80, 0x80, 0x00, 0x00},
		},
		{ // max (sign=0, exp=fe, frac=7fffff) 0x7f7fffff
			math.MaxFloat32,
			[]byte{byte(TypeFloat), 0xff, 0x7f, 0xff, 0xff},
		},
		{ // +Inf (sign=0, exp=ff, frac=000000) 0x7f800000
			float32(math.Inf(1)),
			[]byte{byte(TypeFloat), 0xff, 0x80, 0x00, 0x00},
		},
	}
	for _, test := range tests {
		b := MarshalFloat(test.val)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalFloat(%v):\n%#v\n%#v", test.val, b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if r != test.val || l != len(test.buf) {
			t.Fatalf("Unmarshal = %v (len=%v) wants %v (len=%v)", r, l, test.val, len(test.buf))
		}
		if cmpb != nil {
			if bytes.Compare(cmpb, b) >= 0 {
				t.Fatalf("compare %f (%x) >= %f (%x)", cmpf, cmpb, test.val, b)
			}
		}
		cmpf = test.val
		cmpb = b
	}
}

func TestMarshalDouble(t *testing.T) {
	var cmpf float64
	var cmpb []byte

	tests := []struct {
		val float64
		buf []byte
	}{ // sorted by val.
		{ // -Inf (sign=1, exp=7ff, frac=0 0000 0000 0000) 0xfff0 0000 0000 0000
			math.Inf(-1),
			[]byte{byte(TypeDouble), 0x00, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{ // min (sign=1, exp=7fe, frac=f ffff ffff ffff) 0xffef ffff ffff ffff
			-math.MaxFloat64,
			[]byte{byte(TypeDouble), 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{ // negative normal border (sign=1, exp=1, frac=0 0000 0000 0000) 0x8010 0000 0000 0000
			math.Float64frombits(0x8010000000000000),
			[]byte{byte(TypeDouble), 0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{ // subnormal border (sign=1, exp=0, frac=f ffff ffff ffff) 0x800f ffff ffff ffff
			math.Float64frombits(0x800fffffffffffff),
			[]byte{byte(TypeDouble), 0x7f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{ // subnormal border (sign=1, exp=0, frac=0 0000 0000 0001) 0x8000 0000 0000 0001
			math.Float64frombits(0x8000000000000001),
			[]byte{byte(TypeDouble), 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
		},
		{ // -0 (sign=1, exp=0, frac=0 0000 0000 0000) 0x8000 0000 0000 0000
			math.Copysign(0, -1),
			[]byte{byte(TypeDouble), 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{ // +0 (sign=0, exp=0, frac=0 0000 0000 0000) 0x0000 0000 0000 0000
			0,
			[]byte{byte(TypeDouble), 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{ // subnormal border (sign=0, exp=0, frac=0 0000 0000 0001) 0x0000 0000 0000 0001
			math.Float64frombits(0x0000000000000001),
			[]byte{byte(TypeDouble), 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		},
		{ // subnormal border (sign=0, exp=0, frac=f ffff ffff ffff) 0x000f ffff ffff ffff
			math.Float64frombits(0x000fffffffffffff),
			[]byte{byte(TypeDouble), 0x80, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{ // normal border (sign=0, exp=1, frac=0 0000 0000 0000) 0x0010 0000 0000 0000
			math.Float64frombits(0x0010000000000000),
			[]byte{byte(TypeDouble), 0x80, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{ // max (sign=0, exp=7fe, frac=f ffff ffff ffff) 0x7fef ffff ffff ffff
			math.MaxFloat64,
			[]byte{byte(TypeDouble), 0xff, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{ // +Inf (sign=0, exp=7ff, frac=0 0000 0000 0000) 0x7ff0 0000 0000 0000
			math.Inf(1),
			[]byte{byte(TypeDouble), 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}
	for _, test := range tests {
		b := MarshalDouble(test.val)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalDouble(%v):\n%#v\n%#v", test.val, b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if r != test.val || l != len(test.buf) {
			t.Fatalf("Unmarshal = %v (len=%v) wants %v (len=%v)", r, l, test.val, len(test.buf))
		}
		if cmpb != nil {
			if bytes.Compare(cmpb, b) >= 0 {
				t.Fatalf("compare %f (%x) >= %f (%x)", cmpf, cmpb, test.val, b)
			}
		}
		cmpf = test.val
		cmpb = b
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
	tests := []struct {
		obj *Obj
		buf []byte
	}{
		{
			obj: &Obj{1, []byte{1, 2, 3, 4, 5}},
			buf: []byte{byte(TypeObj), 1, 0, 5, 1, 2, 3, 4, 5},
		},
		{obj: nil, buf: []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := MarshalObj(test.obj)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalObj:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.obj == nil && r == nil) {
			if diff := cmp.Diff(r, test.obj); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalList(t *testing.T) {
	tests := []struct {
		list List
		buf  []byte
	}{
		{
			list: List{
				[]byte{byte(TypeStr8), 3, 'a', 'b', 'c'},
				[]byte{byte(TypeNull)},
				[]byte{byte(TypeByte), 1},
			},
			buf: []byte{byte(TypeList), 3,
				0, 5, byte(TypeStr8), 3, 'a', 'b', 'c',
				0, 1, byte(TypeNull),
				0, 2, byte(TypeByte), 1,
			},
		},
		{list: nil, buf: []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := MarshalList(test.list)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalList:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.list == nil && r == nil) {
			if diff := cmp.Diff(r, test.list); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalDict(t *testing.T) {
	tests := []struct {
		dict Dict
		buf  []byte
	}{
		{
			dict: Dict{
				"int1": []byte{byte(TypeInt), 0x80, 0x00, 0x00, 0x01},
			},
			buf: []byte{byte(TypeDict), 1,
				4, 'i', 'n', 't', '1', 0, 5, byte(TypeInt), 0x80, 0x00, 0x00, 0x01,
			},
		},
		{
			dict: nil,
			buf:  []byte{byte(TypeNull)},
		},
	}
	for _, test := range tests {
		b := MarshalDict(test.dict)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalDict:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.dict == nil && r == nil) {
			if diff := cmp.Diff(r, test.dict); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalBools(t *testing.T) {
	tests := []struct {
		val []bool
		buf []byte
	}{
		{[]bool{}, []byte{byte(TypeBools), 0, 0}},
		{
			[]bool{true, false, true},
			[]byte{byte(TypeBools), 0, 3, 0b10100000},
		},
		{
			[]bool{false, false, true, false, true, true, false, true},
			[]byte{byte(TypeBools), 0, 8, 0b00101101},
		},
		{
			[]bool{true, true, false, true, false, false, true, false, true},
			[]byte{byte(TypeBools), 0, 9, 0b11010010, 0b10000000},
		},
		{nil, []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := MarshalBools(test.val)
		if !reflect.DeepEqual(b, test.buf) {
			t.Fatalf("MarshalBool:\n%#v\n%#v", b, test.buf)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.val == nil && r == nil) {
			if diff := cmp.Diff(r, test.val); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalIntegers(t *testing.T) {
	tests := []struct {
		marshal func([]int) []byte
		in      []int
		out     []int
		buf     []byte
	}{
		{MarshalSBytes, []int{}, []int{}, []byte{byte(TypeSBytes), 0, 0}},
		{MarshalSBytes,
			[]int{0, 1, -128, -129, 127, 128},
			[]int{0, 1, -128, -128, 127, 127},
			[]byte{byte(TypeSBytes), 0, 6, 0x80, 0x81, 0x00, 0x00, 0xff, 0xff},
		},

		{MarshalBytes, []int{}, []int{}, []byte{byte(TypeBytes), 0, 0}},
		{MarshalBytes,
			[]int{0, 1, -1, 128, 255, 256},
			[]int{0, 1, 0, 128, 255, 255},
			[]byte{byte(TypeBytes), 0, 6, 0x00, 0x01, 0x00, 0x80, 0xff, 0xff},
		},

		{MarshalShorts, []int{}, []int{}, []byte{byte(TypeShorts), 0, 0}},
		{MarshalShorts,
			[]int{0, 1, math.MinInt16 - 1, math.MinInt16, math.MaxInt16, math.MaxInt16 + 1},
			[]int{0, 1, math.MinInt16, math.MinInt16, math.MaxInt16, math.MaxInt16},
			[]byte{byte(TypeShorts), 0, 6, 0x80, 0x00, 0x80, 0x01, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff},
		},

		{MarshalUShorts, []int{}, []int{}, []byte{byte(TypeUShorts), 0, 0}},
		{MarshalUShorts,
			[]int{0, 1, -1, math.MaxInt16 + 1, math.MaxUint16, math.MaxUint16 + 1},
			[]int{0, 1, 0, math.MaxInt16 + 1, math.MaxUint16, math.MaxUint16},
			[]byte{byte(TypeUShorts), 0, 6, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0xff, 0xff, 0xff, 0xff},
		},

		{MarshalInts, []int{}, []int{}, []byte{byte(TypeInts), 0, 0}},
		{MarshalInts,
			[]int{0, 1, math.MinInt32 - 1, math.MinInt32, math.MaxInt32, math.MaxInt32 + 1},
			[]int{0, 1, math.MinInt32, math.MinInt32, math.MaxInt32, math.MaxInt32},
			[]byte{byte(TypeInts), 0, 6,
				0x80, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},

		{MarshalUInts, []int{}, []int{}, []byte{byte(TypeUInts), 0, 0}},
		{MarshalUInts,
			[]int{0, 1, -1, math.MaxInt32 + 1, math.MaxUint32, math.MaxUint32 + 1},
			[]int{0, 1, 0, math.MaxInt32 + 1, math.MaxUint32, math.MaxUint32},
			[]byte{byte(TypeUInts), 0, 6,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{MarshalSBytes, nil, nil, []byte{byte(TypeNull)}},
		{MarshalBytes, nil, nil, []byte{byte(TypeNull)}},
		{MarshalShorts, nil, nil, []byte{byte(TypeNull)}},
		{MarshalUShorts, nil, nil, []byte{byte(TypeNull)}},
		{MarshalInts, nil, nil, []byte{byte(TypeNull)}},
		{MarshalUInts, nil, nil, []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := test.marshal(test.in)
		if diff := cmp.Diff(b, test.buf); diff != "" {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%#v): Marshal (-got +want)\n%s", fname, test.in, diff)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%#v): Unmarshal error: %v", fname, test.in, e)
		}
		if !(test.out == nil && r == nil) {
			if diff := cmp.Diff(r, test.out); diff != "" {
				fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
				t.Fatalf("%s(%#v): Unmarshal (-got +want)\n%s", fname, test.in, diff)
			}
		}
		if l != len(test.buf) {
			fname := runtime.FuncForPC(reflect.ValueOf(test.marshal).Pointer()).Name()
			t.Fatalf("%s(%#v): Unmarshal len=%v wants %v", fname, test.in, l, len(test.buf))
		}
	}
}

func TestMarshalChars(t *testing.T) {
	tests := []struct {
		val []rune
		buf []byte
	}{
		{[]rune{}, []byte{byte(TypeChars), 0, 0}},
		{
			[]rune{'\u0000', '\u006a', 'あ'},
			[]byte{byte(TypeChars), 0, 3, 0x00, 0x00, 0x00, 0x6a, 0x30, 0x42},
		},
		{nil, []byte{byte(TypeNull)}},
		{
			[]rune("🍣"),
			[]byte{byte(TypeChars), 0, 2, 0xd8, 0x3c, 0xdf, 0x63},
		},
	}
	for _, test := range tests {
		b := MarshalChars(test.val)
		if diff := cmp.Diff(b, test.buf); diff != "" {
			t.Fatalf("MarshalChars(%#v): Marshal (-got +want)\n%s", test.val, diff)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.val == nil && r == nil) {
			if diff := cmp.Diff(r, test.val); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalLongs(t *testing.T) {
	tests := []struct {
		in  []int64
		out []int64
		buf []byte
	}{
		{[]int64{}, []int64{}, []byte{byte(TypeLongs), 0, 0}},
		{
			[]int64{0, 1, math.MinInt64, math.MaxInt64},
			[]int64{0, 1, math.MinInt64, math.MaxInt64},
			[]byte{byte(TypeLongs), 0, 4,
				0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
		},
		{nil, nil, []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := MarshalLongs(test.in)
		if diff := cmp.Diff(b, test.buf); diff != "" {
			t.Fatalf("MarshalLongs(%#v): Marshal (-got +want)\n%s", test.in, diff)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("MarshalLongs(%#v): Unmarshal error: %v", test.in, e)
		}
		if !(test.out == nil && r == nil) {
			if diff := cmp.Diff(r, test.out); diff != "" {
				t.Fatalf("Unmarshal(%#v) (-got +want)\n%s", test.in, diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal(%#v) len=%v wants %v", test.in, l, len(test.buf))
		}
	}
}

func TestMarshalULongs(t *testing.T) {
	tests := []struct {
		val []uint64
		buf []byte
	}{
		{[]uint64{}, []byte{byte(TypeULongs), 0, 0}},
		{
			[]uint64{0, 1, math.MaxInt64 + 1, math.MaxUint64},
			[]byte{byte(TypeULongs), 0, 4,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
		},
		{nil, []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := MarshalULongs(test.val)
		if diff := cmp.Diff(b, test.buf); diff != "" {
			t.Fatalf("MarshalULongs(%#v): Marshal (-got +want)\n%s", test.val, diff)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.val == nil && r == nil) {
			if diff := cmp.Diff(r, test.val); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalFloats(t *testing.T) {
	tests := []struct {
		val []float32
		buf []byte
	}{
		{[]float32{}, []byte{byte(TypeFloats), 0, 0}},
		{
			[]float32{0, float32(math.Inf(-1)), math.MaxFloat32, 1.25},
			[]byte{byte(TypeFloats), 0, 4,
				0x80, 0x00, 0x00, 0x00,
				0x00, 0x7f, 0xff, 0xff,
				0xff, 0x7f, 0xff, 0xff,
				0xbf, 0xa0, 0x00, 0x00, // sign=0, exp=7f, frac=1.01
			},
		},
		{nil, []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := MarshalFloats(test.val)
		if diff := cmp.Diff(b, test.buf); diff != "" {
			t.Fatalf("MarshalFloats(%#v): Marshal (-got +want)\n%s", test.val, diff)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.val == nil && r == nil) {
			if diff := cmp.Diff(r, test.val); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalDoubles(t *testing.T) {
	tests := []struct {
		val []float64
		buf []byte
	}{
		{[]float64{}, []byte{byte(TypeDoubles), 0, 0}},
		{
			[]float64{0, math.Inf(-1), math.MaxFloat64, 1.25},
			[]byte{byte(TypeDoubles), 0, 4,
				0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xbf, 0xf4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // sign=0, exp=3ff, frac=1.01
			},
		},
		{nil, []byte{byte(TypeNull)}},
	}
	for _, test := range tests {
		b := MarshalDoubles(test.val)
		if diff := cmp.Diff(b, test.buf); diff != "" {
			t.Fatalf("MarshalDoubles(%#v): Marshal (-got +want)\n%s", test.val, diff)
		}
		r, l, e := Unmarshal(b)
		if e != nil {
			t.Fatalf("Unmarshal error: %v", e)
		}
		if !(test.val == nil && r == nil) {
			if diff := cmp.Diff(r, test.val); diff != "" {
				t.Fatalf("Unmarshal (-got +want)\n%s", diff)
			}
		}
		if l != len(test.buf) {
			t.Fatalf("Unmarshal length = %v, wants %v", l, len(test.buf))
		}
	}
}

func TestMarshalStrings(t *testing.T) {
	s := "0123456789abcdef0123456789abcdef" // len=32
	s256 := s + s + s + s + s + s + s + s   // len=256
	s65536 := s256
	for len(s65536) < 65536 {
		s65536 += s65536
	}

	strings := []string{"", "abc", "あいうえお", s256, s65536}

	list := List{}
	for _, s := range strings {
		if len(s) <= math.MaxUint8 {
			list = append(list, MarshalStr8(s))
		} else {
			list = append(list, MarshalStr16(s))
		}
	}
	buf := MarshalList(list)

	b := MarshalStrings(strings)
	if !reflect.DeepEqual(b, buf) {
		t.Fatalf("MarshalStrings:\n%#v\n%#v", b, buf)
	}
}
