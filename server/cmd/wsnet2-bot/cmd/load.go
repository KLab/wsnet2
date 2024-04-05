package cmd

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"wsnet2/pb"
)

var (
	loadRoomCount     int
	loadPlayers       int
	loadWatchers      int
	loadWithWatchable bool

	loadMinLifeTime time.Duration
	loadMaxLifeTime time.Duration
)

// loadCmd runs load test
//
// 負荷試験
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Run load test",
	Long:  `load test: 負荷試験`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLoad(cmd.Context(), loadRoomCount, loadPlayers, loadWatchers, loadWithWatchable, loadMinLifeTime, loadMaxLifeTime)
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)

	loadCmd.Flags().IntVarP(&loadRoomCount, "room-count", "c", 10, "Parallel room count")
	loadCmd.Flags().IntVarP(&loadPlayers, "players", "p", 2, "Players per room")
	loadCmd.Flags().IntVarP(&loadWatchers, "watchers", "w", 5, "Watchers per room")
	loadCmd.Flags().BoolVar(&loadWithWatchable, "with-watchable", false, "With watchable room")

	loadCmd.Flags().DurationVarP(&loadMinLifeTime, "min-life-time", "m", 10*time.Minute, "Minimum life time")
	loadCmd.Flags().DurationVarP(&loadMaxLifeTime, "max-life-time", "M", 20*time.Minute, "Maximum life time")
}

// runLoad runs load test
func runLoad(ctx context.Context, roomCount, players, watchers int, withWatchable bool, minLifeTime, maxLifeTime time.Duration) error {
	if roomCount < 1 {
		return fmt.Errorf("room count must be greater than 0")
	}
	if minLifeTime > maxLifeTime {
		return fmt.Errorf("min life time must be less than max life time")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT)
		logger.Infof("signal: %v", <-s)
		cancel()
	}()

	pid := os.Getpid()

	logger.Infof("load test: pid=%v roomCount=%v players=%v watchers=%v withWatchable=%v", pid, roomCount, players, watchers, withWatchable)

	var wg sync.WaitGroup
	wg.Add(roomCount)
	errch := make(chan error)
	errlist := make([]error, roomCount)

	if roomCount >= 1 && withWatchable {
		roomCount--
		go func(wid int) {
			err := runLoadWatchableRoom(ctx, pid, players, watchers, fmt.Sprintf("watchable[%v]", pid))
			if err != nil && !errors.Is(err, ctx.Err()) {
				errch <- err
				errlist[wid] = err
			}
			wg.Done()
		}(roomCount)
	}
	for i := 0; i < roomCount; i++ {
		go func(i int) {
			wid := fmt.Sprintf("%v-%v", pid, i)
			err := runLoadWorker(ctx, players, watchers, minLifeTime, maxLifeTime, wid)
			if err != nil && !errors.Is(err, ctx.Err()) {
				errch <- err
				errlist[i] = err
			}
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(errch)
	}()

	if err, ok := <-errch; ok {
		logger.Debugf("an error occured: %v", err)
		cancel()
	}

	for range errch {
	}

	logger.Info("load test finished")
	return errors.Join(append(errlist, ctx.Err())...)
}

func runLoadWatchableRoom(ctx context.Context, pid, p, w int, logprefix string) error {
	return runLoadRoom(ctx, p, w, LoadWatchableGroup, 0, fmt.Sprintf("%v-%v", pid, "w"), logprefix)
}

func runLoadWorker(ctx context.Context, p, w int, minLifeTime, maxLifeTime time.Duration, wid string) error {
	lifetimeRange := int(maxLifeTime - minLifeTime)
	n := 0
	for {
		time.Sleep(time.Millisecond * time.Duration(rand.IntN(100)))

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		lifetime := minLifeTime
		if lifetimeRange > 0 {
			lifetime += time.Duration(rand.IntN(lifetimeRange))
		}
		widsuffix := fmt.Sprintf("%v-%v", wid, n)
		logprefix := fmt.Sprintf("room[%v]", widsuffix)
		err := runLoadRoom(ctx, p, w, LoadSearchGroup, lifetime, widsuffix, logprefix)
		if err != nil {
			return fmt.Errorf("load worker %v: %w", widsuffix, err)
		}
		n++
	}
}

func runLoadRoom(ctx context.Context, p, w int, group uint32, lifetime time.Duration, cidsuffix, logprefix string) error {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	masterId := fmt.Sprintf("master-%v", cidsuffix)
	logger.Debugf("%s create %s", logprefix, masterId)
	room, master, err := createRoom(context.Background(), masterId, &pb.RoomOption{
		Visible:     true,
		Joinable:    true,
		Watchable:   true,
		SearchGroup: group,
	})
	if err != nil {
		return err
	}
	logger.Infof("%s start %s", logprefix, room.Id)

	wg.Add(1)
	go func() {
		rttSum, rttCnt, rttMax, avg := runMaster(ctx, master, lifetime, group, logprefix)
		logger.Infof("%s end RTT sum=%v cnt=%v avg=%v max=%v", logprefix, rttSum, rttCnt, avg, rttMax)
		wg.Done()
	}()

	time.Sleep(time.Second)

	for i := 0; i < p; i++ {
		playerId := fmt.Sprintf("player-%v-%v", cidsuffix, i)

		logger.Debugf("%s watch %s", logprefix, playerId)
		_, player, err := joinRoom(context.Background(), playerId, room.Id, nil)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			runPlayer(ctx, player, masterId, logprefix)
			wg.Done()
		}()
	}

	for i := 0; i < w; i++ {
		watcherId := fmt.Sprintf("watcher-%v-%v", cidsuffix, i)

		logger.Debugf("%s join %s", logprefix, watcherId)
		_, watcher, err := watchRoom(context.Background(), watcherId, room.Id, nil)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			runWatcher(ctx, watcher, logprefix)
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}
