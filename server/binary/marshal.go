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
