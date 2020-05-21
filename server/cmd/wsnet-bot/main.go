package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"

	"wsnet2/binary"
	"wsnet2/pb"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal("client connection error:", err)
	}
	defer conn.Close()

	roomOption := &pb.RoomOption{
		MaxPlayers: 10,
		WithNumber: true,
	}

	param, err := proto.Marshal(roomOption)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/rooms", bytes.NewReader(param))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Host", "localhost")
	req.Header.Add("X-App-Id", "testapp")
	req.Header.Add("X-User-Id", "11111")

	httpclient := &http.Client{}
	res, err := httpclient.Do(req)
	if err != nil {
		log.Fatal("http client error:", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("failed to create room: lobby server returned status %v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("read body error:", err)
	}
	room := &pb.Room{}
	proto.Unmarshal(body, room)

	fmt.Printf("%v\n\n\n", room)

	hdr := http.Header{}
	hdr.Add("X-Wsnet-App", "testapp")
	hdr.Add("X-Wsnet-User", "11111")
	hdr.Add("X-Wsnet-LastEventSeq", "0")

	d := websocket.Dialer{
		Subprotocols:    []string{},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, res2, err := d.Dial(room.Url, hdr)
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
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 21, 22, 23, 24, 25})
		//time.Sleep(time.Second)
		//fmt.Println("msg 003")
		//ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 21, 22, 23, 24, 25})
		//		time.Sleep(time.Second)
		//		ws.Close()
	}()

	//	<-done
	time.Sleep(6 * time.Second)

	fmt.Println("reconnect test")
	hdr.Set("X-Wsnet-LastEventSeq", "2")
	ws, res2, err = d.Dial(room.Url, hdr)
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
	ws, res2, err = d.Dial(room.Url, hdr)
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
