package model

import (
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
)

type Team struct {
	Id      string
	Name    string
	PlayerA Player
	PlayerB Player
}

func (t Team) ToProto() *modelpb.Team {
	return &modelpb.Team{
		Id:      t.Id,
		Name:    t.Name,
		PlayerA: t.PlayerA.ToProto(),
		PlayerB: t.PlayerB.ToProto(),
	}
}
