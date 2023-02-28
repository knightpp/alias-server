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

type constantUUIDGen struct{}

func (c constantUUIDGen) NewString() string {
	return TestUUID
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
		err := lis.Close()
		if err != nil {
			log.Err(err).Msg("close listener")
		}
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
