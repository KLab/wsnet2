package cmd

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"
	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/lobby"
	"wsnet2/pb"

	"github.com/spf13/cobra"
)

// scenarioCmd runs scenario test
//
// シナリオテスト（機能テスト）
//   - 部屋検索
//   - 入室
//   - メッセージ送信
//   - Kick
var scenarioCmd = &cobra.Command{
	Use:   "scenario",
	Short: "Run scenario test",
	Long:  `Scenario test: 各機能をテストする`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScenario(cmd.Context())
	},
}

var scenarios = map[string]func(context.Context) error{
	"LobbySearch":   scenarioLobbySearch,
	"JoinRoom":      scenarioJoinRoom,
	"Message":       scenarioMessage,
	"Kick":          scenarioKick,
	"SearchCurrent": scenarioSearchCurrent,
}

func init() {
	rootCmd.AddCommand(scenarioCmd)
}

// runScenario runs scenario test
func runScenario(ctx context.Context) error {
	for _, scenario := range scenarios {
		err := scenario(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// discardEvents : 以降すべてのEventを読み捨てる
func discardEvents(conn *client.Connection) {
	go func() {
		for range conn.Events() {
		}
	}()
}

// clearEventBuffer : 現在バッファに溜まっているイベントを読み捨てる
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

// waitEvent : 指定時間の間、指定したタイプのEventが来るのを待つ。それ以外のEventは読み捨てる。
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

// cleanupConn : Leaveメッセージを送り退室するのを待つ
func cleanupConn(ctx context.Context, conn *client.Connection) {
	if conn == nil {
		return
	}
	discardEvents(conn)
	if err := conn.Leave("done"); err != nil {
		logger.Debugf("cleanupConn(%v): %v", conn.UserId(), err)
		return
	}
	if _, err := conn.Wait(ctx); err != nil {
		logger.Debugf("cleanupConn(%v): %v", conn.UserId(), err)
	}
}

// scenarioLobbySearch : Lobbyの部屋検索のテスト
func scenarioLobbySearch(ctx context.Context) error {
	logger.Infof("=== Scenario Lobby Search ===")
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
	defer cleanupConn(ctx, conn1)

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
	defer cleanupConn(ctx, conn2)

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
	defer cleanupConn(ctx, conn3)

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

// scenarioJoinRoom : 入室テスト
func scenarioJoinRoom(ctx context.Context) error {
	logger.Infof("=== Scenario Join Room ===")
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
	defer cleanupConn(ctx, conn)

	// 正常入室
	_, p1, err := joinRoom(ctx, "joinroom_player1", room.Id, nil)
	if err != nil {
		return fmt.Errorf("join-room: player1: %w", err)
	}
	logger.Infof("join-rooom: player1 ok")
	discardEvents(p1)
	defer cleanupConn(ctx, p1)

	clearEventBuffer(conn)

	// 正常入室
	_, p2, err := joinRoom(ctx, "joinroom_player2", room.Id, nil)
	if err != nil {
		return fmt.Errorf("join-room: player2: %w", err)
	}
	logger.Infof("join-rooom: player2 ok")
	discardEvents(p2)
	defer cleanupConn(ctx, p2)

	clearEventBuffer(conn)

	// 満室のためエラー
	_, p3, err := joinRoom(ctx, "joinroom_player3", room.Id, nil)
	if !errors.Is(err, client.ErrRoomFull) {
		cleanupConn(ctx, p3)
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
	defer cleanupConn(ctx, w1)

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
	logger.Infof("join-room: player4 ok")
	discardEvents(p4)
	defer cleanupConn(ctx, p4)

	// 観戦はエラー
	_, w2, err := watchRoom(ctx, "joinroom_watcher2", room.Id, nil)
	if !errors.Is(err, client.ErrNoRoomFound) {
		cleanupConn(ctx, w2)
		return fmt.Errorf("join-room: watcher2 wants NoRoomFound: %v", err)
	}
	logger.Infof("join-room: watcher2 ok (no room found)")

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
		cleanupConn(ctx, p5)
		return fmt.Errorf("join-room: player5 wants NoRoomFound: %v", err)
	}
	logger.Infof("join-room: player5 ok (no room found)")

	return nil
}

func checkEvMessage(conn *client.Connection, sndrptn string, payload []byte) error {
	logger.Debugf("  check %v", conn.UserId())

	ev, ok := waitEvent(conn, time.Second, binary.EvTypeMessage)
	if !ok {
		return fmt.Errorf("%v: no message event", conn.UserId())
	}

	sndr, pl, err := binary.UnmarshalEvMessage(ev.Payload())
	if err != nil {
		return err
	}
	if ok, _ := regexp.MatchString(sndrptn, sndr); !ok {
		return fmt.Errorf("%v: sender mismatch: %v wants %v", conn.UserId(), sndr, sndrptn)
	}
	if !reflect.DeepEqual(pl, payload) {
		return fmt.Errorf("%v: payload mismatch: %v wants %v", conn.UserId(), pl, payload)
	}
	return nil
}

func checkNoEvMessage(conn *client.Connection) error {
	logger.Debugf("  check %v", conn.UserId())
	ev, ok := waitEvent(conn, time.Second, binary.EvTypeMessage)
	if ok {
		return fmt.Errorf("%v: unexpect message: %v", conn.UserId(), ev)
	}
	return nil
}

// scenarioMessage : メッセージ送信テスト
func scenarioMessage(ctx context.Context) error {
	logger.Infof("=== Scenario Message ===")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// master, player, watcherの3人
	room, master, err := createRoom(ctx, "message_master", &pb.RoomOption{
		Joinable:    true,
		Watchable:   true,
		SearchGroup: ScenarioMessageGroup,
	})
	if err != nil {
		return fmt.Errorf("message: create: %w", err)
	}
	defer cleanupConn(ctx, master)
	_, player, err := joinRoom(ctx, "message_player", room.Id, nil)
	if err != nil {
		return fmt.Errorf("message: join: %w", err)
	}
	defer cleanupConn(ctx, player)
	_, watcher, err := watchRoom(ctx, "message_watcher", room.Id, nil)
	if err != nil {
		return fmt.Errorf("message: watch: %w", err)
	}
	defer cleanupConn(ctx, watcher)

	clearEventBuffer(master)
	clearEventBuffer(player)
	clearEventBuffer(watcher)

	// masterからのbroadcastが全員に届くこと
	logger.Debugf("broadcast from master")
	payload := binary.MarshalStr8("bloadcast from master")
	master.Broadcast(payload)

	err = checkEvMessage(master, master.UserId(), payload)
	if err != nil {
		return fmt.Errorf("bloadcast from master: %w", err)
	}
	err = checkEvMessage(player, master.UserId(), payload)
	if err != nil {
		return fmt.Errorf("bloadcast from master: %w", err)
	}
	err = checkEvMessage(watcher, master.UserId(), payload)
	if err != nil {
		return fmt.Errorf("bloadcast from master: %w", err)
	}

	logger.Infof("broadcast from master: ok")
	clearEventBuffer(master)
	clearEventBuffer(player)
	clearEventBuffer(watcher)

	// masterからのtoTargetsがplayerのみに届きmaster,watcherに届かないこと
	logger.Debugf("message to targets")

	payload = binary.MarshalStr8("message to targets")
	master.ToTargets(payload, player.UserId())

	err = checkEvMessage(player, master.UserId(), payload)
	if err != nil {
		return fmt.Errorf("message to targets: %w", err)
	}
	err = checkNoEvMessage(master)
	if err != nil {
		return fmt.Errorf("message to targets: %w", err)
	}
	err = checkNoEvMessage(watcher)
	if err != nil {
		return fmt.Errorf("message to targets: %w", err)
	}

	logger.Infof("message to targets: ok")
	clearEventBuffer(master)
	clearEventBuffer(player)
	clearEventBuffer(watcher)

	// playerからのtoMasterがmasterのみに届くこと
	logger.Debugf("message to master")

	payload = binary.MarshalStr8("message to master")
	player.ToMaster(payload)

	err = checkEvMessage(master, player.UserId(), payload)
	if err != nil {
		return fmt.Errorf("message to master: %w", err)
	}
	err = checkNoEvMessage(player)
	if err != nil {
		return fmt.Errorf("message to master: %w", err)
	}
	err = checkNoEvMessage(watcher)
	if err != nil {
		return fmt.Errorf("message to master: %w", err)
	}

	logger.Infof("message to master: ok")
	clearEventBuffer(master)
	clearEventBuffer(player)
	clearEventBuffer(watcher)

	// watcherからもbroadcast/toMasterができること
	logger.Debugf("message from watcher")

	payload = binary.MarshalStr8("message from watcher")
	watcher.Broadcast(payload)

	err = checkEvMessage(master, "hub:.*", payload)
	if err != nil {
		return fmt.Errorf("message from watcher: %w", err)
	}
	err = checkEvMessage(player, "hub:.*", payload)
	if err != nil {
		return fmt.Errorf("message from watcher: %w", err)
	}
	err = checkEvMessage(watcher, "hub:.*", payload)
	if err != nil {
		return fmt.Errorf("message from watcher: %w", err)
	}

	watcher.ToMaster(payload)

	err = checkEvMessage(master, "hub:.*", payload)
	if err != nil {
		return fmt.Errorf("message from watcher: %w", err)
	}
	err = checkNoEvMessage(player)
	if err != nil {
		return fmt.Errorf("message from watcher: %w", err)
	}
	err = checkNoEvMessage(watcher)
	if err != nil {
		return fmt.Errorf("message from watcher: %w", err)
	}

	logger.Infof("message from watcher: ok")
	clearEventBuffer(master)
	clearEventBuffer(player)
	clearEventBuffer(watcher)

	// master交代 -> player
	logger.Debugf("switch master")

	master.SwitchMaster(player.UserId())

	for _, conn := range []*client.Connection{master, player, watcher} {
		logger.Debugf("  check %v", conn.UserId())
		_, ok := waitEvent(conn, time.Second, binary.EvTypeMasterSwitched)
		if !ok {
			return fmt.Errorf("switch master: %v: no master-switched event", conn.UserId())
		}
	}

	logger.Infof("switch master: ok")
	clearEventBuffer(master)
	clearEventBuffer(player)
	clearEventBuffer(watcher)

	// masterからのtoMasterがplayerに届きmasterに届かないこと
	logger.Debugf("message to master")

	payload = binary.MarshalStr8("message to master")
	master.ToMaster(payload)

	err = checkEvMessage(player, master.UserId(), payload)
	if err != nil {
		return fmt.Errorf("message to master: %w", err)
	}
	err = checkNoEvMessage(master)
	if err != nil {
		return fmt.Errorf("message to master: %w", err)
	}
	err = checkNoEvMessage(watcher)
	if err != nil {
		return fmt.Errorf("message to master: %w", err)
	}

	return nil
}

// scenarioKick : Kickのテスト
func scenarioKick(ctx context.Context) error {
	logger.Infof("=== Scenario Kick ===")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// master, player
	// master, player, watcherの3人
	room, master, err := createRoom(ctx, "kick_master", &pb.RoomOption{
		Joinable:    true,
		Watchable:   true,
		SearchGroup: ScenarioKickGroup,
	})
	if err != nil {
		return fmt.Errorf("kick: create: %w", err)
	}
	defer cleanupConn(ctx, master)
	_, player, err := joinRoom(ctx, "kick_player", room.Id, nil)
	if err != nil {
		return fmt.Errorf("kick: join: %w", err)
	}

	_, _ = waitEvent(player, time.Second, binary.EvTypePeerReady)
	clearEventBuffer(master)
	clearEventBuffer(player)

	// playerをkick、一定時間内にplayerが終了すること

	msg := "scenario kick"
	master.Kick(player.UserId(), msg)

	type msgerr struct {
		msg string
		err error
	}
	ch := make(chan msgerr, 1)
	go func() {
		msg, err := player.Wait(ctx)
		ch <- msgerr{msg, err}
	}()

	select {
	case <-time.NewTimer(time.Second).C:
		cleanupConn(ctx, player)
		return fmt.Errorf("kick: player is not kicked")
	case r := <-ch:
		if r.err != nil {
			return fmt.Errorf("kick: %w", err)
		}
		if r.msg != msg {
			return fmt.Errorf("kick: msg %v, wants %v", r.msg, msg)
		}
	}

	logger.Infof("kick ok")

	return nil
}

// scenarioSearchCurrent : 入室している部屋一覧取得のテスト
func scenarioSearchCurrent(ctx context.Context) error {
	logger.Infof("=== Scenario SearchCurrent ===")

	p1id := "searchcurrent_p1"

	// room1: masterとして入室
	// room2: 一般playerとして入室
	// room3: 観戦入室
	// => [room1, room2]

	room1, conn1, err := createRoom(ctx, p1id, &pb.RoomOption{
		SearchGroup: ScenarioSearchCurrent,
	})
	if err != nil {
		return fmt.Errorf("search: create room1: %w", err)
	}
	discardEvents(conn1)
	defer cleanupConn(ctx, conn1)
	logger.Infof("room1=%v", room1.Id)

	time.Sleep(time.Second)

	room2, conn2, err := createRoom(ctx, "searchcurrent_p2", &pb.RoomOption{
		Joinable:    true,
		SearchGroup: ScenarioSearchCurrent,
	})
	if err != nil {
		return fmt.Errorf("search: create room2: %w", err)
	}
	discardEvents(conn2)
	defer cleanupConn(ctx, conn2)
	logger.Infof("room2=%v", room2.Id)

	time.Sleep(time.Second)

	room3, conn3, err := createRoom(ctx, "searchcurrent_p3", &pb.RoomOption{
		Watchable:   true,
		SearchGroup: ScenarioSearchCurrent,
	})
	if err != nil {
		return fmt.Errorf("search: create room2: %w", err)
	}
	discardEvents(conn3)
	defer cleanupConn(ctx, conn3)
	logger.Infof("room3=%v", room3.Id)

	_, conn4, err := joinRoom(ctx, p1id, room2.Id, nil)
	if err != nil {
		return fmt.Errorf("search: join room2: %w", err)
	}
	discardEvents(conn4)
	defer cleanupConn(ctx, conn4)

	_, conn5, err := watchRoom(ctx, p1id, room3.Id, nil)
	if err != nil {
		return fmt.Errorf("search: watch room3: %w", err)
	}
	discardEvents(conn5)
	defer cleanupConn(ctx, conn5)

	rooms, err := searchCurrent(ctx, p1id)
	if err != nil {
		return fmt.Errorf("searchCurrent: %w", err)
	}

	ids := make([]string, len(rooms))
	for i, r := range rooms {
		ids[i] = r.Id
	}
	logger.Infof("found: %v", ids)

	wants := []string{room1.Id, room2.Id}
	if !reflect.DeepEqual(ids, wants) {
		return fmt.Errorf("rooms %v, wants %v", ids, wants)
	}

	return nil
}
