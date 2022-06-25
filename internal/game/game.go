package game

import (
	"fmt"
	"sync"

	"github.com/knightpp/alias-server/internal/data"
	"github.com/rs/zerolog"
)

type Game struct {
	log zerolog.Logger

	rooms      map[string]data.Room
	roomsMutex sync.Mutex

	players      map[string]data.Player
	playersMutex sync.Mutex
}

func New(log zerolog.Logger) *Game {
	g := &Game{
		log:          log,
		rooms:        make(map[string]data.Room),
		roomsMutex:   sync.Mutex{},
		players:      make(map[string]data.Player),
		playersMutex: sync.Mutex{},
	}
	return g
}

func (g *Game) RegisterRoom(room data.Room) error {
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

	rooms := make([]data.Room, 0, len(g.rooms))
	for _, room := range g.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (g *Game) RegisterPlayer(player data.Player) {
	g.playersMutex.Lock()
	defer g.playersMutex.Unlock()

	g.players[string(player.ID)] = player
}

func (g *Game) AddPlayerToRoom(playerID data.PlayerID, roomID data.RoomID) error {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	g.playersMutex.Lock()
	defer g.playersMutex.Unlock()

	room, ok := g.rooms[string(roomID)]
	if !ok {
		return fmt.Errorf("no such room")
	}

	player, ok := g.players[string(playerID)]
	if !ok {
		return fmt.Errorf("no such player")
	}

	room.Lobby = append(room.Lobby, player)

	return nil
}

func (g *Game) IsPlayerExists(playerID data.PlayerID) bool {
	g.playersMutex.Lock()
	defer g.playersMutex.Unlock()

	_, ok := g.players[string(playerID)]
	return ok
}

func (g *Game) removeRoom(id []byte) {
	g.roomsMutex.Lock()
	defer g.roomsMutex.Unlock()

	g.log.Debug().Str("id", string(id)).Msg("remove room by id")
	delete(g.rooms, string(id))
}
