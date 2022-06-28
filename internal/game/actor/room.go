package actor

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"github.com/knightpp/alias-server/internal/fp"
)

type Room struct {
	Id        string
	Name      string
	LeaderId  string
	IsPublic  bool
	Langugage string
	Lobby     map[string]Player
	Teams     map[string]Team

	Password *string

	mutex sync.Mutex
}

func NewRoomFromRequest(
	req *serverpb.CreateRoomRequest,
	creatorID string,
) *Room {
	id := uuid.New().String()
	return &Room{
		Id:        id,
		Name:      req.Name,
		LeaderId:  creatorID,
		IsPublic:  req.IsPublic,
		Langugage: req.Language,
		Lobby:     map[string]Player{},
		Teams:     map[string]Team{},
		Password:  req.Password,
		mutex:     sync.Mutex{},
	}
}

func (r *Room) ToProto() *modelpb.Room {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return &modelpb.Room{
		Id:        r.Id,
		Name:      r.Name,
		LeaderId:  r.LeaderId,
		IsPublic:  r.IsPublic,
		Langugage: r.Langugage,
		Lobby:     fp.Map(fp.Values(r.Lobby), func(p Player) *modelpb.Player { return p.ToProto() }),
		Teams:     fp.Map(fp.Values(r.Teams), func(t Team) *modelpb.Team { return t.ToProto() }),
	}
}

func (r *Room) AddPlayerToLobby(p Player) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, ok := r.Lobby[p.Id]
	if ok {
		return fmt.Errorf("player is already in the lobby")
	}

	otherPlayers := fp.Values(r.Lobby)
	r.Lobby[p.Id] = p

	go func() {
		// TODO: parallel
		for _, p := range otherPlayers {
			// TODO: log error
			p.NotifyJoined(p.ToProto())
		}
	}()

	return nil
}
