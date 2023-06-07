package entity

import (
	"errors"
	"sync"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
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

	Mutex        sync.Mutex
	PlayerIDTurn string
	TurnDeadline time.Time
	TotalStats   map[string]*gamesvc.Statistics
	TurnStats    *gamesvc.Statistics

	log zerolog.Logger
}

func NewRoom(
	log zerolog.Logger,
	roomID, leaderID string,
	req *gamesvc.CreateRoomRequest,
) *Room {
	return &Room{
		log:        log.With().Str("room-id", roomID).Logger(),
		Id:         roomID,
		Name:       req.Name,
		LeaderId:   leaderID,
		IsPublic:   req.IsPublic,
		Langugage:  req.Langugage,
		Password:   req.Password,
		TotalStats: make(map[string]*gamesvc.Statistics),
		TurnStats:  &gamesvc.Statistics{},
	}
}

func (r *Room) Announce(ann *gamesvc.Announcement) error {
	var errs []error
	for _, player := range r.GetAllPlayers() {
		errs = append(errs, player.Announce(ann))
	}
	return errors.Join(errs...)
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

func (r *Room) FindPlayer(playerID string) (*Player, bool) {
	for _, player := range r.GetAllPlayers() {
		if player.ID == playerID {
			return player, true
		}
	}
	return nil, false
}

func (r *Room) RemovePlayer(playerID string) (*Player, bool) {
	for i, player := range r.Lobby {
		if player.ID == playerID {
			// TODO: optimize by reusing the slice
			r.Lobby = append(r.Lobby[:i], r.Lobby[i+1:]...)
			return player, true
		}
	}

	var player *Player
	for _, team := range r.Teams {
		if team.PlayerA != nil && team.PlayerA.ID == playerID {
			player = team.PlayerA
			team.PlayerA = nil
		}

		if team.PlayerB != nil && team.PlayerB.ID == playerID {
			player = team.PlayerA
			team.PlayerB = nil
		}

		if player != nil {
			return player, true
		}
	}

	return nil, false
}
