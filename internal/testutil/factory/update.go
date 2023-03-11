package factory

import (
	clone "github.com/huandu/go-clone/generic"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
)

type UpdateRoomOption func(*gamesvc.UpdateRoom)

type Room struct {
	req  *gamesvc.CreateRoomRequest
	opts []UpdateRoomOption
}

func NewRoom(req *gamesvc.CreateRoomRequest) *Room {
	return &Room{
		req:  req,
		opts: nil,
	}
}

func (r *Room) Clone() *Room {
	return clone.Clone(r)
}

func (r *Room) Build() *gamesvc.Message {
	msg := &gamesvc.UpdateRoom{
		Room: &gamesvc.Room{
			Id:           testserver.TestUUID,
			Name:         r.req.Name,
			LeaderId:     "",
			IsPublic:     r.req.IsPublic,
			Langugage:    r.req.Langugage,
			Lobby:        nil,
			Teams:        nil,
			IsPlaying:    false,
			PlayerIdTurn: "",
		},
		Password: nil,
	}

	for _, opt := range r.opts {
		opt(msg)
	}

	if msg.Room.LeaderId == "" {
		panic("LeaderId must not be empty")
	}

	return &gamesvc.Message{
		Message: &gamesvc.Message_UpdateRoom{
			UpdateRoom: msg,
		},
	}
}

func (r *Room) WithLeader(leaderID string) *Room {
	return r.append(func(ur *gamesvc.UpdateRoom) {
		ur.Room.LeaderId = leaderID
	})
}

func (r *Room) WithLobby(players ...*gamesvc.Player) *Room {
	return r.append(func(ur *gamesvc.UpdateRoom) {
		ur.Room.Lobby = players
	})
}

func (r *Room) WithTeams(teams ...*gamesvc.Team) *Room {
	return r.append(func(ur *gamesvc.UpdateRoom) {
		ur.Room.Teams = teams
	})
}

func (r *Room) WithStartedGame(started bool) *Room {
	return r.append(func(ur *gamesvc.UpdateRoom) {
		ur.Room.IsGameStarted = started
	})
}

func (r *Room) WithIsPlaying(playing bool) *Room {
	return r.append(func(ur *gamesvc.UpdateRoom) {
		ur.Room.IsPlaying = playing
	})
}

func (r *Room) WithPlayerIDTurn(id string) *Room {
	return r.append(func(ur *gamesvc.UpdateRoom) {
		ur.Room.PlayerIdTurn = id
	})
}

func (r *Room) append(opt UpdateRoomOption) *Room {
	r.opts = append(r.opts, opt)
	return r
}
