package binary

import (
	"wsnet2/pb"
)

//go:generate stringer -type=EvType
type EvType byte

const (
	EvTypeJoined EvType = 1 + iota
	EvTypePong
	EvTypeLeave
	EvTypeRoomProp
	EvTypeClientProp
	EvTypeMessage
)

type Event struct {
	Type    EvType
	Payload []byte
}

func (ev *Event) Serialize(seqNum int) []byte {
	buf := make([]byte, 5, len(ev.Payload)+5)
	buf[0] = byte(ev.Type)
	// sequence number
	buf[1] = byte((seqNum & 0xff000000) >> 24)
	buf[2] = byte((seqNum & 0xff0000) >> 16)
	buf[3] = byte((seqNum & 0xff00) >> 8)
	buf[4] = byte(seqNum & 0xff)

	buf = append(buf, ev.Payload...)
	return buf
}

// NewEvJoind : 入室イベント
func NewEvJoined(cli *pb.ClientInfo) *Event {
	payload := MarshalStr8(cli.Id)
	payload = append(payload, cli.Props...) // cli.Props marshaled as TypeDict

	return &Event{EvTypeJoined, payload}
}
