package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"wsnet2/binary"
)

type staticBot struct {
	name string
}

func NewStaticBot() *staticBot {
	return &staticBot{"static"}
}

func (cmd *staticBot) Name() string {
	return cmd.name
}

func (cmd *staticBot) Execute(args []string) {
	lifetime := 60 * 10
	switch len(args) {
	case 1:
		lifetime, _ = strconv.Atoi(args[0])
	}
	master, rid, err := SpawnMaster(fmt.Sprintf("master-999"))
	if err != nil {
		logger.Errorf("spawn master: %v", err)
		return
	}
	fmt.Println("RoomId:", rid)
	fmt.Println("LifeTime:", lifetime, "seconds")
	wg := &sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(cid int) {
			cmd.SpawnAndRunPlayer(rid, fmt.Sprintf("player-999-%03d", cid), time.Second*time.Duration(lifetime))
			wg.Done()
		}(i)
	}
	wg.Wait()
	master.LeaveAndClose()
	<-master.done
}

func (cmd *staticBot) SpawnAndRunPlayer(roomId, userId string, lifetime time.Duration) {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
	player, err := SpawnPlayer(roomId, userId, nil)
	if err != nil {
		return
	}
	cmd.play(player, lifetime)
	player.LeaveAndClose()
	<-player.done
}

func (cmd *staticBot) play(player *bot, lifetime time.Duration) {
	end := time.NewTimer(lifetime)
	done := make(chan struct{})
	// goroutine1: 1500byteを0.2秒間隔で5秒(25回)、4000byteを1秒間隔で5回
	go func() {
		nxt := time.NewTimer(0)
		for {
			for i := 0; i < 25; i++ {
				select {
				case <-done:
					return
				case <-nxt.C:
					player.SendMessage(binary.MsgTypeBroadcast, make([]byte, 1500))
					nxt.Reset(time.Millisecond * time.Duration(200))
				}
			}
			for i := 0; i < 5; i++ {
				select {
				case <-done:
					return
				case <-nxt.C:
					player.SendMessage(binary.MsgTypeBroadcast, make([]byte, 4000))
					nxt.Reset(time.Millisecond * time.Duration(1000))
				}
			}
		}
	}()
	// goroutine2: 30~60byteをランダムに毎秒
	go func() {
		nxt := time.NewTimer(0)
		for {
			select {
			case <-done:
				return
			case <-nxt.C:
				player.SendMessage(binary.MsgTypeBroadcast, make([]byte, rand.Intn(30)+30))
				nxt.Reset(time.Millisecond * time.Duration(1000))
			}
		}
	}()
	select {
	case <-end.C:
	case <-player.done:
	}
	close(done)
	return
}
