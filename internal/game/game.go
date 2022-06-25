package game

import (
	"fmt"
	"sync"

	"github.com/knightpp/alias-server/internal/data"
	"github.com/rs/zerolog"
)

type Game struct {
	log        zerolog.Logger
	rooms      map[string]data.Room
	roomsMutex sync.Mutex
}

func New(log zerolog.Logger) *Game {
	g := &Game{
		log:        log,
		rooms:      make(map[string]data.Room),
		roomsMutex: sync.Mutex{},
	}
	return g
}

func (g *Game) AddRoom(room data.Room) error {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	g.log.Debug().Interface("room", room).Msg("adding a new room")

	id := string(room.ID)
	_, exists := g.rooms[id]
	if exists {
		return fmt.Errorf("room with id=%s already exists", id)
	}

	g.rooms[id] = room

	return nil
}

func (g *Game) ListRooms() []data.Room {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	rooms := make([]data.Room, len(g.rooms))
	for _, room := range g.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (g *Game) removeRoom(id []byte) {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	g.log.Debug().Str("id", string(id)).Msg("remove room by id")
	delete(g.rooms, string(id))
}
