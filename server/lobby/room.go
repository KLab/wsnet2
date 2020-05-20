package lobby

import (
	"github.com/jmoiron/sqlx"

	"wsnet2/pb"
)

type Room struct {
	*pb.RoomInfo
	HostName   string
	PublicHost string
	URL        string
}

type RoomService struct {
	db       *sqlx.DB
	grpcPort int
	wsPort   int
}

func NewRoomService(db *sqlx.DB, grpcPort, wsPort int) *RoomService {
	rs := &RoomService{
		db:       db,
		grpcPort: grpcPort,
		wsPort:   wsPort,
	}
	return rs
}
