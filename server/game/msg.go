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

type MsgCreate struct{}

func (MsgCreate) msg() {}

type JoinResponse struct {
	Room   *pb.RoomInfo
	Client *pb.ClientInfo
}

type MsgJoin struct {
	Info *pb.ClientInfo
	Res  chan<- JoinResponse
}

func (MsgJoin) msg() {}

type MsgLeave struct {
	ID ClientID
}

func (MsgLeave) msg() {}
