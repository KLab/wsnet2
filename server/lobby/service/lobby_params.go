package service

import (
	"wsnet2/lobby"
	"wsnet2/pb"
)

type CreateParam struct {
	RoomOption pb.RoomOption `json:"room"`
	ClientInfo pb.ClientInfo `json:"client"`
	EncMACKey  string        `json:"emk"`
}

type JoinParam struct {
	Queries    []lobby.PropQueries `json:"query"`
	ClientInfo pb.ClientInfo       `json:"client"`
	EncMACKey  string              `json:"emk"`
}

type SearchParam struct {
	SearchGroup    uint32              `json:"group"`
	Queries        []lobby.PropQueries `json:"query"`
	Limit          uint32              `json:"limit"`
	CheckJoinable  bool                `json:"joinable,omitempty"`
	CheckWatchable bool                `json:"watchable,omitempty"`
}

type SearchByIdsParam struct {
	RoomIDs []string            `json:"ids"`
	Queries []lobby.PropQueries `json:"query"`
}

type SearchByNumbersParam struct {
	RoomNumbers []int32             `json:"numbers"`
	Queries     []lobby.PropQueries `json:"query"`
}
