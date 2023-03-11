package statemachine

import (
	"errors"
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/entity"
)

type Stater interface {
	HandleMessage(message *gamesvc.Message, player *entity.Player, room *entity.Room) (Stater, error)
}

type UnknownMessageTypeError struct {
	T any
}

func (err *UnknownMessageTypeError) Error() string {
	return fmt.Sprintf("unhandled message: %T", err.T)
}

func sendMsgToPlayers(msg *gamesvc.Message, players ...*entity.Player) error {
	var errs []error
	for _, player := range players {
		err := player.SendMsg(msg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
