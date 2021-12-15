package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"hash"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"

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
	muWrite     sync.Mutex
	deadline    time.Duration
	newDeadline chan time.Duration
	done        chan struct{}
	seq         int
	macKey      string
	hmac        hash.Hash
	encMACKey   string
}

func NewBot(appId, appKey, userId string, props binary.Dict) *bot {
	macKey := auth.GenMACKey()
	hmac := hmac.New(sha1.New, []byte(macKey))
	emk, _ := auth.EncryptMACKey(macKey, appKey)

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

		macKey:    macKey,
		hmac:      hmac,
		encMACKey: emk,
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
		EncMACKey: b.encMACKey,
	}

	var res service.LobbyResponse

	err := b.doLobbyRequest("POST", fmt.Sprintf("%s/rooms", lobbyPrefix), param, &res)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("Create failed: %v", res.Msg)
	}

	room := res.Room
	b.deadline = time.Duration(room.Deadline) * time.Second
	logger.Debugf("[bot:%v] Create success, WebSocket=%s", b.userId, room.Url)

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
		EncMACKey: b.encMACKey,
	}

	var res service.LobbyResponse

	var url string
	if watch {
		url = fmt.Sprintf("%s/rooms/watch/id/%s", lobbyPrefix, roomId)
	} else {
		url = fmt.Sprintf("%s/rooms/join/id/%s", lobbyPrefix, roomId)
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
	logger.Debugf("[bot:%v] Join success, WebSocket=%s", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoomByNumber(roomNumber int32, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
		EncMACKey: b.encMACKey,
	}

	var res service.LobbyResponse

	url := fmt.Sprintf("%s/rooms/join/number/%d", lobbyPrefix, roomNumber)
	err := b.doLobbyRequest("POST", url, param, &res)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("Join by room number failed: %v", res.Msg)
	}

	room := res.Room
	b.deadline = time.Duration(room.Deadline) * time.Second
	logger.Debugf("[bot:%v] Join by room number success, WebSocket=%s", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoomAtRandom(searchGroup uint32, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
		EncMACKey: b.encMACKey,
	}

	var res service.LobbyResponse

	url := fmt.Sprintf("%s/rooms/join/random/%d", lobbyPrefix, searchGroup)
	err := b.doLobbyRequest("POST", url, param, &res)
	if err != nil {
		return nil, err
	}
	if res.Room == nil {
		return nil, fmt.Errorf("Join at random failed: %v", res.Msg)
	}

	room := res.Room
	b.deadline = time.Duration(room.Deadline) * time.Second
	logger.Debugf("[bot:%v] Join at random success, WebSocket=%s", b.userId, room.Url)
	return room, nil
}

func (b *bot) SearchRoom(searchGroup uint32, queries []lobby.PropQuery) ([]*pb.RoomInfo, error) {
	param := &service.SearchParam{
		SearchGroup: searchGroup,
		Queries:     []lobby.PropQueries{queries},
	}

	var res service.LobbyResponse

	err := b.doLobbyRequest("POST", fmt.Sprintf("%s/rooms/search", lobbyPrefix), param, &res)
	if err != nil {
		logger.Debugf("error: %v", err)
		return nil, err
	}
	if res.Rooms == nil {
		logger.Debugf("error: %v", res.Msg)
		return nil, fmt.Errorf("Search failed: %v", res.Msg)
	}

	rooms := res.Rooms
	logger.Debugf("[bot:%v] Search success, rooms=%v", b.userId, rooms)
	return rooms, nil
}

func (b *bot) doLobbyRequest(method, url string, param, dst interface{}) error {
	var p bytes.Buffer
	enc := msgpack.NewEncoder(&p)
	enc.SetCustomStructTag("json")
	enc.UseCompactInts(true)
	err := enc.Encode(param)
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

	dec := msgpack.NewDecoder(res.Body)
	dec.SetCustomStructTag("json")
	err = dec.Decode(dst)
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
		logger.Errorf("[bot:%v] generate authdata error: %v", b.userId, err)
		return err
	}
	hdr.Add("Authorization", "Bearer "+authdata)

	conn, res, err := b.ws.Dial(url, hdr)
	if err != nil {
		logger.Errorf("[bot:%v] dial error: %v, %v", b.userId, res, err)
		return err
	}
	logger.Debugf("[bot:%v] response: %v", b.userId, res)

	b.conn = conn
	go b.pinger()

	return nil
}

func (b *bot) WriteMessage(messageType int, data []byte) error {
	b.muWrite.Lock()
	defer b.muWrite.Unlock()
	return b.conn.WriteMessage(messageType, data)
}

func (b *bot) SendMessage(msgType binary.MsgType, payload []byte) error {
	b.seq++
	msg := binary.BuildRegularMsgFrame(msgType, b.seq, payload, b.hmac)
	logger.Debugf("[bot:%v] %v: seq=%v, %v", b.userId, msgType, b.seq, payload)
	return b.WriteMessage(websocket.BinaryMessage, msg)
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
			if err := b.WriteMessage(websocket.BinaryMessage, msg.Marshal(b.hmac)); err != nil {
				logger.Debugf("pinger: WrteMessage error: %v", err)
				return
			}
		case newDeadline := <-b.newDeadline:
			logger.Debugf("pinger: update deadline: %v to %v", deadline, newDeadline)
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
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logger.Debugf("[bot:%v] ReadMessage: %v", b.userId, err)
			} else {
				if !errors.Is(err, net.ErrClosed) {
					logger.Errorf("[bot:%v] ReadMessage error: %v", b.userId, err)
				}
			}
			return
		}

		ev, seq, err := binary.UnmarshalEvent(p)
		if err != nil {
			logger.Errorf("[bot:%v] Failed to UnmarshalEvent: err=%v, binary=%v", b.userId, err, p)
			continue
		}

		ty := ev.Type()
		lg := logger.With("userId", b.userId, "seq", seq, "event", ty.String())

		switch ty {
		case binary.EvTypeJoined:
			namelen := int(p[6])
			name := string(p[7 : 7+namelen])
			props, _, err := binary.Unmarshal(p[7+namelen:])
			if err != nil {
				panic(err)
			}
			lg.Debugf("name=%v props=%v", name, props)
		case binary.EvTypePermissionDenied:
			lg.Debugf("%v", string(ev.Payload()))
		case binary.EvTypeTargetNotFound:
			list, _, err := binary.UnmarshalAs(p[5:], binary.TypeList, binary.TypeNull)
			if err != nil {
				lg.Errorf("error: failed to unmarshal EvTypeTargetNotFound: %v", err)
				break
			}
			lg.Debugf("%v", list)
		case binary.EvTypeMessage:
			senderId, body, err := binary.UnmarshalEvMessage(ev.Payload())
			if err != nil {
				lg.Errorf("error: failed to unmarshal EvTypeMessage: %v", err)
				break
			}
			val, _, err := binary.Unmarshal(body)
			if err != nil {
				lg.Debugf("sender=%v body=%q", senderId, body)
			} else {
				lg.Debugf("sender=%v value=%+v", senderId, val)
			}
		case binary.EvTypeLeft:
			left, err := binary.UnmarshalEvLeftPayload(ev.Payload())
			if err != nil {
				lg.Errorf("Failed to UnmarshalEvLeftPayload: err=%v, payload=% x", err, ev.Payload())
				break
			}
			lg.Debugf("left=%q master=%q", left.ClientId, left.MasterId)
		case binary.EvTypePong:
			pongPayload, err := binary.UnmarshalEvPongPayload(ev.Payload())
			if err != nil {
				lg.Errorf("failed to unmarshal EvPongPayload: %v", err)
				break
			}
			lg.Debugf("ts=%v watchers=%v", time.Unix(int64(pongPayload.Timestamp), 0), pongPayload.Watchers)
		case binary.EvTypeMasterSwitched:
			newMasterId, err := binary.UnmarshalEvMasterSwitchedPayload(ev.Payload())
			if err != nil {
				panic(err)
			}
			lg.Debugf("new masterId=%v", newMasterId)
		case binary.EvTypeClientProp:
			cp, err := binary.UnmarshalEvClientPropPayload(ev.Payload())
			if err != nil {
				panic(err)
			}
			lg.Debugf("id=%v, props=%v", cp.Id, cp.Props)
		default:
			lg.Debugf("%#v", p)
		}
	}
}

func SpawnMaster(name string) (*bot, string, <-chan bool, error) {
	bot := NewBot(appID, appKey, name, binary.Dict{})

	logger.Debugf("spawnMaster: %v", name)
	room, err := bot.CreateRoom(binary.Dict{})
	if err != nil {
		logger.Errorf("create room error: %v", err)
		return nil, "", nil, err
	}
	logger.Debugf("CreateRoom: %v", room.RoomInfo.Id)
	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		logger.Errorf("dial game error: %v", err)
		return nil, "", nil, err
	}
	done := make(chan bool)
	go bot.EventLoop(done)

	return bot, room.RoomInfo.Id, done, nil
}

func SpawnPlayer(roomId, userId string, queries []lobby.PropQuery) (*bot, <-chan bool, error) {
	bot := NewBot(appID, appKey, userId, binary.Dict{})

	room, err := bot.JoinRoom(roomId, queries)
	if err != nil {
		logger.Errorf("[bot:%v] join room error: %v", userId, err)
		return nil, nil, err
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		logger.Errorf("[bot:%v] dial game error: %v", userId, err)
		return nil, nil, err
	}

	done := make(chan bool)
	go bot.EventLoop(done)

	return bot, done, nil
}

func SpawnWatcher(roomId, userId string) (*bot, <-chan bool, error) {
	bot := NewBot(appID, appKey, userId, binary.Dict{})

	room, err := bot.WatchRoom(roomId, nil)
	if err != nil {
		logger.Errorf("[bot:%v] watch room error: %v", userId, err)
		return nil, nil, err
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		logger.Errorf("[bot:%v] dial watch error: %v", userId, err)
		return nil, nil, err
	}

	done := make(chan bool)
	go bot.EventLoop(done)

	return bot, done, nil
}

func SpawnPlayerByNumber(roomNumber int32, userId string, queries []lobby.PropQuery) (*bot, <-chan bool, error) {
	bot := NewBot(appID, appKey, userId, binary.Dict{})

	room, err := bot.JoinRoomByNumber(roomNumber, queries)
	if err != nil {
		logger.Errorf("[bot:%v] join room error: %v", userId, err)
		return nil, nil, err
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		logger.Errorf("[bot:%v] dial game error: %v", userId, err)
		return nil, nil, err
	}

	done := make(chan bool)
	go bot.EventLoop(done)

	return bot, done, nil
}

func SpawnPlayerAtRandom(userId string, searchGroup uint32, queries []lobby.PropQuery) (*bot, <-chan bool, error) {
	logger.Infof("SpawnPlayerAtRandom(%v,%v,%v)", userId, searchGroup, queries)
	bot := NewBot(appID, appKey, userId, binary.Dict{})

	room, err := bot.JoinRoomAtRandom(searchGroup, queries)
	if err != nil {
		logger.Errorf("[bot:%v] join room error: %v", userId, err)
		return nil, nil, err
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		logger.Errorf("[bot:%v] dial game error: %v", userId, err)
		return nil, nil, err
	}

	done := make(chan bool)
	go bot.EventLoop(done)

	return bot, done, nil
}
