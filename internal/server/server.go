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
	"github.com/knightpp/alias-server/internal/ws"
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

func (s *Server) CreateRoom(c *gin.Context) {
	log := s.log
	log.Trace().Caller().Send()

	var createRequest serverpb.CreateRoomRequest

	err := c.MustBindWith(&createRequest, s.protoBind)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid request: %s", err)
		return
	}

	creatorID := c.GetString(middleware.UserIDKey)
	room := actor.NewRoomFromRequest(&createRequest, creatorID)

	_, err = s.game.CreateRoom(room)
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
	log := s.log
	log.Trace().Caller().Send()

	roomID := c.Param("room_id")
	playerID := c.GetString(middleware.UserIDKey)

	room, ok := s.game.GetRoom(roomID)
	if !ok {
		c.String(http.StatusNotFound, "no such room")
		return
	}

	playerInfo, err := s.game.GetPlayerInfo(playerID)
	if err != nil {
		c.String(http.StatusNotFound, "no such player")
		return
	}

	sock, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Err(err).Msg("upgrade to websocket")
		return
	}

	conn := ws.Wrap(sock)
	defer conn.Close()

	player := actor.NewPlayerFromPB(playerInfo, conn)

	playersWebsocketCurrent.Inc()
	defer playersWebsocketCurrent.Dec()
	playersWebsocketTotal.Inc()

	err = room.AddPlayerToLobby(player)
	if err != nil {
		_ = conn.SendFatal(&serverpb.FatalMessage{
			Error: err.Error(),
		})
		log.Err(err).Msg("AddPlayerToLobby failed")
		return
	}

	defer room.RemovePlayer(playerID)

	err = player.RunLoop(log)
	if err != nil {
		log.Err(err).Msg("player loop failed")
		return
	}
}

func (s *Server) CreateTeam(c *gin.Context) {
	log := s.log
	log.Trace().Caller().Send()

	var req serverpb.CreateTeamRequest

	err := c.MustBindWith(&req, s.protoBind)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid request: %s", err)
		return
	}

	roomID := c.Param("room_id")

	room, ok := s.game.GetRoom(roomID)
	if !ok {
		c.String(http.StatusNotFound, "no such room")
		return
	}

	team, err := room.CreateTeam(&req)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't create team: %s", err)
		return
	}

	c.ProtoBuf(http.StatusOK, &serverpb.CreateTeamResponse{
		Team: team.ToProto(),
	})
}

func (s *Server) ListRooms(c *gin.Context) {
	log := s.log
	log.Trace().Caller().Send()

	rooms := s.game.ListRooms()
	roomsPb := fp.Map(rooms, func(r *actor.Room) *modelpb.Room { return r.ToProto() })

	c.ProtoBuf(http.StatusOK, &serverpb.ListRoomsResponse{Rooms: roomsPb})
}

func (s *Server) UserLogin(c *gin.Context) {
	log := s.log
	log.Trace().Caller().Send()

	var options serverpb.UserSimpleLoginRequest

	err := c.MustBindWith(&options, s.protoBind)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid request: %s", err)
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
