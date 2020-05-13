package binary

import (
	"golang.org/x/xerrors"
)

// Msg from client via websocket
//
// binary format:
// | 8bit MsgType | 24bit-be sequence number | payload ... |
//
type Msg interface {
	Type() MsgType
}

//go:generate stringer -type=MsgType
type MsgType byte

const (
	MsgTypePing MsgType = 1 + iota
	MsgTypeLeave
	MsgTypeRoomProp
	MsgTypeClientProp
	MsgTypeTarget
	MsgTypeBroadcast
	MsgTypeKick
)

// MsgPing request pong response
// payload: empty
type MsgPing struct{}

func (*MsgPing) Type() MsgType { return MsgTypePing }

func newMsgPing(data []byte) (Msg, error) {
	return &MsgPing{}, nil
}

// MsgLeave request leave from room
// payload: empty
type MsgLeave struct{}

func (*MsgLeave) Type() MsgType { return MsgTypeLeave }

func newMsgLeave(data []byte) (Msg, error) {
	return &MsgLeave{}, nil
}

// MsgRoomProp request change room property
// payload: Map {
//   "visible": bool,
//   "joinable": bool,
//   "watchable": bool,
//   "searchGroup": uint32,
//   "clientDeadline": uint32,
//   "maxPlayers": uint32,
//   "publicProps": Map,
//   "privateProps": Map,
// }
type MsgRoomProp struct {
}

func (*MsgRoomProp) Type() MsgType { return MsgTypeRoomProp }

func newMsgRoomProp(data []byte) (Msg, error) {
	return nil, xerrors.New("not implemented")
}

// MsgClientProp
type MsgClientProp struct {
}

func (*MsgClientProp) Type() MsgType { return MsgTypeClientProp }

func newMsgClientProp(data []byte) (Msg, error) {
	return nil, xerrors.New("not implemented")
}

// MsgTarget
type MsgTarget struct {
}

func (*MsgTarget) Type() MsgType { return MsgTypeTarget }

func newMsgTarget(data []byte) (Msg, error) {
	return nil, xerrors.New("not implemented")
}

// MsgBroadcast
type MsgBroadcast struct {
}

func (*MsgBroadcast) Type() MsgType { return MsgTypeBroadcast }

func newMsgBroadcast(data []byte) (Msg, error) {
	return nil, xerrors.New("not implemented")
}

// MsgKick
type MsgKick struct {
}

func (*MsgKick) Type() MsgType { return MsgTypeKick }

func newMsgKick(data []byte) (Msg, error) {
	return nil, xerrors.New("not implemented")
}

// ParseMsg parse binary data, returns sequence number and Msg struct.
func UnmarshalMsg(data []byte) (int, Msg, error) {
	if len(data) < 4 {
		return 0, nil, xerrors.Errorf("data length not enough: %v", len(data))
	}

	mt := MsgType(data[0])
	seq := int(data[1])<<16 + int(data[2])<<8 + int(data[3])
	data = data[4:]

	var msg Msg
	var err error
	switch mt {
	case MsgTypePing:
		msg, err = newMsgPing(data)
	case MsgTypeLeave:
		msg, err = newMsgLeave(data)
	case MsgTypeRoomProp:
		msg, err = newMsgRoomProp(data)
	case MsgTypeClientProp:
		msg, err = newMsgClientProp(data)
	case MsgTypeTarget:
		msg, err = newMsgTarget(data)
	case MsgTypeBroadcast:
		msg, err = newMsgBroadcast(data)
	case MsgTypeKick:
		msg, err = newMsgKick(data)
	default:
		err = xerrors.Errorf("unknown MsgType: %v", mt)
	}

	return seq, msg, err
}
