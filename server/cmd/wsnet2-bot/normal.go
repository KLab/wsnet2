package main

import (
	"fmt"
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

func (cmd *normalBot) Execute(args []string) {
	bot := NewBot(appID, appKey, "12345", binary.Dict{})

	room, err := bot.CreateRoom(binary.Dict{"key1": binary.MarshalInt(1024), "key2": binary.MarshalStr16("hoge")})
	if err != nil {
		logger.Errorf("create room error: %v", err)
		return
	}

	var queries []lobby.PropQuery
	logger.Debug("key1 =")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	logger.Debug("key1 !")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpNot, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	logger.Debug("key1 <")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	logger.Debug("key1 <=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	logger.Debug("key1 >")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	logger.Debug("key1 >=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)

	err = bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		logger.Errorf("dial game error: %v", err)
		return
	}

	go bot.EventLoop()

	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpEqual, Val: binary.MarshalInt(1024)}}

	go SpawnPlayer(room.RoomInfo.Id, "23456", nil)
	go SpawnPlayer(room.RoomInfo.Id, "34567", queries)
	go SpawnPlayerByNumber(room.RoomInfo.Number.Number, "45678", nil)
	go SpawnPlayerByNumber(room.RoomInfo.Number.Number, "56789", queries)
	go SpawnPlayerAtRandom("67890", 1, queries)

	for i := 0; i < 5; i++ {
		go func(id int) {
			watcher, err := SpawnWatcher(room.RoomInfo.Id, fmt.Sprintf("watcher-%d", id))
			if err != nil {
				return
			}
			time.Sleep(time.Second)
			// MsgTypeToMaster
			watcher.SendMessage(binary.MsgTypeToMaster, binary.MarshalStr8("MsgTypeToMaster from watcher"))
			time.Sleep(time.Second)
			// MsgTypeTargets: 存在するターゲットと存在しないターゲットに対してメッセージを送る
			targets := []string{"23456", "goblin"}
			watcher.SendMessage(binary.MsgTypeTargets, MarshalTargetsAndData(targets, binary.MarshalStr8("MsgTypeTargets from watcher")))
			time.Sleep(time.Second)
			// MsgTypeBroadcast
			watcher.SendMessage(binary.MsgTypeBroadcast, binary.MarshalStr8("MsgTypeBroadcast from watcher"))
			time.Sleep(time.Second)
			watcher.Close()
			<-watcher.done
		}(i)
		time.Sleep(time.Millisecond * 10)
	}

	go func() {
		time.Sleep(time.Second * 1)
		SpawnPlayer(room.RoomInfo.Id, "78901", nil)
	}()
	go func() {
		time.Sleep(time.Second * 1)
		SpawnPlayer(room.RoomInfo.Id, "89012", nil)
	}()

	go func() {
		time.Sleep(time.Second * 2)
		logger.Debug("msg 001")
		bot.SendMessage(binary.MsgTypeSwitchMaster, binary.MarshalStr8("23456"))
		time.Sleep(time.Second)
		logger.Debug("msg 002")
		bot.SendMessage(binary.MsgTypeSwitchMaster, binary.MarshalStr8("34567"))
		time.Sleep(time.Second)
		logger.Debug("msg 003")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{1, 2, 3, 4, 5})
		time.Sleep(time.Second)
		logger.Debug("msg 004")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{11, 12, 13, 14, 15})
		time.Sleep(time.Second)
		logger.Debug("msg 005")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{21, 22, 23, 24, 25})
		//time.Sleep(time.Second)
		//logger.Debug("msg 003")
		//ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 21, 22, 23, 24, 25})
		//		time.Sleep(time.Second)
		//		ws.Close()
		time.Sleep(time.Second)
		logger.Debug("msg 006")
		props := binary.MarshalDict(binary.Dict{
			"p1": binary.MarshalUShort(20),
			"p2": []byte{},
		})
		bot.SendMessage(binary.MsgTypeClientProp, props)
	}()

	//	<-bot.done
	time.Sleep(8 * time.Second)

	logger.Debug("reconnect test")
	err = bot.DialGame(room.Url, room.AuthKey, 2)
	if err != nil {
		logger.Errorf("dial game error: %v", err)
		return
	}

	go bot.EventLoop()

	go SpawnPlayer(room.RoomInfo.Id, "99999", nil)
	go func() {
		time.Sleep(time.Second * 3)
		logger.Debug("msg 007")
		bot.SendMessage(binary.MsgTypeKick, binary.MarshalStr8("99999"))
		time.Sleep(time.Second)
		logger.Debug("msg 008")
		bot.SendMessage(binary.MsgTypeKick, binary.MarshalStr8("00000"))
		time.Sleep(time.Second)
		logger.Debug("msg 009")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{31, 32, 33, 34, 35})
		time.Sleep(time.Second)
		logger.Debug("msg 010")
		bot.SendMessage(binary.MsgTypeBroadcast, []byte{41, 42, 43, 44, 45})
		time.Sleep(time.Second)
		logger.Debug("msg 011")
		bot.SendMessage(binary.MsgTypeSwitchMaster, binary.MarshalStr8("23456"))
		time.Sleep(time.Second)
		logger.Debug("msg 012 (leave)")
		bot.SendMessage(binary.MsgTypeLeave, []byte{})
		time.Sleep(time.Second)
		bot.Close()
	}()

	<-bot.done

	time.Sleep(3 * time.Second)
	logger.Debug("reconnect test after leave")
	err = bot.DialGame(room.Url, room.AuthKey, 4)
	if err != nil {
		logger.Errorf("dial game error: %v", err)
		return
	}

	go bot.EventLoop()
	<-bot.done
}
