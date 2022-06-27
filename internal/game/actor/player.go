package actor

import (
	"github.com/gorilla/websocket"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
)

type Player struct {
	Id          string
	Name        string
	GravatarUrl string

	conn *websocket.Conn
}

func NewPlayerFromPB(p *modelpb.Player, conn *websocket.Conn) Player {
	return Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
		conn:        conn,
	}
}

func (p Player) ToProto() *modelpb.Player {
	return &modelpb.Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}

func (p Player) Conn() *websocket.Conn {
	return p.conn
}

func (p Player) NotifyJoined(otherPlayer *modelpb.Player) error {
	// TODO:
	panic("TODO: send a message to websocket")
}
