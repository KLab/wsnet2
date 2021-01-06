package binary

import (
	"time"

	"golang.org/x/xerrors"
)

// Msg from client via websocket
//
// regular message binary format:
// | 8bit MsgType | 24bit-be sequence number | payload ... |
//
// nonregular message (without sequence number)
// - MsgTypePing
// binary format:
// | 8bit MsgType | payload ... |
//
type Msg interface {
	Type() MsgType
	Payload() []byte
	Marshal() []byte
}

type RegularMsg interface {
	Msg
	SequenceNum() int
}

//go:generate stringer -type=MsgType
type MsgType byte

const regularMsgType = 30
const (
	// nonregular msg

	// MsgTypePing : 定期通信.
	// タイムアウトしないように
	// payload:
	// - 64bit-be: unix timestamp (milli seconds)
	MsgTypePing MsgType = 1 + iota

	// MsgTypeNodeCount : NodeCountの更新
	// payload:
	// - UInt: node count
	MsgTypeNodeCount
)
const (
	// regular msg

	// MsgTypeLeave : クライアントの自発的な退室
	// payload: (empty)
	MsgTypeLeave MsgType = regularMsgType + iota

	// MsgTypeRoomProp : 部屋情報の変更
	// MasterClientからのみ有効
	// payload:
	// - Byte: flags (1=visible, 2=joinable, 4=watchable)
	// - UInt: search group
	// - UShort: max players
	// - UShort: client deadline (second)
	// - Dict: public props (modified keys only)
	// - Dict: private props (modified keys only)
	MsgTypeRoomProp

	// MsgTypeClientProp : 自身のプロパティの変更
	// payload:
	// - Dict: properties (modified keys only)
	MsgTypeClientProp

	// MsgTypeSwitchMaster : Masterクライアントの切替え
	// payload:
	// - str8: client id
	MsgTypeSwitchMaster

	// MsgTypeTargets : 特定のクライアントへ送信
	// payload:
	//  - List: user ids
	//  - marshaled data...
	MsgTypeTargets

	// MsgTypeToMaster : 部屋のMasterクライアントへ送信
	// payload: marshaled data...
	MsgTypeToMaster

	// MsgTypeBroadcast : 全員に送信する
	// payload: marshaled data...
	MsgTypeBroadcast

	// MsgTypeKick
	// payload:
	// - str8: client id
	MsgTypeKick
)

type nonregularMsg struct {
	mtype   MsgType
	payload []byte
}

func (m *nonregularMsg) Type() MsgType   { return m.mtype }
func (m *nonregularMsg) Payload() []byte { return m.payload }
func (m *nonregularMsg) Marshal() []byte {
	data := make([]byte, 1+len(m.payload))
	data[0] = byte(m.mtype)
	copy(data[1:], m.payload)
	return data
}

type regularMsg struct {
	mtype   MsgType
	seqNum  int
	payload []byte
}

func (m *regularMsg) Type() MsgType    { return m.mtype }
func (m *regularMsg) Payload() []byte  { return m.payload }
func (m *regularMsg) SequenceNum() int { return m.seqNum }
func (m *regularMsg) Marshal() []byte {
	return BuildRegularMsgFrame(m.mtype, m.seqNum, m.payload)
}

func BuildRegularMsgFrame(t MsgType, seq int, payload []byte) []byte {
	data := make([]byte, 1+3+len(payload))
	data[0] = byte(t)
	put24(data[1:4], seq)
	copy(data[4:], payload)
	return data
}

// ParseMsg parse binary data to Msg struct
func UnmarshalMsg(data []byte) (Msg, error) {
	if len(data) < 1 {
		return nil, xerrors.Errorf("data length not enough: %v", len(data))
	}

	mt := MsgType(data[0])
	data = data[1:]

	if mt < regularMsgType {
		return &nonregularMsg{mt, data}, nil
	}

	if len(data) < 3 {
		return nil, xerrors.Errorf("data length not enough: %v", len(data))
	}
	seq := get24(data)
	data = data[3:]

	return &regularMsg{mt, seq, data}, nil
}

func NewMsgPing(timestamp time.Time) Msg {
	payload := make([]byte, 8)
	put64(payload, uint64(timestamp.Unix()))
	return &nonregularMsg{
		mtype:   MsgTypePing,
		payload: payload,
	}
}

// UnmarshalPingPayload parses payload of MsgPing
func UnmarshalPingPayload(payload []byte) (uint64, error) {
	if len(payload) < 8 {
		return 0, xerrors.Errorf("data length not enough: %v", len(payload))
	}

	return get64(payload), nil
}

func NewMsgNodeCount(count uint32) Msg {
	payload := MarshalUInt(int(count))
	return &nonregularMsg{
		mtype:   MsgTypeNodeCount,
		payload: payload,
	}
}

// UnmarshalNodeCountPayload parses payload of MsgTypeNodeCount
func UnmarshalNodeCountPayload(payload []byte) (uint32, error) {
	d, _, e := UnmarshalAs(payload, TypeUInt)
	if e != nil {
		return 0, xerrors.Errorf("Invalid MsgNodeCount payload (node count): %w", e)
	}

	return d.(uint32), nil
}

type MsgRoomPropPayload struct {
	EventPayload []byte

	Visible        bool
	Joinable       bool
	Watchable      bool
	SearchGroup    uint32
	MaxPlayer      uint32
	ClientDeadline uint32
	PublicProps    Dict
	PrivateProps   Dict
}

// flags (1=visible, 2=joinable, 4=watchable)
const (
	roomPropFlagsVisible   = 1
	roomPropFlagsJoinable  = 2
	roomPropFlagsWatchable = 4
)

// UnmarshalRoomPropPayload parses payload of MsgTypeRoomProp.
func UnmarshalRoomPropPayload(payload []byte) (*MsgRoomPropPayload, error) {
	rpp := MsgRoomPropPayload{
		EventPayload: payload,
	}

	// flags
	d, l, e := UnmarshalAs(payload, TypeByte)
	if e != nil {
		return nil, xerrors.Errorf("Invalid MsgRoomProp payload (flags): %w", e)
	}
	flags := d.(int)
	rpp.Visible = (flags & roomPropFlagsVisible) != 0
	rpp.Joinable = (flags & roomPropFlagsJoinable) != 0
	rpp.Watchable = (flags & roomPropFlagsWatchable) != 0
	payload = payload[l:]

	// search group
	d, l, e = UnmarshalAs(payload, TypeUInt)
	if e != nil {
		return nil, xerrors.Errorf("Invalid MsgRoomProp payload (search group): %w", e)
	}
	rpp.SearchGroup = uint32(d.(int))
	payload = payload[l:]

	// max players
	d, l, e = UnmarshalAs(payload, TypeUShort)
	if e != nil {
		return nil, xerrors.Errorf("Invalid MsgRoomProp payload (max players): %w", e)
	}
	rpp.MaxPlayer = uint32(d.(int))
	payload = payload[l:]

	// client deadline
	d, l, e = UnmarshalAs(payload, TypeUShort)
	if e != nil {
		return nil, xerrors.Errorf("Invalid MsgRoomProp payload (client deadline): %w", e)
	}
	rpp.ClientDeadline = uint32(d.(int))
	payload = payload[l:]

	// public props
	d, l, e = UnmarshalAs(payload, TypeDict, TypeNull)
	if e != nil {
		return nil, xerrors.Errorf("Invalid MsgRoomProp payload (public props): %w", e)
	}
	if d != nil {
		rpp.PublicProps = d.(Dict)
	}
	payload = payload[l:]

	// public props
	d, l, e = UnmarshalAs(payload, TypeDict, TypeNull)
	if e != nil {
		return nil, xerrors.Errorf("Invalid MsgRoomProp payload (private props): %w", e)
	}
	if d != nil {
		rpp.PrivateProps = d.(Dict)
	}

	return &rpp, nil
}

// UnmarshalClientProp parses payload of MsgTypeClientProp.
func UnmarshalClientProp(payload []byte) (Dict, error) {
	d, _, e := UnmarshalAs(payload, TypeDict)
	if e != nil {
		return nil, xerrors.Errorf("Invalid MsgClientProp payload (props): %w", e)
	}
	return d.(Dict), nil
}

// UnmarshalSwitchMasterPayload parses payload of MsgTypeSwitchMaster
func UnmarshalSwitchMasterPayload(payload []byte) (string, error) {
	d, _, e := UnmarshalAs(payload, TypeStr8)
	if e != nil {
		return "", xerrors.Errorf("Invalid MsgSwitchMaster payload (client id): %w", e)
	}

	return d.(string), nil
}

// UnmarshalTargetsAndData parses payload of MsgTypeTargets
func UnmarshalTargetsAndData(payload []byte) ([]string, []byte, error) {
	t, l, e := UnmarshalAs(payload, TypeList)
	if e != nil {
		return nil, nil, xerrors.Errorf("Invalid MsgTargets payload (targets): %w", e)
	}
	ls := t.(List)
	targets := make([]string, len(ls))
	for i, p := range t.(List) {
		t, _, e := Unmarshal(p)
		if e != nil {
			return nil, nil, xerrors.Errorf("Invalid MsgTargets payload (target[%v]): %w", i, e)
		}
		var ok bool
		targets[i], ok = t.(string)
		if !ok {
			return nil, nil, xerrors.Errorf("Invalid MsgTargets payload (target[%v]): %T %v", i, t, t)
		}
	}

	return targets, payload[l:], nil
}

// UnmarshalKickPayload parses payload of MsgTypeKick
func UnmarshalKickPayload(payload []byte) (string, error) {
	d, _, e := UnmarshalAs(payload, TypeStr8)
	if e != nil {
		return "", xerrors.Errorf("Invalid MsgKick payload (client id): %w", e)
	}

	return d.(string), nil
}
