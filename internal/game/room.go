package game

import (
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/tuple"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/life4/genesis/slices"
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
	Lobby     []*Player
	Teams     []*Team

	allMsgChan chan tuple.T2[*gamesvc.Message, *Player]
	actorChan  chan func(*Room)
	done       chan struct{}
	log        zerolog.Logger
	gen        uuidgen.Generator
}

func runFn1[R1 any](r *Room, fn func(r *Room) R1) R1 {
	var r1 R1
	wait := make(chan struct{})
	r.Do(func(r *Room) {
		defer close(wait)

		r1 = fn(r)
	})
	<-wait

	return r1
}

func NewRoom(
	log zerolog.Logger,
	roomID, leaderID string,
	req *gamesvc.CreateRoomRequest,
	gen uuidgen.Generator,
) *Room {
	return &Room{
		gen:        gen,
		log:        log.With().Str("room-id", roomID).Logger(),
		actorChan:  make(chan func(*Room)),
		done:       make(chan struct{}),
		allMsgChan: make(chan tuple.T2[*gamesvc.Message, *Player]),

		Id:        roomID,
		Name:      req.Name,
		LeaderId:  leaderID,
		IsPublic:  req.IsPublic,
		Langugage: req.Langugage,
		Password:  req.Password,
	}
}

func (r *Room) Start() {
	for {
		select {
		case tuple := <-r.allMsgChan:
			err := r.handleMessage(tuple.A, tuple.B)
			if err != nil {
				_ = tuple.B.SendMsg(&gamesvc.Message{
					Message: &gamesvc.Message_Error{
						Error: &gamesvc.MsgError{
							Error: fmt.Sprintf("handle message: %s", err),
						},
					},
				})
			}

		case fn := <-r.actorChan:
			fn(r)

		case <-r.done:
			return
		}
	}
}

func (r *Room) Cancel() {
	close(r.done)
}

func (r *Room) handleMessage(msg *gamesvc.Message, p *Player) error {
	switch v := msg.Message.(type) {
	case *gamesvc.Message_CreateTeam:
		// TODO: return error if no such user
		r.removePlayer(p.ID)

		team := &Team{
			ID:      p.uuidGen.NewString(),
			Name:    v.CreateTeam.Name,
			PlayerA: p,
			PlayerB: nil,
		}
		r.Teams = append(r.Teams, team)
		r.announceChange()
		return nil
	case *gamesvc.Message_JoinTeam:
		team, ok := slices.Find(r.Teams, func(t *Team) bool {
			return t.ID == v.JoinTeam.TeamId
		})
		if ok != nil {
			return fmt.Errorf("TODO: team not found")
		}

		r.removePlayer(p.ID)
		switch {
		case team.PlayerA == nil:
			team.PlayerA = p
		case team.PlayerB == nil:
			team.PlayerB = p
		default:
			return fmt.Errorf("TODO: team is full")
		}

		r.announceChange()
		return nil
	case *gamesvc.Message_TransferLeadership:
		id := v.TransferLeadership.PlayerId
		exists := r.hasPlayer(id)
		if !exists {
			return fmt.Errorf("could not transfer leadership: no player with id=%s", id)
		}

		r.LeaderId = id
		r.announceChange()

		return nil
	default:
		return fmt.Errorf("unhandled message: %T", msg.Message)
	}
}

func (r *Room) getAllPlayers() []*Player {
	count := len(r.Lobby)
	for _, t := range r.Teams {
		if t.PlayerA != nil {
			count += 1
		}
		if t.PlayerB != nil {
			count += 1
		}
	}

	players := make([]*Player, 0, count)

	for _, p := range r.Lobby {
		players = append(players, p)
	}
	for _, t := range r.Teams {
		if t.PlayerA != nil {
			players = append(players, t.PlayerA)
		}
		if t.PlayerB != nil {
			players = append(players, t.PlayerB)
		}
	}

	return players
}

func (r *Room) GetProto() *gamesvc.Room {
	return runFn1(r, func(r *Room) *gamesvc.Room {
		return r.getProto()
	})
}

func (r *Room) getProto() *gamesvc.Room {
	lobby := make([]*gamesvc.Player, len(r.Lobby))
	for i, p := range r.Lobby {
		lobby[i] = p.ToProto()
	}

	teams := make([]*gamesvc.Team, len(r.Teams))
	for i, t := range r.Teams {
		teams[i] = t.ToProto()
	}

	return &gamesvc.Room{
		Id:        r.Id,
		Name:      r.Name,
		LeaderId:  r.LeaderId,
		IsPublic:  r.IsPublic,
		Langugage: r.Langugage,
		Lobby:     lobby,
		Teams:     teams,
	}
}

func (r *Room) getLobbyProto() []*gamesvc.Player {
	return nil
}

func (r *Room) AddAndStartPlayer(socket gamesvc.GameService_JoinServer, proto *gamesvc.Player) error {
	log := r.log.With().Str("player-id", proto.Id).Str("player-name", proto.Name).Logger()
	player := newPlayer(log, r.gen, socket, proto)

	r.Do(func(r *Room) {
		r.Lobby = append(r.Lobby, player)
		r.announceChange()
	})

	go func() {
		for {
			select {
			case <-r.done:
				player.Cancel()
				return

			case msg, ok := <-player.msgChan:
				if !ok {
					return
				}

				select {
				case r.allMsgChan <- tuple.NewT2(msg, player):
					continue
				case <-r.done:
					return
				}
			}
		}
	}()

	err := player.Start()
	if err != nil {
		r.log.
			Err(err).
			Stringer("status_code", status.Code(err)).
			Interface("player", player).
			Msg("tried to send message and something went wrong")

		r.Do(func(r *Room) {
			ok := r.removePlayer(player.ID)
			if !ok {
				return
			}

			r.announceChange()
		})
		return fmt.Errorf("player loop: %w", err)
	}

	return nil
}

func (r *Room) Do(fn func(r *Room)) {
	select {
	case r.actorChan <- fn:
	case <-r.done:
	}
}

func (r *Room) HasPlayer(playerID string) bool {
	return runFn1(r, func(r *Room) bool {
		return r.hasPlayer(playerID)
	})
}

func (r *Room) hasPlayer(playerID string) bool {
	for _, player := range r.Lobby {
		if player.ID == playerID {
			return true
		}
	}

	for _, team := range r.Teams {
		if team.PlayerA != nil && team.PlayerA.ID == playerID {
			return true
		}

		if team.PlayerB != nil && team.PlayerB.ID == playerID {
			return true
		}
	}
	return false
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
		if team.PlayerA != nil && team.PlayerA.ID == playerID {
			changed = true
			team.PlayerA = nil
		}

		if team.PlayerB != nil && team.PlayerB.ID == playerID {
			changed = true
			team.PlayerB = nil
		}
	}

	return changed || (oldLobbyLen != newLobbyLen)
}

func (r *Room) announceChange() {
	send := func(p *Player) {
		if p == nil {
			return
		}

		p.SendMsg(&gamesvc.Message{
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
