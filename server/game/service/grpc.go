package service

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wsnet2/game"
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
	logger := log.GetLoggerWith(
		log.KeyHandler, "grpc:Create",
		log.KeyApp, in.AppId,
		log.KeyClient, in.MasterInfo.Id,
		log.KeyRequestedAt, float64(time.Now().UnixMilli())/1000,
	)
	sv.fillRoomOption(in.RoomOption)
	logger.Debugf("gRPC Create: %v %v", in.RoomOption, in.MasterInfo)

	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.NotFound, "Invalid app_id: %v", in.AppId)
	}

	res, err := repo.CreateRoom(ctx, in.RoomOption, in.MasterInfo, in.MacKey)
	if err != nil {
		logEWC(logger, "repo.CreateRoom", err)
		return nil, status.Errorf(err.Code(), "CreateRoom failed: %s", err)
	}

	res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)

	logger.With(log.KeyRoom, res.RoomInfo.Id).Infof("gRPC Create OK: room=%v", res.RoomInfo.Id)

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
	logger := log.GetLoggerWith(
		log.KeyHandler, "grpc:Join",
		log.KeyApp, in.AppId,
		log.KeyClient, in.ClientInfo.Id,
		log.KeyRoom, in.RoomId,
		log.KeyRequestedAt, float64(time.Now().UnixMilli())/1000,
	)
	logger.Debugf("gRPC Join: %v %v", in.RoomId, in.ClientInfo)

	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}

	res, err := repo.JoinRoom(ctx, in.RoomId, in.ClientInfo, in.MacKey)
	if err != nil {
		logEWC(logger, "repo.JoinRoom", err)
		return nil, status.Errorf(err.Code(), "JoinRoom failed: %s", err)
	}

	res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)

	logger.Infof("gRPC Join OK: room=%v user=%v", res.RoomInfo.Id, in.ClientInfo.Id)

	return res, nil
}

func (sv *GameService) Watch(ctx context.Context, in *pb.JoinRoomReq) (*pb.JoinedRoomRes, error) {
	logger := log.GetLoggerWith(
		log.KeyHandler, "grpc:Watch",
		log.KeyApp, in.AppId,
		log.KeyClient, in.ClientInfo.Id,
		log.KeyRoom, in.RoomId,
		log.KeyRequestedAt, float64(time.Now().UnixMilli())/1000,
	)
	logger.Debugf("gRPC Watch: %v %v", in.RoomId, in.ClientInfo)

	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}

	res, err := repo.WatchRoom(ctx, in.RoomId, in.ClientInfo, in.MacKey)
	if err != nil {
		logEWC(logger, "repo.WatchRoom", err)
		return nil, status.Errorf(err.Code(), "WatchRoom failed: %s", err)
	}

	res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)

	logger.Infof("gRPC Watch OK: room=%v user=%v", res.RoomInfo.Id, in.ClientInfo.Id)

	return res, nil
}

func (sv *GameService) GetRoomInfo(ctx context.Context, in *pb.GetRoomInfoReq) (*pb.GetRoomInfoRes, error) {
	logger := log.GetLoggerWith(
		log.KeyHandler, "grpc:GetRoomInfo",
		log.KeyApp, in.AppId,
		log.KeyRoom, in.RoomId,
		log.KeyRequestedAt, float64(time.Now().UnixMilli())/1000,
	)
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

	return res, nil
}

func (sv *GameService) CurrentRooms(ctx context.Context, in *pb.CurrentRoomsReq) (*pb.RoomIdsRes, error) {
	logger := log.GetLoggerWith(
		log.KeyHandler, "grpc:CurrentRooms",
		log.KeyApp, in.AppId,
		log.KeyClient, in.ClientId,
		log.KeyRequestedAt, float64(time.Now().UnixMilli())/1000,
	)
	logger.Debugf("gRPC CurrentRooms: %v", in.ClientId)
	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}
	res, err := repo.GetCurrentRoomIds(ctx, in.ClientId)
	if err != nil {
		logger.Errorf("repo.CurrentRooms: %+v", err)
		return nil, err
	}

	logger.Infof("gRPC CurrentRooms OK: %v", res.RoomIds)

	return res, nil
}

func (sv *GameService) Kick(ctx context.Context, in *pb.KickReq) (*pb.Empty, error) {
	logger := log.GetLoggerWith(
		log.KeyHandler, "grcp:Kick",
		log.KeyApp, in.AppId,
		log.KeyRoom, in.RoomId,
		log.KeyClient, in.ClientId,
		log.KeyRequestedAt, float64(time.Now().UnixMilli())/1000,
	)
	logger.Debugf("gRPC Kick: %v %v", in.RoomId, in.ClientId)
	repo, ok := sv.repos[in.AppId]
	if !ok {
		logger.Errorf("invalid app_id: %v", in.AppId)
		return nil, status.Errorf(codes.Internal, "Invalid app_id: %v", in.AppId)
	}
	err := repo.AdminKick(ctx, in.RoomId, in.ClientId, logger)
	if err != nil {
		logger.Errorf("repo.AdminKick: %+v", err)
		return nil, err
	}

	logger.Infof("gRPC Kick OK: room=%q user=%q", in.RoomId, in.ClientId)

	return &pb.Empty{}, nil
}

func logEWC(logger log.Logger, msg string, err game.ErrorWithCode) {
	if err.IsNormal() {
		logger.Infof("%s: %v", msg, err)
	} else {
		logger.Errorf("%s: %+v", msg, err)
	}
}
