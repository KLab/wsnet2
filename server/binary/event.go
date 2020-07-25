package binary

import (
	"wsnet2/pb"
)

//go:generate stringer -type=EvType
type EvType byte

const regularEvType = 30
const (
	// NewEvPeerReady : Peer準備完了イベント
	// payload:
	// | 24bit-be msg sequence number |
	EvTypePeerReady EvType = 1 + iota
	EvTypePong
)
const (
	// EvTypeJoined : クライアントが入室した
	// payload:
	//  - str8: client ID
	//  - Dict: properties
	EvTypeJoined EvType = regularEvType + iota

	// EvTypeLeaved : クライアントが退室した
	// payload:
	//  - str8: client ID
	EvTypeLeaved

	// EvTypeRoomProp : 部屋情報の変更
	// payload:
	//  - UInt: client deadline
	//  - Dict: public props
	//  - Dict: private props
	EvTypeRoomProp

	// EvTypeClientProp : クライアント情報の変更
	// payload:
	//  - Dict: properties
	EvTypeClientProp

	// EvTypeMessage : その他の通常メッセージ
	// payload: (any)
	EvTypeMessage
)

// Event from wsnet to client via websocket
//
// regular event binary format:
// | 8bit EvType | 32bit-be sequence number | payload ... |
//
type Event struct {
	Type    EvType
	Payload []byte
}

func (ev *Event) Marshal(seqNum int) []byte {
	buf := make([]byte, len(ev.Payload)+5)
	buf[0] = byte(ev.Type)
	put32(buf[1:], seqNum)
	copy(buf[5:], ev.Payload)
	return buf
}

// SystemEvent (without sequence number)
// - EvTypePeerReady
// - EvTypePong
// binary format:
// | 8bit MsgType | payload ... |
//
type SystemEvent struct {
	Type    EvType
	Payload []byte
}

func (ev *SystemEvent) Marshal() []byte {
	buf := make([]byte, len(ev.Payload)+1)
	buf[0] = byte(ev.Type)
	copy(buf[1:], ev.Payload)
	return buf
}

// NewEvPeerReady : Peer準備完了イベント
// wsnetが受信済みのMsgシーケンス番号を通知.
// これを受信後、クライアントはMsgを該当シーケンス番号から送信する.
// payload:
// | 24bit-be msg sequence number |
func NewEvPeerReady(seqNum int) *SystemEvent {
	payload := make([]byte, 3)
	put24(payload, seqNum)
	return &SystemEvent{
		Type:    EvTypePeerReady,
		Payload: payload,
	}
}

// NewEvJoind : 入室イベント
func NewEvJoined(cli *pb.ClientInfo) *Event {
	payload := MarshalStr8(cli.Id)
	payload = append(payload, cli.Props...) // cli.Props marshaled as TypeDict

	return &Event{EvTypeJoined, payload}
}

func NewEvLeave(cliId string) *Event {
	return &Event{EvTypeLeaved, MarshalStr8(cliId)}
}

func NewEvRoomProp(cliId string, rpp *MsgRoomPropPayload) *Event {
	return &Event{EvTypeRoomProp, rpp.EventPayload}
}

func NewEvMessage(cliId string, body []byte) *Event {
	payload := make([]byte, 0, len(cliId)+1+len(body))
	payload = append(payload, MarshalStr8(cliId)...)
	payload = append(payload, body...)
	return &Event{EvTypeMessage, payload}
}
