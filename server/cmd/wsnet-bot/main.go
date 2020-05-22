package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v4"

	"wsnet2/binary"
	"wsnet2/lobby"
)

var (
	appID  = "testapp"
	userID = "12345"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal("client connection error:", err)
	}
	defer conn.Close()

	p := map[string]interface{}{"visible": true, "open": true, "mplayers": 4}

	param, err := msgpack.Marshal(p)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/rooms", bytes.NewReader(param))
	req.Header.Add("Content-Type", "application/x-msgpack")
	req.Header.Add("Host", "localhost")
	req.Header.Add("X-App-Id", appID)
	req.Header.Add("X-User-Id", userID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("http client error:", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("failed to create room: lobby server returned status %v", res.StatusCode)
	}

	room := &lobby.Room{}
	err = msgpack.NewDecoder(res.Body).UseJSONTag(true).Decode(room)
	if err != nil {
		log.Fatal("Unmarshal error:", err)
	}
	fmt.Println("url:", room.URL)

	hdr := http.Header{}
	hdr.Add("X-Wsnet-App", appID)
	hdr.Add("X-Wsnet-User", userID)
	hdr.Add("X-Wsnet-LastEventSeq", "0")

	d := websocket.Dialer{
		Subprotocols:    []string{},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, res2, err := d.Dial(room.URL, hdr)
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
	ws, res2, err = d.Dial(room.URL, hdr)
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
	ws, res2, err = d.Dial(room.URL, hdr)
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
