package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/knightpp/alias-server/internal/middleware"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	logger := log.Logger
	logger = logger.Output(os.Stderr)
	logger = logger.Level(zerolog.TraceLevel)
	log.Logger = logger

	gameServer := server.New(logger)

	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(middleware.ZerologLogger(logger))
	r.Use(gin.Recovery())

	log.Info().Msg("starting server")

	r.POST("/user/simple/login", gameServer.UserLogin)
	{
		group := r.Group("/", middleware.Authorized(gameServer.Game()))
		group.GET("rooms", gameServer.ListRooms)
		group.POST("room/join", gameServer.JoinRoom)
		group.POST("room", gameServer.CreateRoom)
	}

	if err := r.Run(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
