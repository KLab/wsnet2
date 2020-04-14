package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
			Visible:   true,
			Watchable: true,
			LogLevel:  4,
		},
		MasterInfo: &pb.ClientInfo{
			Id: "11111",
		},
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		fmt.Printf("create room error: %v", err)
	}

	h := &http.Client{}
	hreq, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost:8000/room/%s", res.RoomInfo.Id),
		nil)
	hreq.Header.Set("x-wsnet-app", "testapp")
	hreq.Header.Set("x-wsnet-user", "11111")
	hres, err := h.Do(hreq)

	fmt.Printf("result:%#v \n", hres)
	fmt.Printf("error::%+v \n", err)
}
