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

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Err(err).Msg("websocket ReadMessage")
			return fmt.Errorf("websocket read message: %w", err)
		}

		if mt != websocket.BinaryMessage {
			log.Error().Msg("unxpected message type")
			continue
		}

		log.Trace().Bytes("message", message).Msg("received websocket message")
	}
}
