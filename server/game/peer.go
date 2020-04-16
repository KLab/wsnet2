package game

import (
	"context"

	"github.com/gorilla/websocket"
)

type Peer struct {
	client *Client
	conn   *websocket.Conn
	msgCh  chan Msg
	done   chan struct{}
}

func NewPeer(ctx context.Context, cli *Client, conn *websocket.Conn) *Peer {
	peer := &Peer{
		client: cli,
		msgCh:  make(chan Msg),

		done: make(chan struct{}),
	}
	return peer
}

func (p *Peer) MsgCh() <-chan Msg {
	return p.msgCh
}

func (p *Peer) Done() <-chan struct{} {
	return p.done
}

func (p *Peer) SendEvent([]Event, int) {
	// memo: errorが起きたら readしてるgoroutineに通知、そこからclientに通知
}
