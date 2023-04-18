package cli

import (
	"golang.org/x/xerrors"

	"wsnet2/binary"
	"wsnet2/pb"
)

type Room struct {
	Id             string
	Number         *int32
	Visible        bool
	Joinable       bool
	Watchable      bool
	SearchGroup    uint32
	MaxPlayers     uint32
	Watchers       uint32
	PublicProps    binary.Dict
	PrivateProps   binary.Dict
	ClientDeadline uint32
	Players        map[string]*Player
	Me             *Player
	Master         *Player
}

type Player struct {
	Id    string
	Props binary.Dict
}

func newRoom(joined *pb.JoinedRoomRes, myid string) (*Room, error) {
	var num *int32 = nil
	if joined.RoomInfo.Number != nil {
		n := joined.RoomInfo.Number.Number
		num = &n
	}

	pubProps, _, err := binary.UnmarshalNullDict(joined.RoomInfo.PublicProps)
	if err != nil {
		return nil, xerrors.Errorf("public props: %w", err)
	}

	privProps, _, err := binary.UnmarshalNullDict(joined.RoomInfo.PrivateProps)
	if err != nil {
		return nil, xerrors.Errorf("private props: %w", err)
	}

	players := make(map[string]*Player, len(joined.Players))
	for _, p := range joined.Players {
		props, _, err := binary.UnmarshalNullDict(p.Props)
		if err != nil {
			return nil, xerrors.Errorf("player[%v] props: %w", p.Id, err)
		}
		players[p.Id] = &Player{
			Id:    p.Id,
			Props: props,
		}
	}

	return &Room{
		Id:             joined.RoomInfo.Id,
		Number:         num,
		Visible:        joined.RoomInfo.Visible,
		Joinable:       joined.RoomInfo.Joinable,
		Watchable:      joined.RoomInfo.Watchable,
		SearchGroup:    joined.RoomInfo.SearchGroup,
		MaxPlayers:     joined.RoomInfo.MaxPlayers,
		Watchers:       joined.RoomInfo.Watchers,
		PublicProps:    pubProps,
		PrivateProps:   privProps,
		ClientDeadline: joined.Deadline,
		Players:        players,
		Me:             players[myid],
		Master:         players[joined.MasterId],
	}, nil
}

// Update Room using an Event
//
// 届いたEventを順序通りに適用することでRoom情報を更新できます
func (r *Room) Update(ev binary.Event) error {
	switch ev.Type() {
	case binary.EvTypeJoined:
		return r.onEvJoined(ev)
	case binary.EvTypeLeft:
		return r.onEvLeft(ev)
	case binary.EvTypeRoomProp:
		return r.onEvRoomProp(ev)
	case binary.EvTypeClientProp:
		return r.onEvClientProp(ev)
	case binary.EvTypeMasterSwitched:
		return r.onEvMasterSwitched(ev)
	case binary.EvTypeRejoined:
		return r.onEvRejoined(ev)
	case binary.EvTypePong:
		return r.onEvPong(ev)
	}
	return nil
}

func (r *Room) onEvJoined(ev binary.Event) error {
	clinfo, err := binary.UnmarshalEvJoinedPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Room.onEvJoined: payload: %w", err)
	}
	props, _, err := binary.UnmarshalNullDict(clinfo.Props)
	if err != nil {
		return xerrors.Errorf("Room.onEvJoined: player(%v) props: %w", clinfo.Id, err)
	}
	r.Players[clinfo.Id] = &Player{
		Id:    clinfo.Id,
		Props: props,
	}
	return nil
}

func (r *Room) onEvLeft(ev binary.Event) error {
	p, err := binary.UnmarshalEvLeftPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Room.onEvLeft: payload: %w", err)
	}
	r.Master = r.Players[p.MasterId]
	delete(r.Players, p.ClientId)
	return nil
}

func (r *Room) onEvRoomProp(ev binary.Event) error {
	p, err := binary.UnmarshalEvRoomPropPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Room.onEvRoomProp: payload: %w", err)
	}
	r.Visible = p.Visible
	r.Joinable = p.Joinable
	r.Watchable = p.Watchable
	r.SearchGroup = p.SearchGroup
	r.MaxPlayers = p.MaxPlayer
	r.ClientDeadline = p.ClientDeadline
	for k, v := range p.PublicProps {
		r.PublicProps[k] = v
	}
	for k, v := range p.PrivateProps {
		r.PrivateProps[k] = v
	}
	return nil
}

func (r *Room) onEvClientProp(ev binary.Event) error {
	p, err := binary.UnmarshalEvClientPropPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Room.onEvClientProp: payload: %w", err)
	}
	for k, v := range p.Props {
		r.Players[p.Id].Props[k] = v
	}
	return nil
}

func (r *Room) onEvMasterSwitched(ev binary.Event) error {
	mid, err := binary.UnmarshalEvMasterSwitchedPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Room.onEvMasterSwitched: payload: %w", err)
	}
	r.Master = r.Players[mid]
	return nil
}

func (r *Room) onEvRejoined(ev binary.Event) error {
	p, err := binary.UnmarshalEvRejoinedPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Room.onEvRejoined: payload: %w", err)
	}
	props, _, err := binary.UnmarshalNullDict(p.Props)
	if err != nil {
		return xerrors.Errorf("Room.onEvRejoined: player(%v) props: %w", p.Id, err)
	}
	r.Players[p.Id] = &Player{
		Id:    p.Id,
		Props: props,
	}
	return nil
}

func (r *Room) onEvPong(ev binary.Event) error {
	p, err := binary.UnmarshalEvPongPayload(ev.Payload())
	if err != nil {
		return xerrors.Errorf("Room.onEvPong: payload: %w", err)
	}
	r.Watchers = p.Watchers
	return nil
}
