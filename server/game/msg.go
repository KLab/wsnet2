package game

import (
	"time"

	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/pb"
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
	Room     *pb.RoomInfo
	Players  []*pb.ClientInfo
	Client   *Client
	MasterId ClientID
	Deadline time.Duration
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
	binary.RegularMsg
	Sender *Client
}

func (*MsgLeave) msg() {}

func msgLeave(sender *Client, msg binary.RegularMsg) (Msg, error) {
	return &MsgLeave{
		RegularMsg: msg,
		Sender:     sender,
	}, nil
}

// MsgRoomProp : 部屋情報の変更
// MasterClientからのみ受け付ける.
type MsgRoomProp struct {
	binary.RegularMsg
	*binary.MsgRoomPropPayload
	Sender *Client
}

func (*MsgRoomProp) msg() {}

func msgRoomProp(sender *Client, msg binary.RegularMsg) (Msg, error) {
	rpp, err := binary.UnmarshalRoomPropPayload(msg.Payload())
	if err != nil {
		return nil, err
	}
	return &MsgRoomProp{
		RegularMsg:         msg,
		MsgRoomPropPayload: rpp,
		Sender:             sender,
	}, nil
}

// MsgTargets : 特定プレイヤーに送る
type MsgTargets struct {
	binary.RegularMsg
	Sender  *Client
	Targets []string
	Data    []byte
}

func (*MsgTargets) msg() {}

func msgTargets(sender *Client, msg binary.RegularMsg) (Msg, error) {
	targets, data, err := binary.UnmarshalTargetsAndData(msg.Payload())
	if err != nil {
		return nil, err
	}
	return &MsgTargets{
		RegularMsg: msg,
		Sender:     sender,
		Targets:    targets,
		Data:       data,
	}, nil
}

// MsgToMaster : MasterClientに送る
type MsgToMaster struct {
	binary.RegularMsg
	Sender *Client
	Data   []byte
}

func (*MsgToMaster) msg() {}

func msgToMaster(sender *Client, msg binary.RegularMsg) (Msg, error) {
	return &MsgToMaster{
		RegularMsg: msg,
		Sender:     sender,
		Data:       msg.Payload(),
	}, nil
}

// MsgBroadcast : 全員に送る
type MsgBroadcast struct {
	binary.RegularMsg
	Sender *Client
	Data   []byte
}

func (*MsgBroadcast) msg() {}

func msgBroadcast(sender *Client, msg binary.RegularMsg) (Msg, error) {
	return &MsgBroadcast{
		RegularMsg: msg,
		Sender:     sender,
		Data:       msg.Payload(),
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
		return msgLeave(cli, m.(binary.RegularMsg))
	case binary.MsgTypeRoomProp:
		return msgRoomProp(cli, m.(binary.RegularMsg))
	case binary.MsgTypeTargets:
		return msgTargets(cli, m.(binary.RegularMsg))
	case binary.MsgTypeToMaster:
		return msgToMaster(cli, m.(binary.RegularMsg))
	case binary.MsgTypeBroadcast:
		return msgBroadcast(cli, m.(binary.RegularMsg))
	}
	return nil, xerrors.Errorf("unknown msg type: %T %v", m, m)
}
