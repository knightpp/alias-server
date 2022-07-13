package game

import (
	"context"
	"fmt"
	"time"

	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game/actor"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/rs/zerolog"
)

type Game struct {
	log zerolog.Logger

	rooms fp.Locker[map[string]*actor.Room]

	playerDB storage.PlayerDB
}

func New(log zerolog.Logger, playerDB storage.PlayerDB) *Game {
	g := &Game{
		log:      log,
		rooms:    fp.NewLocker(make(map[string]*actor.Room)),
		playerDB: playerDB,
	}
	return g
}

func (g *Game) CreateRoom(room *actor.Room) (*actor.Room, error) {
	return fp.Lock2(g.rooms, func(rooms map[string]*actor.Room) (*actor.Room, error) {
		g.log.Debug().Interface("room", room).Msg("adding a new room")

		id := room.Id
		_, exists := rooms[id]
		if exists {
			return nil, fmt.Errorf("room with id=%s already exists", id)
		}

		room.SetLogger(g.log)
		rooms[id] = room
		return room, nil
	})
}

func (g *Game) ListRooms() []*actor.Room {
	return fp.Lock1(g.rooms, func(rooms map[string]*actor.Room) []*actor.Room {
		return fp.Values(rooms)
	})
}

func (g *Game) GetRoom(roomID string) (*actor.Room, bool) {
	return fp.Lock2(g.rooms, func(rooms map[string]*actor.Room) (*actor.Room, bool) {
		room, ok := rooms[roomID]
		return room, ok
	})
}

func (g *Game) GetPlayerInfo(playerID string) (*modelpb.Player, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return g.playerDB.GetPlayer(ctx, playerID)
}
