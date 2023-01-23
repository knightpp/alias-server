package game

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/status"
)

type Room struct {
	Id        string
	Name      string
	LeaderId  string
	IsPublic  bool
	Langugage string
	Password  *string

	Lobby []*Player
	Teams []*Team

	actorChan chan func(*Room)
	log       zerolog.Logger
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
		Id:        roomID,
		Name:      req.Name,
		LeaderId:  leaderID,
		IsPublic:  req.IsPublic,
		Langugage: req.Langugage,
		Password:  req.Password,
	}
}

func (r *Room) Start() {
	for fn := range r.actorChan {
		fn(r)
	}
}

func (r *Room) GetProto() *gamesvc.Room {
	return runFn1(r.actorChan, func(r *Room) *gamesvc.Room {
		return r.getProto()
	})
}

func (r *Room) getProto() *gamesvc.Room {
	return &gamesvc.Room{
		Id:        r.Id,
		Name:      r.Name,
		LeaderId:  r.LeaderId,
		IsPublic:  r.IsPublic,
		Langugage: r.Langugage,
		Lobby:     []*gamesvc.Player{},
		Teams:     []*gamesvc.Team{},
	}
}

func (r *Room) getLobbyProto() []*gamesvc.Player {
	return nil
}

func (r *Room) AddAndStartPlayer(socket gamesvc.GameService_JoinServer, proto *gamesvc.Player) error {
	player := runFn1(r.actorChan, func(r *Room) *Player {
		log := r.log.With().Str("player-id", proto.Id).Str("player-name", proto.Name).Logger()
		player := newPlayer(log, socket, proto)
		r.Lobby = append(r.Lobby, player)
		return player
	})

	runFn0(r.actorChan, func(r *Room) {
		r.announceNewPlayer()
	})

	err := player.Start(r.actorChan)
	if err != nil {
		r.errorCallback(player, err)
	}

	return err
}

func (r *Room) HasPlayer(playerID string) bool {
	return runFn1(r.actorChan, func(r *Room) bool {
		for _, player := range r.Lobby {
			if player.ID == playerID {
				return true
			}
		}

		for _, team := range r.Teams {
			if team.PlayerA.ID == playerID {
				return true
			}

			if team.PlayerB.ID == playerID {
				return true
			}
		}
		return false
	})
}

func (r *Room) removePlayer(playerID string) bool {
	oldLobbyLen := len(r.Lobby)
	r.Lobby = fp.FilterInPlace(r.Lobby, func(p *Player) bool {
		// TODO: potential data races if player struct accesses itself
		return p.ID != playerID
	})
	newLobbyLen := len(r.Lobby)

	var changed bool
	for _, team := range r.Teams {
		if team.PlayerA.ID == playerID {
			changed = true
			team.PlayerA = nil
		}

		if team.PlayerB.ID == playerID {
			changed = true
			team.PlayerB = nil
		}
	}

	return changed || (oldLobbyLen != newLobbyLen)
}

func (r *Room) announceNewPlayer() {
	send := func(p *Player) {
		if p == nil {
			return
		}

		p.QueueMsg(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: &gamesvc.UpdateRoom{
					Room:     r.getProto(),
					Password: r.Password,
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
		Interface("player", player).
		Msg("tried to send message and something went wrong")

	runFn0(r.actorChan, func(r *Room) {
		ok := r.removePlayer(player.ID)
		if !ok {
			return
		}

		r.announceNewPlayer()
	})
}
