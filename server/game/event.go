package game

import (
	"wsnet2/pb"
)

type Event interface {
	Encode() []byte
}

var _ Event = EvJoined{}

type EvJoined struct {
	Client *pb.ClientInfo
}

func (e EvJoined) Encode() []byte {
	return []byte{}
}
