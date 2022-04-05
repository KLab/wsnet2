package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var (
	oldroomsAt    string
	oldroomsLimit int
)

// oldroomsCmd represents the oldrooms command
var oldroomsCmd = &cobra.Command{
	Use:   "oldrooms",
	Short: "Show closed room list",
	Long:  "Show closed room list",
	RunE: func(cmd *cobra.Command, args []string) error {

		at, err := parseAt(oldroomsAt)
		if err != nil {
			return err
		}

		rooms, err := selectRoomHistoryForList(cmd.Context(), oldroomsLimit, at)
		if err != nil {
			return err
		}

		hosts, err := hostMap(cmd.Context())
		if err != nil {
			return err
		}

		cmd.SetOut(os.Stdout)
		if verbose {
			printOldRoomsHeader(cmd)
		}
		for _, r := range rooms {
			err := printOldRoom(cmd, r, hosts)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(oldroomsCmd)

	oldroomsCmd.Flags().StringVarP(&oldroomsAt, "at", "", "", "Show the rooms existed at the specified time")
	oldroomsCmd.Flags().IntVarP(&oldroomsLimit, "limit", "l", 100, "Upper limit of the room count to be shown")
}

func parseAt(at string) (*time.Time, error) {
	if at == "" {
		return nil, nil
	}

	// today's time
	t, err := time.Parse("15:04:05", at)
	if err == nil {
		d := t.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, t.Location()))
		now := time.Now()
		t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(d)
		return &t, nil
	}

	for _, layout := range []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	} {
		t, err := time.Parse(layout, at)
		if err == nil {
			return &t, nil
		}
	}
	return nil, xerrors.Errorf("Invalid time format: %v", at)
}

func selectRoomHistoryForList(ctx context.Context, limit int, at *time.Time) ([]*roomHistory, error) {
	q := "SELECT app_id, host_id, room_id, number, search_group, max_players, public_props, private_props, player_logs, created, closed FROM room_history"
	p := []interface{}{}
	if at != nil {
		q += " WHERE created <= ? AND closed >= ?"
		p = append(p, at)
		p = append(p, at)
	}
	q += " ORDER BY created DESC LIMIT ?"
	p = append(p, limit)

	var rooms []*roomHistory
	err := db.SelectContext(ctx, &rooms, q, p...)
	return rooms, err
}

func playerIds(data []byte) ([]string, error) {
	var logs []map[string]any
	err := json.Unmarshal(data, &logs)
	if err != nil {
		return nil, err
	}

	m := make(map[string]struct{})
	ids := []string{}
	for _, l := range logs {
		pid, ok := l["player_id"].(string)
		if !ok {
			continue
		}
		if _, ok := m[pid]; ok {
			continue
		}
		m[pid] = struct{}{}
		ids = append(ids, pid)
	}

	return ids, nil
}

func printOldRoomsHeader(cmd *cobra.Command) {
	cmd.Println("id\tapp\thost\tnumber\tgroup\tmax_players\tplayers\tcreated\tclosed\tprops")
}

func printOldRoom(cmd *cobra.Command, r *roomHistory, hosts map[uint32]*server) error {
	host := "-"
	if h, ok := hosts[r.HostID]; ok {
		host = h.HostName
	}

	var number int32
	if r.Number.Valid {
		number = r.Number.Int32
	}

	players, err := playerIds(r.PlayerLogs)
	if err != nil {
		return err
	}

	props, err := parsePropsSimple(r.PublicProps)
	if err != nil {
		return err
	}

	cmd.Printf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
		r.RoomID,
		r.AppID,
		host,
		number,
		r.SearchGroup,
		r.MaxPlayers,
		strings.Join(players, ","),
		r.Created,
		r.Closed,
		props,
	)

	return nil
}
