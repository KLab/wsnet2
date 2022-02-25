/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"wsnet2/binary"
	"wsnet2/pb"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// roomCmd represents the room command
var roomCmd = &cobra.Command{
	Use:   "room",
	Short: "Show room info",
	Long:  "Show room info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return xerrors.Errorf("need roomid\n")
		}
		rid := args[0]

		var svr struct {
			App  string `db:"app_id"`
			Host string `db:"hostname"`
			Port int    `db:"grpc_port"`
		}
		err := db.GetContext(cmd.Context(), &svr,
			"SELECT r.app_id, s.hostname, s.grpc_port FROM room r JOIN game_server s ON r.host_id = s.id WHERE r.id = ?", rid)
		if err != nil {
			return err
		}

		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", svr.Host, svr.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		res, err := pb.NewGameClient(conn).GetRoomInfo(
			cmd.Context(), &pb.GetRoomInfoReq{AppId: svr.App, RoomId: rid})
		if err != nil {
			return err
		}

		out, err := formatRoom(res)
		if err != nil {
			return err
		}

		j, err := json.Marshal(out)
		if err != nil {
			return err
		}

		cmd.SetOut(os.Stdout)
		cmd.Print(string(j))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(roomCmd)
}

func formatRoom(res *pb.GetRoomInfoRes) (map[string]interface{}, error) {
	r := res.RoomInfo
	cs := res.ClientInfos

	m := map[string]interface{}{
		"id":             r.Id,
		"app_id":         r.AppId,
		"host_id":        r.HostId,
		"visible":        r.Visible,
		"joinable":       r.Joinable,
		"watchable":      r.Watchable,
		"number":         r.Number.Number,
		"search_group":   r.SearchGroup,
		"max_players":    r.MaxPlayers,
		"watchers_count": r.Watchers,
		"created":        r.Created.Time(),
	}
	var err error
	m["public_props"], err = binary.UnmarshalRecursive(r.PublicProps)
	if err != nil {
		return nil, err
	}
	m["private_props"], err = binary.UnmarshalRecursive(r.PrivateProps)
	if err != nil {
		return nil, err
	}

	ps := make([]map[string]interface{}, 0)
	for _, c := range cs {
		// todo: remove
		if c.Id == "" {
			continue
		}

		props, err := binary.UnmarshalRecursive(c.Props)
		if err != nil {
			return nil, err
		}

		p := map[string]interface{}{
			"id": c.Id,
			// todo: is_master flag
			"props": props,
		}

		ps = append(ps, p)
	}
	m["players"] = ps

	return m, nil
}
