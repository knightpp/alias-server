package storage

import (
	"context"
	"errors"

	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
)

var ErrNotFound = errors.New("player not found")

//go:generate mockery --name Player --with-expecter
type Player interface {
	SetPlayer(ctx context.Context, token string, p *gamesvc.Player) error
	GetPlayer(ctx context.Context, token string) (*gamesvc.Player, error)
}
