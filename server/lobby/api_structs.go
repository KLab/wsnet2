package lobby

import (
	"fmt"

	"wsnet2/pb"
)

type CreateParam struct {
	RoomOption *pb.RoomOption `json:"room"`
	ClientInfo *pb.ClientInfo `json:"client"`
	EncMACKey  string         `json:"emk"`
}

type JoinParam struct {
	Queries    []PropQueries  `json:"query"`
	ClientInfo *pb.ClientInfo `json:"client"`
	EncMACKey  string         `json:"emk"`
}

type SearchParam struct {
	SearchGroup    uint32        `json:"group"`
	Queries        []PropQueries `json:"query"`
	Limit          uint32        `json:"limit"`
	CheckJoinable  bool          `json:"joinable,omitempty"`
	CheckWatchable bool          `json:"watchable,omitempty"`
}

type SearchByIdsParam struct {
	RoomIDs []string      `json:"ids"`
	Queries []PropQueries `json:"query"`
}

type SearchByNumbersParam struct {
	RoomNumbers []int32       `json:"numbers"`
	Queries     []PropQueries `json:"query"`
}

type AdminKickParam struct {
	TargetID string `json:"target_id"`
}

type Response struct {
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
		return fmt.Sprintf("UnknownType(%v)", byte(r))
	}
}
