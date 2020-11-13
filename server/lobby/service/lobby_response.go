package service

import "wsnet2/pb"

type LobbyResponse struct {
	Msg   string            `json:"msg"`
	Room  *pb.JoinedRoomRes `json:"room,omitempty"`
	Rooms []*pb.RoomInfo    `json:"rooms,omitempty"`
}
