package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"

	"wsnet2/binary"
	"wsnet2/pb"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:19000", grpc.WithInsecure())
	if err != nil {
		log.Fatal("client connection error:", err)
	}
	defer conn.Close()

	client := pb.NewGameClient(conn)
	req := &pb.CreateRoomReq{
		AppId: "testapp",
		RoomOption: &pb.RoomOption{
			Visible:     true,
			Watchable:   true,
			SearchGroup: 50,
			LogLevel:    4,
			PublicProps: binary.MarshalDict(binary.Dict{
				"hoge": binary.MarshalLong(10),
				"fuga": binary.MarshalUShort(20),
			}),
			PrivateProps: binary.MarshalDict(binary.Dict{
				"foo": binary.MarshalStr8("foobar"),
				"bar": binary.MarshalBool(true),
			}),
		},
		MasterInfo: &pb.ClientInfo{
			Id: "11111",
		},
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		fmt.Printf("create room error: %v", err)
	}

	url := fmt.Sprintf("ws://localhost:8000/room/%s", res.RoomInfo.Id)
	hdr := http.Header{}
	hdr.Add("X-Wsnet-App", "testapp")
	hdr.Add("X-Wsnet-User", "11111")
	hdr.Add("X-Wsnet-LastEventSeq", "0")

	d := websocket.Dialer{
		Subprotocols:    []string{},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, res2, err := d.Dial(url, hdr)
	if err != nil {
		fmt.Printf("dial error: %v, %v\n", res2, err)
		return
	}
	fmt.Println("response:", res2)

	done := make(chan bool)
	go eventloop(ws, done)

	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("msg 001")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 1, 1, 2, 3, 4, 5})
		time.Sleep(time.Second)
		fmt.Println("msg 002")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 2, 11, 12, 13, 14, 15})
		time.Sleep(time.Second)
		fmt.Println("msg 003")

		payload := []byte{byte(binary.MsgTypeRoomProp), 0, 0, 3}
		payload = append(payload, binary.MarshalByte(3)...)       // flags
		payload = append(payload, binary.MarshalUInt(1000)...)    // search group
		payload = append(payload, binary.MarshalUShort(13)...)    // maxplayer
		payload = append(payload, binary.MarshalUShort(15)...)    // deadline
		payload = append(payload, binary.MarshalDict(binary.Dict{ // public props
			"fuga": binary.MarshalUShort(20),
			"hige": binary.MarshalNull(),
		})...)
		payload = append(payload, binary.MarshalDict(binary.Dict{ // private props
			"bar": []byte{},
			"baz": binary.MarshalInt(10),
		})...)
		ws.WriteMessage(websocket.BinaryMessage, payload)
	}()

	//	<-done
	time.Sleep(6 * time.Second)

	fmt.Println("reconnect test")
	hdr.Set("X-Wsnet-LastEventSeq", "2")
	ws, res2, err = d.Dial(url, hdr)
	if err != nil {
		fmt.Printf("dial error: %v, %v\n", res2, err)
		return
	}
	fmt.Println("response:", res2)

	done = make(chan bool)
	go eventloop(ws, done)

	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("msg 004")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 4, 31, 32, 33, 34, 35})
		time.Sleep(time.Second)
		fmt.Println("msg 005")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 5, 41, 42, 43, 44, 45})
		time.Sleep(time.Second)
		fmt.Println("msg 006 (leave)")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeLeave), 0, 0, 6})
		time.Sleep(time.Second)
		ws.Close()
	}()

	<-done

	time.Sleep(3 * time.Second)
	fmt.Println("reconnect test after leave")
	hdr.Set("X-Wsnet-LastEventSeq", "4")
	ws, res2, err = d.Dial(url, hdr)
	if err != nil {
		fmt.Printf("dial error: %v, %v\n", res2, err)
		return
	}
	fmt.Println("response:", res2)

	done = make(chan bool)
	go eventloop(ws, done)
	<-done
}

func eventloop(ws *websocket.Conn, done chan bool) {
	defer close(done)
	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("ReadMessage error: %v\n", err)
			return
		}

		switch ty := binary.EvType(b[0]); ty {
		case binary.EvTypeJoined:
			seqnum := (int(b[1]) << 24) + (int(b[2]) << 16) + (int(b[3]) << 8) + int(b[4])
			namelen := int(b[6])
			name := string(b[7 : 7+namelen])
			props := b[7+namelen:]
			fmt.Printf("%s: %v %#v, %v, %v\n", ty, seqnum, name, props, b)
		default:
			fmt.Printf("ReadMessage: %v, %v\n", ty, b)
		}
	}
}
