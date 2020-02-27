package game

type Msg interface{}

type JoinResponse struct {
	Client      *Client
	MasterID    ClientID
	RoomProps   []byte
	ClientProps map[ClientID][]byte
}

type MsgJoin struct {
	ID  ClientID
	Res chan<- JoinResponse
}

type MsgLeave struct {
	ID ClientID
}
