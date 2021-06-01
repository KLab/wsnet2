package pb

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (src *RoomInfo) Clone() *RoomInfo {
	dst := &RoomInfo{}
	*dst = *src
	dst.Number = &RoomNumber{Number: src.Number.Number}
	dst.PublicProps = make([]byte, len(src.PublicProps))
	dst.PrivateProps = make([]byte, len(src.PrivateProps))
	copy(dst.PublicProps, src.PublicProps)
	copy(dst.PrivateProps, src.PrivateProps)
	dst.Created = src.Created.Clone()
	return dst
}

func (src *ClientInfo) Clone() *ClientInfo {
	dst := &ClientInfo{}
	*dst = *src
	dst.Props = make([]byte, len(src.Props))
	copy(dst.Props, src.Props)

	return dst
}

func (src *Timestamp) Clone() *Timestamp {
	dst := &Timestamp{}
	*dst = *src
	dst.Timestamp = timestamppb.New(src.Timestamp.AsTime())
	return dst
}
