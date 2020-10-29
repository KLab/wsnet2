package service

import "wsnet2/pb"

type LobbyResponse struct {
	Msg   string            `json:"msg"`
	Room  *pb.JoinedRoomRes `json:"room,ommitempty"`
	Rooms []pb.RoomInfo     `json:"rooms,ommitempty"`
}
