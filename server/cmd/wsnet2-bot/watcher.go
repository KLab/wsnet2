package main

import (
	"fmt"
	"math"
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
	var slice []*bot
	var mu sync.Mutex
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			watcher, err := SpawnWatcher(rid, fmt.Sprintf("watcher-%s:%d:%03d", rid, pid, cid))
			if err != nil {
				return
			}
			mu.Lock()
			slice = append(slice, watcher)
			mu.Unlock()
			<-watcher.done
		}(i)
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			mu.Lock()
			stats := statics{min: math.MaxInt64}
			for _, watcher := range slice {
				watcher.muStat.Lock()
				if watcher.stat.received > 0 {
					stats.received += watcher.stat.received
					stats.sum += watcher.stat.sum
					stats.sum2 += watcher.stat.sum2
					if stats.min > watcher.stat.min {
						stats.min = watcher.stat.min
					}
					if stats.max < watcher.stat.max {
						stats.max = watcher.stat.max
					}
					watcher.stat = statics{min: math.MaxInt64}
				}
				watcher.muStat.Unlock()
			}
			mu.Unlock()
			if stats.received > 0 {
				avg := stats.sum / stats.received
				mdev := math.Sqrt(float64(stats.sum2/stats.received - avg*avg))
				logger.Infof("pong received: %d, min: %d.%03d ms, avg: %d.%03d ms, max: %d.%03d ms, mdev: %.3f ms",
					stats.received,
					stats.min/1000, stats.min%1000,
					avg/1000, avg%1000,
					stats.max/1000, stats.max%1000,
					mdev/1000)
			}
		}
	}()
	wg.Wait()
	logger.Info("watcher bot finished.")
}
