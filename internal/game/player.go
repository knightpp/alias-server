package game

import (
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
)

type Player struct {
	proto *gamesvc.Player

	socket  gamesvc.GameService_JoinServer
	msgChan chan func(gamesvc.GameService_JoinServer) error
	log     zerolog.Logger
}

func newPlayer(
	log zerolog.Logger,
	socket gamesvc.GameService_JoinServer,
	proto *gamesvc.Player,
) *Player {
	ch := make(chan func(gamesvc.GameService_JoinServer) error, 1)
	player := &Player{
		log:     log,
		proto:   proto,
		socket:  socket,
		msgChan: ch,
	}

	return player
}

func (p *Player) Start() error {
	var eg errgroup.Group

	eg.Go(func() error {
		for f := range p.msgChan {
			err := f(p.socket)
			if err != nil {
				return fmt.Errorf("execute func: %w", err)
			}
		}

		return nil
	})

	eg.Go(func() error {
		defer close(p.msgChan)

		for {
			msg, err := p.socket.Recv()
			if err != nil {
				return fmt.Errorf("socket recv: %w", err)
			}

			p.log.Debug().RawJSON("msg", []byte(protojson.Format(msg))).Msg("received a message")

			_ = msg
		}
	})

	return eg.Wait()
}

// QueueMsg is a non-blocking send
func (p *Player) QueueMsg(msg *gamesvc.Message) {
	p.msgChan <- func(gs gamesvc.GameService_JoinServer) error {
		return gs.Send(msg)
	}
}
