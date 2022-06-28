package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/knightpp/alias-server/internal/middleware"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/knightpp/alias-server/internal/storage/redis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	logger := log.Logger
	logger = logger.Output(selectLogOutput())
	logger = logger.Level(zerolog.TraceLevel)
	log.Logger = logger

	if err := run(logger); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

func run(logger zerolog.Logger) error {
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

	gameServer := server.New(logger, playerDB)

	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(middleware.ZerologLogger(logger))
	r.Use(gin.Recovery())

	log.Info().Msg("starting server")

	r.POST("/user/login/simple", gameServer.UserLogin)
	{
		group := r.Group("/", middleware.Authorized(playerDB))
		group.GET("rooms", gameServer.ListRooms)
		group.Any("room/join/:room_id", gameServer.JoinRoom)
		group.POST("room", gameServer.CreateRoom)
	}

	if err := r.Run(); err != nil {
		return fmt.Errorf("gin run: %w", err)
	}

	return nil
}

func selectLogOutput() io.Writer {
	if os.Getenv("GIN_MODE") == "release" {
		return os.Stderr
	}
	return zerolog.ConsoleWriter{Out: os.Stderr}
}
