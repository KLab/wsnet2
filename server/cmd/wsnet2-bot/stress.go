package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"wsnet2/binary"

	"github.com/gorilla/websocket"
)

type stressBot struct {
	name string
}

func NewStressBot() *stressBot {
	return &stressBot{"stress"}
}

func (cmd *stressBot) Name() string {
	return cmd.name
}

func (cmd *stressBot) Execute(args []string) {
	n := 10000
	c := 10
	switch len(args) {
	case 2:
		c, _ = strconv.Atoi(args[1])
		fallthrough
	case 1:
		n, _ = strconv.Atoi(args[0])
	}
	logger.Infof("n=%v, c=%v", n, c)
	queue := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		queue <- struct{}{}
	}
	close(queue)
	wg := &sync.WaitGroup{}
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(mid int) {
			cmd.Run(mid, queue)
			wg.Done()
		}(i)
	}
	wg.Wait()
	logger.Info("stress bot finished.")
}

func (cmd *stressBot) Run(mid int, queue <-chan struct{}) {
	seq := 0
	for {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
		_, ok := <-queue
		if !ok {
			break
		}
		master, rid, done, err := SpawnMaster(fmt.Sprintf("master-%03d:%03d", mid, seq))
		if err != nil {
			continue
		}
		wgPlayers := &sync.WaitGroup{}
		for i := 0; i < 4; i++ {
			wgPlayers.Add(1)
			go func(cid int) {
				cmd.SpawnAndRunPlayer(rid, fmt.Sprintf("player-%03d:%03d-%03d", mid, seq, cid))
				wgPlayers.Done()
			}(i)
		}
		wgWatchers := &sync.WaitGroup{}
		for i := 0; i < rand.Intn(20); i++ {
			wgWatchers.Add(1)
			go func(cid int) {
				defer wgWatchers.Done()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				_, done, err := SpawnWatcher(rid, fmt.Sprintf("watcher-%03d:%03d-%03d", mid, seq, cid))
				if err != nil {
					return
				}
				<-done
			}(i)
		}
		wgPlayers.Wait()
		master.SendMessage(binary.MsgTypeLeave, []byte{})
		time.Sleep(time.Millisecond * time.Duration(100)) // MsgLeaveが処理される前にPeerが切断されるとGameにエラーログが出力されるので少し待つ
		master.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""))
		master.Close()
		<-done
		wgWatchers.Wait()
		seq++
	}
}

func (cmd *stressBot) SpawnAndRunPlayer(roomId, userId string) {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
	player, done, err := SpawnPlayer(roomId, userId, nil)
	if err != nil {
		return
	}
	play(player)
	player.SendMessage(binary.MsgTypeLeave, []byte{})
	time.Sleep(time.Millisecond * time.Duration(100)) // MsgLeaveが処理される前にPeerが切断されるとGameにエラーログが出力されるので少し待つ
	player.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""))
	player.Close()
	<-done
}

func play(player *bot) {
	end := time.NewTimer(time.Second * time.Duration(rand.Intn(5)))
	nxt := time.NewTimer(0)
	for {
		select {
		case <-end.C:
			return
		case <-nxt.C:
			player.SendMessage(binary.MsgTypeBroadcast, []byte{})
			nxt.Reset(time.Millisecond * time.Duration(rand.Intn(500)))
		}
	}
}
