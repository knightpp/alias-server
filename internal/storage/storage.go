package storage

//go:generate mockery --name PlayerDB --with-expecter

import (
	"context"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
)

type Player interface {
	SetPlayer(ctx context.Context, token string, p *gamesvc.Player) error
	GetPlayer(ctx context.Context, token string) (*gamesvc.Player, error)
}
