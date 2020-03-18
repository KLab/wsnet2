package pb

func (src *RoomInfo) Clone() *RoomInfo {
	dst := &RoomInfo{}
	*dst = *src
	dst.PublicProps = make([]byte, len(src.PublicProps))
	dst.PrivateProps = make([]byte, len(src.PrivateProps))
	copy(dst.PublicProps, src.PublicProps)
	copy(dst.PrivateProps, src.PrivateProps)
	return dst
}

func (src *ClientInfo) Clone() *ClientInfo {
	dst := &ClientInfo{}
	*dst = *src
	dst.Props = make([]byte, len(src.Props))
	copy(dst.Props, src.Props)

	return dst
}
