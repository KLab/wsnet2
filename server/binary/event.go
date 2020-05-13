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
	buf := make([]byte, len(ev.Payload)+5)
	buf[0] = byte(ev.Type)
	PutInt32(buf[1:], seqNum)
	copy(buf[5:], ev.Payload)
	return buf
}

// NewEvJoind : 入室イベント
func NewEvJoined(cli *pb.ClientInfo) *Event {
	payload := MarshalStr8(cli.Id)
	payload = append(payload, cli.Props...) // cli.Props marshaled as TypeDict

	return &Event{EvTypeJoined, payload}
}

func NewEvLeave(cliId string) *Event {
	return &Event{EvTypeLeave, MarshalStr8(cliId)}
}
