package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"time"
	"wsnet2/binary"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

type roomHistory struct {
	ID           int           `db:"id"`
	AppID        string        `db:"app_id"`
	HostID       uint32        `db:"host_id"`
	RoomID       string        `db:"room_id"`
	Number       sql.NullInt32 `db:"number"`
	SearchGroup  uint32        `db:"search_group"`
	MaxPlayers   uint32        `db:"max_players"`
	PublicProps  []byte        `db:"public_props"`
	PrivateProps []byte        `db:"private_props"`
	PlayerLogs   []byte        `db:"player_logs"` // JSON Type
	Created      time.Time     `db:"created"`
	Closed       time.Time     `db:"closed"`
}

// oldroomCmd represents the oldroom command
var oldroomCmd = &cobra.Command{
	Use:   "oldroom <roomid>...",
	Short: "Show closed room info",
	Long:  "Show closed room info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return xerrors.Errorf("need roomid\n")
		}

		rooms, err := selectRoomHistory(cmd.Context(), args)
		if err != nil {
			return err
		}

		for _, id := range args {
			room, ok := rooms[id]
			if !ok {
				return xerrors.Errorf("room not found: %v", id)
			}

			out, err := formatRoomHistory(room)
			if err != nil {
				return err
			}

			j, err := json.Marshal(out)
			if err != nil {
				return err
			}

			cmd.SetOut(os.Stdout)
			cmd.Println(string(j))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(oldroomCmd)
}

func selectRoomHistory(ctx context.Context, ids []string) (map[string]*roomHistory, error) {
	q, p, err := sqlx.In("SELECT * FROM room_history WHERE room_id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	var rooms []*roomHistory
	err = db.SelectContext(ctx, &rooms, q, p...)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*roomHistory)
	for _, r := range rooms {
		m[r.RoomID] = r
	}
	return m, nil
}

func formatRoomHistory(r *roomHistory) (_ map[string]any, err error) {
	var number int32
	if r.Number.Valid {
		number = r.Number.Int32
	}
	var publicProps any
	if len(r.PublicProps) > 0 {
		publicProps, err = binary.UnmarshalRecursive(r.PublicProps)
		if err != nil {
			return nil, err
		}
	}
	var privateProps any
	if len(r.PrivateProps) > 0 {
		privateProps, err = binary.UnmarshalRecursive(r.PrivateProps)
		if err != nil {
			return nil, err
		}
	}
	var playerLogs any
	if len(r.PlayerLogs) > 0 {
		err = json.Unmarshal(r.PlayerLogs, &playerLogs)
		if err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"id":            r.RoomID,
		"app_id":        r.AppID,
		"host_id":       r.HostID,
		"room_id":       r.RoomID,
		"number":        number,
		"search_group":  r.SearchGroup,
		"max_players":   r.MaxPlayers,
		"public_props":  publicProps,
		"private_props": privateProps,
		"player_logs":   playerLogs,
		"created":       r.Created,
		"closed":        r.Closed,
	}, nil
}
