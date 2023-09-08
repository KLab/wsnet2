package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/lobby"

	"github.com/spf13/cobra"
)

const (
	ScenarioLobbySearchGroup = uint32(101)
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
	err := scenarioLobbySearch(ctx)

	return err
}

func discardEvents(conn *client.Connection) {
	go func() {
		for range conn.Events() {
		}
	}()
}

// Lobbyの部屋検索のテスト
func scenarioLobbySearch(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	room1, conn1, err := createRoom(
		ctx, "lobbysearch_owner", true, true, true, ScenarioLobbySearchGroup,
		binary.Dict{
			"key1": binary.MarshalInt(1024),
		})
	if err != nil {
		return err
	}
	discardEvents(conn1)

	room2, conn2, err := createRoom(
		ctx, "lobbysearch_owner", true, false, false, ScenarioLobbySearchGroup,
		binary.Dict{
			"key1": binary.MarshalInt(1025),
		})
	if err != nil {
		return err
	}
	discardEvents(conn2)

	room3, conn3, err := createRoom(
		ctx, "lobbysearch_owner", false, true, true, ScenarioLobbySearchGroup,
		binary.Dict{
			"key1": binary.MarshalInt(1024),
		})
	if err != nil {
		return err
	}
	discardEvents(conn3)

	logger.Infof("lobby-search: room1=%v room2=%v room3=%v", room1.Id, room2.Id, room3.Id)
	time.Sleep(time.Second)

	var serr error

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
		if err != nil {
			logger.Infof("%T %+v", err, err)
			serr = err
			break
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
				serr = fmt.Errorf("search[%v] wants: %v", name, cond.expect)
				break
			}
		}
	}

	conn1.Leave("done")
	conn2.Leave("done")
	conn3.Leave("done")
	_, err1 := conn1.Wait(ctx)
	_, err2 := conn2.Wait(ctx)
	_, err3 := conn3.Wait(ctx)
	return errors.Join(serr, err1, err2, err3)
}
