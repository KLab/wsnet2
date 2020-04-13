package game

import (
	"wsnet2/pb"
)

type Msg interface {
	msg()
}

var _ Msg = MsgCreate{}
var _ Msg = MsgJoin{}
var _ Msg = MsgLeave{}

type JoinedInfo struct {
	Room   *pb.RoomInfo
	Client *pb.ClientInfo
}

type MsgCreate struct {
	Joined chan<- JoinedInfo
}

func (MsgCreate) msg() {}

type MsgJoin struct {
	Info   *pb.ClientInfo
	Joined chan<- JoinedInfo
}

func (MsgJoin) msg() {}

type MsgLeave struct {
	ID ClientID
}

func (MsgLeave) msg() {}
