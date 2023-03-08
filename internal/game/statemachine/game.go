package statemachine

import (
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/entity"
)

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
	if p.ID != r.PlayerIDTurn {
		return g, fmt.Errorf("only player with %s id can start next turn", r.PlayerIDTurn)
	}

	r.IsPlaying = true

	r.AnnounceChange()

	return g, nil
}
