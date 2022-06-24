package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	serverpb "github.com/knightpp/alias-server/pkg/server/v1"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	log      zerolog.Logger
	upgrader websocket.Upgrader
}

func New(log zerolog.Logger) *Server {
	return &Server{
		log:      log,
		upgrader: websocket.Upgrader{},
	}
}

func (s *Server) JoinRoom(ctx *gin.Context) {
	log := s.log.With().
		Str("remote_ip", ctx.RemoteIP()).
		Str("room_id", ctx.Param("id")).
		Logger()
	log.Trace().Msg("JoinRoom")
	w := ctx.Writer
	r := ctx.Request

	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Print("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Print("write:", err)
			break
		}
	}
}

func (s *Server) ListRooms(ctx *gin.Context) {
	w := ctx.Writer
	log := s.log.With().Interface("remote_ip", ctx.RemoteIP()).Logger()
	log.Trace().Msg("ListRooms")

	id := uuid.New()
	rooms := serverpb.Rooms{
		Rooms: []*serverpb.Room{
			{
				Id:       id[:],
				Name:     "room 1",
				IsPublic: true,
				Language: serverpb.Language_LANGUAGE_UKR,
				Password: nil,
				Players:  []*serverpb.Player{},
			},
		},
	}

	resp, err := proto.Marshal(&rooms)
	if err != nil {
		log.Err(err).Msg("couldn't marshal protobuf")
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Err(err).Msg("couldn't write response")
		return
	}
}
