package main

import (
	"bytes"
	"fmt"
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

var (
	appID  = "testapp"
	appKey = "testapppkey"
)

type bot struct {
	appId  string
	appKey string
	userId string
	ws     *websocket.Dialer
}

func NewBot(appId, appKey, userId string) *bot {
	return &bot{
		appId:  appId,
		appKey: appKey,
		userId: userId,
		ws: &websocket.Dialer{
			Subprotocols:    []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (b *bot) CreateRoom() (*pb.JoinedRoomRes, error) {
	props := binary.Dict{}
	props["key1"] = binary.MarshalInt(1024)
	props["key2"] = binary.MarshalStr16("hoge")

	param := &service.CreateParam{
		RoomOption: pb.RoomOption{
			Visible:     true,
			Watchable:   false,
			MaxPlayers:  4,
			SearchGroup: 1,
			PublicProps: binary.MarshalDict(props),
		},
		ClientInfo: pb.ClientInfo{
			Id: b.userId,
		},
	}

	room := &pb.JoinedRoomRes{}

	err := b.doLobbyRequest("POST", "http://localhost:8080/rooms", param, room)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[bot:%v] Create success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoom(roomId string) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		RoomId: roomId,
		ClientInfo: pb.ClientInfo{
			Id: b.userId,
		},
	}

	room := &pb.JoinedRoomRes{}

	err := b.doLobbyRequest("POST", "http://localhost:8080/rooms/join", param, room)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[bot:%v] Join success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) SearchRoom(queries []lobby.PropQuery) ([]pb.RoomInfo, error) {
	param := &service.SearchParam{
		SearchGroup: 1,
		Queries:     []lobby.PropQueries{queries},
	}

	rooms := []pb.RoomInfo{}

	err := b.doLobbyRequest("POST", "http://localhost:8080/rooms/search", param, &rooms)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[bot:%v] Search success, rooms=%v\n", b.userId, rooms)
	return nil, nil
}

func (b *bot) doLobbyRequest(method, url string, param, dst interface{}) error {
	var p bytes.Buffer
	err := msgpack.NewEncoder(&p).UseJSONTag(true).Encode(param)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, &p)
	req.Header.Add("Content-Type", "application/x-msgpack")
	req.Header.Add("Host", "localhost")
	req.Header.Add("X-App-Id", b.appId)
	req.Header.Add("X-User-Id", b.userId)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := auth.GenerateNonce()
	if err != nil {
		return err
	}
	req.Header.Add("X-Auth-Timestamp", timestamp)
	req.Header.Add("X-Auth-Nonce", nonce)
	hash := auth.GenerateHash(b.userId, timestamp, b.appKey, nonce)
	fmt.Printf("[bot:%v] hash: %v\n", b.userId, hash)
	req.Header.Add("X-Auth-Hash", hash)

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

func (b *bot) DialGame(url string, seq int) (*websocket.Conn, error) {
	hdr := http.Header{}
	hdr.Add("X-Wsnet-App", b.appId)
	hdr.Add("X-Wsnet-User", b.userId)
	hdr.Add("X-Wsnet-LastEventSeq", strconv.Itoa(seq))

	ws, res, err := b.ws.Dial(url, hdr)
	if err != nil {
		fmt.Printf("[bot:%v] dial error: %v, %v\n", b.userId, res, err)
		return nil, err
	}
	fmt.Printf("[bot:%v] response: %v\n", b.userId, res)

	return ws, nil
}

func main() {
	bot := NewBot(appID, appKey, "12345")

	room, err := bot.CreateRoom()
	if err != nil {
		fmt.Printf("create room error: %v\n", err)
		return
	}

	var queries []lobby.PropQuery
	fmt.Println("key1 =")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(queries)
	fmt.Println("key1 !")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpNot, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(queries)
	fmt.Println("key1 <")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(queries)
	fmt.Println("key1 <=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(queries)
	fmt.Println("key1 >")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(queries)
	fmt.Println("key1 >=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(queries)

	ws, err := bot.DialGame(room.Url, 0)
	if err != nil {
		fmt.Printf("dial game error: %v\n", err)
		return
	}

	done := make(chan bool)
	go eventloop(ws, bot.userId, done)

	go spawnPlayer(room.RoomInfo.Id, "23456")
	go spawnPlayer(room.RoomInfo.Id, "34567")
	go spawnPlayer(room.RoomInfo.Id, "45678")
	go spawnPlayer(room.RoomInfo.Id, "56789")
	go func() {
		time.Sleep(time.Second * 1)
		spawnPlayer(room.RoomInfo.Id, "67890")
	}()

	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("msg 001")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 1, 1, 2, 3, 4, 5})
		time.Sleep(time.Second)
		fmt.Println("msg 002")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 2, 11, 12, 13, 14, 15})
		time.Sleep(time.Second)
		fmt.Println("msg 003")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 21, 22, 23, 24, 25})
		//time.Sleep(time.Second)
		//fmt.Println("msg 003")
		//ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 21, 22, 23, 24, 25})
		//		time.Sleep(time.Second)
		//		ws.Close()
	}()

	//	<-done
	time.Sleep(6 * time.Second)

	fmt.Println("reconnect test")
	ws, err = bot.DialGame(room.Url, 2)
	if err != nil {
		fmt.Printf("dial game error: %v\n", err)
		return
	}

	done = make(chan bool)
	go eventloop(ws, bot.userId, done)

	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("msg 004")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 4, 31, 32, 33, 34, 35})
		time.Sleep(time.Second)
		fmt.Println("msg 005")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 5, 41, 42, 43, 44, 45})
		time.Sleep(time.Second)
		fmt.Println("msg 006 (leave)")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeLeave), 0, 0, 6})
		time.Sleep(time.Second)
		ws.Close()
	}()

	<-done

	time.Sleep(3 * time.Second)
	fmt.Println("reconnect test after leave")
	ws, err = bot.DialGame(room.Url, 4)
	if err != nil {
		fmt.Printf("dial game error: %v\n", err)
		return
	}

	done = make(chan bool)
	go eventloop(ws, bot.userId, done)
	<-done
}

func spawnPlayer(roomId, userId string) {
	bot := NewBot(appID, appKey, userId)
	room, err := bot.JoinRoom(roomId)
	if err != nil {
		fmt.Printf("[bot:%v] join room error: %v\n", userId, err)
		return
	}

	ws, err := bot.DialGame(room.Url, 0)
	if err != nil {
		fmt.Printf("[bot:%v] dial game error: %v\n", userId, err)
		return
	}

	done := make(chan bool)
	go eventloop(ws, userId, done)
}

func eventloop(ws *websocket.Conn, userId string, done chan bool) {
	defer close(done)
	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("[bot:%v] ReadMessage error: %v\n", userId, err)
			return
		}

		switch ty := binary.EvType(b[0]); ty {
		case binary.EvTypeJoined:
			seqnum := (int(b[1]) << 24) + (int(b[2]) << 16) + (int(b[3]) << 8) + int(b[4])
			namelen := int(b[6])
			name := string(b[7 : 7+namelen])
			props := b[7+namelen:]
			fmt.Printf("[bot:%v] %s: %v %#v, %v, %v\n", userId, ty, seqnum, name, props, b)
		default:
			fmt.Printf("[bot:%v] ReadMessage: %v, %v\n", userId, ty, b)
		}
	}
}
