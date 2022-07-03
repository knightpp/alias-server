package actor

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/rs/zerolog"
)

type Room struct {
	Id       string
	Name     string
	LeaderId string
	IsPublic bool
	Language string
	Lobby    map[string]Player
	Teams    map[string]Team

	Password *string

	mutex sync.Mutex
	log   zerolog.Logger
}

func NewRoomFromRequest(
	req *serverpb.CreateRoomRequest,
	creatorID string,
) *Room {
	id := uuid.New().String()
	return &Room{
		Id:       id,
		Name:     req.Name,
		LeaderId: creatorID,
		IsPublic: req.IsPublic,
		Language: req.Language,
		Lobby:    make(map[string]Player),
		Teams:    make(map[string]Team),
		Password: req.Password,
	}
}

func (r *Room) ToProto() *modelpb.Room {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.toProtoUnsafe()
}

func (r *Room) toProtoUnsafe() *modelpb.Room {
	return &modelpb.Room{
		Id:        r.Id,
		Name:      r.Name,
		LeaderId:  r.LeaderId,
		IsPublic:  r.IsPublic,
		Langugage: r.Language,
		Lobby:     fp.Map(fp.Values(r.Lobby), func(p Player) *modelpb.Player { return p.ToProto() }),
		Teams:     fp.Map(fp.Values(r.Teams), func(t Team) *modelpb.Team { return t.ToProto() }),
	}
}

func (r *Room) SetLogger(log zerolog.Logger) {
	r.log = log.With().
		Str("component", "game.room").
		Str("room.id", r.Id).
		Str("room.name", r.Name).
		Logger()
}

func (r *Room) RemovePlayer(playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.Lobby, playerID)

	left := &serverpb.PlayerLeftMessage{
		PlayerId: playerID,
	}

	for _, p := range r.Lobby {

		err := p.conn.SendPlayerLeft(left)
		if err != nil {
			r.log.Err(err).Msg("failed to send player left msg")
		}
	}

	return nil
}

func (r *Room) AddPlayerToLobby(p Player) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, ok := r.Lobby[p.Id]
	if ok {
		return fmt.Errorf("player is already in the lobby")
	}

	otherPlayers := fp.Values(r.Lobby)
	r.Lobby[p.Id] = p

	err := p.conn.SendInitRoom(&serverpb.InitRoomMessage{
		Room: r.toProtoUnsafe(),
	})
	if err != nil {
		return fmt.Errorf("websocket send: %w", err)
	}

	if len(otherPlayers) > 0 {
		go func() {
			// TODO: parallel
			for _, p := range otherPlayers {
				err := p.conn.SendPlayerJoined(&serverpb.PlayerJoinedMessage{
					Player: p.ToProto(),
				})
				if err != nil {
					r.log.Err(err).Msg("NotifyJoined failed")
				}
			}
		}()
	}

	return nil
}
