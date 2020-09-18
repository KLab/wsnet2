package binary

import (
	"math"

	"golang.org/x/xerrors"
)

//go:generate stringer -type=Type -trimprefix=Type
type Type byte

const (
	TypeNull    Type = iota // C#:null
	TypeFalse               // C#:bool (false)
	TypeTrue                // C#:bool (true)
	TypeSByte               // C#:sbyte
	TypeByte                // C#:byte
	TypeChar                // C#:char
	TypeShort               // C#:short (16bit)
	TypeUShort              // C#:ushort
	TypeInt                 // C#:int (32bit)
	TypeUInt                // C#:uint
	TypeLong                // C#:long (64bit)
	TypeULong               // C#:ulong
	TypeFloat               // C#:float
	TypeDouble              // C#:double
	TypeDecimal             // C#:decimal
	TypeStr8                // C#:string; lenght < 256
	TypeStr16               // C#:string; lenght >= 256
	TypeObj                 // C#:object
	TypeList                // C#:List<object>; count < 256
	TypeDict                // C#:Dictionary<string, object>; count < 256; key length < 256

	TypeBools    // C#:bool[]
	TypeSBytes   // C#:sbyte[]
	TypeBytes    // C#:byte[]
	TypeChars    // C#:char[]
	TypeShorts   // C#:short[]
	TypeUShorts  // C#:ushort[]
	TypeInts     // C#:int[]
	TypeUInts    // C#:uint[]
	TypeLongs    // C#:long[]
	TypeULongs   // C#:ulong[]
	TypeFloats   // C#:float[]
	TypeDoubles  // C#:double[]
	TypeDecimals // C#:decimal[]
)

type Obj struct {
	ClassId byte   // specified by app
	Body    []byte // marshaled bytes
}

type List [][]byte

type Dict map[string][]byte

// MarshalNull marshals null
func MarshalNull() []byte {
	return []byte{byte(TypeNull)}
}

// MarshalBool marshals boolean value
func MarshalBool(b bool) []byte {
	if b {
		return []byte{byte(TypeTrue)}
	}
	return []byte{byte(TypeFalse)}
}

// MarshalByte marshals unsigned 8bit integer
func MarshalByte(val int) []byte {
	val = clamp(val, 0, math.MaxUint8)
	buf := make([]byte, 2)
	buf[0] = byte(TypeByte)
	put8(buf[1:], val)
	return buf
}

func unmarshalByte(src []byte) (int, int, error) {
	if len(src) < 2 {
		return 0, 0, xerrors.Errorf("Unmarshal Byte error: not enough data (%v)", len(src))
	}
	return get8(src[1:]), 2, nil
}

// MarshalSByte marshals signed 8bit integer comparably
//
// This func maps the value -128..127 to unsigned value 0..255
// to make the dst array comparable byte-by-byte directly.
func MarshalSByte(val int) []byte {
	val = clamp(val, math.MinInt8, math.MaxInt8)
	buf := make([]byte, 2)
	buf[0] = byte(TypeSByte)
	put8(buf[1:], val-math.MinInt8)
	return buf
}

func unmarshalSByte(src []byte) (int, int, error) {
	if len(src) < 2 {
		return 0, 0, xerrors.Errorf("Unmarshal SByte error: not enough data (%v)", len(src))
	}
	return get8(src[1:]) + math.MinInt8, 2, nil
}

// MarshalChar marshals 16bit code-point.
//
// if the val is larger than \uffff, it is replaced to \uffff.
func MarshalChar(val rune) []byte {
	v := clamp(int(val), 0, math.MaxUint16)
	buf := make([]byte, 3)
	buf[0] = byte(TypeChar)
	put16(buf[1:], v)
	return buf
}

func unmarshalChar(src []byte) (rune, int, error) {
	if len(src) < 3 {
		return 0, 0, xerrors.Errorf("Unmarshal Char error: not enough data (%v)", len(src))
	}
	return rune(get16(src[1:])), 3, nil
}

// MarshalUShort marshals unsigned 16bit integer
func MarshalUShort(val int) []byte {
	val = clamp(val, 0, math.MaxUint16)
	buf := make([]byte, 3)
	buf[0] = byte(TypeUShort)
	put16(buf[1:], val)
	return buf
}

func unmarshalUShort(src []byte) (int, int, error) {
	if len(src) < 3 {
		return 0, 0, xerrors.Errorf("Unmarshal UShort error: not enough data (%v)", len(src))
	}
	return get16(src[1:]), 3, nil
}

// MarshalUShort marshals signed 16bit integer comparably
func MarshalShort(val int) []byte {
	val = clamp(val, math.MinInt16, math.MaxInt16)
	buf := make([]byte, 3)
	buf[0] = byte(TypeShort)
	put16(buf[1:], val-math.MinInt16)
	return buf
}

func unmarshalShort(src []byte) (int, int, error) {
	if len(src) < 3 {
		return 0, 0, xerrors.Errorf("Unmarshal Short error: not enough data (%v)", len(src))
	}
	return get16(src[1:]) + math.MinInt16, 3, nil
}

// MarshalUInt marshals unsigned 32bit integer
func MarshalUInt(val int) []byte {
	val = clamp(val, 0, math.MaxUint32)
	buf := make([]byte, 5)
	buf[0] = byte(TypeUInt)
	put32(buf[1:], val)
	return buf
}

func unmarshalUInt(src []byte) (int, int, error) {
	if len(src) < 5 {
		return 0, 0, xerrors.Errorf("Unmarshal UInt error: not enough data (%v)", len(src))
	}
	return get32(src[1:]), 5, nil
}

// MarshalInt marshals signed 32bit integer comparably
func MarshalInt(val int) []byte {
	val = clamp(val, math.MinInt32, math.MaxInt32)
	buf := make([]byte, 5)
	buf[0] = byte(TypeInt)
	put32(buf[1:], val-math.MinInt32)
	return buf
}

func unmarshalInt(src []byte) (int, int, error) {
	if len(src) < 5 {
		return 0, 0, xerrors.Errorf("Unmarshal Int error: not enough data (%v)", len(src))
	}
	return get32(src[1:]) + math.MinInt32, 5, nil
}

// MarshalULong marshals unsigned 64bit integer
func MarshalULong(val uint64) []byte {
	buf := make([]byte, 9)
	buf[0] = byte(TypeULong)
	put64(buf[1:], val)
	return buf
}

func unmarshalULong(src []byte) (uint64, int, error) {
	if len(src) < 9 {
		return 0, 0, xerrors.Errorf("Unmarshal ULong error: not enough data (%v)", len(src))
	}
	return get64(src[1:]), 9, nil
}

// MarshalLong marshals signed 64bit integer comparably
func MarshalLong(val int) []byte {
	var v uint64
	if val >= 0 {
		v = uint64(val) + -math.MinInt64
	} else {
		v = uint64(val - math.MinInt64)
	}
	buf := make([]byte, 9)
	buf[0] = byte(TypeLong)
	put64(buf[1:], v)
	return buf
}

func unmarshalLong(src []byte) (int, int, error) {
	if len(src) < 9 {
		return 0, 0, xerrors.Errorf("Unmarshal Long error: not enough data (%v)", len(src))
	}
	v := get64(src[1:])
	if v >= -math.MinInt64 {
		return int(v - -math.MinInt64), 9, nil
	}
	return int(v) + math.MinInt64, 9, nil
}

// MarshalFloat marshals IEEE 754 single value as comparably.
// The sign-bit (MSB) is inverted to make the positive value greater than the negative value.
// All exponent and fraction bits on the negative value are inverted to make it natural order.
func MarshalFloat(val float32) []byte {
	v := math.Float32bits(val)
	buf := make([]byte, 5)
	buf[0] = byte(TypeFloat)
	if v&(1<<31) == 0 {
		v ^= 1 << 31
	} else {
		v = ^v
	}
	put32(buf[1:], int(v))
	return buf
}

func unmarshalFloat(src []byte) (float32, int, error) {
	if len(src) < 5 {
		return 0, 0, xerrors.Errorf("Unmarshal Float error: not enough data (%v)", len(src))
	}
	v := uint32(get32(src[1:]))
	if v&(1<<31) != 0 {
		v ^= 1 << 31
	} else {
		v = ^v
	}
	return math.Float32frombits(v), 5, nil
}

// MarshalFloat marshals IEEE 754 double value as comparably.
func MarshalDouble(val float64) []byte {
	v := math.Float64bits(val)
	buf := make([]byte, 9)
	buf[0] = byte(TypeDouble)
	if v&(1<<63) == 0 {
		v ^= 1 << 63
	} else {
		v = ^v
	}
	put64(buf[1:], v)
	return buf
}

func unmarshalDouble(src []byte) (float64, int, error) {
	if len(src) < 9 {
		return 0, 0, xerrors.Errorf("Unmarshal Double error: not enough data (%v)", len(src))
	}
	v := get64(src[1:])
	if v&(1<<63) != 0 {
		v ^= 1 << 63
	} else {
		v = ^v
	}
	return math.Float64frombits(v), 9, nil
}

// MarshalStr8 marshals short string (len <= 255)
func MarshalStr8(str string) []byte {
	len := len(str)
	if len >= math.MaxUint8 {
		len = math.MaxUint8
		str = str[:len]
	}
	buf := make([]byte, len+2)
	buf[0] = byte(TypeStr8)
	put8(buf[1:], len)
	copy(buf[2:], []byte(str))
	return buf
}

func unmarshalStr8(src []byte) (string, int, error) {
	if len(src) < 2 {
		return "", 0, xerrors.Errorf("Unmarshal Str8 error: not enough data (%v)", len(src))
	}
	l := get8(src[1:])
	if len(src) < 2+l {
		return "", 0, xerrors.Errorf("Unmarshal Str8(%v) error: not enough data (%v)", l, len(src))
	}
	return string(src[2 : 2+l]), 2 + l, nil
}

// MarshalStr16 marshals long string (len > 255..65545)
func MarshalStr16(str string) []byte {
	len := len(str)
	if len >= math.MaxUint16 {
		len = math.MaxUint16
		str = str[:len]
	}
	buf := make([]byte, len+3)
	buf[0] = byte(TypeStr16)
	put16(buf[1:], len)
	copy(buf[3:], []byte(str))
	return buf
}

func unmarshalStr16(src []byte) (string, int, error) {
	if len(src) < 3 {
		return "", 0, xerrors.Errorf("Unmarshal Str16 error: not enough data (%v)", len(src))
	}
	l := get16(src[1:])
	if len(src) < 3+l {
		return "", 0, xerrors.Errorf("Unmarshal Str16(%v) error: not enough data (%v)", l, len(src))
	}
	return string(src[3 : 3+l]), 3 + l, nil
}

// MarshalObj marshals Obj
// format:
//  - TypeObj
//  - 8bit class id (specified by app)
//  - 16bit body length
//  - body
func MarshalObj(obj *Obj) []byte {
	if obj == nil {
		return MarshalNull()
	}
	len := len(obj.Body)
	buf := make([]byte, len+4)
	buf[0] = byte(TypeObj)
	buf[1] = obj.ClassId
	put16(buf[2:], len)
	copy(buf[4:], obj.Body)
	return buf
}

func unmarshalObj(src []byte) (*Obj, int, error) {
	if len(src) < 4 {
		return nil, 0, xerrors.Errorf("Unmarshal Obj error: not enough data (%v)", len(src))
	}
	l := get16(src[2:])
	if len(src) < 4+l {
		return nil, 0, xerrors.Errorf("Unmarshal Obj(%v) error: not enough data (%v)", l, len(src))
	}
	obj := &Obj{
		ClassId: src[1],
		Body:    src[4 : 4+l],
	}
	return obj, 4 + l, nil
}

// MarshalList marshals List
// format:
//  - TypeList
//  - 8bit count
//  - repeat:
//    - 16bit body length
//    - marshaled body
func MarshalList(list List) []byte {
	if list == nil {
		return MarshalNull()
	}
	buf := make([]byte, 2)
	buf[0] = byte(TypeList)
	buf[1] = byte(len(list))
	sizebuf := make([]byte, 2)
	for _, b := range list {
		put16(sizebuf, len(b))
		buf = append(buf, sizebuf...)
		buf = append(buf, b...)
	}
	return buf
}

func unmarshalList(src []byte) (List, int, error) {
	if len(src) < 2 {
		return nil, 0, xerrors.Errorf("Unmarshal List error: not enough data (%v)", len(src))
	}
	count := get8(src[1:])
	l := 2
	list := make(List, count)
	for i := 0; i < count; i++ {
		if len(src) < l+2 {
			return nil, 0, xerrors.Errorf("Unmarshal List[%v](%v..) error: not enough data (%v)", i, l, len(src))
		}
		ll := get16(src[l:])
		l += 2
		if len(src) < l+ll {
			return nil, 0, xerrors.Errorf("Unmarshal List[%v](%v+%v) error: not enough data (%v)", i, l, ll, len(src))
		}
		list[i] = src[l : l+ll]
		l += ll
	}
	return list, l, nil
}

// MarshalDict marshals Dict
// format:
//  - TypeDict
//  - 8bit count
//  - repeat:
//    - 8bit key length
//    - key string
//    - 16bit body length
//    - marshaled body
func MarshalDict(dict Dict) []byte {
	if dict == nil {
		return MarshalNull()
	}
	buf := make([]byte, 2)
	buf[0] = byte(TypeDict)
	buf[1] = byte(len(dict))
	sizebuf := make([]byte, 2)
	for k, v := range dict {
		buf = append(buf, byte(len(k)))
		buf = append(buf, []byte(k)...)
		put16(sizebuf, len(v))
		buf = append(buf, sizebuf...)
		buf = append(buf, v...)
	}
	return buf
}

func unmarshalDict(src []byte) (Dict, int, error) {
	if len(src) < 2 {
		return nil, 0, xerrors.Errorf("Unmarshal Dict error: not enough data (%v)", len(src))
	}
	count := get8(src[1:])
	l := 2
	dict := make(Dict)
	for i := 0; i < count; i++ {
		if len(src) < l+1 {
			return nil, 0, xerrors.Errorf("Unmarshal Dict[%v](%v..) error: not enough data (%v)", i, l, len(src))
		}
		lk := get8(src[l:])
		l += 1
		if len(src) < l+lk+2 {
			return nil, 0, xerrors.Errorf("Unmarshal Dict[%v](%v..%v..2) error: not enough data (%v)", i, l, lk, len(src))
		}
		key := src[l : l+lk]
		l += lk
		lv := get16(src[l:])
		l += 2
		if len(src) < l+lv {
			return nil, 0, xerrors.Errorf("Unmarshal Dict[%q](%v..%v) error: not enough data (%v)", key, l, lv, len(src))
		}
		dict[string(key)] = src[l : l+lv]
		l += lv
	}
	return dict, l, nil
}

// MarshalBools marshals bool array
// format:
//  - TypeBools
//  - 16bit count
//  - repeat: bits...
func MarshalBools(bs []bool) []byte {
	if bs == nil {
		return MarshalNull()
	}

	count := len(bs)
	if count > math.MaxInt16 {
		count = math.MaxInt16
	}

	l := (len(bs) + 7) / 8
	buf := make([]byte, l+3)
	buf[0] = byte(TypeBools)
	put16(buf[1:], len(bs))

	for i, b := range bs {
		if b {
			buf[2+(i+8)/8] += 1 << (7 - byte(i%8))
		}
	}

	return buf
}

func unmarshalBools(src []byte) ([]bool, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal Bools error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + (count+7)/8
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal Bools error: not enough data (%v) wants %v", len(src), l)
	}

	bs := make([]bool, count)
	for i := range bs {
		bs[i] = src[2+(i+8)/8]&(1<<(7-i%8)) != 0
	}

	return bs, l, nil
}

// MarshalSBytes marshals sbytes array
// format:
//  - TypeSBytes
//  - 16bit count
//  - repeat: sbyte...
func MarshalSBytes(vals []int) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count)
	buf[0] = byte(TypeSBytes)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		buf[3+i] = byte(clamp(vals[i], math.MinInt8, math.MaxInt8) - math.MinInt8)
	}

	return buf
}

func unmarshalSBytes(src []byte) ([]int, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal SBytes error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal SBytes error: not enough data (%v)", len(src))
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		vals[i] = get8(src[3+i:]) + math.MinInt8
	}
	return vals, l, nil
}

// MarshalBytes marshals bytes array
// format:
//  - TypeBytes
//  - 16bit count
//  - repeat: byte...
func MarshalBytes(vals []int) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count)
	buf[0] = byte(TypeBytes)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		buf[3+i] = byte(clamp(vals[i], 0, math.MaxUint8))
	}

	return buf
}

func unmarshalBytes(src []byte) ([]int, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal Bytes error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal Bytes error: not enough data (%v)", len(src))
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		vals[i] = get8(src[3+i:])
	}
	return vals, l, nil
}

// MarshalChars marshals 16bit code-point array
// format:
//  - TypeChars
//  - 16bit count
//  - repeat: 16bit BE integer...
func MarshalChars(vals []rune) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+(count*2))
	buf[0] = byte(TypeChars)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		// todo: support surrogate pair
		v := clamp(int(vals[i]), 0, math.MaxUint16)
		put16(buf[3+i*2:], v)
	}

	return buf
}

func unmarshalChars(src []byte) ([]rune, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal UShorts error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*2
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal UShorts error: not enough data (%v)", len(src))
	}
	vals := make([]rune, count)
	for i := 0; i < count; i++ {
		// todo: support surrogate pair
		vals[i] = rune(get16(src[3+i*2:]))
	}
	return vals, l, nil
}

// MarshalShorts marshals signed 16bit integer array
// format:
//  - TypeShorts
//  - 16bit count
//  - repeat: 16bit BE integer...
func MarshalShorts(vals []int) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count*2)
	buf[0] = byte(TypeShorts)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		v := clamp(vals[i], math.MinInt16, math.MaxInt16) - math.MinInt16
		put16(buf[3+i*2:], v)
	}

	return buf
}

func unmarshalShorts(src []byte) ([]int, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal Shorts error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*2
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal Shorts error: not enough data (%v)", len(src))
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		vals[i] = get16(src[3+i*2:]) + math.MinInt16
	}
	return vals, l, nil
}

// MarshalUShort marshals unsigned 16bit integer array
// format:
//  - TypeUShort
//  - 16bit count
//  - repeat: 16bit BE integer...
func MarshalUShorts(vals []int) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+(count*2))
	buf[0] = byte(TypeUShorts)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		v := clamp(vals[i], 0, math.MaxUint16)
		put16(buf[3+i*2:], v)
	}

	return buf
}

func unmarshalUShorts(src []byte) ([]int, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal UShorts error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*2
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal UShorts error: not enough data (%v)", len(src))
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		vals[i] = get16(src[3+i*2:])
	}
	return vals, l, nil
}

// MarshalInts marshals signed 32bit integer array
// format:
//  - TypeInts
//  - 16bit count
//  - repeat: 32bit BE integer...
func MarshalInts(vals []int) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count*4)
	buf[0] = byte(TypeInts)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		v := clamp(vals[i], math.MinInt32, math.MaxInt32) - math.MinInt32
		put32(buf[3+i*4:], v)
	}

	return buf
}

func unmarshalInts(src []byte) ([]int, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal Intts error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*4
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal Ints error: not enough data (%v)", len(src))
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		vals[i] = get32(src[3+i*4:]) + math.MinInt32
	}
	return vals, l, nil
}

// MarshalInts marshals unsigned 32bit integer array
// format:
//  - TypeUInts
//  - 16bit count
//  - repeat: 32bit BE integer...
func MarshalUInts(vals []int) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+(count*4))
	buf[0] = byte(TypeUInts)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		v := clamp(vals[i], 0, math.MaxUint32)
		put32(buf[3+i*4:], v)
	}

	return buf
}

func unmarshalUInts(src []byte) ([]int, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal UInts error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*4
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal UInts error: not enough data (%v)", len(src))
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		vals[i] = get32(src[3+i*4:])
	}
	return vals, l, nil
}

// MarshalLongs marshals signed 64bit integer array
// format:
//  - TypeInts
//  - 16bit count
//  - repeat: 64bit BE integer...
func MarshalLongs(vals []int) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count*8)
	buf[0] = byte(TypeLongs)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		var v uint64
		if vals[i] >= 0 {
			v = uint64(vals[i]) + -math.MinInt64
		} else {
			v = uint64(vals[i] - math.MinInt64)
		}
		put64(buf[3+i*8:], v)
	}

	return buf
}

func unmarshalLongs(src []byte) ([]int, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal Longs error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*8
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal Longs error: not enough data (%v)", len(src))
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		v := get64(src[3+i*8:])
		if v >= -math.MinInt64 {
			vals[i] = int(v - -math.MinInt64)
		} else {
			vals[i] = int(v) + math.MinInt64
		}
	}
	return vals, l, nil
}

// MarshalULongs marshals unsigned 64bit integer array
// format:
//  - TypeInts
//  - 16bit count
//  - repeat: 64bit BE integer...
func MarshalULongs(vals []uint64) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count*8)
	buf[0] = byte(TypeULongs)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		put64(buf[3+i*8:], vals[i])
	}

	return buf
}

func unmarshalULongs(src []byte) ([]uint64, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal ULongs error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*8
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal ULongs error: not enough data (%v)", len(src))
	}
	vals := make([]uint64, count)
	for i := 0; i < count; i++ {
		vals[i] = get64(src[3+i*8:])
	}
	return vals, l, nil
}

// MarshalFloats marshals IEEE754 single array
func MarshalFloats(vals []float32) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count*4)
	buf[0] = byte(TypeFloats)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		v := math.Float32bits(vals[i])
		if v&(1<<31) == 0 {
			v ^= 1 << 31
		} else {
			v = ^v
		}
		put32(buf[3+i*4:], int(v))
	}

	return buf
}

func unmarshalFloats(src []byte) ([]float32, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal Floats error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*4
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal Floats error: not enough data (%v)", len(src))
	}
	vals := make([]float32, count)
	for i := 0; i < count; i++ {
		v := uint32(get32(src[3+i*4:]))
		if v&(1<<31) != 0 {
			v ^= 1 << 31
		} else {
			v = ^v
		}
		vals[i] = math.Float32frombits(v)
	}
	return vals, l, nil
}

// MarshalDoubles marshals IEEE754 single array
func MarshalDoubles(vals []float64) []byte {
	if vals == nil {
		return MarshalNull()
	}

	count := len(vals)
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	buf := make([]byte, 3+count*8)
	buf[0] = byte(TypeDoubles)
	put16(buf[1:], count)

	for i := 0; i < count; i++ {
		v := math.Float64bits(vals[i])
		if v&(1<<63) == 0 {
			v ^= 1 << 63
		} else {
			v = ^v
		}
		put64(buf[3+i*8:], v)
	}

	return buf
}

func unmarshalDoubles(src []byte) ([]float64, int, error) {
	if len(src) < 3 {
		return nil, 0, xerrors.Errorf("Unmarshal Doubles error: not enough data (%v)", len(src))
	}
	count := get16(src[1:])
	l := 3 + count*8
	if len(src) < l {
		return nil, 0, xerrors.Errorf("Unmarshal Doubles error: not enough data (%v)", len(src))
	}
	vals := make([]float64, count)
	for i := 0; i < count; i++ {
		v := get64(src[3+i*8:])
		if v&(1<<63) != 0 {
			v ^= 1 << 63
		} else {
			v = ^v
		}
		vals[i] = math.Float64frombits(v)
	}
	return vals, l, nil
}

func MarshalStrings(vals []string) []byte {
	buf := make([]byte, 2)
	buf[0] = byte(TypeList)
	buf[1] = byte(len(vals))
	sizebuf := make([]byte, 2)
	strbuf := make([]byte, 3)
	for _, v := range vals {
		var sz int
		n := len(v)
		if n <= math.MaxUint8 {
			sz = 1
			strbuf[0] = byte(TypeStr8)
			put8(strbuf[1:], n)
		} else {
			if n > math.MaxUint16 {
				n = math.MaxUint16
				v = v[:math.MaxUint16]
			}
			sz = 2
			strbuf[0] = byte(TypeStr16)
			put16(strbuf[1:], n)
		}
		put16(sizebuf, 1+sz+n)
		buf = append(buf, sizebuf...)
		buf = append(buf, strbuf[:1+sz]...)
		buf = append(buf, []byte(v)...)
	}
	return buf
}

// Unmarshal serialized bytes
func Unmarshal(src []byte) (interface{}, int, error) {
	if len(src) == 0 {
		return nil, 0, xerrors.Errorf("Unmarshal error: empty")
	}
	switch Type(src[0]) {
	case TypeNull:
		return nil, 1, nil
	case TypeFalse:
		return false, 1, nil
	case TypeTrue:
		return true, 1, nil
	case TypeByte:
		return unmarshalByte(src)
	case TypeSByte:
		return unmarshalSByte(src)
	case TypeChar:
		return unmarshalChar(src)
	case TypeUShort:
		return unmarshalUShort(src)
	case TypeShort:
		return unmarshalShort(src)
	case TypeUInt:
		return unmarshalUInt(src)
	case TypeInt:
		return unmarshalInt(src)
	case TypeULong:
		return unmarshalULong(src)
	case TypeLong:
		return unmarshalLong(src)
	case TypeFloat:
		return unmarshalFloat(src)
	case TypeDouble:
		return unmarshalDouble(src)
	case TypeStr8:
		return unmarshalStr8(src)
	case TypeStr16:
		return unmarshalStr16(src)
	case TypeObj:
		return unmarshalObj(src)
	case TypeList:
		return unmarshalList(src)
	case TypeDict:
		return unmarshalDict(src)
	case TypeBools:
		return unmarshalBools(src)
	case TypeSBytes:
		return unmarshalSBytes(src)
	case TypeBytes:
		return unmarshalBytes(src)
	case TypeChars:
		return unmarshalChars(src)
	case TypeShorts:
		return unmarshalShorts(src)
	case TypeUShorts:
		return unmarshalUShorts(src)
	case TypeInts:
		return unmarshalInts(src)
	case TypeUInts:
		return unmarshalUInts(src)
	case TypeLongs:
		return unmarshalLongs(src)
	case TypeULongs:
		return unmarshalULongs(src)
	case TypeFloats:
		return unmarshalFloats(src)
	case TypeDoubles:
		return unmarshalDoubles(src)
	}
	return nil, 0, xerrors.Errorf("Unknown type: %v", Type(src[0]))
}

// Unmarshal bytes as specified type
func UnmarshalAs(src []byte, types ...Type) (interface{}, int, error) {
	if len(src) == 0 {
		return nil, 0, xerrors.Errorf("Unmarshal error: empty")
	}
	st := Type(src[0])
	for _, t := range types {
		if st == t {
			return Unmarshal(src)
		}
	}

	return nil, 0, xerrors.Errorf("Unmarshal type mismatch: %v != %v", Type(src[0]), types)
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

func put8(dst []byte, val int) {
	dst[0] = byte(val)
}

func get8(src []byte) int {
	return int(src[0])
}

func put16(dst []byte, val int) {
	dst[0] = byte((val & 0xff00) >> 8)
	dst[1] = byte(val & 0xff)
}

func get16(src []byte) int {
	v := int(src[0]) << 8
	v += int(src[1])
	return v
}

func put24(dst []byte, val int) {
	dst[0] = byte((val & 0xff0000) >> 16)
	dst[1] = byte((val & 0xff00) >> 8)
	dst[2] = byte(val & 0xff)
}

func get24(src []byte) int {
	i := int(src[0]) << 16
	i += int(src[1]) << 8
	i += int(src[2])
	return i
}

func put32(dst []byte, val int) {
	dst[0] = byte((val & 0xff000000) >> 24)
	dst[1] = byte((val & 0xff0000) >> 16)
	dst[2] = byte((val & 0xff00) >> 8)
	dst[3] = byte(val & 0xff)
}
func get32(src []byte) int {
	i := int(src[0]) << 24
	i += int(src[1]) << 16
	i += int(src[2]) << 8
	i += int(src[3])
	return i
}

func put64(dst []byte, val uint64) {
	dst[0] = byte((val & 0xff00000000000000) >> 56)
	dst[1] = byte((val & 0xff000000000000) >> 48)
	dst[2] = byte((val & 0xff0000000000) >> 40)
	dst[3] = byte((val & 0xff00000000) >> 32)
	dst[4] = byte((val & 0xff000000) >> 24)
	dst[5] = byte((val & 0xff0000) >> 16)
	dst[6] = byte((val & 0xff00) >> 8)
	dst[7] = byte(val & 0xff)
}

func get64(src []byte) uint64 {
	i := uint64(src[0]) << 56
	i += uint64(src[1]) << 48
	i += uint64(src[2]) << 40
	i += uint64(src[3]) << 32
	i += uint64(src[4]) << 24
	i += uint64(src[5]) << 16
	i += uint64(src[6]) << 8
	i += uint64(src[7])
	return i
}
