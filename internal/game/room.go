package game

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/mitchellh/copystructure"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/status"
)

type Room struct {
	actorChan chan func(*Room)
	proto     *gamesvc.Room
	password  *string
	log       zerolog.Logger

	Lobby []*Player
	Teams []*Team
}

func runFn0[T any](actorChan chan func(T), fn func(r T)) {
	wait := make(chan struct{})
	actorChan <- func(r T) {
		defer close(wait)

		fn(r)
	}
	<-wait
}

func runFn1[T any, R1 any](actorChan chan func(T), fn func(r T) R1) R1 {
	var r1 R1
	wait := make(chan struct{})
	actorChan <- func(r T) {
		defer close(wait)

		r1 = fn(r)
	}
	<-wait

	return r1
}

func NewRoom(log zerolog.Logger, roomID, leaderID string, req *gamesvc.CreateRoomRequest) *Room {
	return &Room{
		log:       log.With().Str("room-id", roomID).Logger(),
		actorChan: make(chan func(*Room)),
		proto: &gamesvc.Room{
			Id:        roomID,
			Name:      req.Name,
			LeaderId:  leaderID,
			IsPublic:  req.IsPublic,
			Langugage: req.Langugage,
		},
	}
}

func (r *Room) Start() {
	for fn := range r.actorChan {
		fn(r)
	}
}

func (r *Room) GetProto() *gamesvc.Room {
	return runFn1(r.actorChan, func(r *Room) *gamesvc.Room {
		copied, err := copystructure.Copy(r.proto)
		if err != nil {
			panic(err)
		}

		return copied.(*gamesvc.Room)
	})
}

func (r *Room) AddAndStartPlayer(socket gamesvc.GameService_JoinServer, proto *gamesvc.Player) error {
	player := runFn1(r.actorChan, func(r *Room) *Player {
		r.proto.Lobby = append(r.proto.Lobby, proto)

		log := r.log.With().Str("player-id", proto.Id).Str("player-name", proto.Name).Logger()
		player := newPlayer(log, socket, proto)
		r.Lobby = append(r.Lobby, player)
		return player
	})

	runFn0(r.actorChan, func(r *Room) {
		r.announceNewPlayer()
	})

	err := player.Start()
	if err != nil {
		r.errorCallback(player, err)
	}

	return err
}

func (r *Room) HasPlayer(playerID string) bool {
	return runFn1(r.actorChan, func(r *Room) bool {
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
	})
}

func (r *Room) removePlayer(playerID string) bool {
	oldLobbyLen := len(r.proto.Lobby)
	r.proto.Lobby = fp.FilterInPlace(r.proto.Lobby, func(p *gamesvc.Player) bool {
		return p.Id != playerID
	})
	newLobbyLen := len(r.proto.Lobby)

	r.Lobby = fp.FilterInPlace(r.Lobby, func(p *Player) bool {
		// TODO: potential data races if player struct accesses itself
		return p.proto.Id == playerID
	})

	// TODO: filter r.Teams

	var changed bool
	for _, team := range r.proto.Teams {
		if team.PlayerA.Id == playerID {
			changed = true
			team.PlayerA = nil
		}

		if team.PlayerB.Id == playerID {
			changed = true
			team.PlayerB = nil
		}
	}

	return changed || (oldLobbyLen != newLobbyLen)
}

func (r *Room) announceNewPlayer() {
	send := func(p *Player) {
		p.QueueMsg(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: &gamesvc.UpdateRoom{
					Room:     r.proto,
					Password: r.password,
				},
			},
		})
	}

	for _, p := range r.Lobby {
		send(p)
	}

	for _, team := range r.Teams {
		send(team.PlayerA)
		send(team.PlayerB)
	}
}

func (r *Room) errorCallback(player *Player, err error) {
	r.log.
		Err(err).
		Stringer("status_code", status.Code(err)).
		Interface("player", player.proto).
		Msg("tried to send message and something went wrong")

	runFn0(r.actorChan, func(r *Room) {
		ok := r.removePlayer(player.proto.Id)
		if !ok {
			return
		}

		r.announceNewPlayer()
	})
}
