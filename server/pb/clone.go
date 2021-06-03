package pb

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (src *RoomInfo) Clone() *RoomInfo {
	dst := &RoomInfo{}
	dst.Id = src.Id
	dst.AppId = src.AppId
	dst.HostId = src.HostId
	dst.Visible = src.Visible
	dst.Joinable = src.Joinable
	dst.Watchable = src.Watchable
	dst.Number = &RoomNumber{Number: src.Number.Number}
	dst.SearchGroup = src.SearchGroup
	dst.MaxPlayers = src.MaxPlayers
	dst.Players = src.Players
	dst.Watchers = src.Watchers
	dst.PublicProps = make([]byte, len(src.PublicProps))
	copy(dst.PublicProps, src.PublicProps)
	dst.PrivateProps = make([]byte, len(src.PrivateProps))
	copy(dst.PrivateProps, src.PrivateProps)
	dst.Created = src.Created.Clone()
	return dst
}

func (src *ClientInfo) Clone() *ClientInfo {
	dst := &ClientInfo{}
	dst.Id = src.Id
	dst.IsHub = src.IsHub
	dst.Props = make([]byte, len(src.Props))
	copy(dst.Props, src.Props)
	return dst
}

func (src *Timestamp) Clone() *Timestamp {
	dst := &Timestamp{}
	dst.Timestamp = timestamppb.New(src.Timestamp.AsTime())
	return dst
}
