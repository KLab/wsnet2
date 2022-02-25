/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"wsnet2/binary"
	"wsnet2/pb"
)

// roomsCmd represents the rooms command
var roomsCmd = &cobra.Command{
	Use:   "rooms",
	Short: "Show active room list",
	Long:  "Show active room list",
	RunE: func(cmd *cobra.Command, args []string) error {
		hosts, err := hostMap(cmd.Context())
		if err != nil {
			return err
		}

		const roomsql = "SELECT * FROM room WHERE number is not null"
		var rooms []*pb.RoomInfo
		err = db.SelectContext(cmd.Context(), &rooms, roomsql)
		if err != nil {
			return err
		}

		if verbose {
			printRoomsHeader()
		}

		for _, r := range rooms {
			err := printRoom(r, hosts)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(roomsCmd)
}

func hostMap(ctx context.Context) (map[uint32]*server, error) {
	const hostsql = "SELECT * FROM game_server"
	var hosts []*server
	err := db.SelectContext(ctx, &hosts, hostsql)
	if err != nil {
		return nil, err
	}

	m := make(map[uint32]*server)
	for _, h := range hosts {
		m[uint32(h.Id)] = h
	}

	return m, nil
}

func printRoomsHeader() {
	fmt.Println("id\tapp\thost\tflags\tnumber\tgroup\tmax_players\tplayers\twatchers\tcreated\tprops")
}

func printRoom(r *pb.RoomInfo, h map[uint32]*server) error {
	p, err := parsePropsSimple(r.PublicProps)
	if err != nil {
		return err
	}

	fmt.Printf("%v\t%v\t%v\t%06d\t%d\t%d\t%d\t%d\t%v\t%s\n",
		r.Id, h[r.HostId].HostName, roomFlags(r), r.Number.Number,
		r.SearchGroup, r.MaxPlayers, r.Players, r.Watchers,
		r.Created.Time(), p)

	return nil
}

func roomFlags(r *pb.RoomInfo) string {
	f := []byte("---")
	if r.Visible {
		f[0] = 'v'
	}
	if r.Joinable {
		f[1] = 'j'
	}
	if r.Watchable {
		f[2] = 'w'
	}
	return string(f)
}

func parsePropsSimple(data []byte) (string, error) {
	u, _, err := binary.UnmarshalAs(data, binary.TypeDict)
	if err != nil {
		return "", err
	}
	dic := u.(binary.Dict)
	out := []byte{'{'}
	for k, d := range dic {
		if len(d) == 0 {
			return "", fmt.Errorf("No payload: key=%v", k)
		}

		out = append(out, []byte(k)...)
		out = append(out, ':')

		t := binary.Type(d[0])
		switch t {
		case binary.TypeNull:
			out = append(out, []byte("nil, ")...)
		case binary.TypeTrue:
			out = append(out, []byte("true, ")...)
		case binary.TypeFalse:
			out = append(out, []byte("false, ")...)
		case binary.TypeSByte, binary.TypeByte, binary.TypeChar, binary.TypeShort, binary.TypeUShort,
			binary.TypeInt, binary.TypeUInt, binary.TypeLong, binary.TypeULong,
			binary.TypeFloat, binary.TypeDouble, binary.TypeDecimal:
			v, _, err := binary.Unmarshal(d)
			if err != nil {
				return "", err
			}
			out = append(out, []byte(fmt.Sprintf("%v, ", v))...)
		case binary.TypeStr8, binary.TypeStr16:
			v, _, err := binary.Unmarshal(d)
			if err != nil {
				return "", err
			}
			out = append(out, []byte(fmt.Sprintf("%q, ", v))...)
		case binary.TypeObj:
			out = append(out, []byte(fmt.Sprintf("Obj(%d), ", d[1]))...)
		default:
			out = append(out, []byte(fmt.Sprintf("%v, ", t))...)
		}
	}
	out[len(out)-2] = '}'
	out = out[:len(out)-1]

	return string(out), nil
}
