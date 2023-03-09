package statemachine

import (
	"errors"
	"fmt"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/entity"
)

var _ Stater = Game{}

type Game struct{}

func (g Game) HandleMessage(message *gamesvc.Message, p *entity.Player, r *entity.Room) (Stater, error) {
	switch msg := message.Message.(type) {
	case *gamesvc.Message_StartTurn:
		return g.handleStartTurn(msg, p, r)
	default:
		return Game{}, &UnknownMessageTypeError{T: message.Message}
	}
}

func (g Game) handleStartTurn(msg *gamesvc.Message_StartTurn, p *entity.Player, r *entity.Room) (Stater, error) {
	if r.IsPlaying {
		return g, errors.New("turn already started")
	}
	if p.ID != r.PlayerIDTurn {
		return g, fmt.Errorf("only player with %s id can start next turn", r.PlayerIDTurn)
	}
	if msg.StartTurn.DurationMs == 0 {
		return g, errors.New("could not start turn with 0 duration")
	}

	r.IsPlaying = true

	r.AnnounceChange()

	return Turn{
		turnDeadline: time.Now().Add(time.Duration(msg.StartTurn.DurationMs) * time.Millisecond),
	}, nil
}
