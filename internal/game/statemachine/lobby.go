package statemachine

import (
	"errors"
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/entity"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/life4/genesis/slices"
)

type Lobby struct{}

func (l Lobby) HandleMessage(message *gamesvc.Message, p *entity.Player, r *entity.Room) (Stater, error) {
	switch v := message.Message.(type) {
	case *gamesvc.Message_CreateTeam:
		return handleCreateTeam(v, p, r)
	case *gamesvc.Message_JoinTeam:
		return handleJoinTeam(v, p, r)
	case *gamesvc.Message_TransferLeadership:
		return handleTransferLeadership(v, p, r)
	case *gamesvc.Message_StartGame:
		return handleStartGame(v, p, r)
	default:
		return l, &UnknownMessageTypeError{T: message.Message}
	}
}

func handleCreateTeam(msg *gamesvc.Message_CreateTeam, p *entity.Player, r *entity.Room) (Stater, error) {
	// TODO: return error if no such user
	r.RemovePlayer(p.ID)

	team := &entity.Team{
		ID:      uuidgen.NewString(),
		Name:    msg.CreateTeam.Name,
		PlayerA: p,
		PlayerB: nil,
	}
	r.Teams = append(r.Teams, team)
	r.AnnounceChange()
	return Lobby{}, nil
}

func handleJoinTeam(msg *gamesvc.Message_JoinTeam, p *entity.Player, r *entity.Room) (Stater, error) {
	team, ok := slices.Find(r.Teams, func(t *entity.Team) bool {
		return t.ID == msg.JoinTeam.TeamId
	})
	if ok != nil {
		return Lobby{}, fmt.Errorf("TODO: team not found")
	}

	r.RemovePlayer(p.ID)
	switch {
	case team.PlayerA == nil:
		team.PlayerA = p
	case team.PlayerB == nil:
		team.PlayerB = p
	default:
		return Lobby{}, fmt.Errorf("TODO: team is full")
	}

	r.AnnounceChange()
	return Lobby{}, nil
}

func handleTransferLeadership(msg *gamesvc.Message_TransferLeadership, p *entity.Player, r *entity.Room) (Stater, error) {
	id := msg.TransferLeadership.PlayerId
	exists := r.HasPlayer(id)
	if !exists {
		return Lobby{}, fmt.Errorf("could not transfer leadership: no player with id=%s", id)
	}

	r.LeaderId = id
	r.AnnounceChange()

	return Lobby{}, nil
}

func handleStartGame(msg *gamesvc.Message_StartGame, p *entity.Player, r *entity.Room) (Stater, error) {
	if r.LeaderId != p.ID {
		return Lobby{}, errors.New("only leader id can start game")
	}

	if len(r.Teams) == 0 {
		return Lobby{}, entity.ErrStartNoTeams
	}
	for _, team := range r.Teams {
		if team.PlayerA == nil || team.PlayerB == nil {
			return Lobby{}, entity.ErrStartIncompleteTeam
		}
	}

	firstTeam := r.Teams[0]
	switch {
	case firstTeam.PlayerA != nil:
		r.PlayerIDTurn = firstTeam.PlayerA.ID
	case firstTeam.PlayerB != nil:
		r.PlayerIDTurn = firstTeam.PlayerB.ID
	default:
		return Lobby{}, errors.New("no players in the first team")
	}

	r.IsGameStarted = true

	r.AnnounceChange()

	return Game{}, nil

}
