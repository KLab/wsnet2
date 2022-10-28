package game

import (
	"time"

	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/pb"
)

type Msg interface {
	msg()
	SenderID() ClientID
}

var _ Msg = &MsgCreate{}
var _ Msg = &MsgJoin{}
var _ Msg = &MsgWatch{}
var _ Msg = &MsgPing{}
var _ Msg = &MsgNodeCount{}
var _ Msg = &MsgLeave{}
var _ Msg = &MsgRoomProp{}
var _ Msg = &MsgClientProp{}
var _ Msg = &MsgBroadcast{}
var _ Msg = &MsgSwitchMaster{}
var _ Msg = &MsgKick{}
var _ Msg = &MsgClientError{}
var _ Msg = &MsgClientTimeout{}

const adminClientID = ClientID("")

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
	MACKey string
	Joined chan<- *JoinedInfo
	Err    chan<- ErrorWithCode
}

func (*MsgCreate) msg() {}

func (m *MsgCreate) SenderID() ClientID {
	return ClientID(m.Info.Id)
}

// MsgJoin : 入室メッセージ
// gRPCリクエストよりwsnet内で発生
type MsgJoin struct {
	Info   *pb.ClientInfo
	MACKey string
	Joined chan<- *JoinedInfo
	Err    chan<- ErrorWithCode
}

func (*MsgJoin) msg() {}

func (m *MsgJoin) SenderID() ClientID {
	return ClientID(m.Info.Id)
}

// MsgWatch : 観戦入室メッセージ
// gRPCリクエストよりwsnet内で発生
type MsgWatch struct {
	Info   *pb.ClientInfo
	MACKey string
	Joined chan<- *JoinedInfo
	Err    chan<- ErrorWithCode
}

func (*MsgWatch) msg() {}

func (m *MsgWatch) SenderID() ClientID {
	return ClientID(m.Info.Id)
}

// MsgPing : タイムアウト防止定期通信.
// nonregular message
type MsgPing struct {
	Sender    *Client
	Timestamp uint64
}

func (*MsgPing) msg() {}

func (m *MsgPing) SenderID() ClientID {
	return m.Sender.ID()
}

func msgPing(sender *Client, m binary.Msg) (Msg, error) {
	ts, err := binary.UnmarshalPingPayload(m.Payload())
	if err != nil {
		return nil, err
	}
	return &MsgPing{
		Sender:    sender,
		Timestamp: ts,
	}, nil
}

// MsgNodeCount : 観戦者数の更新
type MsgNodeCount struct {
	Sender *Client
	Count  uint32
}

func (*MsgNodeCount) msg() {}

func (m *MsgNodeCount) SenderID() ClientID {
	return m.Sender.ID()
}

func msgNodeCount(sender *Client, m binary.Msg) (Msg, error) {
	count, err := binary.UnmarshalNodeCountPayload(m.Payload())
	if err != nil {
		return nil, err
	}
	return &MsgNodeCount{
		Sender: sender,
		Count:  count,
	}, nil
}

// MsgGetRoomInfo : 部屋情報の取得
// gRPCから実行される
type MsgGetRoomInfo struct {
	Res chan<- *pb.GetRoomInfoRes
}

func (*MsgGetRoomInfo) msg() {}
func (m *MsgGetRoomInfo) SenderID() ClientID {
	return adminClientID
}

// MsgAdmingKick : 指定したClientをKickする
// gRPCから実行される
type MsgAdminKick struct {
	Target ClientID
	Res    chan<- error
}

func (*MsgAdminKick) msg() {}
func (m *MsgAdminKick) SenderID() ClientID {
	return adminClientID
}

// MsgLeave : 退室メッセージ
// クライアントの自発的な退室リクエスト
type MsgLeave struct {
	binary.RegularMsg
	Sender  *Client
	Message string
}

func (*MsgLeave) msg() {}

func (m *MsgLeave) SenderID() ClientID {
	return m.Sender.ID()
}

func msgLeave(sender *Client, msg binary.RegularMsg) (Msg, error) {
	m := "client leave"
	s, _, err := binary.UnmarshalAs(msg.Payload(), binary.TypeStr8)
	if err != nil {
		return nil, err
	}
	if ss, _ := s.(string); len(ss) > 0 {
		m = ss
	}
	return &MsgLeave{
		RegularMsg: msg,
		Sender:     sender,
		Message:    m,
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

func (m *MsgRoomProp) SenderID() ClientID {
	return m.Sender.ID()
}

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

// MsgClientProp : 自身のプロパティの変更
type MsgClientProp struct {
	binary.RegularMsg
	Sender *Client
	Props  binary.Dict
}

func (*MsgClientProp) msg() {}

func (m *MsgClientProp) SenderID() ClientID {
	return m.Sender.ID()
}

func msgClientProp(sender *Client, msg binary.RegularMsg) (Msg, error) {
	props, err := binary.UnmarshalClientProp(msg.Payload())
	if err != nil {
		return nil, err
	}
	return &MsgClientProp{
		RegularMsg: msg,
		Sender:     sender,
		Props:      props,
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

func (m *MsgTargets) SenderID() ClientID {
	return m.Sender.ID()
}

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

func (m *MsgToMaster) SenderID() ClientID {
	return m.Sender.ID()
}

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

func (m *MsgBroadcast) SenderID() ClientID {
	return m.Sender.ID()
}

func msgBroadcast(sender *Client, msg binary.RegularMsg) (Msg, error) {
	return &MsgBroadcast{
		RegularMsg: msg,
		Sender:     sender,
		Data:       msg.Payload(),
	}, nil
}

// MsgSwitchMaster : MasterClientの切替え
// MasterClientからのみ受け付ける.
type MsgSwitchMaster struct {
	binary.RegularMsg
	Sender *Client
	Target ClientID
}

func (*MsgSwitchMaster) msg() {}

func (m *MsgSwitchMaster) SenderID() ClientID {
	return m.Sender.ID()
}

func msgSwitchMaster(sender *Client, msg binary.RegularMsg) (Msg, error) {
	target, err := binary.UnmarshalSwitchMasterPayload(msg.Payload())
	if err != nil {
		return nil, err
	}
	return &MsgSwitchMaster{
		RegularMsg: msg,
		Sender:     sender,
		Target:     ClientID(target),
	}, nil
}

// MsgKick : ClientをKick
// MasterClientからのみ受け付ける.
type MsgKick struct {
	binary.RegularMsg
	Sender *Client
	Target ClientID
}

func (*MsgKick) msg() {}

func (m *MsgKick) SenderID() ClientID {
	return m.Sender.ID()
}

func msgKick(sender *Client, msg binary.RegularMsg) (Msg, error) {
	target, err := binary.UnmarshalKickPayload(msg.Payload())
	if err != nil {
		return nil, err
	}
	return &MsgKick{
		RegularMsg: msg,
		Sender:     sender,
		Target:     ClientID(target),
	}, nil
}

// MsgClientError : Client内部エラー（内部で発生）
type MsgClientError struct {
	Sender *Client
	ErrMsg string
}

func (*MsgClientError) msg() {}

func (m *MsgClientError) SenderID() ClientID {
	return m.Sender.ID()
}

// MsgClientTimeout : タイムアウトによるClientの退室
type MsgClientTimeout struct {
	Sender *Client
}

func (*MsgClientTimeout) msg() {}

func (m *MsgClientTimeout) SenderID() ClientID {
	return m.Sender.ID()
}

func ConstructMsg(cli *Client, m binary.Msg) (msg Msg, err error) {
	switch m.Type() {
	case binary.MsgTypePing:
		return msgPing(cli, m)
	case binary.MsgTypeNodeCount:
		return msgNodeCount(cli, m)
	case binary.MsgTypeLeave:
		return msgLeave(cli, m.(binary.RegularMsg))
	case binary.MsgTypeRoomProp:
		return msgRoomProp(cli, m.(binary.RegularMsg))
	case binary.MsgTypeClientProp:
		return msgClientProp(cli, m.(binary.RegularMsg))
	case binary.MsgTypeTargets:
		return msgTargets(cli, m.(binary.RegularMsg))
	case binary.MsgTypeToMaster:
		return msgToMaster(cli, m.(binary.RegularMsg))
	case binary.MsgTypeBroadcast:
		return msgBroadcast(cli, m.(binary.RegularMsg))
	case binary.MsgTypeSwitchMaster:
		return msgSwitchMaster(cli, m.(binary.RegularMsg))
	case binary.MsgTypeKick:
		return msgKick(cli, m.(binary.RegularMsg))
	}
	return nil, xerrors.Errorf("unknown msg type: %T %v", m, m)
}
