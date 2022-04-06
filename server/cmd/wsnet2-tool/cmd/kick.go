package cmd

import (
	"wsnet2/pb"

	"golang.org/x/xerrors"

	"github.com/spf13/cobra"
)

// kickCmd represents the kick command
var kickCmd = &cobra.Command{
	Use:   "kick <player> <room>",
	Short: "Kick the player",
	Long:  `Kick the player from the specified room`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return xerrors.Errorf("need player and room")
		}

		svrs, err := selectGrpcServers(cmd.Context(), args[1:2])
		if err != nil {
			return err
		}
		svr, ok := svrs[args[1]]
		if !ok {
			return xerrors.Errorf("room not found: %v", args[1])
		}

		conn, err := svr.Dial()
		if err != nil {
			return err
		}

		_, err = pb.NewGameClient(conn).Kick(cmd.Context(), &pb.KickReq{
			AppId:    svr.App,
			RoomId:   svr.Room,
			ClientId: args[0],
		})
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(kickCmd)
}
