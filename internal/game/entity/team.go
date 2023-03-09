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

func (t *Team) OponentOf(playerID string) (*Player, bool) {
	if t.PlayerA != nil && t.PlayerA.ID == playerID {
		return t.PlayerB, true
	}

	if t.PlayerB != nil && t.PlayerB.ID == playerID {
		return t.PlayerA, true
	}

	return nil, false
}
