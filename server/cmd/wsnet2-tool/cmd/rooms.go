package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

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

		const roomsql = "SELECT * FROM room"
		var rooms []*pb.RoomInfo
		err = db.SelectContext(cmd.Context(), &rooms, roomsql)
		if err != nil {
			return err
		}

		cmd.SetOut(os.Stdout)
		if verbose {
			printRoomsHeader(cmd)
		}

		for _, r := range rooms {
			err := printRoom(cmd, r, hosts)
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

func printRoomsHeader(cmd *cobra.Command) {
	cmd.Println("id\tapp\thost\tflags\tnumber\tgroup\tmax_players\tplayers\twatchers\tcreated\tprops")
}

func printRoom(cmd *cobra.Command, r *pb.RoomInfo, h map[uint32]*server) error {
	var num int32
	if r.Number != nil {
		num = r.Number.Number
	}

	p, err := parsePropsSimple(r.PublicProps)

	cmd.Printf("%v\t%v\t%v\t%v\t%06d\t%d\t%d\t%d\t%d\t%v\t%s\n",
		r.Id, r.AppId, h[r.HostId].HostName, roomFlags(r), num,
		r.SearchGroup, r.MaxPlayers, r.Players, r.Watchers,
		r.Created.Time(), p)

	return err
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
	u, _, err := binary.UnmarshalAs(data, binary.TypeDict, binary.TypeNull)
	if err != nil {
		return "", err
	}
	dic, _ := u.(binary.Dict)
	out := []byte{'{'}
	for k, d := range dic {
		if len(d) == 0 {
			return string(out), xerrors.Errorf("No payload: key=%v", k)
		}

		out = fmt.Appendf(out, "%q:", k)

		t := binary.Type(d[0])
		switch t {
		case binary.TypeNull:
			out = fmt.Append(out, "null,")
		case binary.TypeTrue:
			out = fmt.Append(out, "true,")
		case binary.TypeFalse:
			out = fmt.Append(out, "false,")
		case binary.TypeSByte, binary.TypeByte, binary.TypeChar, binary.TypeShort, binary.TypeUShort,
			binary.TypeInt, binary.TypeUInt, binary.TypeLong, binary.TypeULong,
			binary.TypeFloat, binary.TypeDouble, binary.TypeDecimal:
			v, _, err := binary.Unmarshal(d)
			if err != nil {
				return string(out), err
			}
			out = fmt.Appendf(out, "%v,", v)
		case binary.TypeStr8, binary.TypeStr16:
			v, _, err := binary.Unmarshal(d)
			if err != nil {
				return string(out), err
			}
			out = fmt.Appendf(out, "%q,", v)
		case binary.TypeObj:
			out = fmt.Appendf(out, `"Obj(%d)",`, d[1])
		case binary.TypeBools:
			out, err = appendPrimitiveArraySimple[bool](out, d)
			if err != nil {
				return string(out), err
			}
		case binary.TypeSBytes, binary.TypeBytes, binary.TypeShorts, binary.TypeUShorts,
			binary.TypeInts, binary.TypeUInts, binary.TypeLongs:
			out, err = appendPrimitiveArraySimple[int](out, d)
			if err != nil {
				return string(out), err
			}
		case binary.TypeChars:
			out, err = appendPrimitiveArraySimple[rune](out, d)
			if err != nil {
				return string(out), err
			}
		case binary.TypeULongs:
			out, err = appendPrimitiveArraySimple[uint64](out, d)
			if err != nil {
				return string(out), err
			}
		case binary.TypeFloats:
			out, err = appendPrimitiveArraySimple[float32](out, d)
			if err != nil {
				return string(out), err
			}
		case binary.TypeDoubles:
			out, err = appendPrimitiveArraySimple[float64](out, d)
			if err != nil {
				return string(out), err
			}
		case binary.TypeList:
			out = fmt.Appendf(out, `"List[%d]",`, d[1])
		default:
			out = fmt.Appendf(out, "%q,", t)
		}
	}
	if len(out) > 1 {
		out = out[:len(out)-1]
	}
	out = append(out, '}')

	return string(out), nil
}

func appendPrimitiveArraySimple[T any](out, data []byte) ([]byte, error) {
	if n := int(data[1])<<8 + int(data[2]); n > 4 {
		return fmt.Appendf(out, "\"%v[%d]\",", binary.Type(data[0]), n), nil
	}

	u, _, err := binary.Unmarshal(data)
	if err != nil {
		return out, err
	}
	l, _ := u.([]T)
	out = append(out, '[')
	for _, v := range l {
		out = fmt.Appendf(out, "%v,", v)
	}
	if len(out) > 1 {
		out = out[:len(out)-1]
	}
	return fmt.Append(out, "],"), nil
}
