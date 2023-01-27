package memory

import (
	"context"
	"sync"

	clone "github.com/huandu/go-clone/generic"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
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

	m.players[token] = clone.Clone(p)

	return nil
}

func (m *Memory) GetPlayer(ctx context.Context, token string) (*gamesvc.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, ok := m.players[token]
	if !ok {
		return nil, storage.ErrNotFound
	}

	return clone.Clone(player), nil
}
