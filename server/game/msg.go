package game

import (
	"wsnet2/pb"
)

//go:generate stringer -type=MsgType
type MsgType byte

const (
	MsgTypeCreate MsgType = 1 + iota
	MsgTypeJoin
	MsgTypeLeave
	MsgTypeRoomInfo
	MsgTypeRoomProp
	MsgTypeClientProp
	MsgTypeBroadcast
	MsgTypeTarget

	MsgTypeClientError MsgType = 100
)

type Msg interface {
	msg()
}

var _ Msg = MsgCreate{}
var _ Msg = MsgJoin{}
var _ Msg = MsgLeave{}
var _ Msg = MsgClientError{}

// JoinedInfo : MsgCreate/MsgJoin成功時点の情報
type JoinedInfo struct {
	Room   *pb.RoomInfo
	Client *pb.ClientInfo
}

// MsgCreate : 部屋作成メッセージ
type MsgCreate struct {
	Joined chan<- JoinedInfo
}

func (MsgCreate) msg() {}

// MsgJoin : 入室メッセージ
type MsgJoin struct {
	Info   *pb.ClientInfo
	Joined chan<- JoinedInfo
}

func (MsgJoin) msg() {}

// MsgLeave : 退室メッセージ
type MsgLeave struct {
	ID ClientID
}

func (MsgLeave) msg() {}

// MsgClientError : Client内部エラー（内部で発生）
type MsgClientError struct {
	Client *Client
	Err    error
}

func (MsgClientError) msg() {}


func DecodeMsg(cli *Client, data []byte) (int, Msg, error) {

	return 0, nil, nil
}
