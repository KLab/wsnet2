package service

import (
	"fmt"

	"wsnet2/pb"
)

type LobbyResponse struct {
	Msg   string            `json:"msg"`
	Type  ResponseType      `json:"type"`
	Room  *pb.JoinedRoomRes `json:"room,omitempty"`
	Rooms []*pb.RoomInfo    `json:"rooms,omitempty"`
}

type ResponseType byte

const (
	ResponseTypeOK = ResponseType(iota)
	ResponseTypeRoomLimit
	ResponseTypeNoRoomFound
	ResponseTypeRoomFull
)

func (r ResponseType) String() string {
	switch r {
	case ResponseTypeOK:
		return "OK"
	case ResponseTypeRoomLimit:
		return "RoomLimit"
	case ResponseTypeNoRoomFound:
		return "NoRoomFound"
	case ResponseTypeRoomFull:
		return "RoomFull"
	default:
		return fmt.Sprintf("Unknown(%v)", r)
	}
}
