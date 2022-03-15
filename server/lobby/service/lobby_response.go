package service

import "wsnet2/pb"

type LobbyResponse struct {
	Msg   string            `json:"msg"`
	Type  byte              `json:"type"`
	Room  *pb.JoinedRoomRes `json:"room,omitempty"`
	Rooms []*pb.RoomInfo    `json:"rooms,omitempty"`
}

const (
	ResponseTypeOK = byte(iota)
	ResponseTypeRoomLimit
	ResponseTypeNoRoomFound
	ResponseTypeRoomFull
)
