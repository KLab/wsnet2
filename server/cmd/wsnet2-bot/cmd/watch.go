package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"wsnet2/lobby"
)

var (
	watchWatcherCount int
)

// watchCmd runs load watcher test
//
// 負荷試験 : Watcher大量投入
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Run watcher load test",
	Long:  "Watcher load test: Watcher大量投入負荷試験",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLoadWatcher(cmd.Context(), watchWatcherCount)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().IntVarP(&watchWatcherCount, "watchers", "w", 1000, "Watchers count")
}

func runLoadWatcher(ctx context.Context, watcherCount int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT)
		logger.Infof("signal: %v", <-s)
		cancel()
	}()

	pid := os.Getpid()

	logger.Infof("watcher load test: pid=%v watchers=%v", pid, watcherCount)

	rooms, err := searchRooms(ctx, "watcher-load-test", &lobby.SearchParam{
		SearchGroup:    LoadWatchableGroup,
		Limit:          1,
		CheckWatchable: true,
	})
	if err != nil {
		return fmt.Errorf("searchRooms: %w", err)
	}

	logger.Infof("room: %v", rooms[0].Id)

	var wg sync.WaitGroup
	wg.Add(watcherCount)
	errch := make(chan error)

	for i := 0; i < watcherCount; i++ {
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(900)+100) * time.Millisecond)

			watcherId := fmt.Sprintf("watcher-%v-%v", pid, i)
			logprefix := fmt.Sprintf("watcher[%v]", i)
			logger.Debugf("%v join %s", logprefix, watcherId)
			_, watcher, err := watchRoom(context.Background(), watcherId, rooms[0].Id, nil)
			if err != nil {
				errch <- err
				return
			}
			runWatcher(ctx, watcher, logprefix)
		}(i)
	}

	go func() {
		wg.Wait()
		close(errch)
	}()
	if e, ok := <-errch; ok {
		err = e
		logger.Infof("watcher stopped: %v", err)
		cancel()
	}
	for range errch {
	}

	logger.Info("watcher load test finished")
	return err
}
