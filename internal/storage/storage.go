package storage

import (
	"context"

	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
)

type PlayerDB interface {
	SetPlayer(ctx context.Context, p *modelpb.Player) error
	GetPlayer(ctx context.Context, playerID string) (*modelpb.Player, error)
}
