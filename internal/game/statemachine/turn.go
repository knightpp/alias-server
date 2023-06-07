package statemachine

import (
	"errors"
	"fmt"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game/entity"
)

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
