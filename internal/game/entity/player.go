package entity

import (
	"context"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
)

type Player struct {
	ID          string
	Name        string
	GravatarUrl string

	Room *Room

	msgChan chan *gamesvc.Message
	socket  gamesvc.GameService_JoinServer
	log     zerolog.Logger
}

func NewPlayer(
	log zerolog.Logger,
	socket gamesvc.GameService_JoinServer,
	proto *gamesvc.Player,
	room *Room,
) *Player {
	return &Player{
		ID:          proto.Id,
		Name:        proto.Name,
		GravatarUrl: proto.GravatarUrl,
		Room:        room,

		log: log.With().
			Str("player-id", proto.Id).
			Str("player-name", proto.Name).
			Logger(),
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

func (p *Player) Start(ctx context.Context) error {
	for {
		msg, err := p.socket.Recv()
		if err != nil {
			return err
		}

		evt := p.log.Debug()
		if evt.Enabled() {
			json, _ := protojson.Marshal(msg)
			evt.RawJSON("msg", json).Msg("received a message")
		}

		select {
		case <-ctx.Done():
			return nil
		case p.msgChan <- msg:
			continue
		}
	}
}

func (p *Player) SendMsg(msg *gamesvc.Message) error {
	return p.socket.Send(msg)
}

func (p *Player) SendError(err string) error {
	return p.SendMsg(&gamesvc.Message{
		Message: &gamesvc.Message_Error{
			Error: &gamesvc.MsgError{
				Error: err,
			},
		},
	})
}

func (p *Player) Chan() <-chan *gamesvc.Message {
	return p.msgChan
}
