package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"wsnet2/auth"
	"wsnet2/lobby"
	"wsnet2/pb"
)

var (
	// LobbyTransport : Lobbyへのリクエストに使う http.Client.Transport
	// プロキシやTLS設定などを変更できる
	LobbyTransport http.RoundTripper

	// LobbyTimeout : Lobbyへのリクエストのタイムアウト時間
	LobbyTimeout time.Duration = time.Second * 5

	ErrRoomLimit   = xerrors.Errorf(lobby.ResponseTypeRoomLimit.String())
	ErrNoRoomFound = xerrors.Errorf(lobby.ResponseTypeNoRoomFound.String())
	ErrRoomFull    = xerrors.Errorf(lobby.ResponseTypeRoomFull.String())
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
		return nil, nil, xerrors.Errorf("lobbyRequest: %w", err)
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
		return nil, nil, xerrors.Errorf("lobbyRequest: %w", err)
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
		return nil, nil, xerrors.Errorf("lobbyRequest: %w", err)
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
		return nil, nil, xerrors.Errorf("lobbyRequest: %w", err)
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
		return nil, nil, xerrors.Errorf("lobbyRequest: %w", err)
	}

	return connectToRoom(ctx, accinfo, res.Room, warn)
}

// WatchDirect : gameサーバに直接接続して観戦する（hub->game用）
func WatchDirect(ctx context.Context, grpccon *grpc.ClientConn, wshost, appid, roomid string, clinfo *pb.ClientInfo, warn func(error)) (*Room, *Connection, error) {
	accinfo := &AccessInfo{
		AppId:  appid,
		UserId: clinfo.Id,
		MACKey: auth.GenMACKey(),
	}

	req := &pb.JoinRoomReq{
		AppId:      accinfo.AppId,
		RoomId:     roomid,
		ClientInfo: clinfo,
		MacKey:     accinfo.MACKey,
	}

	res, err := pb.NewGameClient(grpccon).Watch(ctx, req)
	if err != nil {
		return nil, nil, xerrors.Errorf("gRPC Watch: %w", err)
	}
	wsurl, err := url.Parse(res.Url)
	if err != nil {
		return nil, nil, xerrors.Errorf("parse url(%v): %w", res.Url, err)
	}
	wsurl.Host = wshost
	res.Url = wsurl.String()

	return connectToRoom(ctx, accinfo, res, warn)
}

// Search 部屋を検索する
func Search(ctx context.Context, accinfo *AccessInfo, param *lobby.SearchParam) ([]*pb.RoomInfo, error) {
	res, err := lobbyRequest(ctx, accinfo, "/rooms/search", param)
	if err != nil {
		return nil, err
	}

	return res.Rooms, nil
}

func lobbyRequest(ctx context.Context, accinfo *AccessInfo, path string, param interface{}) (*lobby.Response, error) {
	var p bytes.Buffer
	enc := msgpack.NewEncoder(&p)
	enc.SetCustomStructTag("json")
	enc.UseCompactInts(true)
	err := enc.Encode(param)
	if err != nil {
		return nil, xerrors.Errorf("encode param: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", accinfo.LobbyURL+path, &p)
	if err != nil {
		return nil, xerrors.Errorf("new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-msgpack")
	req.Header.Add("Wsnet2-App", accinfo.AppId)
	req.Header.Add("Wsnet2-User", accinfo.UserId)
	req.Header.Add("Authorization", "Bearer "+accinfo.Bearer)

	client := &http.Client{
		Transport: LobbyTransport,
		Timeout:   LobbyTimeout,
	}
	r, err := client.Do(req)
	if err != nil {
		return nil, xerrors.Errorf("do request: %w", err)
	}
	if r.StatusCode != 200 {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		return nil, xerrors.Errorf("do request: %v: %v", r.Status, string(body))
	}

	var res lobby.Response
	dec := msgpack.NewDecoder(r.Body)
	dec.SetCustomStructTag("json")
	err = dec.Decode(&res)
	r.Body.Close()
	if err != nil {
		return nil, xerrors.Errorf("decode body: %w", err)
	}

	switch res.Type {
	case lobby.ResponseTypeOK:
		return &res, nil

	case lobby.ResponseTypeRoomLimit:
		return &res, ErrRoomLimit
	case lobby.ResponseTypeNoRoomFound:
		return &res, ErrNoRoomFound
	case lobby.ResponseTypeRoomFull:
		return &res, ErrRoomFull
	default:
		return &res, xerrors.Errorf("response type: %s: %v", res.Type, res.Msg)
	}
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
