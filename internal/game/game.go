package game

import (
	"fmt"
	"sync"

	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/model"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/rs/zerolog"
)

type Game struct {
	log zerolog.Logger

	rooms      map[string]*model.Room
	roomsMutex sync.Mutex

	playerDB storage.PlayerDB
}

func New(log zerolog.Logger, playerDB storage.PlayerDB) *Game {
	g := &Game{
		log:        log,
		rooms:      make(map[string]*model.Room),
		roomsMutex: sync.Mutex{},
		playerDB:   playerDB,
	}
	return g
}

func (g *Game) RegisterRoom(room *model.Room) error {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	g.log.Debug().Interface("room", room).Msg("adding a new room")

	id := string(room.Id)
	_, exists := g.rooms[id]
	if exists {
		return fmt.Errorf("room with id=%s already exists", id)
	}

	g.rooms[id] = room

	return nil
}

func (g *Game) ListRooms() []*model.Room {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	return fp.Values(g.rooms)
}

func (g *Game) GetRoom(roomID string) (*model.Room, bool) {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Lock()

	room, ok := g.rooms[roomID]
	return room, ok
}
