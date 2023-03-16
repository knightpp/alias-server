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
		return t.handeEndTurn(msg.EndTurn, sender, r)
	case *gamesvc.Message_Word:
		return t.handleWord(msg.Word, sender, r)
	default:
		return t, &UnknownMessageTypeError{T: message.Message}
	}
}

func (t Turn) handeEndTurn(msg *gamesvc.MsgEndTurn, sender *entity.Player, r *entity.Room) (Stater, error) {
	if sender.ID != t.prev.playerIDTurn {
		return t, fmt.Errorf("only %q player can end turn", t.prev.playerIDTurn)
	}

	players := fp.FilterInPlace(r.GetAllPlayers(), func(p *entity.Player) bool {
		return p.ID != sender.ID
	})

	err := sendMsgToPlayers(&gamesvc.Message{
		Message: &gamesvc.Message_EndTurn{
			EndTurn: msg,
		},
	}, players...)
	if err != nil {
		return t, err
	}

	team, ok := r.FindTeamWithPlayer(sender.ID)
	if ok {
		prevStats, ok := t.prev.stats[team.ID]
		if ok {
			prevStats.Rights += msg.Stats.Rights
			prevStats.Wrongs += msg.Stats.Wrongs
		} else {
			prevStats = msg.Stats
		}

		t.prev.stats[team.ID] = prevStats
	}

	return t.prev, nil
}

func (t Turn) handleWord(msg *gamesvc.MsgWord, sender *entity.Player, r *entity.Room) (Stater, error) {
	if t.prev.playerIDTurn != sender.ID {
		return t, fmt.Errorf("only player %q can send word", t.prev.playerIDTurn)
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
			Word: msg.GetWord(),
		}},
	}, players...)

	return t, err
}
