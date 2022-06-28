package game

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	"github.com/knightpp/alias-server/internal/game/actor"
)

func (g *Game) JoinRoom(conn *websocket.Conn, playerID, roomID string) error {
	defer conn.Close()
	log := g.log
	log.Trace().Str("player_id", playerID).Str("room_id", roomID).Msg("JoinRoom")

	room, ok := g.GetRoom(roomID)
	if !ok {
		return fmt.Errorf("no such room")
	}

	playerPb, err := func() (*modelpb.Player, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return g.playerDB.GetPlayer(ctx, playerID)
	}()
	if err != nil {
		return fmt.Errorf("get player from database: %w", err)
	}

	player := actor.NewPlayerFromPB(playerPb, conn)

	err = room.AddPlayerToLobby(player)
	if err != nil {
		return fmt.Errorf("add player to lobby: %w", err)
	}

	err = player.NotifyLeft("Test uuid")
	if err != nil {
		return err
	}

	err = player.RunLoop()
	return err
}