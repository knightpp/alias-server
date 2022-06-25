package main

import (
	"flag"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/knightpp/alias-server/internal/middleware"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	flag.Parse()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.Logger.Level(zerolog.TraceLevel)

	gameServer := server.New(log.Logger.With().Str("component", "server").Logger())

	r := gin.Default()
	r.SetTrustedProxies(nil)

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
