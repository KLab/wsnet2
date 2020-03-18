package service

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wsnet2/game"
	"wsnet2/pb"
)

type GameService struct {}

func (r *GameService) Create(ctx context.Context, in *pb.CreateRoomReq, opts ...grpc.CallOption) (*pb.CreateRoomRes, error) {
	repo := game.GetRepo(in.AppId)
	if repo == nil {
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
