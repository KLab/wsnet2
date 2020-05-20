package binary

import (
	"math"

	"golang.org/x/xerrors"
)

//go:generate stringer -type=Type -trimprefix=Type
type Type byte

const (
	TypeNull   Type = iota // C#:null
	TypeFalse              // C#:bool (false)
	TypeTrue               // C#:bool (true)
	TypeSByte              // C#:sbyte
	TypeByte               // C#:byte
	TypeShort              // C#:short (16bit)
	TypeUShort             // C#:ushort
	TypeInt                // C#:int (32bit)
	TypeUInt               // C#:uint
	TypeLong               // C#:long (64bit)
	TypeULong              // C#:ulong
	TypeFloat              // C#:float -- not implemented yet
	TypeDouble             // C#:double -- not implemented yet
	TypeStr8               // C#:string; lenght < 256
	TypeStr16              // C#:string; lenght >= 256
	TypeObj                // C#:object
	TypeList               // C#:List<object>; count < 256
	TypeDict               // C#:Dictionary<string, object>; count < 256; key length < 256
	TypeBytes              // C#:byte array -- not implemented yet
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

// Unmarshal bytes
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
	}
	return nil, 0, xerrors.Errorf("Unknown type: %v", Type(src[0]))
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
