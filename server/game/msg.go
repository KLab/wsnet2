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
	Room    *pb.RoomInfo
	Players []*pb.ClientInfo
	Client  *Client
}

// MsgCreate : 部屋作成メッセージ
// gRPCリクエストよりwsnet内で発生
type MsgCreate struct {
	Info   *pb.ClientInfo
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

func msgLeave(sender *Client) (Msg, error) {
	return &MsgLeave{sender}, nil
}

// MsgRoomProp : 部屋情報の変更
// MasterClientからのみ受け付ける.
type MsgRoomProp struct {
	*binary.MsgRoomPropPayload
	Sender *Client
}

func (*MsgRoomProp) msg() {}

func msgRoomProp(sender *Client, payload []byte) (Msg, error) {
	rpp, err := binary.UnmarshalRoomPropPayload(payload)
	if err != nil {
		return nil, err
	}
	return &MsgRoomProp{
		MsgRoomPropPayload: rpp,
		Sender:             sender,
	}, nil
}

// MsgBroadcast : 全員に送る
type MsgBroadcast struct {
	Sender  *Client
	Payload []byte
}

func (*MsgBroadcast) msg() {}

func msgBroadcast(sender *Client, payload []byte) (Msg, error) {
	return &MsgBroadcast{
		Sender:  sender,
		Payload: payload,
	}, nil
}

// MsgClientError : Client内部エラー（内部で発生）
type MsgClientError struct {
	Sender *Client
	Err    error
}

func (*MsgClientError) msg() {}

func ConstructMsg(cli *Client, m binary.Msg) (msg Msg, err error) {
	switch m.Type() {
	case binary.MsgTypeLeave:
		return msgLeave(cli)
	case binary.MsgTypeRoomProp:
		return msgRoomProp(cli, m.Payload())
	case binary.MsgTypeBroadcast:
		return msgBroadcast(cli, m.Payload())
	}
	return nil, xerrors.Errorf("unknown msg type: %T %v", m, m)
}
