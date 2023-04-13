package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"wsnet2/lobby"
	"wsnet2/pb"
)

// Create : Roomを作成して入室
func Create(ctx context.Context, accinfo *AccessInfo, roomopt *pb.RoomOption, clinfo *pb.ClientInfo, warn func(error)) (*Room, *Connection, error) {
	param := lobby.CreateParam{
		RoomOption: roomopt,
		ClientInfo: clinfo,
		EncMACKey:  accinfo.EncMACKey,
	}

	res, err := lobbyRequest(ctx, accinfo, "/rooms", param)
	if err != nil {
		return nil, nil, xerrors.Errorf("Create: %w", err)
	}

	return connectToRoom(ctx, accinfo, res.Room, warn)
}

// Join : RoomIDを指定して入室
func Join(ctx context.Context, accinfo *AccessInfo, roomid string, query *Query, clinfo *pb.ClientInfo, warn func(error)) (*Room, *Connection, error) {
	param := lobby.JoinParam{
		Queries:    []lobby.PropQueries(*query),
		ClientInfo: clinfo,
		EncMACKey:  accinfo.EncMACKey,
	}

	res, err := lobbyRequest(ctx, accinfo, "/rooms/join/id/"+roomid, param)
	if err != nil {
		return nil, nil, xerrors.Errorf("Join: %w", err)
	}

	return connectToRoom(ctx, accinfo, res.Room, warn)
}

// JoinByNumber : 部屋番号で入室
func JoinByNumber(ctx context.Context, accinfo *AccessInfo, number int32, query *Query, clinfo *pb.ClientInfo, warn func(error)) (*Room, *Connection, error) {
	param := lobby.JoinParam{
		Queries:    []lobby.PropQueries(*query),
		ClientInfo: clinfo,
		EncMACKey:  accinfo.EncMACKey,
	}

	res, err := lobbyRequest(ctx, accinfo, fmt.Sprintf("/rooms/join/number/%d", number), param)
	if err != nil {
		return nil, nil, xerrors.Errorf("Join: %w", err)
	}

	return connectToRoom(ctx, accinfo, res.Room, warn)
}

// RandomJoin : 部屋をgroup検索してランダム入室
func RandomJoin(ctx context.Context, accinfo *AccessInfo, group uint32, query *Query, clinfo *pb.ClientInfo, warn func(error)) (*Room, *Connection, error) {
	param := lobby.JoinParam{
		Queries:    []lobby.PropQueries(*query),
		ClientInfo: clinfo,
		EncMACKey:  accinfo.EncMACKey,
	}

	res, err := lobbyRequest(ctx, accinfo, fmt.Sprintf("/rooms/join/random/%d", group), param)
	if err != nil {
		return nil, nil, xerrors.Errorf("RandomJoin: %w", err)
	}

	return connectToRoom(ctx, accinfo, res.Room, warn)
}

// Watch : RoomIDを指定して観戦入室
func Watch(ctx context.Context, accinfo *AccessInfo, roomid string, query *Query, warn func(error)) (*Room, *Connection, error) {
	var q []lobby.PropQueries
	if query != nil {
		q = []lobby.PropQueries(*query)
	}
	param := lobby.JoinParam{
		Queries:    q,
		ClientInfo: &pb.ClientInfo{Id: accinfo.UserId},
		EncMACKey:  accinfo.EncMACKey,
	}

	res, err := lobbyRequest(ctx, accinfo, "/rooms/watch/id/"+roomid, param)
	if err != nil {
		return nil, nil, xerrors.Errorf("Watch: %w", err)
	}

	return connectToRoom(ctx, accinfo, res.Room, warn)
}

// WatchDirect : gameサーバに直接gRPCで観戦リクエストする
func WatchDirect(ctx context.Context, grpccon *grpc.ClientConn, accinfo *AccessInfo, roomid string, warn func(error)) (*Room, *Connection, error) {
	req := &pb.JoinRoomReq{
		AppId:      accinfo.AppId,
		RoomId:     roomid,
		ClientInfo: &pb.ClientInfo{Id: accinfo.UserId},
		MacKey:     accinfo.MACKey,
	}

	res, err := pb.NewGameClient(grpccon).Watch(ctx, req)
	if err != nil {
		return nil, nil, xerrors.Errorf("WatchDirect: %w", err)
	}

	return connectToRoom(ctx, accinfo, res, warn)
}

func lobbyRequest(ctx context.Context, accinfo *AccessInfo, path string, param interface{}) (*lobby.Response, error) {
	var p bytes.Buffer
	enc := msgpack.NewEncoder(&p)
	enc.SetCustomStructTag("json")
	enc.UseCompactInts(true)
	err := enc.Encode(param)
	if err != nil {
		return nil, xerrors.Errorf("lobbyRequest: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", accinfo.LobbyURL+path, &p)
	if err != nil {
		return nil, xerrors.Errorf("lobbyRequest: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-msgpack")
	req.Header.Add("Wsnet2-App", accinfo.AppId)
	req.Header.Add("Wsnet2-User", accinfo.UserId)
	req.Header.Add("Authorization", "Bearer "+accinfo.Bearer)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, xerrors.Errorf("lobbyRequest: %w", err)
	}

	var res lobby.Response
	dec := msgpack.NewDecoder(r.Body)
	dec.SetCustomStructTag("json")
	err = dec.Decode(&res)
	r.Body.Close()
	if err != nil {
		return nil, xerrors.Errorf("lobbyRequest: %w", err)
	}
	if res.Type != lobby.ResponseTypeOK {
		return nil, xerrors.Errorf("lobbyRequest: %s: %v", res.Type, res.Msg)
	}

	return &res, nil
}

func connectToRoom(ctx context.Context, accinfo *AccessInfo, joined *pb.JoinedRoomRes, warn func(error)) (*Room, *Connection, error) {
	room, err := newRoom(joined, accinfo.UserId)
	if err != nil {
		return nil, nil, xerrors.Errorf("new room: %w", err)
	}

	conn, err := newConn(ctx, accinfo, joined, warn)
	if err != nil {
		return nil, nil, xerrors.Errorf("new connection: %w", err)
	}

	return room, conn, nil
}
