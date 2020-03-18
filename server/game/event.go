package game

import (
	"wsnet2/pb"
)

type Event interface {
	event()
}

var _ Event = EvJoined{}

type EvJoined struct {
	Client *pb.ClientInfo
}

func (EvJoined) event() {}
