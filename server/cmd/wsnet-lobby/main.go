package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	"wsnet2/pb"
)
func main() {
	conn, err := grpc.Dial("127.0.0.1:19000", grpc.WithInsecure())
	if err != nil {
		log.Fatal("client connection error:", err)
	}
	defer conn.Close()

	client := pb.NewGameClient(conn)
	req := &pb.CreateRoomReq{
		AppId: "testapp",
		RoomOption: &pb.RoomOption{
			Visible: true,
			Watchable: true,
			Debug: true,
		},
		MasterInfo: &pb.ClientInfo{
			Id: "11111",
		},
	}


	res, err := client.Create(context.TODO(), req)
	fmt.Printf("result:%+v \n", res)
	fmt.Printf("error::%+v \n", err)
}
