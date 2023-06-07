package entity

import (
	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	"github.com/rs/zerolog"
)

type Player struct {
	ID          string
	Name        string
	GravatarUrl string

	done chan struct{}
	srv  gamesvc.GameService_JoinRoomServer
	log  zerolog.Logger
}

func NewPlayer(
	log zerolog.Logger,
	srv gamesvc.GameService_JoinRoomServer,
	proto *gamesvc.Player,
) *Player {
	return &Player{
		ID:          proto.Id,
		Name:        proto.Name,
		GravatarUrl: proto.GravatarUrl,

		log: log.With().
			Str("player-id", proto.Id).
			Str("player-name", proto.Name).
			Logger(),
		srv:  srv,
		done: make(chan struct{}),
	}
}

func (p *Player) Done() <-chan struct{} {
	return p.done
}

func (p *Player) Cancel() {
	close(p.done)
}

func (p *Player) ToProto() *gamesvc.Player {
	if p == nil {
		return nil
	}

	return &gamesvc.Player{
		Id:          p.ID,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}

func (p *Player) Announce(ann *gamesvc.Announcement) error {
	return p.srv.Send(&gamesvc.JoinRoomResponse{
		Announcement: ann,
	})
}
