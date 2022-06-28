package game

import (
	"fmt"
	"sync"

	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game/actor"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/rs/zerolog"
)

type Game struct {
	log zerolog.Logger

	rooms      map[string]*actor.Room
	roomsMutex sync.Mutex

	playerDB storage.PlayerDB
}

func New(log zerolog.Logger, playerDB storage.PlayerDB) *Game {
	g := &Game{
		log:        log,
		rooms:      make(map[string]*actor.Room),
		roomsMutex: sync.Mutex{},
		playerDB:   playerDB,
	}
	return g
}

func (g *Game) CreateRoom(room *actor.Room) error {
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

func (g *Game) ListRooms() []*actor.Room {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	return fp.Values(g.rooms)
}

func (g *Game) GetRoom(roomID string) (*actor.Room, bool) {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Lock()

	room, ok := g.rooms[roomID]
	return room, ok
}
