package service

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"wsnet2/hub"
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

	res, err := sv.repo.WatchRoom(ctx, in.AppId, hub.RoomID(in.RoomId), in.ClientInfo)
	if err != nil {
		log.Infof("watch room error: %+v", err)
		return nil, status.Errorf(err.Code(), "WatchRoom failed: %s", err)
	}

	res.Url = fmt.Sprintf(sv.wsURLFormat, res.RoomInfo.Id)

	log.Infof("Watch room: room=%v, client=%v", res.RoomInfo.Id, in.ClientInfo.Id)

	return res, nil
}
