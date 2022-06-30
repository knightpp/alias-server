package actor

import (
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	"github.com/knightpp/alias-server/internal/ws"
	"github.com/rs/zerolog"
)

type Player struct {
	Id          string
	Name        string
	GravatarUrl string

	conn *ws.Conn
}

func NewPlayerFromPB(p *modelpb.Player, conn *ws.Conn) Player {
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

func (p Player) RunLoop(log zerolog.Logger) error {
	log = log.With().Str("player.id", p.Id).Str("component", "game.player").Logger()

	for {
		msg, err := p.conn.ReceiveMessage()
		if err != nil {
			return err
		}

		log.Debug().Interface("msg", msg).Msg("received message")
	}
}
