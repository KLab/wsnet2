package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"wsnet2/pb"
	"wsnet2/game/service"
)

func main() {
	listenPort, err := net.Listen("tcp", ":19000")
	if err != nil {
		log.Fatalln(err)
	}

	server := grpc.NewServer()
	service := &service.GameService{}

	pb.RegisterGameServer(server, service)
	log.Printf("game start")
	server.Serve(listenPort)
}
