package entity

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
)

type Team struct {
	ID   string
	Name string

	PlayerA *Player
	PlayerB *Player
}

func (t *Team) ToProto() *gamesvc.Team {
	return &gamesvc.Team{
		Id:      t.ID,
		Name:    t.Name,
		PlayerA: t.PlayerA.ToProto(),
		PlayerB: t.PlayerB.ToProto(),
	}
}
