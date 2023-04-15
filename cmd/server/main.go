package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
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
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger().
		Level(zerolog.TraceLevel)

	if err := run(log); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

func run(log zerolog.Logger) error {
	var playerDB storage.Player
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
		return fmt.Errorf("listen socket: %w", err)
	}

	log.Info().Str("addr", addr).Msg("starting GRPC server")

	grpcLog := interceptorLogger(log)
	grpcServer := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(grpcLog),
			recovery.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(grpcLog),
			recovery.UnaryServerInterceptor(),
		),
	)
	gamesvc.RegisterGameServiceServer(grpcServer, gameServer)
	loginsvc.RegisterLoginServiceServer(grpcServer, loginservice.New(playerDB))
	return grpcServer.Serve(lis)
}

func interceptorLogger(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l = l.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
