package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type watcherBot struct {
	name string
}

func NewWatcherBot() *watcherBot {
	return &watcherBot{"watcher"}
}

func (cmd *watcherBot) Name() string {
	return cmd.name
}

func (cmd *watcherBot) Execute(args []string) {
	rid := args[0]
	n := 1000
	if len(args) > 1 {
		n, _ = strconv.Atoi(args[1])
	}
	logger.Infof("rid==%v, n=%v", rid, n)
	pid := os.Getpid()
	wg := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			watcher, err := SpawnWatcher(rid, fmt.Sprintf("watcher-%s:%d:%03d", rid, pid, cid))
			if err != nil {
				return
			}
			<-watcher.done
		}(i)
	}
	wg.Wait()
	logger.Info("watcher bot finished.")
}
