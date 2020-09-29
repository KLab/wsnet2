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

func (sv *HubService) serveGRPC(ctx context.Context) <-chan error {
	errCh := make(chan error)

	sv.preparation.Add(1)
	go func() {
		laddr := fmt.Sprintf(":%d", sv.conf.GRPCPort)
		log.Infof("hub grpc: %#v", laddr)

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

func (sv *HubService) Watch(ctx context.Context, in *pb.JoinRoomReq) (*pb.JoinedRoomRes, error) {
	log.Infof("Watch request: %v, room=%v, client=%v", in.AppId, in.RoomId, in.ClientInfo.Id)

	//TODO: 実装
	return nil, status.Errorf(codes.Unimplemented, "method Watch not implemented")

	/* memo
	room.id は app_idと組にしなくてもユニークなので hub では気にする必要がない？
	hub -> game で Watch 接続するためには認証を通らないといけない
	*/

	/*
		res, err := repo.WatchRoom(ctx, in.RoomId, in.ClientInfo)
		if err != nil {
			log.Infof("join room error: %+v", err)
			return nil, status.Errorf(codes.Internal, "JoinRoom failed: %s", err)
		}

		res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)
		log.Infof("Join room: room=%v, client=%v", res.RoomInfo.Id, in.ClientInfo.Id)
		return res, nil
	*/
}
