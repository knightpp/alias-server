package game

import (
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
)

type Player struct {
	ID          string
	Name        string
	GravatarUrl string

	msgChan chan *gamesvc.Message
	uuidGen uuidgen.Generator
	socket  gamesvc.GameService_JoinServer
	log     zerolog.Logger
}

func newPlayer(
	log zerolog.Logger,
	gen uuidgen.Generator,
	socket gamesvc.GameService_JoinServer,
	proto *gamesvc.Player,
) *Player {
	return &Player{
		ID:          proto.Id,
		Name:        proto.Name,
		GravatarUrl: proto.GravatarUrl,

		log:     log,
		uuidGen: gen,
		socket:  socket,
		msgChan: make(chan *gamesvc.Message),
	}
}

func (p *Player) ToProto() *gamesvc.Player {
	if p == nil {
		return nil
	}

	return &gamesvc.Player{
		Id:          p.ID,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}

func (p *Player) Start() error {
	for {
		msg, err := p.socket.Recv()
		if err != nil {
			return fmt.Errorf("socket recv: %w", err)
		}

		p.log.Debug().RawJSON("msg", []byte(protojson.Format(msg))).Msg("received a message")

		p.msgChan <- msg
	}
}

// SendMsg is a non-blocking send
func (p *Player) SendMsg(msg *gamesvc.Message) error {
	return p.socket.Send(msg)
}
