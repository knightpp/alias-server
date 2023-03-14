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
	stats map[string]*gamesvc.Statistics
}

func (g Game) HandleMessage(message *gamesvc.Message, p *entity.Player, r *entity.Room) (Stater, error) {
	switch msg := message.Message.(type) {
	case *gamesvc.Message_StartTurn:
		return g.handleStartTurn(msg, p, r)
	case *gamesvc.Message_EndGame:
		return g.handleEndGame(msg, p, r)
	default:
		return g, &UnknownMessageTypeError{T: message.Message}
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

	err := r.AnnounceChange()
	if err != nil {
		return nil, err
	}

	if g.stats == nil {
		g.stats = make(map[string]*gamesvc.Statistics, len(r.Teams))
	}

	return Turn{
		turnDeadline: time.Now().Add(time.Duration(msg.StartTurn.DurationMs) * time.Millisecond),
		prev:         g,
	}, nil
}

func (g Game) handleEndGame(msg *gamesvc.Message_EndGame, sender *entity.Player, r *entity.Room) (Stater, error) {
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
