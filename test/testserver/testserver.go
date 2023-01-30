package testserver

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/knightpp/alias-server/internal/storage/memory"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const TestUUID = "00000000-0000-0000-0000-000000000000"

type TestServer struct {
	playerDB storage.Player
	lis      net.Listener
	service  *server.GameService
	log      zerolog.Logger
}

type constantUUIDGen struct{}

func (c constantUUIDGen) NewString() string {
	return TestUUID
}

func CreateAndStart(t *testing.T) (*TestServer, error) {
	t.Helper()

	playerDB := memory.New()
	log := zerolog.New(zerolog.NewTestWriter(t))
	gameServer := server.New(log, playerDB, constantUUIDGen{})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen socket: %w", err)
	}

	log.Info().Str("addr", lis.Addr().String()).Msg("starting GRPC server")

	grpcServer := grpc.NewServer()
	go func() {
		gamesvc.RegisterGameServiceServer(grpcServer, gameServer)
		_ = grpcServer.Serve(lis)
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
	})

	return &TestServer{
		playerDB: playerDB,
		service:  gameServer,
		lis:      lis,
		log:      log,
	}, nil
}

func (ts *TestServer) NewPlayer(ctx context.Context, player *gamesvc.Player) (*TestPlayer, error) {
	token := uuid.NewString()
	err := ts.playerDB.SetPlayer(ctx, token, player)
	if err != nil {
		return nil, fmt.Errorf("set player: %w", err)
	}

	conn, err := grpc.DialContext(ctx, ts.lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	client := gamesvc.NewGameServiceClient(conn)

	return newTestPlayer(client, player, token), nil
}
