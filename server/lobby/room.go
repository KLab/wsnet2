package lobby

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"wsnet2/log"
	"wsnet2/pb"
)

type RoomService struct {
	db       *sqlx.DB
	grpcPort int
	wsPort   int
}

func NewRoom(info *pb.RoomInfo) *pb.Room {
	return &pb.Room{
		RoomInfo: info,
	}
}

func NewRoomService(db *sqlx.DB, grpcPort, wsPort int) *RoomService {
	rs := &RoomService{
		db:       db,
		grpcPort: grpcPort,
		wsPort:   wsPort,
	}
	return rs
}

func makeRoomURL(host, roomID string, port int) string {
	return fmt.Sprintf("ws://%s:%d/room/%s", host, port, roomID)
}

func makeRoomHost(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

type gameServer struct {
	id         int
	hostName   string
	publicHost string
}

func (rs *RoomService) Create(appID string, roomOption *pb.RoomOption, clientInfo *pb.ClientInfo) (*pb.Room, error) {
	// TODO: check App

	// TODO: check Max Rooms

	// TODO: select game server

	gs := &gameServer{
		id:         1,
		hostName:   "localhost",
		publicHost: "localhost",
	}
	grpcAddr := fmt.Sprintf("%s:%d", gs.hostName, rs.grpcPort)

	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("client connection error:", err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewGameClient(conn)

	req := &pb.CreateRoomReq{
		AppId:      appID,
		RoomOption: roomOption,
		MasterInfo: clientInfo,
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		fmt.Printf("create room error: %v", err)
	}

	// TODO: check response

	log.Infof("Created room: %v", res.RoomInfo)

	room := &pb.Room{
		RoomInfo: res.RoomInfo,
	}
	room.Hostname = makeRoomHost(gs.publicHost, rs.wsPort)
	room.Url = makeRoomURL(gs.publicHost, room.RoomInfo.Id, rs.wsPort)

	return room, nil
}
