package storage

//go:generate mockery --name PlayerDB --with-expecter

import (
	"context"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
)

type PlayerDB interface {
	SetPlayer(ctx context.Context, p *gamesvc.Player) error
	GetPlayer(ctx context.Context, playerID string) (*gamesvc.Player, error)
}
