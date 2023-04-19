# Client Library for Go

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/pb"
)

func main() {
	ctx := context.Background()

	lobbyUrl := "http://localhost:8080"
	appId := "testapp"
	appKey := "testapppkey"
	userId := "testuser"

	accessInfo, err := client.GenAccessInfo(lobbyUrl, appId, appKey, userId)
	if err != nil {
		panic(err)
	}

	roomOption := &pb.RoomOption{
		Visible:     true,
		Joinable:    true,
		SearchGroup: 1,
		PublicProps: binary.MarshalDict(binary.Dict{
			"RoomType": binary.MarshalByte(1),
		}),
	}

	clientInfo := &pb.ClientInfo{
		Id:    userId,
		Props: binary.MarshalDict(binary.Dict{"Name": binary.MarshalStr8("TestUser1")}),
	}

	room, conn, err := client.Create(
		ctx, accessInfo, roomOption, clientInfo,
		func(err error) { fmt.Printf("warning: %+v\n", err) })
	if err != nil {
		panic(err)
	}

	go func() {
		for ev := range conn.Events() {
			room.Update(ev)
			if binary.IsRegularEvent(ev) {
				u, err := binary.UnmarshalRecursive(ev.Payload())
				fmt.Printf("%v: %#v, %v\n", ev.Type(), u, err)
			} else {
				fmt.Println(ev.Type())
			}
		}
	}()

	for i := 0; i < 5; i++ {
		conn.Send(binary.MsgTypeBroadcast, binary.MarshalStr8("Hello, WSNet2!"))
		time.Sleep(time.Second)
	}

	conn.Send(binary.MsgTypeLeave, binary.MarshalLeavePayload("bye"))

	msg, err := conn.Wait(ctx)
	fmt.Println(msg, err)
}
```
