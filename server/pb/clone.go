package pb

import (
	"google.golang.org/protobuf/proto"
)

func (src *RoomInfo) Clone() *RoomInfo {
	return proto.Clone(src).(*RoomInfo)
}

func (src *ClientInfo) Clone() *ClientInfo {
	return proto.Clone(src).(*ClientInfo)
}

func (src *Timestamp) Clone() *Timestamp {
	return proto.Clone(src).(*Timestamp)
}
