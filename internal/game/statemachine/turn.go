package statemachine

import (
	"errors"
	"fmt"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game/entity"
)

var _ Stater = Turn{}

type Turn struct {
	turnDeadline time.Time
}

func (t Turn) HandleMessage(message *gamesvc.Message, sender *entity.Player, r *entity.Room) (Stater, error) {
	switch msg := message.Message.(type) {
	case *gamesvc.Message_EndTurn:
		if sender.ID != r.PlayerIDTurn {
			return t, fmt.Errorf("only %q player can end turn", r.PlayerIDTurn)
		}

		players := fp.FilterInPlace(r.GetAllPlayers(), func(p *entity.Player) bool {
			return p.ID != sender.ID
		})

		var errs []error
		for _, p := range players {
			err := p.SendMsg(&gamesvc.Message{
				Message: &gamesvc.Message_EndTurn{
					EndTurn: msg.EndTurn,
				},
			})
			if err != nil {
				errs = append(errs, err)
			}
		}

		r.IsPlaying = false
		errs = append(errs, r.AnnounceChange())

		return Game{}, errors.Join(errs...)
	case *gamesvc.Message_Word:
		if time.Now().After(t.turnDeadline) {
			return t, errors.New("turn deadline exceeded")
		}

		team, ok := r.FindTeamWithPlayer(sender.ID)
		if !ok {
			return t, fmt.Errorf("could not find team with player %q", sender.ID)
		}

		oponent, ok := team.OponentOf(sender.ID)
		if !ok {
			return t, fmt.Errorf("could not find oponent in %q team", team.ID)
		}

		players := fp.FilterInPlace(r.GetAllPlayers(), func(p *entity.Player) bool {
			return p.ID != oponent.ID && p.ID != sender.ID
		})

		var errs []error
		for _, p := range players {
			err := p.SendMsg(&gamesvc.Message{
				Message: &gamesvc.Message_Word{Word: &gamesvc.MsgWord{}},
			})
			if err != nil {
				errs = append(errs, err)
			}
		}

		return t, errors.Join(errs...)
	default:
		return t, &UnknownMessageTypeError{T: message.Message}
	}
}
