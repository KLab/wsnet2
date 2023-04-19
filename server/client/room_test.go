package client_test

import (
	"reflect"
	"testing"
	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/pb"
)

func newRoom() *client.Room {
	players := map[string]*client.Player{
		"user1": {
			Id:    "user1",
			Props: binary.Dict{"cli1": binary.MarshalInt(100)},
		},
		"user2": {
			Id:    "user2",
			Props: binary.Dict{},
		},
	}
	return &client.Room{
		Id:             "room1",
		Number:         nil,
		Visible:        false,
		Joinable:       true,
		Watchable:      false,
		SearchGroup:    10,
		MaxPlayers:     5,
		Watchers:       20,
		PublicProps:    binary.Dict{"pub1": binary.MarshalBool(true), "pub2": binary.MarshalNull()},
		PrivateProps:   binary.Dict{},
		ClientDeadline: 30,
		Players:        players,
		Me:             players["user2"],
		Master:         players["user1"],
	}
}

func TestRoom_Update_onEvJoined(t *testing.T) {
	name := "user3"
	props := binary.Dict{"cli3": binary.MarshalInt(200)}
	ev := binary.NewEvJoined(
		&pb.ClientInfo{
			Id:    name,
			Props: binary.MarshalDict(props),
		})

	room := newRoom()
	err := room.Update(ev)
	if err != nil {
		t.Fatalf("%v", err)
	}

	p, ok := room.Players[name]
	if !ok {
		t.Fatalf("player %v not found", name)
	}
	if !reflect.DeepEqual(p.Props, props) {
		t.Fatalf("prop = %v, wants %v", p.Props, props)
	}
}

func TestRoom_Update_onEvLeft(t *testing.T) {
	nameleft := "user1"
	namemaster := "user2"

	ev := binary.NewEvLeft(nameleft, namemaster, "leave")

	room := newRoom()
	err := room.Update(ev)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if _, ok := room.Players[nameleft]; ok {
		t.Fatalf("found %v", nameleft)
	}
	if room.Master.Id != namemaster {
		t.Fatalf("master = %q, wants %q", room.Master.Id, namemaster)
	}
}

func TestRoom_Update_onEvRoomProp(t *testing.T) {
	const v = true
	const j = false
	const w = true
	const sgrp = 42
	const maxp = 10
	const cdl = 120
	pubp := binary.Dict{"pub2": binary.MarshalStr8("abc"), "pub3": binary.MarshalByte(3)}
	prvp := binary.Dict{"prv1": binary.MarshalInt(15)}

	expPubp := binary.Dict{
		"pub1": binary.MarshalBool(true),
		"pub2": binary.MarshalStr8("abc"),
		"pub3": binary.MarshalByte(3),
	}
	expPrivp := binary.Dict{
		"prv1": binary.MarshalInt(15),
	}

	ev := binary.NewRegularEvent(
		binary.EvTypeRoomProp,
		binary.MarshalRoomPropPayload(v, j, w, sgrp, maxp, cdl, pubp, prvp))

	room := newRoom()
	err := room.Update(ev)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if room.Visible != v {
		t.Fatalf("Visible = %v, wants %v", room.Visible, v)
	}
	if room.Joinable != j {
		t.Fatalf("Joinable = %v, wants %v", room.Joinable, j)
	}
	if room.Watchable != w {
		t.Fatalf("Watchable = %v, wants %v", room.Watchable, w)
	}
	if room.SearchGroup != sgrp {
		t.Fatalf("SearchGroup = %v, wants %v", room.SearchGroup, sgrp)
	}
	if room.MaxPlayers != maxp {
		t.Fatalf("MaxPlayers = %v, wants %v", room.MaxPlayers, maxp)
	}
	if room.ClientDeadline != cdl {
		t.Fatalf("ClientDeadline = %v, wants %v", room.ClientDeadline, cdl)
	}
	if !reflect.DeepEqual(room.PublicProps, expPubp) {
		t.Fatalf("PublicProps = %v, wants %v", room.PublicProps, expPubp)
	}
	if !reflect.DeepEqual(room.PrivateProps, expPrivp) {
		t.Fatalf("PrivateProps = %v, wants %v", room.PrivateProps, expPrivp)
	}
}

func TestRoom_Update_onEvClientProp(t *testing.T) {
	user := "user1"
	ev := binary.NewEvClientProp(user, binary.MarshalDict(binary.Dict{
		"cli2": binary.MarshalBool(false),
	}))
	exp := binary.Dict{
		"cli1": binary.MarshalInt(100),
		"cli2": binary.MarshalBool(false),
	}

	room := newRoom()
	err := room.Update(ev)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(room.Players[user].Props, exp) {
		t.Fatalf("player[%v] prop: %v, wants %v", user, room.Players[user].Props, exp)
	}
}

func TestRoom_Update_onEvMasterSwitched(t *testing.T) {
	newmaster := "user2"
	ev := binary.NewEvMasterSwitched("user1", "user2")

	room := newRoom()
	err := room.Update(ev)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if room.Master.Id != newmaster {
		t.Fatalf("new master: %v, wants %v", room.Master.Id, newmaster)
	}
}

func TestRoom_Update_onRejoined(t *testing.T) {
	user := "user1"
	props := binary.Dict{
		"cli2": binary.MarshalBool(true),
	}
	ev := binary.NewEvRejoined(&pb.ClientInfo{
		Id:    user,
		Props: binary.MarshalDict(props),
	})

	room := newRoom()
	err := room.Update(ev)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(room.Players[user].Props, props) {
		t.Fatalf("client props: %v, wants %v", room.Players[user].Props, props)
	}
}

func TestRoom_Update_onEvPong(t *testing.T) {
	const watchers = 17
	ev := binary.NewEvPong(10000, watchers, binary.Dict{})

	room := newRoom()
	err := room.Update(ev)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if room.Watchers != watchers {
		t.Fatalf("Watchers = %v, wants %v", room.Watchers, watchers)
	}
}
