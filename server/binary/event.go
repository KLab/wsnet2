package binary

import (
	"wsnet2/pb"
)

//go:generate stringer -type=EvType
type EvType byte

const regularEvType = 30
const errorEvType = 128
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

	// EvTypeLeft : クライアントが退室した
	// payload:
	//  - str8: client ID
	//  - str8: master client ID
	EvTypeLeft

	// EvTypeRoomProp : 部屋情報の変更
	// payload:
	// - Byte: flags (1=visible, 2=joinable, 4=watchable)
	// - UInt: search group
	// - UShort: max players
	// - UShort: client deadline (second)
	// - Dict: public props (modified keys only)
	// - Dict: private props (modified keys only)
	EvTypeRoomProp

	// EvTypeClientProp : クライアント情報の変更
	// payload:
	//  - str8: client ID
	//  - Dict: properties (modified keys only)
	EvTypeClientProp

	// EvTypeMasterSwitched : Masterクライアントが切替わった
	// payload:
	//  - str8: new master client ID
	EvTypeMasterSwitched

	// EvTypeMessage : その他の通常メッセージ
	// payload: (any)
	EvTypeMessage
)
const (
	// EvTypePermissionDenied : 権限エラー
	// payload:
	//  - RegularMsg: original msg
	EvTypePermissionDenied EvType = errorEvType + iota

	// EvTypeTargetNotFound : あて先不明
	// payload:
	//  - List: client IDs
	//  - RegularMsg: original msg
	EvTypeTargetNotFound
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

// NewEvPong : Pongイベント
// payload:
// - unsigned 64bit-be: timestamp on ping sent.
// - unsigned 32bit-be: watcher count in the room.
// - dict: last msg timestamps of each player.
func NewEvPong(pingtime uint64, watchers uint32, lastMsg Dict) *SystemEvent {
	payload := MarshalULong(pingtime)
	payload = append(payload, MarshalUInt(int(watchers))...)
	payload = append(payload, MarshalDict(lastMsg)...)

	return &SystemEvent{
		Type:    EvTypePong,
		Payload: payload,
	}
}

// NewEvJoind : 入室イベント
func NewEvJoined(cli *pb.ClientInfo) *Event {
	payload := MarshalStr8(cli.Id)
	payload = append(payload, cli.Props...) // cli.Props marshaled as TypeDict

	return &Event{EvTypeJoined, payload}
}

func NewEvLeft(cliId, masterId string) *Event {
	payload := MarshalStr8(cliId)
	payload = append(payload, MarshalStr8(masterId)...)

	return &Event{EvTypeLeft, payload}
}

func NewEvRoomProp(cliId string, rpp *MsgRoomPropPayload) *Event {
	return &Event{EvTypeRoomProp, rpp.EventPayload}
}

func NewEvClientProp(cliId string, props []byte) *Event {
	payload := make([]byte, 0, len(cliId)+1+len(props))
	payload = append(payload, MarshalStr8(cliId)...)
	payload = append(payload, props...)

	return &Event{EvTypeClientProp, payload}
}

func NewEvMasterSwitched(cliId, masterId string) *Event {
	return &Event{EvTypeMasterSwitched, MarshalStr8(masterId)}
}

func NewEvMessage(cliId string, body []byte) *Event {
	payload := make([]byte, 0, len(cliId)+1+len(body))
	payload = append(payload, MarshalStr8(cliId)...)
	payload = append(payload, body...)
	return &Event{EvTypeMessage, payload}
}

// NewEvPermissionDenied : 権限エラー
// エラー発生の原因となったメッセージをそのまま返す
func NewEvPermissionDenied(msg RegularMsg) *Event {
	payload := make([]byte, 1+3+len(msg.Payload()))
	payload[0] = byte(msg.Type())
	put24(payload[1:], msg.SequenceNum())
	copy(payload[5:], msg.Payload())
	return &Event{EvTypePermissionDenied, payload}
}

// NewEvTargetNotFound : あて先不明
// 不明なClientのリストとエラー発生の原因となったメッセージをそのまま返す
func NewEvTargetNotFound(msg RegularMsg, cliIds []string) *Event {
	payload := MarshalStrings(cliIds)
	payload = append(payload, byte(msg.Type()))
	seq := make([]byte, 3)
	put24(seq, msg.SequenceNum())
	payload = append(payload, seq...)
	payload = append(payload, msg.Payload()...)
	return &Event{EvTypeTargetNotFound, payload}
}
