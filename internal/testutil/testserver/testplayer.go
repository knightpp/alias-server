// go:build test
package testserver

import (
	"context"
	"fmt"

	clone "github.com/huandu/go-clone/generic"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/onsi/ginkgo/v2"
	"google.golang.org/grpc/metadata"
)

type TestPlayer struct {
	authToken string
	player    *gamesvc.Player
	client    gamesvc.GameServiceClient
}

func newTestPlayer(client gamesvc.GameServiceClient, player *gamesvc.Player, auth string) *TestPlayer {
	return &TestPlayer{
		client:    client,
		authToken: auth,
		player:    clone.Clone(player),
	}
}

func (tp *TestPlayer) Proto() *gamesvc.Player {
	return tp.player
}

func (tp *TestPlayer) Join(roomID string) (*TestPlayerInRoom, error) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, server.RoomIDKey, roomID, server.AuthKey, tp.authToken)
	ctx, cancel := context.WithCancel(ctx)

	sock, err := tp.client.Join(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("join socket: %w", err)
	}

	playerInRoom := &TestPlayerInRoom{
		done:      make(chan struct{}),
		C:         make(chan *gamesvc.Message),
		sock:      sock,
		authToken: tp.authToken,
		player:    tp.player,
		cancel:    cancel,
	}

	ginkgo.DeferCleanup(func() {
		playerInRoom.Cancel()
	})

	go func() {
		err := playerInRoom.Start()
		if err != nil {
			// TODO: Fail
			// tp.t.Fail()
		}
	}()

	return playerInRoom, nil
}

func (tp *TestPlayer) CreateRoom(ctx context.Context, req *gamesvc.CreateRoomRequest) (string, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, server.AuthKey, tp.authToken)

	resp, err := tp.client.CreateRoom(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

func (tp *TestPlayer) CreateRoomAndJoin(ctx context.Context, req *gamesvc.CreateRoomRequest) (*TestPlayerInRoom, error) {
	roomID, err := tp.CreateRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	return tp.Join(roomID)
}
