package main

import (
	"fmt"
	"log"
	"time"

	"wsnet2/binary"
	"wsnet2/lobby"
)

type normalBot struct {
	name string
}

func NewNormalBot() *normalBot {
	return &normalBot{"normal"}
}

func (cmd *normalBot) Name() string {
	return cmd.name
}

func (cmd *normalBot) Execute() {
	bot := NewBot(appID, appKey, "12345", binary.Dict{})

	room, err := bot.CreateRoom(binary.Dict{"key1": binary.MarshalInt(1024), "key2": binary.MarshalStr16("hoge")})
	if err != nil {
		log.Printf("create room error: %v", err)
		return
	}

	var queries []lobby.PropQuery
	fmt.Println("key1 =")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 !")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpNot, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 <")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 <=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 >")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 >=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		log.Printf("dial game error: %v", err)
		return
	}

	done := make(chan bool)
	go bot.EventLoop(done)

	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpEqual, Val: binary.MarshalInt(1024)}}

	go spawnPlayer(room.RoomInfo.Id, "23456", nil)
	go spawnPlayer(room.RoomInfo.Id, "34567", queries)
	go spawnPlayerByNumber(room.RoomInfo.Number.Number, "45678", nil)
	go spawnPlayerByNumber(room.RoomInfo.Number.Number, "56789", queries)
	go spawnPlayerAtRandom("67890", 1, queries)

	for i := 0; i < 5; i++ {
		go spawnWatcher(room.RoomInfo.Id, fmt.Sprintf("watcher-%d", i))
		time.Sleep(time.Millisecond * 10)
	}

	go func() {
		time.Sleep(time.Second * 1)
		spawnPlayer(room.RoomInfo.Id, "78901", nil)
	}()
	go func() {
		time.Sleep(time.Second * 1)
		spawnPlayer(room.RoomInfo.Id, "89012", nil)
	}()

	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("msg 001")
		bot.SendMessage(binary.MsgTypeSwitchMaster, binary.MarshalStr8("23456"))
		time.Sleep(time.Second)
		fmt.Println("msg 002")
		bot.SendMessage(binary.MsgTypeSwitchMaster, binary.MarshalStr8("34567"))
		time.Sleep(time.Second)
		fmt.Println("msg 003")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{1, 2, 3, 4, 5})
		time.Sleep(time.Second)
		fmt.Println("msg 004")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{11, 12, 13, 14, 15})
		time.Sleep(time.Second)
		fmt.Println("msg 005")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{21, 22, 23, 24, 25})
		//time.Sleep(time.Second)
		//fmt.Println("msg 003")
		//ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 21, 22, 23, 24, 25})
		//		time.Sleep(time.Second)
		//		ws.Close()
		time.Sleep(time.Second)
		fmt.Println("msg 006")
		props := binary.MarshalDict(binary.Dict{
			"p1": binary.MarshalUShort(20),
			"p2": []byte{},
		})
		bot.SendMessage(binary.MsgTypeClientProp, props)
	}()

	//	<-done
	time.Sleep(8 * time.Second)

	fmt.Println("reconnect test")
	err = bot.DialGame(room.Url, room.AuthKey, 2)
	if err != nil {
		log.Printf("dial game error: %v", err)
		return
	}

	done = make(chan bool)
	go bot.EventLoop(done)

	go spawnPlayer(room.RoomInfo.Id, "99999", nil)
	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("msg 007")
		bot.SendMessage(binary.MsgTypeKick, binary.MarshalStr8("99999"))
		time.Sleep(time.Second)
		fmt.Println("msg 008")
		bot.SendMessage(binary.MsgTypeKick, binary.MarshalStr8("00000"))
		time.Sleep(time.Second)
		fmt.Println("msg 009")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{31, 32, 33, 34, 35})
		time.Sleep(time.Second)
		fmt.Println("msg 010")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{41, 42, 43, 44, 45})
		time.Sleep(time.Second)
		fmt.Println("msg 011")
		bot.SendMessage(binary.MsgTypeSwitchMaster, binary.MarshalStr8("23456"))
		time.Sleep(time.Second)
		fmt.Println("msg 012 (leave)")
		bot.SendMessage(binary.MsgTypeLeave, []byte{})
		time.Sleep(time.Second)
		bot.Close()
	}()

	<-done

	time.Sleep(3 * time.Second)
	fmt.Println("reconnect test after leave")
	err = bot.DialGame(room.Url, room.AuthKey, 4)
	if err != nil {
		log.Printf("dial game error: %v", err)
		return
	}

	done = make(chan bool)
	go bot.EventLoop(done)
	<-done
}

func spawnPlayer(roomId, userId string, queries []lobby.PropQuery) {
	bot := NewBot(appID, appKey, userId, binary.Dict{})

	room, err := bot.JoinRoom(roomId, queries)
	if err != nil {
		log.Printf("[bot:%v] join room error: %v", userId, err)
		return
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		log.Printf("[bot:%v] dial game error: %v", userId, err)
		return
	}

	done := make(chan bool)
	go bot.EventLoop(done)
}

func spawnWatcher(roomId, userId string) {
	bot := NewBot(appID, appKey, userId, binary.Dict{})

	room, err := bot.WatchRoom(roomId, nil)
	if err != nil {
		log.Printf("[bot:%v] watch room error: %v", userId, err)
		return
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		log.Printf("[bot:%v] dial watch error: %v", userId, err)
		return
	}

	done := make(chan bool)
	go bot.EventLoop(done)

	time.Sleep(time.Second)

	// MsgTypeToMaster
	bot.SendMessage(binary.MsgTypeToMaster, binary.MarshalStr8("MsgTypeToMaster from watcher"))
	time.Sleep(time.Second)

	// MsgTypeTargets: 存在するターゲットと存在しないターゲットに対してメッセージを送る
	payload := MarshalTargetsAndData([]byte{},
		[]string{"23456", "goblin"},
		binary.MarshalStr8("MsgTypeTargets from watcher"))
	bot.SendMessage(binary.MsgTypeTargets, payload)
	time.Sleep(time.Second)

	// MsgTypeBroadcast
	bot.SendMessage(binary.MsgTypeBroadcast, binary.MarshalStr8("MsgTypeBroadcast from watcher"))
	time.Sleep(time.Second)
}

func spawnPlayerByNumber(roomNumber int32, userId string, queries []lobby.PropQuery) {
	bot := NewBot(appID, appKey, userId, binary.Dict{})

	room, err := bot.JoinRoomByNumber(roomNumber, queries)
	if err != nil {
		log.Printf("[bot:%v] join room error: %v", userId, err)
		return
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		log.Printf("[bot:%v] dial game error: %v", userId, err)
		return
	}

	done := make(chan bool)
	go bot.EventLoop(done)
}

func spawnPlayerAtRandom(userId string, searchGroup uint32, queries []lobby.PropQuery) {
	bot := NewBot(appID, appKey, userId, binary.Dict{})
	room, err := bot.JoinRoomAtRandom(searchGroup, queries)
	if err != nil {
		log.Printf("[bot:%v] join room error: %v", userId, err)
		return
	}

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		log.Printf("[bot:%v] dial game error: %v", userId, err)
		return
	}

	done := make(chan bool)
	go bot.EventLoop(done)
}
