package game

import (
	"wsnet2/binary"
	"wsnet2/pb"

	"golang.org/x/xerrors"
)

type Msg interface {
	msg()
}

var _ Msg = &MsgCreate{}
var _ Msg = &MsgJoin{}
var _ Msg = &MsgLeave{}
var _ Msg = &MsgRoomProp{}
var _ Msg = &MsgBroadcast{}
var _ Msg = &MsgClientError{}

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

func (*MsgCreate) msg() {}

// MsgJoin : 入室メッセージ
// gRPCリクエストよりwsnet内で発生
type MsgJoin struct {
	Info   *pb.ClientInfo
	Joined chan<- JoinedInfo
}

func (*MsgJoin) msg() {}

// MsgLeave : 退室メッセージ
// クライアントの自発的な退室リクエスト
type MsgLeave struct {
	Sender *Client
}

func (*MsgLeave) msg() {}

// MsgRoomProp : 部屋情報の変更
type MsgRoomProp struct {
	Sender  *Client
	Payload []byte
	//todo: mapも必要
}

func (*MsgRoomProp) msg() {}

// MsgBroadcast : 全員に送る
type MsgBroadcast struct {
	Sender  *Client
	Payload []byte
}

func (*MsgBroadcast) msg() {}

// MsgClientError : Client内部エラー（内部で発生）
type MsgClientError struct {
	Sender *Client
	Err    error
}

func (*MsgClientError) msg() {}

func ConstructMsg(cli *Client, m binary.Msg) (msg Msg, err error) {
	switch m.Type() {
	case binary.MsgTypeLeave:
		msg = &MsgLeave{cli}
	case binary.MsgTypeBroadcast:
		msg = &MsgBroadcast{
			Sender:  cli,
			Payload: m.Payload(),
		}
	default:
		err = xerrors.Errorf("unknown msg type: %T %v", m, m)
	}
	return msg, err
}
