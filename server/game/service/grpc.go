package service

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wsnet2/log"
	"wsnet2/pb"
)

func (sv *GameService) grpcServe() <-chan error {
	errCh := make(chan error)

	go func() {
		laddr := sv.conf.GRPCAddr
		listenPort, err := net.Listen("tcp", laddr)
		if err != nil {
			errCh <- err
			return
		}

		server := grpc.NewServer()
		pb.RegisterGameServer(server, sv)

		log.Infof("game grpc: %#v", laddr)
		errCh <- server.Serve(listenPort)
	}()

	return errCh
}

func (sv *GameService) Create(ctx context.Context, in *pb.CreateRoomReq) (*pb.CreateRoomRes, error) {
	repo, ok := sv.repos[in.AppId]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid app_id: %v", in.AppId)
	}

	room, err := repo.CreateRoom(ctx, in.RoomOption, in.MasterInfo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateRoom failed: %s", err)
	}

	res := &pb.CreateRoomRes{
		Room: room,
	}

	return res, nil
}
