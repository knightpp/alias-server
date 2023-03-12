// go:build test
package testserver

import (
	"context"
	"fmt"
	"net"

	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/knightpp/alias-server/internal/storage/memory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const TestUUID = "00000000-0000-0000-0000-000000000000"

type TestServer struct {
	playerDB storage.Player
	addr     string
	service  *server.GameService
	log      zerolog.Logger
}

func CreateAndStart() (*TestServer, error) {
	playerDB := memory.New()
	log := zerolog.New(zerolog.TestWriter{
		T:     GinkgoT(),
		Frame: 4,
	})
	gameServer := server.New(log, playerDB)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen socket: %w", err)
	}

	log.Info().Str("addr", lis.Addr().String()).Msg("starting GRPC server")

	grpcServer := grpc.NewServer()
	gamesvc.RegisterGameServiceServer(grpcServer, gameServer)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	DeferCleanup(func() {
		grpcServer.GracefulStop()
	})

	return &TestServer{
		playerDB: playerDB,
		service:  gameServer,
		log:      log,
		addr:     lis.Addr().String(),
	}, nil
}

func (ts *TestServer) NewPlayer(ctx context.Context, player *gamesvc.Player) (*TestPlayer, error) {
	token := uuid.NewString()
	err := ts.playerDB.SetPlayer(ctx, token, player)
	if err != nil {
		return nil, fmt.Errorf("set player: %w", err)
	}

	conn, err := grpc.DialContext(ctx, ts.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	client := gamesvc.NewGameServiceClient(conn)

	log := ts.log.With().
		Str("player.name", player.Name).
		Str("player.id", player.Id).
		Logger()
	return newTestPlayer(client, player, token, log), nil
}

func (ts *TestServer) JoinPlayers(ctx context.Context, roomID string, players ...*TestPlayer) []*TestPlayerInRoom {
	inRoom := make([]*TestPlayerInRoom, 0, len(players))
	for _, player := range players {
		conn, err := player.Join(roomID)
		Expect(err).ShouldNot(HaveOccurred())

		inRoom = append(inRoom, conn)

		for _, inRoomPlayer := range inRoom {
			Expect(inRoomPlayer.NextMsg(ctx).GetUpdateRoom()).ShouldNot(BeNil())
		}
	}

	return inRoom
}

func (ts *TestServer) CreatePlayers(ctx context.Context, n int, gen func(n int) *gamesvc.Player) []*TestPlayer {
	players := make([]*TestPlayer, n)
	for i := 0; i < n; i++ {
		player, err := ts.NewPlayer(ctx, gen(i+1))
		Expect(err).To(BeNil())

		players[i] = player
	}
	return players
}
