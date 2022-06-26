package model

import (
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
	Lobby     []Player
	Teams     []Team

	Password *string
}

func NewRoomFromRequest(
	req *serverpb.CreateRoomRequest,
	creatorID string,
) Room {
	id := uuid.New().String()
	return Room{
		Id:        id,
		Name:      req.Name,
		IsPublic:  req.IsPublic,
		Langugage: req.Language,
		LeaderId:  creatorID,
		Password:  req.Password,
	}
}

func (r Room) ToProto() *modelpb.Room {
	return &modelpb.Room{
		Id:        r.Id,
		Name:      r.Name,
		LeaderId:  r.LeaderId,
		IsPublic:  r.IsPublic,
		Langugage: r.Langugage,
		Lobby:     fp.Map(r.Lobby, func(p Player) *modelpb.Player { return p.ToProto() }),
		Teams:     fp.Map(r.Teams, func(t Team) *modelpb.Team { return t.ToProto() }),
	}
}

type Player struct {
	Id          string
	Name        string
	GravatarUrl string
}

func NewPlayerFromPB(p *modelpb.Player) Player {
	return Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}

func (p Player) ToProto() *modelpb.Player {
	return &modelpb.Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}

type Team struct {
	Id      string
	Name    string
	PlayerA Player
	PlayerB Player
}

func NewTeamFromPB(t *modelpb.Team) Team {
	return Team{
		Id:      t.Id,
		Name:    t.Name,
		PlayerA: NewPlayerFromPB(t.PlayerA),
		PlayerB: NewPlayerFromPB(t.PlayerB),
	}
}

func (t Team) ToProto() *modelpb.Team {
	return &modelpb.Team{
		Id:      t.Id,
		Name:    t.Name,
		PlayerA: t.PlayerA.ToProto(),
		PlayerB: t.PlayerB.ToProto(),
	}
}
