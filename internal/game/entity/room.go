package entity

import (
	"context"
	"errors"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/tuple"
	"github.com/rs/zerolog"
)

var (
	ErrStartNoTeams        = errors.New("cannot start game without a single team")
	ErrStartIncompleteTeam = errors.New("cannot start game with incomplete team")
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

	ctx        context.Context
	cancel     func()
	allMsgChan chan tuple.T2[*gamesvc.Message, *Player]
	actorChan  chan func(*Room)
	log        zerolog.Logger
}

func NewRoom(
	log zerolog.Logger,
	roomID, leaderID string,
	req *gamesvc.CreateRoomRequest,
) *Room {
	ctx, cancel := context.WithCancel(context.Background())
	return &Room{
		ctx:        ctx,
		cancel:     cancel,
		log:        log.With().Str("room-id", roomID).Logger(),
		actorChan:  make(chan func(*Room)),
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
		case fn := <-r.actorChan:
			fn(r)

		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Room) Cancel() {
	r.cancel()
}

func (r *Room) FindTeamWithPlayer(playerID string) (*Team, bool) {
	for _, t := range r.Teams {
		if t.PlayerA != nil && t.PlayerA.ID == playerID {
			return t, true
		}

		if t.PlayerB != nil && t.PlayerB.ID == playerID {
			return t, true
		}
	}
	return nil, false
}

func (r *Room) GetAllPlayers() []*Player {
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
	players = append(players, r.Lobby...)
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

func (r *Room) Ctx() context.Context {
	return r.ctx
}

func (r *Room) AggregationChan() chan tuple.T2[*gamesvc.Message, *Player] {
	return r.allMsgChan
}

// Do queues for execution. It's async method.
func (r *Room) Do(fn func(r *Room)) {
	select {
	case r.actorChan <- fn:
	case <-r.ctx.Done():
	}
}

func (r *Room) IsEmpty() bool {
	if len(r.Lobby) != 0 {
		return false
	}

	for _, team := range r.Teams {
		if team.PlayerA != nil || team.PlayerB != nil {
			return false
		}
	}

	return true
}

func (r *Room) HasPlayer(playerID string) bool {
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

func (r *Room) RemovePlayer(playerID string) bool {
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

func (r *Room) AnnounceChange() error {
	send := func(p *Player) error {
		if p == nil {
			return nil
		}

		return p.SendMsg(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: &gamesvc.UpdateRoom{
					Room:     r.GetProto(),
					Password: r.Password,
				},
			},
		})
	}

	var errs []error

	for _, p := range r.Lobby {
		err := send(p)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, team := range r.Teams {
		err := send(team.PlayerA)
		if err != nil {
			errs = append(errs, err)
		}

		err = send(team.PlayerB)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
