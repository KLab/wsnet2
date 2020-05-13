package game

import (
	"wsnet2/binary"
	"wsnet2/pb"

	"golang.org/x/xerrors"
)

type Msg interface {
	msg()
}

var _ Msg = MsgCreate{}
var _ Msg = MsgJoin{}
var _ Msg = MsgLeave{}
var _ Msg = MsgRoomProp{}
var _ Msg = MsgClientError{}

// JoinedInfo : MsgCreate/MsgJoin成功時点の情報
type JoinedInfo struct {
	Room   *pb.RoomInfo
	Client *pb.ClientInfo
}

// MsgCreate : 部屋作成メッセージ
// gRPCリクエストよりwsnet内で発生
type MsgCreate struct {
	Joined chan<- JoinedInfo
}

func (MsgCreate) msg() {}

// MsgJoin : 入室メッセージ
// gRPCリクエストよりwsnet内で発生
type MsgJoin struct {
	Info   *pb.ClientInfo
	Joined chan<- JoinedInfo
}

func (MsgJoin) msg() {}

// MsgLeave : 退室メッセージ
// クライアントから
type MsgLeave struct {
	Sender *Client
}

func (MsgLeave) msg() {}

// MsgRoomProp : 部屋情報の変更
type MsgRoomProp struct {
	Sender *Client
}

func (MsgRoomProp) msg() {}

// MsgClientError : Client内部エラー（内部で発生）
type MsgClientError struct {
	Sender *Client
	Err    error
}

func (MsgClientError) msg() {}

func UnmarshalMsg(cli *Client, data []byte) (int, Msg, error) {
	seq, m, err := binary.UnmarshalMsg(data)
	if err != nil {
		return 0, nil, err
	}

	var msg Msg
	switch m := m.(type) {
	case *binary.MsgLeave:
		msg = &MsgLeave{Sender: cli}
	default:
		err = xerrors.Errorf("unknown msg type: %T %v", m, m)
	}

	return seq, msg, err
}
