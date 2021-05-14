package main

import (
	"wsnet2/binary"
)

func MarshalTargetsAndData(targets []string, payload []byte) []byte {
	return append(binary.MarshalStrings(targets), payload...)
}
