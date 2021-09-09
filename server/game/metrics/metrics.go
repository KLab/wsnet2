package metrics

import (
	"expvar"
)

var (
	expmap      = expvar.NewMap("wsnet2.game")
	Conns       = new(expvar.Int)
	Rooms       = new(expvar.Int)
	MessageSent = new(expvar.Int)
	MessageRecv = new(expvar.Int)
)

func init() {
	expmap.Set("conns", Conns)
	expmap.Set("rooms", Rooms)
	expmap.Set("message_sent", MessageSent)
	expmap.Set("message_recv", MessageRecv)
}
