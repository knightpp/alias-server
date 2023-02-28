package statemachine

import (
	"errors"
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/player"
	"github.com/knightpp/alias-server/internal/game/room"
	"github.com/knightpp/alias-server/internal/game/team"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/life4/genesis/slices"
)

type Handler interface {
	HandleMessage(message *gamesvc.Message, player *player.Player, room *room.Room) error
}

type Lobby struct{}

func (Lobby) HandleMessage(message *gamesvc.Message, p *player.Player, r *room.Room) error {
	switch v := message.Message.(type) {
	case *gamesvc.Message_CreateTeam:
		// TODO: return error if no such user
		r.RemovePlayer(p.ID)

		team := &team.Team{
			ID:      uuidgen.NewString(),
			Name:    v.CreateTeam.Name,
			PlayerA: p,
			PlayerB: nil,
		}
		r.Teams = append(r.Teams, team)
		r.AnnounceChange()
		return nil
	case *gamesvc.Message_JoinTeam:
		team, ok := slices.Find(r.Teams, func(t *team.Team) bool {
			return t.ID == v.JoinTeam.TeamId
		})
		if ok != nil {
			return fmt.Errorf("TODO: team not found")
		}

		r.RemovePlayer(p.ID)
		switch {
		case team.PlayerA == nil:
			team.PlayerA = p
		case team.PlayerB == nil:
			team.PlayerB = p
		default:
			return fmt.Errorf("TODO: team is full")
		}

		r.AnnounceChange()
		return nil
	case *gamesvc.Message_TransferLeadership:
		id := v.TransferLeadership.PlayerId
		exists := r.HasPlayer(id)
		if !exists {
			return fmt.Errorf("could not transfer leadership: no player with id=%s", id)
		}

		r.LeaderId = id
		r.AnnounceChange()

		return nil
	case *gamesvc.Message_StartGame:
		if r.LeaderId != p.ID {
			return errors.New("only leader id can start game")
		}

		if len(r.Teams) == 0 {
			return room.ErrStartNoTeams
		}
		for _, team := range r.Teams {
			if team.PlayerA == nil || team.PlayerB == nil {
				return room.ErrStartIncompleteTeam
			}
		}

		firstTeam := r.Teams[0]
		switch {
		case firstTeam.PlayerA != nil:
			r.PlayerIDTurn = firstTeam.PlayerA.ID
		case firstTeam.PlayerB != nil:
			r.PlayerIDTurn = firstTeam.PlayerB.ID
		default:
			return errors.New("no players in the first team")
		}

		r.IsPlaying = true

		r.AnnounceChange()

		return nil
	default:
		return &UnknownMessageTypeError{T: message.Message}
	}
}

type UnknownMessageTypeError struct {
	T any
}

func (err *UnknownMessageTypeError) Error() string {
	return fmt.Sprintf("unhandled message: %T", err.T)
}
