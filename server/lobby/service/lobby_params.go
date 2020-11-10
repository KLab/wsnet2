package service

import (
	"wsnet2/lobby"
	"wsnet2/pb"
)

type CreateParam struct {
	RoomOption pb.RoomOption `json:"room"`
	ClientInfo pb.ClientInfo `json:"client"`
}

type JoinParam struct {
	Queries    []lobby.PropQueries `json:"query"`
	ClientInfo pb.ClientInfo       `json:"client"`
}

type SearchParam struct {
	SearchGroup    uint32              `json:"group"`
	Queries        []lobby.PropQueries `json:"query"`
	Limit          uint32              `json:"limit"`
	CheckJoinable  bool                `json:"joinable,omitempty"`
	CheckWatchable bool                `json:"watchable,omitempty"`
}
