package statemachine

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/player"
	"github.com/knightpp/alias-server/internal/game/room"
)

type Handler interface {
	HandleMessage(message *gamesvc.Message, player *player.Player, room *room.Room) (bool, error)
}

type Lobby struct{}

func (Lobby) HandleMessage(message *gamesvc.Message, player *player.Player) (bool, error) {
	return false, nil
}
