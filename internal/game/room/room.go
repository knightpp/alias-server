package room

import (
	"errors"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game/player"
	"github.com/knightpp/alias-server/internal/game/team"
	"github.com/knightpp/alias-server/internal/tuple"
	"github.com/rs/zerolog"
)

var (
	ErrStartNoTeams        = errors.New("cannot start game without a single team")
	ErrStartIncompleteTeam = errors.New("cannot start game with incomplete team")
)

type Room struct {
	Id           string
	Name         string
	LeaderId     string
	IsPublic     bool
	Langugage    string
	Password     *string
	Lobby        []*player.Player
	Teams        []*team.Team
	IsPlaying    bool
	PlayerIDTurn string

	allMsgChan chan tuple.T2[*gamesvc.Message, *player.Player]
	actorChan  chan func(*Room)
	done       chan struct{}
	log        zerolog.Logger
}

func NewRoom(
	log zerolog.Logger,
	roomID, leaderID string,
	req *gamesvc.CreateRoomRequest,
) *Room {
	return &Room{
		log:        log.With().Str("room-id", roomID).Logger(),
		actorChan:  make(chan func(*Room)),
		done:       make(chan struct{}),
		allMsgChan: make(chan tuple.T2[*gamesvc.Message, *player.Player]),

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

		case <-r.done:
			return
		}
	}
}

func (r *Room) Cancel() {
	close(r.done)
}

func (r *Room) findTeamOfPlayer(id string) *team.Team {
	for _, team := range r.Teams {
		if team.PlayerA != nil && team.PlayerA.ID == id {
			return team
		}

		if team.PlayerB != nil && team.PlayerB.ID == id {
			return team
		}
	}
	return nil
}

func (r *Room) getAllPlayers() []*player.Player {
	count := len(r.Lobby)
	for _, t := range r.Teams {
		if t.PlayerA != nil {
			count += 1
		}
		if t.PlayerB != nil {
			count += 1
		}
	}

	players := make([]*player.Player, 0, count)

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
	lobby := make([]*gamesvc.Player, len(r.Lobby))
	for i, p := range r.Lobby {
		lobby[i] = p.ToProto()
	}

	teams := make([]*gamesvc.Team, len(r.Teams))
	for i, t := range r.Teams {
		teams[i] = t.ToProto()
	}

	return &gamesvc.Room{
		Id:           r.Id,
		Name:         r.Name,
		LeaderId:     r.LeaderId,
		IsPublic:     r.IsPublic,
		Langugage:    r.Langugage,
		Lobby:        lobby,
		Teams:        teams,
		IsPlaying:    r.IsPlaying,
		PlayerIdTurn: r.PlayerIDTurn,
	}
}

func (r *Room) Done() chan struct{} {
	return r.done
}

func (r *Room) AggregationChan() chan tuple.T2[*gamesvc.Message, *player.Player] {
	return r.allMsgChan
}

func (r *Room) Do(fn func(r *Room)) {
	select {
	case r.actorChan <- fn:
	case <-r.done:
	}
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
	r.Lobby = fp.FilterInPlace(r.Lobby, func(p *player.Player) bool {
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

func (r *Room) AnnounceChange() {
	send := func(p *player.Player) {
		if p == nil {
			return
		}

		p.SendMsg(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: &gamesvc.UpdateRoom{
					Room:     r.GetProto(),
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
