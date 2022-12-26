package main

import (
	"fmt"
	"net"
	"os"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	loginsvc "github.com/knightpp/alias-proto/go/login_service"
	"github.com/knightpp/alias-server/internal/loginservice"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/knightpp/alias-server/internal/storage/redis"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.TraceLevel)

	if err := run(log); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

func run(log zerolog.Logger) error {
	var playerDB storage.PlayerDB
	if url, ok := os.LookupEnv("REDIS_URL"); ok {
		pdb, err := redis.NewFromURL(url)
		if err != nil {
			return err
		}

		playerDB = pdb
	} else if addr, ok := os.LookupEnv("REDIS_ADDR"); ok {
		playerDB = redis.New(addr)
	} else {
		return fmt.Errorf("REDIS_ADDR must not be empty")
	}

	gameServer := server.New(log, playerDB)

	addr := fmt.Sprintf("localhost:%d", 8080)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}

	log.Info().Str("addr", addr).Msg("starting GRPC server")

	grpcServer := grpc.NewServer()
	gamesvc.RegisterGameServiceServer(grpcServer, gameServer)
	loginsvc.RegisterLoginServiceServer(grpcServer, loginservice.New(playerDB))
	return grpcServer.Serve(lis)
}
