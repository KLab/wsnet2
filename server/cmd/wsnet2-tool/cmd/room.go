package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"wsnet2/binary"
	"wsnet2/pb"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcServer struct {
	Room string `db:"room_id"`
	App  string `db:"app_id"`
	Host string `db:"hostname"`
	Port int    `db:"grpc_port"`
}

func selectGrpcServers(ctx context.Context, ids []string) (map[string]*grpcServer, error) {
	q, p, err := sqlx.In(
		"SELECT r.id room_id, r.app_id, s.hostname, s.grpc_port FROM room r JOIN game_server s ON r.host_id = s.id WHERE r.id IN (?)", ids)
	if err != nil {
		return nil, xerrors.Errorf("build query: %w", err)
	}
	var svrs []*grpcServer
	err = db.SelectContext(ctx, &svrs, q, p...)
	if err != nil {
		return nil, xerrors.Errorf("select grpc servers: %w", err)
	}

	m := make(map[string]*grpcServer)
	for _, s := range svrs {
		m[s.Room] = s
	}

	return m, nil
}

func (s *grpcServer) Dial() (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("%s:%d", s.Host, s.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// roomCmd represents the room command
var roomCmd = &cobra.Command{
	Use:   "room <roomid>...",
	Short: "Show room info",
	Long:  "Show room info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return xerrors.Errorf("need roomid\n")
		}

		svrs, err := selectGrpcServers(cmd.Context(), args)
		if err != nil {
			return err
		}

		for _, id := range args {
			svr, ok := svrs[id]
			if !ok {
				return xerrors.Errorf("room not found: %v", id)
			}

			conn, err := svr.Dial()
			if err != nil {
				return err
			}

			res, err := pb.NewGameClient(conn).GetRoomInfo(
				cmd.Context(), &pb.GetRoomInfoReq{AppId: svr.App, RoomId: id})
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
			cmd.Println(string(j))
		}

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
		props, err := binary.UnmarshalRecursive(c.Props)
		if err != nil {
			return nil, err
		}

		p := map[string]interface{}{
			"id":        c.Id,
			"is_master": c.Id == res.MasterId,
			"props":     props,
		}

		ps = append(ps, p)
	}
	m["players"] = ps

	return m, nil
}
