package server

import (
	"net/http"

	"github.com/knightpp/alias-server/internal/gravatar"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/middleware"
	"github.com/knightpp/alias-server/internal/model"
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

	var createRequest serverpb.CreateRoomRequest

	err := c.BindJSON(&createRequest)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json: %s", err)
		return
	}

	creatorID := c.GetString(middleware.UserIDKey)
	room := model.NewRoomFromRequest(&createRequest, creatorID)

	err = s.game.RegisterRoom(room)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't add new room: %s", err)
		return
	}

	c.JSON(http.StatusOK, serverpb.CreateRoomResponse{
		RoomId: room.Id,
	})
}

func (s *Server) JoinRoom(c *gin.Context) {
	log := s.log.With().Str("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("JoinRoom")

	var options serverpb.JoinRoomRequest

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
	roomsPb := fp.Map(rooms, func(r model.Room) *modelpb.Room { return r.ToProto() })

	c.JSON(http.StatusOK, serverpb.ListRoomsResponse{Rooms: roomsPb})
}

func (s *Server) UserLogin(c *gin.Context) {
	log := s.log.With().Interface("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("UserLogin")

	var options serverpb.UserSimpleLoginRequest

	err := c.BindJSON(&options)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json: %s", err)
		return
	}

	id := uuid.New().String()
	playerPb := &modelpb.Player{
		Id:          id,
		Name:        options.Name,
		GravatarUrl: gravatar.GetUrlOrDefault(options.Email),
	}
	player := model.NewPlayerFromPB(playerPb)
	s.game.RegisterPlayer(player)

	c.JSON(http.StatusOK, serverpb.UserSimpleLoginResponse{
		Player: playerPb,
	})
}
