package game

import (
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/mitchellh/copystructure"
	"github.com/rs/zerolog/log"
)

type Room struct {
	protoMu  sync.Mutex
	proto    *gamesvc.Room
	password *string

	mu    sync.Mutex
	Lobby []*Player
	Teams []*Team
}

func NewRoom(roomID, leaderID string, req *gamesvc.CreateRoomRequest) *Room {
	return &Room{
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

	player, done := newPlayer(socket, proto, r.errorCallback)
	r.Lobby = append(r.Lobby, player)

	r.announceNewPlayer()

	return done
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
	log.Err(err).Interface("player", player.proto).
		Msg("tried to send message and something went wrong")
}
