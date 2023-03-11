package statemachine

import (
	"errors"
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/entity"
	"github.com/life4/genesis/slices"
)

var _ Stater = Game{}

type Lobby struct{}

func (l Lobby) HandleMessage(message *gamesvc.Message, p *entity.Player, r *entity.Room) (Stater, error) {
	switch v := message.Message.(type) {
	case *gamesvc.Message_CreateTeam:
		return l.handleCreateTeam(v, p, r)
	case *gamesvc.Message_JoinTeam:
		return l.handleJoinTeam(v, p, r)
	case *gamesvc.Message_TransferLeadership:
		return l.handleTransferLeadership(v, p, r)
	case *gamesvc.Message_StartGame:
		return l.handleStartGame(v, p, r)
	default:
		return l, &UnknownMessageTypeError{T: message.Message}
	}
}

func (l Lobby) handleCreateTeam(msg *gamesvc.Message_CreateTeam, p *entity.Player, r *entity.Room) (Stater, error) {
	team := &entity.Team{
		ID:      uuid.NewString(),
		Name:    msg.CreateTeam.Name,
		PlayerA: nil,
		PlayerB: nil,
	}
	if team.Name == "" {
		team.Name = gofakeit.Vegetable()
	}
	r.Teams = append(r.Teams, team)

	resp := &gamesvc.Message{
		Message: &gamesvc.Message_TeamCreated{
			TeamCreated: &gamesvc.MsgTeamCreated{
				Team: team.ToProto(),
			},
		},
	}
	err := sendMsgToPlayers(resp, r.GetAllPlayers()...)
	if err != nil {
		return l, err
	}

	return l, nil
}

func (l Lobby) handleJoinTeam(msg *gamesvc.Message_JoinTeam, p *entity.Player, r *entity.Room) (Stater, error) {
	team, ok := slices.Find(r.Teams, func(t *entity.Team) bool {
		return t.ID == msg.JoinTeam.TeamId
	})
	if ok != nil {
		return l, fmt.Errorf("TODO: team not found")
	}

	r.RemovePlayer(p.ID)
	switch {
	case team.PlayerA == nil:
		team.PlayerA = p
	case team.PlayerB == nil:
		team.PlayerB = p
	default:
		return l, fmt.Errorf("TODO: team is full")
	}

	r.AnnounceChange()
	return l, nil
}

func (l Lobby) handleTransferLeadership(msg *gamesvc.Message_TransferLeadership, p *entity.Player, r *entity.Room) (Stater, error) {
	id := msg.TransferLeadership.PlayerId
	exists := r.HasPlayer(id)
	if !exists {
		return l, fmt.Errorf("could not transfer leadership: no player with id=%s", id)
	}

	r.LeaderId = id
	r.AnnounceChange()

	return l, nil
}

func (l Lobby) handleStartGame(msg *gamesvc.Message_StartGame, p *entity.Player, r *entity.Room) (Stater, error) {
	if r.LeaderId != p.ID {
		return l, errors.New("only leader id can start game")
	}

	if len(r.Teams) == 0 {
		return l, entity.ErrStartNoTeams
	}
	for _, team := range r.Teams {
		if team.PlayerA == nil || team.PlayerB == nil {
			return l, entity.ErrStartIncompleteTeam
		}
	}

	firstTeam := r.Teams[0]
	switch {
	case firstTeam.PlayerA != nil:
		r.PlayerIDTurn = firstTeam.PlayerA.ID
	case firstTeam.PlayerB != nil:
		r.PlayerIDTurn = firstTeam.PlayerB.ID
	default:
		return l, errors.New("no players in the first team")
	}

	r.IsGameStarted = true

	r.AnnounceChange()

	return Game{}, nil
}
