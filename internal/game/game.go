package game

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/room"
	"github.com/rs/zerolog"
)

var (
	ErrRoomNoTeams = room.ErrStartNoTeams
)

type Game struct {
}

func New() *Game {
	return &Game{}
}

func (g *Game) SpawnRoom(
	log zerolog.Logger,
	roomID, leaderID string,
	req *gamesvc.CreateRoomRequest,
) *room.Room {
	return room.NewRoom(log, roomID, leaderID, req)
}
