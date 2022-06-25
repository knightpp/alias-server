package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/knightpp/alias-server/internal/data"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/rs/zerolog"
)

type Server struct {
	log      zerolog.Logger
	upgrader websocket.Upgrader
	game     *game.Game
}

func New(log zerolog.Logger) *Server {
	return &Server{
		log:      log,
		upgrader: websocket.Upgrader{},
		game:     game.New(log.With().Str("component", "game").Logger()),
	}
}

func (s *Server) CreateRoom(c *gin.Context) {
	log := s.log.With().Str("remote_ip", c.RemoteIP()).Logger()

	log.Trace().Msg("CreateRoom")

	var options data.CreateRoomOptions

	err := c.BindJSON(&options)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json: %s", err)
		return
	}

	id := uuid.New()
	room := data.Room{
		ID:       id[:],
		Name:     options.Name,
		IsPublic: options.IsPublic,
		Language: options.Language,
		Lobby:    []data.Player{},
		Teams:    []data.Team{},
	}

	err = s.game.AddRoom(room)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't add new room: %s", err)
		return
	}

	c.JSON(http.StatusOK, data.CreateRoomResponse{
		ID: id[:],
	})
}

func (s *Server) JoinRoom(c *gin.Context) {
	log := s.log.With().
		Str("remote_ip", c.RemoteIP()).
		Str("room_id", c.Param("id")).
		Logger()
	log.Trace().Msg("JoinRoom")

	sock, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Err(err).Msg("upgrade to websocket")
		return
	}

	defer sock.Close()
	for {
		mt, message, err := sock.ReadMessage()
		if err != nil {
			log.Err(err).Msg("websocket ReadMessage")
			break
		}

		log.Debug().Bytes("message", message).Msg("received websocket message")

		err = sock.WriteMessage(mt, message)
		if err != nil {
			log.Err(err).Msg("websocket WriteMessage")
			break
		}
	}
}

func (s *Server) ListRooms(c *gin.Context) {
	log := s.log.With().Interface("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("ListRooms")

	rooms := s.game.ListRooms()

	c.JSON(http.StatusOK, data.ListRoomsResponse{Rooms: rooms})
}
