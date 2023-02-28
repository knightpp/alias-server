package team

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/player"
)

type Team struct {
	ID   string
	Name string

	PlayerA *player.Player
	PlayerB *player.Player
}

func (t *Team) ToProto() *gamesvc.Team {
	return &gamesvc.Team{
		Id:      t.ID,
		Name:    t.Name,
		PlayerA: t.PlayerA.ToProto(),
		PlayerB: t.PlayerB.ToProto(),
	}
}
