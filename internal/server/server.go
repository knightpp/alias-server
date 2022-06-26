package server

import (
	"context"
	"net/http"
	"time"

	"github.com/knightpp/alias-server/internal/gravatar"
	"github.com/knightpp/alias-server/internal/storage"

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
	playerDB storage.PlayerDB
}

func New(log zerolog.Logger, playerDB storage.PlayerDB) *Server {
	gameLogger := log.With().Str("component", "game").Logger()
	serverLogger := log.With().Str("component", "server").Logger()
	return &Server{
		log: serverLogger,
		upgrader: websocket.Upgrader{
			EnableCompression: true,
		},
		game:     game.New(gameLogger, playerDB),
		playerDB: playerDB,
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
		log.Err(err).Msg("coudln't bind json")
		c.String(http.StatusBadRequest, "couldn't parse json")
		return
	}

	sock, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Err(err).Msg("upgrade to websocket")
		// upgrader sends response for us
		return
	}

	playerID := c.GetString(middleware.UserIDKey)

	err = s.game.JoinRoom(sock, playerID, options.RoomId)
	if err != nil {
		log.Err(err).Msg("join room failed")
		c.String(http.StatusInternalServerError, "failed to join")
		return
	}
}

func (s *Server) ListRooms(c *gin.Context) {
	log := s.log.With().Interface("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("ListRooms")

	rooms := s.game.ListRooms()
	roomsPb := fp.Map(rooms, func(r *model.Room) *modelpb.Room { return r.ToProto() })

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.playerDB.SetPlayer(ctx, playerPb)
	if err != nil {
		log.Err(err).Msg("failed to create a user")
		c.String(http.StatusInternalServerError, "couldn't create a user")
		return
	}

	c.JSON(http.StatusOK, serverpb.UserSimpleLoginResponse{
		Player: playerPb,
	})
}
