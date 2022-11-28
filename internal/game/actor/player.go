package actor

import (
	"fmt"

	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"github.com/knightpp/alias-server/pkg/ws"
	"github.com/rs/zerolog"
)

type Player struct {
	Id          string
	Name        string
	GravatarUrl string

	conn ws.Conn
	room *Room
	team *Team
}

func NewPlayerFromPB(p *modelpb.Player, conn ws.Conn) *Player {
	return &Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,

		conn: conn,
		room: nil,
		team: nil,
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
		log.Trace().Interface("msg", msg).Msg("received message")

		switch m := msg.Message.(type) {
		case *serverpb.Message_Fatal:
			return fmt.Errorf("receive message: %s", m.Fatal.Error)
		default:
			log.Warn().Interface("msg", msg).Msg("unhandled message")
		}
	}
}
