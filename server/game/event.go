package game

type Event interface{}

type EvJoined struct{
	Client *Client
}

