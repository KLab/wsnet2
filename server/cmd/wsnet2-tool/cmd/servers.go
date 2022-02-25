/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	serversGameOnly bool
	serversHubOnly  bool

	serverStatusStr = []string{"Starting", "Running", "Closing"}
)

type server struct {
	Id            int    `db:"id"`
	HostName      string `db:"hostname"`
	PublicName    string `db:"public_name"`
	GRPCPort      int    `db:"grpc_port"`
	WebSocketPort int    `db:"ws_port"`
	Status        int    `db:"status"`
	HeartBeat     int64  `db:"heartbeat"`
}

// serversCmd represents the servers command
var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "Show all game/hub servers",
	Long:  "Show all game and/or hub servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(os.Stdout)

		if verbose {
			printServersHeader(cmd)
		}

		if !serversHubOnly {
			const sql = "select * from game_server"
			var servers []server
			err := db.SelectContext(cmd.Context(), &servers, sql)
			if err != nil {
				return err
			}
			for _, s := range servers {
				printServer(cmd, "game", s)
			}
		}
		if !serversGameOnly {
			const sql = "select * from hub_server"
			var servers []server
			err := db.SelectContext(cmd.Context(), &servers, sql)
			if err != nil {
				return err
			}
			for _, s := range servers {
				printServer(cmd, "hub", s)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serversCmd)

	serversCmd.Flags().BoolVarP(&serversGameOnly, "game", "g", false, "show game servers only")
	serversCmd.Flags().BoolVarP(&serversHubOnly, "hub", "u", false, "show hub servers only")
}

func printServersHeader(cmd *cobra.Command) {
	cmd.Println("type\tid\thost\tpublic\tgrpc\twebsocket\tstatus\theartbeat")
}

func printServer(cmd *cobra.Command, typ string, s server) {
	st := serverStatusStr[s.Status]
	hb := time.Unix(s.HeartBeat, 0)
	v := time.Duration(conf.Lobby.ValidHeartBeat)
	ok := "Available"
	if hb.Before(time.Now().Add(-v)) {
		ok = "Dead"
	}

	cmd.Printf("%s\t%d\t%s\t%s\t%d\t%d\t%s:%s\t%v\n",
		typ, s.Id, s.HostName, s.PublicName, s.GRPCPort, s.WebSocketPort, st, ok, hb)
}
