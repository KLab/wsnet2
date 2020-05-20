package binary

import (
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
	MsgTypePing MsgType = 1 + iota
)
const (
	// regular msg
	MsgTypeLeave MsgType = regularMsgType + iota
	MsgTypeRoomProp
	MsgTypeClientProp
	MsgTypeTarget
	MsgTypeBroadcast
	MsgTypeKick
)

type nonregularMsg struct {
	mtype   MsgType
	payload []byte
}

func (m *nonregularMsg) Type() MsgType   { return m.mtype }
func (m *nonregularMsg) Payload() []byte { return m.payload }

type regularMsg struct {
	mtype   MsgType
	seqNum  int
	payload []byte
}

func (m *regularMsg) Type() MsgType    { return m.mtype }
func (m *regularMsg) Payload() []byte  { return m.payload }
func (m *regularMsg) SequenceNum() int { return m.seqNum }

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
