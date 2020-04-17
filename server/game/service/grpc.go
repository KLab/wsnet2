package service

import (
	"context"
	"net"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wsnet2/log"
	"wsnet2/pb"
)

func (sv *GameService) serveGRPC(ctx context.Context) <-chan error {
	errCh := make(chan error)

	go func() {
		laddr := sv.conf.GRPCAddr
		log.Infof("game grpc: %#v", laddr)

		listenPort, err := net.Listen("tcp", laddr)
		if err != nil {
			errCh <- xerrors.Errorf("listen error: %w", errCh)
			return
		}

		server := grpc.NewServer()
		pb.RegisterGameServer(server, sv)

		c := make(chan error)
		go func() {
			c <- server.Serve(listenPort)
		}()
		select {
		case <-ctx.Done():
			server.Stop()
			log.Infof("gRPC server stop")
		case err := <-c:
			errCh <- err
			log.Infof("gRPC server error: %v", err)
		}
	}()

	return errCh
}

func (sv *GameService) Create(ctx context.Context, in *pb.CreateRoomReq) (*pb.CreateRoomRes, error) {
	log.Infof("Create request: %v, master=%v", in.AppId, in.MasterInfo.Id)
	sv.fillRoomOption(in.RoomOption)
	log.Debugf("option: %v", in.RoomOption)
	log.Debugf("master: %v", in.MasterInfo)

	repo, ok := sv.repos[in.AppId]
	if !ok {
		log.Infof("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid app_id: %v", in.AppId)
	}

	room, master, err := repo.CreateRoom(ctx, in.RoomOption, in.MasterInfo)
	if err != nil {
		log.Infof("create room error: %+v", err)
		return nil, status.Errorf(codes.Internal, "CreateRoom failed: %s", err)
	}

	res := &pb.CreateRoomRes{
		RoomInfo:   room,
		MasterInfo: master,
	}

	log.Infof("New room: room=%v, master=%v", room.Id, master.Id)

	return res, nil
}

func (sv *GameService) fillRoomOption(op *pb.RoomOption) {
	if op.ClientDeadline == 0 {
		op.ClientDeadline = sv.conf.DefaultDeadline
	}
	if op.MaxPlayers == 0 {
		op.MaxPlayers = sv.conf.DefaultMaxPlayers
	}
	if op.LogLevel == 0 {
		op.LogLevel = sv.conf.DefaultLoglevel
	}
}
