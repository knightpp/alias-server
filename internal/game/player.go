package game

import (
	"fmt"
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
)

type Player struct {
	proto *gamesvc.Player

	socket      gamesvc.GameService_JoinServer
	msgChan     chan func(gamesvc.GameService_JoinServer) error
	errCallback func(player *Player, err error)
}

func newPlayer(
	socket gamesvc.GameService_JoinServer,
	proto *gamesvc.Player,
	errCallback func(player *Player, err error),
) (*Player, *sync.WaitGroup) {
	ch := make(chan func(gamesvc.GameService_JoinServer) error, 1)
	player := &Player{
		proto:       proto,
		socket:      socket,
		msgChan:     ch,
		errCallback: errCallback,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for f := range ch {
			err := f(socket)
			if err != nil {
				errCallback(player, fmt.Errorf("execute func: %w", err))
				break
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			msg, err := socket.Recv()
			if err != nil {
				errCallback(player, fmt.Errorf("socket recv: %w", err))
				return
			}

			log.Debug().RawJSON("msg", []byte(protojson.Format(msg))).Msg("received a message")

			_ = msg
		}
	}()

	return player, wg
}

// SendMsg is a non-blocking send
func (p *Player) SendMsg(msg *gamesvc.Message) {
	p.msgChan <- func(gs gamesvc.GameService_JoinServer) error {
		return gs.Send(msg)
	}
}
