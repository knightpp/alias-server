// go:build test
package testserver

import (
	"context"
	"fmt"

	clone "github.com/huandu/go-clone/generic"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-proto/go/mdkey"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TestPlayer struct {
	authToken string
	player    *gamesvc.Player
	client    gamesvc.GameServiceClient
	log       zerolog.Logger
}

func newTestPlayer(
	client gamesvc.GameServiceClient,
	player *gamesvc.Player,
	auth string,
	log zerolog.Logger,
) *TestPlayer {
	return &TestPlayer{
		client:    client,
		authToken: auth,
		player:    clone.Clone(player),
		log:       log,
	}
}

func (tp *TestPlayer) Proto() *gamesvc.Player {
	return tp.player
}

func (tp *TestPlayer) Join(roomID string) (*TestPlayerInRoom, error) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, mdkey.RoomID, roomID, mdkey.Auth, tp.authToken)
	ctx, cancel := context.WithCancel(ctx)

	sock, err := tp.client.Join(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("join socket: %w", err)
	}

	playerInRoom := &TestPlayerInRoom{
		done:   make(chan struct{}),
		C:      make(chan *gamesvc.Message),
		logger: tp.log.With().Str("room.id", roomID).Logger(),
		sock:   sock,
		player: tp.player,
		cancel: cancel,
	}

	ginkgo.DeferCleanup(func() {
		playerInRoom.Cancel()
	})

	go func() {
		defer ginkgo.GinkgoRecover()

		err := playerInRoom.Start()
		if status.Code(err) == codes.Canceled {
			return
		}

		Expect(err).ShouldNot(HaveOccurred())
	}()

	return playerInRoom, nil
}

func (tp *TestPlayer) CreateRoom(ctx context.Context, req *gamesvc.CreateRoomRequest) (string, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, mdkey.Auth, tp.authToken)

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
