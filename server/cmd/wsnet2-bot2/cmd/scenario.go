package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/lobby"
	"wsnet2/pb"

	"github.com/spf13/cobra"
)

const (
	ScenarioLobbySearchGroup = uint32(101) + iota
	ScenarioJoinRoomGroup
)

// scenarioCmd runs scenario test
//
// 機能テスト
var scenarioCmd = &cobra.Command{
	Use:   "scenario",
	Short: "Run scenario test",
	Long:  `Scenario test: 各機能をテストする`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScenario(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(scenarioCmd)

}

// runScenario runs scenario test
func runScenario(ctx context.Context) error {
	for _, test := range []func(context.Context) error{
		scenarioLobbySearch,
		scenarioJoinRoom,
	} {
		err := test(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func discardEvents(conn *client.Connection) {
	go func() {
		for range conn.Events() {
		}
	}()
}

func clearEventBuffer(conn *client.Connection) {
	for {
		select {
		case ev := <-conn.Events():
			logger.Debugf("discard event: %v", ev.Type())
		default:
			return
		}
	}
}

func waitEvent(conn *client.Connection, d time.Duration, evtypes ...binary.EvType) (binary.Event, bool) {
	t := time.NewTimer(d)
	for {
		select {
		case <-t.C:
			return nil, false

		case ev, ok := <-conn.Events():
			if !ok {
				return nil, false
			}
			for _, t := range evtypes {
				if ev.Type() == t {
					return ev, true
				}
			}
			logger.Debugf("discard event: %v", ev.Type())
		}
	}
}

// Lobbyの部屋検索のテスト
func scenarioLobbySearch(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	room1, conn1, err := createRoom(ctx, "lobbysearch_master", &pb.RoomOption{
		Visible:     true,
		Joinable:    true,
		Watchable:   true,
		SearchGroup: ScenarioLobbySearchGroup,
		PublicProps: binary.MarshalDict(binary.Dict{
			"key1": binary.MarshalInt(1024),
		}),
	})
	if err != nil {
		return fmt.Errorf("search: create room1: %w", err)
	}
	discardEvents(conn1)
	defer conn1.Wait(ctx)
	defer conn1.Leave("done")

	room2, conn2, err := createRoom(ctx, "lobbysearch_master", &pb.RoomOption{
		Visible:     true,
		Joinable:    false,
		Watchable:   false,
		SearchGroup: ScenarioLobbySearchGroup,
		PublicProps: binary.MarshalDict(binary.Dict{
			"key1": binary.MarshalInt(1025),
		}),
	})
	if err != nil {
		return fmt.Errorf("search: create room2: %w", err)
	}
	discardEvents(conn2)
	defer conn2.Wait(ctx)
	defer conn2.Leave("done")

	room3, conn3, err := createRoom(ctx, "lobbysearch_master", &pb.RoomOption{
		Visible:     false,
		SearchGroup: ScenarioLobbySearchGroup,
		PublicProps: binary.MarshalDict(binary.Dict{
			"key1": binary.MarshalInt(1024),
		}),
	})
	if err != nil {
		return fmt.Errorf("search: create room3: %w", err)
	}
	discardEvents(conn3)
	defer conn3.Wait(ctx)
	defer conn3.Leave("done")

	logger.Infof("lobby-search: room1=%v room2=%v room3=%v", room1.Id, room2.Id, room3.Id)
	time.Sleep(time.Second)

	for name, cond := range map[string]struct {
		query  *client.Query
		expect []string
	}{
		"key==1024": {
			query:  client.NewQuery().Equal("key1", binary.MarshalInt(1024)),
			expect: []string{room1.Id},
		},
		"key!=1024": {
			query:  client.NewQuery().Not("key1", binary.MarshalInt(1024)),
			expect: []string{room2.Id},
		},
		"key1<1024": {
			query:  client.NewQuery().LessThan("key1", binary.MarshalInt(1024)),
			expect: []string{},
		},
		"key1<=1024": {
			query:  client.NewQuery().LessEqual("key1", binary.MarshalInt(1024)),
			expect: []string{room1.Id},
		},
		"key1>1024": {
			query:  client.NewQuery().GreaterThan("key1", binary.MarshalInt(1024)),
			expect: []string{room2.Id},
		},
		"key1>=1024": {
			query:  client.NewQuery().GreaterEqual("key1", binary.MarshalInt(1024)),
			expect: []string{room1.Id, room2.Id},
		},
	} {
		param := &lobby.SearchParam{
			SearchGroup: ScenarioLobbySearchGroup,
			Queries:     *cond.query,
		}

		rooms, err := searchRooms(ctx, "searcher", param)
		if err != nil && !errors.Is(err, client.ErrNoRoomFound) {
			return fmt.Errorf("search[%v]: %w", name, err)
		}

		ids := make([]string, 0, len(rooms))
		cnts := make(map[string]int)
		for _, r := range rooms {
			ids = append(ids, r.Id)
			cnts[r.Id]++
		}

		logger.Infof("search[%v] %v", name, ids)

		for _, expid := range cond.expect {
			cnts[expid]--
		}
		for _, c := range cnts {
			if c != 0 {
				return fmt.Errorf("search[%v] wants: %v", name, cond.expect)
			}
		}
	}

	return nil
}

func scenarioJoinRoom(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	room, conn, err := createRoom(ctx, "joinroom_master", &pb.RoomOption{
		Joinable:    true,
		Watchable:   true,
		SearchGroup: ScenarioJoinRoomGroup,
		WithNumber:  true,
		MaxPlayers:  3,
	})
	if err != nil {
		return fmt.Errorf("join-room: create: %w", err)
	}
	logger.Infof("join-room: %v", room.Id)
	defer conn.Wait(ctx)
	defer conn.Leave("done")

	// 正常入室
	_, p1, err := joinRoom(ctx, "joinroom_player1", room.Id, nil)
	if err != nil {
		return fmt.Errorf("join-room: player1: %w", err)
	}
	logger.Infof("join-rooom: player1 ok")
	discardEvents(p1)
	defer p1.Wait(ctx)
	defer p1.Leave("done")

	clearEventBuffer(conn)

	// 正常入室
	_, p2, err := joinRoom(ctx, "joinroom_player2", room.Id, nil)
	if err != nil {
		return fmt.Errorf("join-room: player2: %w", err)
	}
	logger.Infof("join-rooom: player2 ok")
	discardEvents(p2)
	defer p2.Wait(ctx)
	defer p2.Leave("done")

	clearEventBuffer(conn)

	// 満室のためエラー
	_, p3, err := joinRoom(ctx, "joinroom_player3", room.Id, nil)
	if !errors.Is(err, client.ErrRoomFull) {
		if p3 != nil {
			p3.Leave("done")
			p3.Wait(ctx)
		}
		return fmt.Errorf("join-room: player3 wants RoomFull: %v", err)
	}
	logger.Infof("join-rooom: player3 ok (room full)")

	clearEventBuffer(conn)

	// 満室でも観戦は可能
	_, w1, err := watchRoom(ctx, "joinroom_watcher1", room.Id, nil)
	if err != nil {
		return fmt.Errorf("join-room: watcher1: %w", err)
	}
	logger.Infof("join-room: watcher1 ok")
	discardEvents(w1)
	defer w1.Wait(ctx)
	defer w1.Leave("done")

	clearEventBuffer(conn)

	// MaxPlayerを+2増やしwatchable=falseに
	err = conn.Send(binary.MsgTypeRoomProp, binary.MarshalRoomPropPayload(
		true, true, false, ScenarioJoinRoomGroup, 5, 0, nil, nil))
	if err != nil {
		return fmt.Errorf("join-room: roomprop: %w", err)
	}
	_, ok := waitEvent(conn, time.Second, binary.EvTypeRoomProp)
	if !ok {
		return fmt.Errorf("join-room: wait EvRoomProp failed")
	}
	time.Sleep(time.Second) // DBへの書き込みが非同期なのでちょっと待つ

	// 入室可能
	_, p4, err := joinRoom(ctx, "joinroom_player4", room.Id, nil)
	if err != nil {
		return fmt.Errorf("join-room: player4: %w", err)
	}
	logger.Infof("join-rooom: player4 ok")
	discardEvents(p4)
	defer p4.Wait(ctx)
	defer p4.Leave("done")

	// 観戦はエラー
	_, w2, err := watchRoom(ctx, "joinroom_watcher2", room.Id, nil)
	if !errors.Is(err, client.ErrNoRoomFound) {
		if w2 != nil {
			w2.Leave("done")
			w2.Wait(ctx)
		}
		return fmt.Errorf("join-room: watcher2 wants NoRoomFound: %v", err)
	}
	logger.Infof("join-rooom: watcher2 ok (no room found)")

	clearEventBuffer(conn)

	// joinable=falseに
	err = conn.Send(binary.MsgTypeRoomProp, binary.MarshalRoomPropPayload(
		true, false, false, ScenarioJoinRoomGroup, 5, 0, nil, nil))
	if err != nil {
		return fmt.Errorf("join-room: roomprop: %w", err)
	}
	_, ok = waitEvent(conn, time.Second, binary.EvTypeRoomProp)
	if !ok {
		return fmt.Errorf("join-room: wait EvRoomProp failed")
	}
	time.Sleep(time.Second)

	// 入室できない
	_, p5, err := joinRoom(ctx, "joinroom_player5", room.Id, nil)
	if !errors.Is(err, client.ErrNoRoomFound) {
		if p5 != nil {
			p5.Leave("done")
			p5.Wait(ctx)
		}
		return fmt.Errorf("join-room: player5 wants NoRoomFound: %v", err)
	}
	logger.Infof("join-rooom: player5 ok (no room found)")

	discardEvents(conn)

	return nil
}
