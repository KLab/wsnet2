package game

import (
	"context"

	"github.com/gorilla/websocket"
)

type Peer struct {
	client *Client
	MsgCh chan Msg

	done  chan struct{}
}

func NewPeer(ctx context.Context, cli *Client, conn *websocket.Conn) *Peer {
	peer := &Peer{
		client: cli,
		MsgCh: make(chan Msg),

		done: make(chan struct{}),
	}
	return peer
}

func (p *Peer) Done() <-chan struct{} {
	return p.done
}
