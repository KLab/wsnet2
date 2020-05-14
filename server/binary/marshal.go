package binary

import (
	"math"
)

type Type byte

const (
	TypeNil    Type = 0 + iota // C#:null
	TypeBool                   // C#:bool
	TypeSByte                  // C#:sbyte
	TypeByte                   // C#:byte
	TypeShort                  // C#:short (16bit)
	TypeUShort                 // C#:ushort
	TypeInt                    // C#:int (32bit)
	TypeUInt                   // C#:uint
	TypeLong                   // C#:long (64bit)
	TypeULong                  // C#:ulong
	TypeStr8                   // C#:string; lenght < 256
	TypeStr16                  // C#:string; lenght >= 256
	TypeObj                    // C#:object;
	TypeList                   // C#:List<object>
	TypeDict                   // C#:Dictionary<string, object>; key length < 128
)

func PutInt8(dst []byte, val int) {
	dst[0] = byte(val & 0xff)
}
func PutInt16(dst []byte, val int) {
	dst[0] = byte((val & 0xff00) >> 8)
	dst[1] = byte(val & 0xff)
}
func PutInt24(dst []byte, val int) {
	dst[0] = byte((val & 0xff0000) >> 16)
	dst[1] = byte((val & 0xff00) >> 8)
	dst[2] = byte(val & 0xff)
}
func PutInt32(dst []byte, val int) {
	dst[0] = byte((val & 0xff000000) >> 24)
	dst[1] = byte((val & 0xff0000) >> 16)
	dst[2] = byte((val & 0xff00) >> 8)
	dst[3] = byte(val & 0xff)
}

func Int8(src []byte) int {
	i := int(src[0])
	return i
}

func Int16(src []byte) int {
	i := int(src[0]) << 8
	i += int(src[1])
	return i
}
func Int24(src []byte) int {
	i := int(src[0]) << 16
	i += int(src[1]) << 8
	i += int(src[2])
	return i
}
func Int32(src []byte) int {
	i := int(src[0]) << 24
	i += int(src[1]) << 16
	i += int(src[2]) << 8
	i += int(src[3])
	return i
}

func MarshalStr8(str string) []byte {
	len := len(str)
	if len >= math.MaxUint8 {
		len = math.MaxUint8
		str = str[:len]
	}
	buf := make([]byte, len+2)
	buf[0] = byte(TypeStr8)
	buf[1] = byte(len & 0xff)
	copy(buf[2:], []byte(str))

	return buf
}
