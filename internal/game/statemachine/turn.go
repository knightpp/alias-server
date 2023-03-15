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
	prev         Game
}

func newTurn(deadline time.Time, prev Game) Turn {
	return Turn{
		turnDeadline: deadline,
		prev:         prev,
	}
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

		err := sendMsgToPlayers(&gamesvc.Message{
			Message: &gamesvc.Message_EndTurn{
				EndTurn: msg.EndTurn,
			},
		}, players...)
		if err != nil {
			return t, err
		}

		team, ok := r.FindTeamWithPlayer(r.PlayerIDTurn)
		if ok {
			prevStats, ok := t.prev.stats[team.ID]
			if ok {
				prevStats.Rights += msg.EndTurn.Stats.Rights
				prevStats.Wrongs += msg.EndTurn.Stats.Wrongs
			} else {
				prevStats = msg.EndTurn.Stats
			}

			t.prev.stats[team.ID] = prevStats
		}

		return t.prev, nil
	case *gamesvc.Message_Word:
		if r.PlayerIDTurn != sender.ID {
			return t, fmt.Errorf("only player %q can send word", r.PlayerIDTurn)
		}
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

		err := sendMsgToPlayers(&gamesvc.Message{
			Message: &gamesvc.Message_Word{Word: &gamesvc.MsgWord{
				Word: msg.Word.GetWord(),
			}},
		}, players...)

		return t, err
	default:
		return t, &UnknownMessageTypeError{T: message.Message}
	}
}
