package main

import (
	"flag"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

func main() {
	flag.Parse()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.Logger.Level(zerolog.TraceLevel)

	gameServer := server.New(log.Logger)

	r := gin.Default()
	log.Info().Msg("starting server")
	r.GET("/room/join/:id", gameServer.JoinRoom)
	r.GET("/rooms", gameServer.ListRooms)

	if err := r.Run(*addr); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
