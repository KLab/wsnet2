package main

import (
	"wsnet2/binary"
)

func MarshalTargetsAndData(buf []byte, targets []string, payload []byte) []byte {
	buf = append(buf, binary.MarshalStrings(targets)...)
	buf = append(buf, payload...)
	return buf
}
