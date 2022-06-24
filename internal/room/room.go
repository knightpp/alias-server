package room

import serverpb "github.com/knightpp/alias-server/pkg/server/v1"

type Room struct{}

func New() *Room {
	return &Room{}
}

func (r *Room) Test() {
	_ = serverpb.Player{}
}
