package game

import (
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/mitchellh/copystructure"
	"github.com/rs/zerolog"
)

type Room struct {
	protoMu  sync.Mutex
	proto    *gamesvc.Room
	password *string
	log      zerolog.Logger

	mu    sync.Mutex
	Lobby []*Player
	Teams []*Team
}

func NewRoom(log zerolog.Logger, roomID, leaderID string, req *gamesvc.CreateRoomRequest) *Room {
	return &Room{
		log:     log.With().Str("room-id", roomID).Logger(),
		protoMu: sync.Mutex{},
		mu:      sync.Mutex{},
		proto: &gamesvc.Room{
			Id:        roomID,
			Name:      req.Name,
			LeaderId:  leaderID,
			IsPublic:  req.IsPublic,
			Langugage: req.Langugage,
		},
	}
}

func (r *Room) GetProto() *gamesvc.Room {
	r.protoMu.Lock()
	defer r.protoMu.Unlock()

	copied, err := copystructure.Copy(r.proto)
	if err != nil {
		panic(err)
	}

	return copied.(*gamesvc.Room)
}

func (r *Room) AddPlayer(socket gamesvc.GameService_JoinServer, proto *gamesvc.Player) *sync.WaitGroup {
	r.protoMu.Lock()
	defer r.protoMu.Unlock()

	r.proto.Lobby = append(r.proto.Lobby, proto)

	log := r.log.With().Str("player-id", proto.Id).Str("player-name", proto.Name).Logger()
	player, done := newPlayer(log, socket, proto, r.errorCallback)
	r.Lobby = append(r.Lobby, player)

	r.announceNewPlayer()

	return done
}

func (r *Room) HasPlayer(playerID string) bool {
	r.protoMu.Lock()
	defer r.protoMu.Unlock()

	for _, player := range r.proto.Lobby {
		if player.Id == playerID {
			return true
		}
	}

	for _, team := range r.proto.Teams {
		if team.PlayerA.Id == playerID {
			return true
		}

		if team.PlayerB.Id == playerID {
			return true
		}
	}

	return false
}

func (r *Room) announceNewPlayer() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.Lobby {
		p.SendMsg(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: &gamesvc.UpdateRoom{
					Room:     r.proto,
					Password: r.password,
				},
			},
		})
	}
}

func (r *Room) errorCallback(player *Player, err error) {
	r.log.Err(err).Interface("player", player.proto).
		Msg("tried to send message and something went wrong")

	// TODO: remove player and announce it
}
