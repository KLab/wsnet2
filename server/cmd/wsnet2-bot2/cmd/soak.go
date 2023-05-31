package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/pb"
)

const (
	SearchGroup = 10

	RttThreshold = 30 // millisecond
)

var (
	roomCount   int
	minLifeTime time.Duration
	maxLifeTime time.Duration
)

// soakCmd runs soak test
var soakCmd = &cobra.Command{
	Use:   "soak",
	Short: "Run soak test",
	Long:  `Run soak test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSoak(cmd.Context(), roomCount, minLifeTime, maxLifeTime)
	},
}

func init() {
	rootCmd.AddCommand(soakCmd)

	soakCmd.Flags().IntVarP(&roomCount, "room-count", "c", 10, "Room count")
	soakCmd.Flags().DurationVarP(&minLifeTime, "min-life-time", "m", 10*time.Minute, "Minimum life time")
	soakCmd.Flags().DurationVarP(&maxLifeTime, "max-life-time", "M", 20*time.Minute, "Maximum life time")
}

// runSoak runs soak test
func runSoak(ctx context.Context, roomCount int, minLifeTime, maxLifeTime time.Duration) error {
	if roomCount < 1 {
		return fmt.Errorf("room count must be greater than 0")
	}
	if minLifeTime > maxLifeTime {
		return fmt.Errorf("min life time must be less than max life time")
	}
	lifetimeRange := int(maxLifeTime - minLifeTime)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ech := make(chan error)
	counter := make(chan struct{}, roomCount)

	var wg sync.WaitGroup
	for n := 0; ; n++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-ech:
			cancel()
			wg.Wait()
			return err
		case counter <- struct{}{}:
		}

		wg.Add(1)
		go func() {
			lifetime := minLifeTime
			if lifetimeRange != 0 {
				lifetime += time.Duration(rand.Intn(lifetimeRange))
			}
			err := runRoom(ctx, n, lifetime)
			if err != nil {
				ech <- err
			}
			wg.Done()
			<-counter
		}()

		time.Sleep(time.Second)
	}
}

// runRoom runs a room
func runRoom(ctx context.Context, n int, lifetime time.Duration) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	roomOpt := &pb.RoomOption{
		Visible:        true,
		Joinable:       true,
		Watchable:      true,
		SearchGroup:    SearchGroup,
		MaxPlayers:     10,
		ClientDeadline: 25,
		PublicProps: binary.MarshalDict(binary.Dict{
			"room":  binary.MarshalStr8(fmt.Sprintf("soak-%d", n)),
			"score": binary.MarshalInt(0),
		}),
	}
	masterId := fmt.Sprintf("master-%d", n)
	eci, err := client.GenAccessInfo(lobbyURL, appId, appKey, masterId)
	if err != nil {
		return err
	}

	room, master, err := client.Create(ctx, eci, roomOpt, &pb.ClientInfo{Id: masterId}, nil)
	if err != nil {
		return err
	}
	log.Printf("room[%d]: start %v lifetime=%v", n, room.Id, lifetime)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		runMaster(ctx, n, master, lifetime)
		wg.Done()
	}()

	time.Sleep(time.Second) // wait for refleshing cache of the lobby

	for i := 0; i < 2; i++ {
		playerId := fmt.Sprintf("player-%v-%v", n, i)
		aci, err := client.GenAccessInfo(lobbyURL, appId, appKey, playerId)
		if err != nil {
			return err
		}

		q := client.NewQuery()
		q.Equal("name", room.PublicProps["name"])

		_, player, err := client.RandomJoin(ctx, aci, SearchGroup, q, &pb.ClientInfo{Id: playerId}, nil)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			runPlayer(ctx, player, n, playerId, masterId)
			wg.Done()
		}()
	}

	for i := 0; i < 5; i++ {
		watcherId := fmt.Sprintf("watcher-%v-%v", n, i)
		aci, err := client.GenAccessInfo(lobbyURL, appId, appKey, watcherId)
		if err != nil {
			return err
		}

		_, watcher, err := client.Watch(ctx, aci, room.Id, nil, nil)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			runWatcher(ctx, watcher)
			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}

// runMaster runs a master
func runMaster(ctx context.Context, n int, conn *client.Connection, lifetime time.Duration) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(lifetime):
			conn.Send(binary.MsgTypeLeave, binary.MarshalLeavePayload("done"))
			cancel()
		}
	}()

	sender := func() {
		// goroutine1: 1500byteを0.2秒間隔で5秒(25回)、4000byteを1秒間隔で5回 broadcast
		go func() {
			for {
				for i := 0; i < 25; i++ {
					select {
					case <-ctx.Done():
						return
					default:
					}
					conn.Send(binary.MsgTypeBroadcast, msgBody[:1500])
					time.Sleep(200 * time.Millisecond)
				}
				for i := 0; i < 5; i++ {
					select {
					case <-ctx.Done():
						return
					default:
					}
					conn.Send(binary.MsgTypeBroadcast, msgBody[:4000])
					time.Sleep(time.Second)
				}
			}
		}()
		// goroutine2: 30~60byteをランダムに毎秒 broadcast
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				conn.Send(binary.MsgTypeBroadcast, msgBody[:rand.Intn(30)+30])
				time.Sleep(time.Second)
			}
		}()
		// groutine3: 5秒に1回PublicPropを書きかえ
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				conn.Send(binary.MsgTypeRoomProp, binary.MarshalRoomPropPayload(
					true, true, true, SearchGroup, 10, 0,
					binary.Dict{"score": binary.MarshalInt(rand.Intn(1024))}, binary.Dict{}))
				time.Sleep(5 * time.Second)
			}
		}()
	}

	rttSum := int64(0)
	rttMax := int64(0)
	rttCnt := 0

	go func() {
		for ev := range conn.Events() {
			switch ev.Type() {
			case binary.EvTypePeerReady:
				sender()

			case binary.EvTypePong:
				p, _ := binary.UnmarshalEvPongPayload(ev.Payload())
				rtt := time.Now().UnixMilli() - int64(p.Timestamp)
				if rtt > RttThreshold {
					log.Printf("room[%d]: master rtt=%d", n, rtt)
				}
				rttSum += rtt
				rttCnt++
				if rttMax < rtt {
					rttMax = rtt
				}

			case binary.EvTypeLeft:
				p, _ := binary.UnmarshalEvLeftPayload(ev.Payload())
				log.Printf("room[%d]: player %v left: %v", n, p.ClientId, p.Cause)
			}
		}
	}()

	_, err := conn.Wait(ctx)
	if err != nil {
		log.Printf("room[%d]: master error: %v", n, err)
	}

	avg := float64(rttSum) / float64(rttCnt)
	log.Printf("room[%d]: end RTT sum=%v cnt=%v avg=%v max=%v", n, rttSum, rttCnt, avg, rttMax)
}

func runPlayer(ctx context.Context, conn *client.Connection, n int, myId, masterId string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sender := func() {
		// goroutine1: 1500byteを0.2秒間隔で5秒(25回)、4000byteを1秒間隔で5回 ToMaster
		go func() {
			for {
				for i := 0; i < 25; i++ {
					select {
					case <-ctx.Done():
						return
					default:
					}
					conn.Send(binary.MsgTypeToMaster, msgBody[:1500])
					time.Sleep(200 * time.Millisecond)
				}
				for i := 0; i < 5; i++ {
					select {
					case <-ctx.Done():
						return
					default:
					}
					conn.Send(binary.MsgTypeToMaster, msgBody[:4000])
					time.Sleep(time.Second)
				}
			}
		}()
		// goroutine2: 30~60byteをランダムに毎秒 ToMaster
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				conn.Send(binary.MsgTypeToMaster, msgBody[:rand.Intn(30)+30])
				time.Sleep(time.Second)
			}
		}()
	}

	go func() {
		for ev := range conn.Events() {
			switch ev.Type() {
			case binary.EvTypePeerReady:
				sender()

			case binary.EvTypeLeft:
				p, err := binary.UnmarshalEvLeftPayload(ev.Payload())
				if err != nil {
					log.Printf("room[%v]: %v: UnmarshalEvLeftPayload: %v", n, myId, err)
					conn.Send(binary.MsgTypeLeave, binary.MarshalLeavePayload("done"))
					cancel()
				}

				if p.ClientId == masterId {
					conn.Send(binary.MsgTypeLeave, binary.MarshalLeavePayload("done"))
					cancel()
				}
			}
		}
	}()

	conn.Wait(ctx)
}

func runWatcher(ctx context.Context, conn *client.Connection) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sender := func() {
		// goroutine1: 30~60byteをランダムに10秒毎 ToMaster
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				conn.Send(binary.MsgTypeToMaster, msgBody[:rand.Intn(30)+30])
				time.Sleep(10 * time.Second)
			}
		}()
	}

	go func() {
		for ev := range conn.Events() {
			switch ev.Type() {
			case binary.EvTypePeerReady:
				sender()
			}
		}
	}()

	// 部屋が自然消滅するまで居続ける

	conn.Wait(ctx)
}
