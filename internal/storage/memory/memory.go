package memory

import (
	"context"
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	"github.com/knightpp/alias-server/internal/storage"
)

var _ storage.Player = (*Memory)(nil)

type Memory struct {
	players map[string]*gamesvc.Player
	mu      sync.Mutex
}

func New() *Memory {
	return &Memory{
		players: make(map[string]*gamesvc.Player),
	}
}

func (m *Memory) SetPlayer(ctx context.Context, token string, p *gamesvc.Player) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.players[token] = clonePlayer(p)

	return nil
}

func (m *Memory) GetPlayer(ctx context.Context, token string) (*gamesvc.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, ok := m.players[token]
	if !ok {
		return nil, storage.ErrNotFound
	}

	return clonePlayer(player), nil
}

func clonePlayer(p *gamesvc.Player) *gamesvc.Player {
	return &gamesvc.Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}
