package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v4"

	"wsnet2/auth"
	"wsnet2/binary"
	"wsnet2/lobby"
	"wsnet2/lobby/service"
	"wsnet2/pb"
)

type bot struct {
	appId       string
	appKey      string
	userId      string
	props       binary.Dict
	ws          *websocket.Dialer
	conn        *websocket.Conn
	deadline    time.Duration
	newDeadline chan time.Duration
	ready       chan struct{}
	done        chan struct{}
}

func NewBot(appId, appKey, userId string, props binary.Dict) *bot {
	return &bot{
		appId:  appId,
		appKey: appKey,
		userId: userId,
		props:  props,
		ws: &websocket.Dialer{
			Subprotocols:    []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (b *bot) CreateRoom(props binary.Dict) (*pb.JoinedRoomRes, error) {
	param := &service.CreateParam{
		RoomOption: pb.RoomOption{
			Visible:     true,
			Joinable:    true,
			Watchable:   true,
			WithNumber:  true,
			MaxPlayers:  6,
			SearchGroup: 1,
			PublicProps: binary.MarshalDict(props),
		},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	var res service.LobbyResponse

	err := b.doLobbyRequest("POST", "http://localhost:8080/rooms", param, &res)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("Create failed: %v", res.Msg)
	}

	room := res.Room
	b.deadline = time.Duration(room.Deadline) * time.Second
	log.Printf("[bot:%v] Create success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoom(roomId string, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	return b.joinRoom(false, roomId, queries)
}

func (b *bot) WatchRoom(roomId string, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	return b.joinRoom(true, roomId, queries)
}

func (b *bot) joinRoom(watch bool, roomId string, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	var res service.LobbyResponse

	var url string
	if watch {
		url = fmt.Sprintf("http://localhost:8080/rooms/watch/id/%s", roomId)
	} else {
		url = fmt.Sprintf("http://localhost:8080/rooms/join/id/%s", roomId)
	}
	err := b.doLobbyRequest("POST", url, param, &res)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("Join failed: %v", res.Msg)
	}

	room := res.Room
	b.deadline = time.Duration(room.Deadline) * time.Second
	log.Printf("[bot:%v] Join success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoomByNumber(roomNumber int32, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	var res service.LobbyResponse

	url := fmt.Sprintf("http://localhost:8080/rooms/join/number/%d", roomNumber)
	err := b.doLobbyRequest("POST", url, param, &res)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("Join by room number failed: %v", res.Msg)
	}

	room := res.Room
	b.deadline = time.Duration(room.Deadline) * time.Second
	log.Printf("[bot:%v] Join by room number success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoomAtRandom(searchGroup uint32, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	var res service.LobbyResponse

	url := fmt.Sprintf("http://localhost:8080/rooms/join/random/%d", searchGroup)
	err := b.doLobbyRequest("POST", url, param, &res)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("Join at random failed: %v", res.Msg)
	}

	room := res.Room
	b.deadline = time.Duration(room.Deadline) * time.Second
	log.Printf("[bot:%v] Join at random success, WebSocket=%s\n", b.userId, room.Url)
	return room, nil
}

func (b *bot) SearchRoom(searchGroup uint32, queries []lobby.PropQuery) ([]*pb.RoomInfo, error) {
	param := &service.SearchParam{
		SearchGroup: searchGroup,
		Queries:     []lobby.PropQueries{queries},
	}

	var res service.LobbyResponse

	err := b.doLobbyRequest("POST", "http://localhost:8080/rooms/search", param, &res)
	if err != nil {
		log.Printf("error: %v\n", err)
		return nil, err
	}
	if res.Rooms == nil {
		log.Printf("error: %v\n", res.Msg)
		return nil, fmt.Errorf("Search failed: %v", res.Msg)
	}

	rooms := res.Rooms
	log.Printf("[bot:%v] Search success, rooms=%v\n", b.userId, rooms)
	return rooms, nil
}

func (b *bot) doLobbyRequest(method, url string, param, dst interface{}) error {
	var p bytes.Buffer
	err := msgpack.NewEncoder(&p).UseJSONTag(true).Encode(param)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, &p)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-msgpack")
	req.Header.Add("Host", "localhost")
	req.Header.Add("Wsnet2-App", b.appId)
	req.Header.Add("Wsnet2-User", b.userId)

	authdata, err := auth.GenerateAuthData(b.appKey, b.userId, time.Now())
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+authdata)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to lobby request: lobby server returned status %v", res.StatusCode)
	}

	err = msgpack.NewDecoder(res.Body).UseJSONTag(true).Decode(dst)
	if err != nil {
		return err
	}

	return nil
}

func (b *bot) DialGame(url, authKey string, seq int) error {
	hdr := http.Header{}
	hdr.Add("Wsnet2-App", b.appId)
	hdr.Add("Wsnet2-User", b.userId)
	hdr.Add("Wsnet2-LastEventSeq", strconv.Itoa(seq))

	authdata, err := auth.GenerateAuthData(authKey, b.userId, time.Now())
	if err != nil {
		log.Printf("[bot:%v] generate authdata error: %v\n", b.userId, err)
		return err
	}
	hdr.Add("Authorization", "Bearer "+authdata)

	conn, res, err := b.ws.Dial(url, hdr)
	if err != nil {
		log.Printf("[bot:%v] dial error: %v, %v\n", b.userId, res, err)
		return err
	}
	log.Printf("[bot:%v] response: %v\n", b.userId, res)

	b.conn = conn
	go b.pinger()

	return nil
}

func (b *bot) WriteMessage(messageType int, data []byte) error {
	return b.conn.WriteMessage(messageType, data)
}

func (b *bot) Close() error {
	return b.conn.Close()
}

func calcPingInterval(deadline time.Duration) time.Duration {
	return deadline / 3
}

func (b *bot) pinger() {
	deadline := b.deadline
	t := time.NewTicker(calcPingInterval(deadline))
	defer t.Stop()
	for {
		select {
		case <-t.C:
			msg := binary.NewMsgPing(time.Now())
			if err := b.WriteMessage(websocket.BinaryMessage, msg.Marshal()); err != nil {
				log.Printf("pinger: WrteMessage error: %v\n", err)
				return
			}
		case newDeadline := <-b.newDeadline:
			log.Printf("pinger: update deadline: %v to %v\n", deadline, newDeadline)
			t.Reset(calcPingInterval(newDeadline))
		case <-b.done:
			return
		}
	}
}

func (b *bot) EventLoop(done chan bool) {
	defer close(done)
	for {
		_, p, err := b.conn.ReadMessage()
		if err != nil {
			log.Printf("[bot:%v] ReadMessage error: %v\n", b.userId, err)
			return
		}

		ev, seq, err := binary.UnmarshalEvent(p)
		if err != nil {
			log.Printf("[bot:%v] Failed to UnmarshalEvent: err=%v, binary=%v", b.userId, err, p)
			continue
		}
		switch ty := ev.Type(); ty {
		case binary.EvTypeJoined:
			namelen := int(p[6])
			name := string(p[7 : 7+namelen])
			props := p[7+namelen:]
			log.Printf("[bot:%v] %s: %v %#v, %v\n", b.userId, ty, seq, name, props)
		case binary.EvTypePermissionDenied:
			log.Printf("[bot:%v] %s: %v\n", b.userId, ty, p)
		case binary.EvTypeTargetNotFound:
			list, _, err := binary.UnmarshalAs(p[5:], binary.TypeList, binary.TypeNull)
			if err != nil {
				log.Printf("[bot:%v] %s: error: %v\n", b.userId, ty, err)
				break
			}
			log.Printf("[bot:%v] %s: %v %v\n", b.userId, ty, list, p)
		case binary.EvTypeMessage:
			log.Printf("[bot:%v] %v: %q\n", b.userId, ty, string(ev.Payload()))
		case binary.EvTypeLeft:
			left, err := binary.UnmarshalEvLeftPayload(ev.Payload())
			if err != nil {
				log.Printf("[bot:%v] Failed to UnmarshalEvLeftPayload: err=%v, payload=% x", b.userId, err, ev.Payload())
				break
			}
			log.Printf("[bot:%v] %s: left=%q master=%q", b.userId, ty, left.ClientId, left.MasterId)
		default:
			log.Printf("[bot:%v] ReadMessage: %v, %v\n", b.userId, ty, p)
		}
	}
}
