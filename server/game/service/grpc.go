package service

import (
	"context"
	"fmt"
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

	sv.preparation.Add(1)
	go func() {
		laddr := fmt.Sprintf(":%d", sv.conf.GRPCPort)
		log.Infof("game grpc: %#v", laddr)

		listenPort, err := net.Listen("tcp", laddr)
		if err != nil {
			errCh <- xerrors.Errorf("listen error: %w", err)
			return
		}

		server := grpc.NewServer()
		pb.RegisterGameServer(server, sv)

		c := make(chan error)
		go func() {
			c <- server.Serve(listenPort)
		}()
		sv.preparation.Done()
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

func (sv *GameService) Create(ctx context.Context, in *pb.CreateRoomReq) (*pb.JoinedRoomRes, error) {
	logger := log.GetLoggerWith("grpc", "Create", "app", in.AppId, "user", in.MasterInfo.Id)
	sv.fillRoomOption(in.RoomOption)
	logger.Debugf("gRPC Create: %v %v", in.RoomOption, in.MasterInfo)

	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.NotFound, "Invalid app_id: %v", in.AppId)
	}

	res, err := repo.CreateRoom(ctx, in.RoomOption, in.MasterInfo, in.MacKey)
	if err != nil {
		logger.Errorf("repo.CreateRoom: %+v", err)
		return nil, status.Errorf(err.Code(), "CreateRoom failed: %s", err)
	}

	res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)

	logger.With("room", res.RoomInfo.Id).Infof("gRPC Create OK: room=%v", res.RoomInfo.Id)

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

func (sv *GameService) Join(ctx context.Context, in *pb.JoinRoomReq) (*pb.JoinedRoomRes, error) {
	logger := log.GetLoggerWith("grpc", "Join", "app", in.AppId, "user", in.ClientInfo.Id, "room", in.RoomId)
	logger.Debugf("gRPC Join: %v %v", in.RoomId, in.ClientInfo)

	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}

	res, err := repo.JoinRoom(ctx, in.RoomId, in.ClientInfo, in.MacKey)
	if err != nil {
		logger.Errorf("repo.JoinRoom: %+v", err)
		return nil, status.Errorf(err.Code(), "JoinRoom failed: %s", err)
	}

	res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)

	logger.Infof("gRPC Join OK: room=%v user=%v", res.RoomInfo.Id, in.ClientInfo.Id)

	return res, nil
}

func (sv *GameService) Watch(ctx context.Context, in *pb.JoinRoomReq) (*pb.JoinedRoomRes, error) {
	logger := log.GetLoggerWith("grpc", "Watch", "app", in.AppId, "user", in.ClientInfo.Id, "room", in.RoomId)
	logger.Debugf("gRPC Watch: %v %v", in.RoomId, in.ClientInfo)

	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}

	res, err := repo.WatchRoom(ctx, in.RoomId, in.ClientInfo, in.MacKey)
	if err != nil {
		logger.Errorf("repo.WatchRoom: %+v", err)
		return nil, status.Errorf(err.Code(), "WatchRoom failed: %s", err)
	}

	res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)

	logger.Infof("gRPC Watch OK: room=%v user=%v", res.RoomInfo.Id, in.ClientInfo.Id)

	return res, nil
}

func (sv *GameService) GetRoomInfo(ctx context.Context, in *pb.GetRoomInfoReq) (*pb.GetRoomInfoRes, error) {
	logger := log.GetLoggerWith("grpc", "GetRoomInfo", "app", in.AppId, "room", in.RoomId)
	logger.Debugf("gRPC GetRoomInfo: %v", in.RoomId)
	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}
	res, err := repo.GetRoomInfo(ctx, in.RoomId)
	if err != nil {
		logger.Errorf("repo.GetRoomInfo: %+v", err)
		return nil, err
	}

	logger.Infof("gRPC GetRoomInfo OK: room=%v", res.RoomInfo.Id)

	return res, err
}

func (sv *GameService) Kick(ctx context.Context, in *pb.KickReq) (*pb.Empty, error) {
	logger := log.GetLoggerWith("grpc", "Kick", "app", in.AppId, "room", in.RoomId, "user", in.ClientId)
	logger.Debugf("gRPC Kick: %v %v", in.RoomId, in.ClientId)
	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}
	err := repo.AdminKick(ctx, in.RoomId, in.ClientId)
	if err != nil {
		logger.Errorf("repo.AdminKick: %+v", err)
		return nil, err
	}

	logger.Infof("gRPC Kick OK: room=%v user=%v", in.RoomId, in.ClientId)

	return &pb.Empty{}, nil
}
