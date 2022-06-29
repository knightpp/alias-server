package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	custombindings "github.com/knightpp/alias-server/internal/binding"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/game/actor"
	"github.com/knightpp/alias-server/internal/gravatar"
	"github.com/knightpp/alias-server/internal/middleware"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/rs/zerolog"
)

type Server struct {
	log       zerolog.Logger
	upgrader  websocket.Upgrader
	game      *game.Game
	playerDB  storage.PlayerDB
	protoBind binding.Binding
}

func New(log zerolog.Logger, playerDB storage.PlayerDB) *Server {
	gameLogger := log.With().Str("component", "game").Logger()
	serverLogger := log.With().Str("component", "server").Logger()
	return &Server{
		log:       serverLogger,
		upgrader:  websocket.Upgrader{EnableCompression: true},
		game:      game.New(gameLogger, playerDB),
		playerDB:  playerDB,
		protoBind: custombindings.NewProtobuf(log),
	}
}

func (s *Server) Game() *game.Game {
	return s.game
}

func (s *Server) CreateRoom(c *gin.Context) {
	log := s.log.With().Str("remote_ip", c.RemoteIP()).Logger()

	log.Trace().Msg("CreateRoom")

	var createRequest serverpb.CreateRoomRequest

	err := c.MustBindWith(&createRequest, s.protoBind)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json: %s", err)
		return
	}

	creatorID := c.GetString(middleware.UserIDKey)
	room := actor.NewRoomFromRequest(&createRequest, creatorID)

	err = s.game.CreateRoom(room)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't add new room: %s", err)
		return
	}

	roomsCreatedTotal.Inc()

	c.ProtoBuf(http.StatusOK, &serverpb.CreateRoomResponse{
		RoomId: room.Id,
	})
}

func (s *Server) JoinRoom(c *gin.Context) {
	log := s.log.With().Str("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("JoinRoom")

	sock, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Err(err).Msg("upgrade to websocket")
		// upgrader sends response for us
		return
	}

	defer sock.Close()
	defer playersWebsocketCurrent.Dec()

	roomID := c.Param("room_id")
	playerID := c.GetString(middleware.UserIDKey)

	playersWebsocketCurrent.Inc()
	playersWebsocketTotal.Inc()
	err = s.game.JoinRoom(sock, playerID, roomID)
	if err != nil {
		log.Err(err).Msg("join room failed")
		return
	}
}

func (s *Server) ListRooms(c *gin.Context) {
	log := s.log.With().Interface("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("ListRooms")

	rooms := s.game.ListRooms()
	roomsPb := fp.Map(rooms, func(r *actor.Room) *modelpb.Room { return r.ToProto() })

	c.ProtoBuf(http.StatusOK, &serverpb.ListRoomsResponse{Rooms: roomsPb})
}

func (s *Server) UserLogin(c *gin.Context) {
	log := s.log.With().Interface("remote_ip", c.RemoteIP()).Logger()
	log.Trace().Msg("UserLogin")

	var options serverpb.UserSimpleLoginRequest

	err := c.MustBindWith(&options, s.protoBind)
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

	c.ProtoBuf(http.StatusOK, &serverpb.UserSimpleLoginResponse{
		Player: playerPb,
	})
}
