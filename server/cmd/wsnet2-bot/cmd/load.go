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

	"wsnet2/binary"
	"wsnet2/client"
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
	for i := range roomCount {
		time.Sleep(5 * time.Millisecond)
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

	playerIds := make([]string, 0, p)
	for n := range p {
		playerIds = append(playerIds, fmt.Sprintf("player-%v-%d", cidsuffix, n))
	}
	propKey := "cidsuffix"
	propVal := binary.MarshalStr8(cidsuffix)
	query := client.NewQuery().Equal(propKey, propVal)
	props := binary.MarshalDict(binary.Dict{propKey: propVal})
	ready := make(chan struct{})

	masterId := playerIds[0]
	logger.Debugf("%s create %s", logprefix, masterId)
	room, master, err := createRoom(context.Background(), masterId, &pb.RoomOption{
		Visible:     true,
		Joinable:    true,
		Watchable:   true,
		SearchGroup: group,
		PublicProps: props,
	})
	if err != nil {
		return err
	}
	logger.Infof("%s start %s", logprefix, room.Id)

	wg.Add(1)
	go func() {
		rttSum, rttCnt, rttMax, avg := runLoadMaster(ctx, master, lifetime, ready, playerIds, logprefix)
		logger.Infof("%s end RTT sum=%v cnt=%v avg=%v max=%v", logprefix, rttSum, rttCnt, avg, rttMax)
		wg.Done()
	}()

	time.Sleep(time.Second)

	for _, playerId := range playerIds[1:] {
		time.Sleep(5 * time.Millisecond)
		logger.Debugf("%s join %s", logprefix, playerId)
		_, player, err := joinRandom(context.Background(), playerId, group, query)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			runLoadPlayer(ctx, player, lifetime, ready, playerIds, logprefix)
			wg.Done()
		}()
	}

	for i := range w {
		watcherId := fmt.Sprintf("watcher-%v-%v", cidsuffix, i)

		logger.Debugf("%s watch %s", logprefix, watcherId)
		_, watcher, err := watchRoom(context.Background(), watcherId, room.Id, nil)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			runLoadWatcher(ctx, watcher, logprefix)
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}

// runLoadMaster runs master player for loadtest
//
// 1. 全員入ってくるまで待つ
//   - ready chanで通知
//
// 2. メッセージ送信
//   - size: 300±150b、5%の確率で+1kb
//   - freq: 12msg / 1sec
//   - type: broadcastとtargets(全員)を交互に
//
// 3. lifetime経過でLeave
func runLoadMaster(ctx context.Context, conn *client.Connection, lifetime time.Duration, ready chan struct{}, pids []string, logprefix string) (rttSum, rttCnt, rttMax int64, rttAvg float64) {
	logger.Debugf("%s %s start", logprefix, conn.UserId())
	sendctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		var c <-chan time.Time
		if lifetime > 0 {
			c = time.After(lifetime)
		}
		var msg string
		select {
		case <-ctx.Done():
			msg = "context done"
		case <-c:
			msg = "lifetime elapsed"
		}
		cancel()
		logger.Debugf("%v master leave", logprefix)
		conn.Leave(msg)
	}()

	go func() {
		defer wg.Done()

		for range len(pids) {
			_, ok := waitEvent(conn, lifetime, binary.EvTypeJoined)
			if !ok {
				return
			}
		}
		logger.Debugf("%v all players joined", logprefix)
		close(ready)

		for {
			ev, ok := waitEvent(conn, lifetime, binary.EvTypePong)
			if !ok {
				return
			}
			p, _ := binary.UnmarshalEvPongPayload(ev.Payload())
			rtt := time.Now().UnixMilli() - int64(p.Timestamp)
			if rtt > RttThreshold {
				logger.Warnf("%s master rtt=%d", logprefix, rtt)
			}
			rttSum += rtt
			rttCnt++
			if rttMax < rtt {
				rttMax = rtt
			}
		}
	}()

	go func() {
		defer wg.Done()

		select {
		case <-sendctx.Done():
			return
		case <-ready:
		}

		logger.Debugf("%v change room unjoinable", logprefix)
		conn.Send(binary.MsgTypeRoomProp, binary.MarshalRoomPropPayload(
			false, false, true, LoadSearchGroup, uint32(len(pids)), 0, nil, nil))

		tick := time.NewTicker(time.Second / 12)
		defer tick.Stop()

		broadcast := true
		for {
			select {
			case <-sendctx.Done():
				return
			case <-tick.C:
			}

			// 300±150、300周辺が高頻度になるように
			size := 150 + rand.IntN(101) + rand.IntN(101) + rand.IntN(101)
			if rand.IntN(20) == 0 { // 5%
				size += 1000
			}

			if broadcast {
				conn.Broadcast(msgBody[:size])
			} else {
				conn.ToTargets(msgBody[:size], pids...)
			}
			broadcast = !broadcast
		}
	}()

	msg, err := conn.Wait(ctx)
	if err != nil {
		logger.Errorf("%s %v error: %v", logprefix, conn.UserId(), err)
	}
	logger.Debugf("%s %v end: %v", logprefix, conn.UserId(), msg)

	wg.Wait()
	return rttSum, rttCnt, rttMax, float64(rttSum) / float64(rttCnt)
}

// runLoadPlayer runs player for loadtest
//
// 1. 全員入室を待つ
//   - ready chanを待つ
//
// 2. メッセージ送信
//   - size: 250±150b、5%の確率で+1kb
//   - freq: 10msg / 1sec
//   - type: broadcastとtargets(全員)を交互に
//
// 3. 誰かが抜けたら終了
//   - 正常系ならMasterが最初に抜ける
func runLoadPlayer(ctx context.Context, conn *client.Connection, lifetime time.Duration, ready chan struct{}, pids []string, logprefix string) {
	logger.Debugf("%s %s start", logprefix, conn.UserId())

	sendctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		_, ok := waitEvent(conn, lifetime, binary.EvTypeLeft)
		cancel()
		if !ok {
			return
		}
		conn.Leave("leave")
		discardEvents(conn)
	}()

	go func() {
		select {
		case <-sendctx.Done():
			return
		case <-ready:
		}

		tick := time.NewTicker(time.Second / 10)
		defer tick.Stop()

		broadcast := true
		for {
			select {
			case <-sendctx.Done():
				return
			case <-tick.C:
			}

			// 200±150
			size := 50 + rand.IntN(101) + rand.IntN(101) + rand.IntN(101)
			if rand.IntN(20) == 0 { // 5%
				size += 1000
			}

			logger.Debugf("%v msgsize = %v", logprefix, size)

			if broadcast {
				conn.Broadcast(msgBody[:size])
			} else {
				conn.ToTargets(msgBody[:size], pids...)
			}
			broadcast = !broadcast
		}
	}()

	msg, err := conn.Wait(ctx)
	if err != nil {
		logger.Errorf("%s %v error: %v", logprefix, conn.UserId(), err)
	}
	logger.Debugf("%s %v end: %v", logprefix, conn.UserId(), msg)
}

// runLoadWatcher runs watcher for loadtest
//
// pattern: same as soak test
func runLoadWatcher(ctx context.Context, conn *client.Connection, logprefix string) {
	runSoakWatcher(ctx, conn, logprefix)
}
