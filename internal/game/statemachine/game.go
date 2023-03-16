package statemachine

import (
	"errors"
	"fmt"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/entity"
)

var _ Stater = Game{}

type Game struct {
	stats        map[string]*gamesvc.Statistics
	playerIDTurn string
}

func (g Game) HandleMessage(message *gamesvc.Message, p *entity.Player, r *entity.Room) (Stater, error) {
	switch msg := message.Message.(type) {
	case *gamesvc.Message_StartTurn:
		return g.handleStartTurn(msg.StartTurn, p, r)
	case *gamesvc.Message_EndGame:
		return g.handleEndGame(msg.EndGame, p, r)
	default:
		return g, &UnknownMessageTypeError{T: message.Message}
	}
}

func (g Game) handleStartTurn(msg *gamesvc.MsgStartTurn, p *entity.Player, r *entity.Room) (Stater, error) {
	if p.ID != g.playerIDTurn {
		return g, fmt.Errorf("only player with %s id can start next turn", g.playerIDTurn)
	}
	if msg.DurationMs == 0 {
		return g, errors.New("could not start turn with 0 duration")
	}

	err := sendMsgToPlayers(&gamesvc.Message{
		Message: &gamesvc.Message_StartTurn{
			StartTurn: &gamesvc.MsgStartTurn{
				DurationMs: msg.GetDurationMs(),
			},
		},
	}, r.GetAllPlayers()...)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(time.Duration(msg.DurationMs) * time.Millisecond)
	return newTurn(deadline, g), nil
}

func (g Game) handleEndGame(_ *gamesvc.MsgEndGame, sender *entity.Player, r *entity.Room) (Stater, error) {
	if r.LeaderId != sender.ID {
		return g, errors.New("only leader can end game")
	}

	err := sendMsgToPlayers(&gamesvc.Message{
		Message: &gamesvc.Message_Results{
			Results: &gamesvc.MsgResults{
				TeamIdToStats: g.stats,
			},
		},
	}, r.GetAllPlayers()...)

	return Lobby{}, err
}
