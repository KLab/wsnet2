package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/game"
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
	Created      time.Time     `db:"created"`
	Closed       time.Time     `db:"closed"`

	PlayerLogs []*playerLog
}

type playerLog struct {
	ID       int               `db:"id" json:"-"`
	RoomID   string            `db:"room_id" json:"-"`
	PlayerID string            `db:"player_id" json:"player_id"`
	Message  game.PlayerLogMsg `db:"message" json:"message"`
	Datetime time.Time         `db:"datetime" json:"time"`
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

		rooms, err := selectRoomHistoryByIds(cmd.Context(), args)
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

func selectRoomHistoryByIds(ctx context.Context, ids []string) (map[string]*roomHistory, error) {
	q, p, err := sqlx.In("SELECT * FROM room_history WHERE room_id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	_, rooms, err := selectRoomHistory(ctx, q, p...)
	return rooms, err
}

func selectRoomHistory(ctx context.Context, q string, p ...any) ([]*roomHistory, map[string]*roomHistory, error) {
	var rooms []*roomHistory
	err := db.SelectContext(ctx, &rooms, q, p...)
	if err != nil || len(rooms) == 0 {
		return rooms, map[string]*roomHistory{}, err
	}
	m := make(map[string]*roomHistory, len(rooms))
	rids := make([]string, 0, len(rooms))
	for _, r := range rooms {
		m[r.RoomID] = r
		rids = append(rids, r.RoomID)
	}

	q, p, err = sqlx.In("SELECT * FROM player_log WHERE room_id IN (?)", rids)
	if err != nil {
		return nil, nil, err
	}
	var plogs []*playerLog
	err = db.SelectContext(ctx, &plogs, q, p...)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range plogs {
		rid := p.RoomID
		m[rid].PlayerLogs = append(m[rid].PlayerLogs, p)
	}

	return rooms, m, nil
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
		"player_logs":   r.PlayerLogs,
		"created":       r.Created,
		"closed":        r.Closed,
	}, nil
}
