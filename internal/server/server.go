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
		log:      log.With().Str("component", "server").Logger(),
		upgrader: websocket.Upgrader{},
		game:     game.New(log.With().Str("component", "game").Logger()),
	}
}

func (s *Server) Game() *game.Game {
	return s.game
}

func (s *Server) CreateRoom(c *gin.Context) {
	log := s.log.With().Str("remote_ip", c.RemoteIP()).Logger()

	log.Trace().Msg("CreateRoom")

	var options data.CreateRoomRequest

	err := c.BindJSON(&options)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json: %s", err)
		return
	}

	id := uuid.New().String()
	room := data.Room{
		ID:       data.RoomID(id),
		Name:     options.Name,
		IsPublic: options.IsPublic,
		Language: options.Language,
		Lobby:    []data.Player{},
		Teams:    []data.Team{},
	}

	err = s.game.RegisterRoom(room)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't add new room: %s", err)
		return
	}

	c.JSON(http.StatusOK, data.CreateRoomResponse{
		ID: data.RoomID(id),
	})
}

func (s *Server) JoinRoom(c *gin.Context) {
	log := s.log.With().Str("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("JoinRoom")

	var options data.JoinRoomRequest

	err := c.BindJSON(&options)
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't parse json: %s", err)
		return
	}

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

func (s *Server) UserLogin(c *gin.Context) {
	log := s.log.With().Interface("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("UserLogin")

	var options data.UserSimpleLoginRequest

	err := c.BindJSON(&options)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json: %s", err)
		return
	}

	id := uuid.New().String()
	player := data.Player{
		ID:          data.PlayerID(id),
		Name:        options.Name,
		GravatarURL: "", // TODO: add gravatar
	}
	s.game.RegisterPlayer(player)

	c.JSON(http.StatusOK, data.UserSimpleLoginResponse{
		Player: player,
	})
}
