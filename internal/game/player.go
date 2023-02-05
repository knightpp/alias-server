package game

import (
	"fmt"
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
)

type Player struct {
	ID          string
	Name        string
	GravatarUrl string

	once    sync.Once
	done    chan struct{}
	msgChan chan *gamesvc.Message
	uuidGen uuidgen.Generator
	socket  gamesvc.GameService_JoinServer
	log     zerolog.Logger
}

func newPlayer(
	log zerolog.Logger,
	gen uuidgen.Generator,
	socket gamesvc.GameService_JoinServer,
	proto *gamesvc.Player,
) *Player {
	return &Player{
		ID:          proto.Id,
		Name:        proto.Name,
		GravatarUrl: proto.GravatarUrl,

		log:     log,
		uuidGen: gen,
		socket:  socket,
		done:    make(chan struct{}),
		msgChan: make(chan *gamesvc.Message),
	}
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

func (p *Player) Start() error {
	for {
		msg, err := p.socket.Recv()
		if err != nil {
			return fmt.Errorf("socket recv: %w", err)
		}

		evt := p.log.Debug()
		if evt.Enabled() {
			evt.RawJSON("msg", []byte(protojson.Format(msg))).Msg("received a message")
		}

		select {
		case <-p.done:
			return nil
		case p.msgChan <- msg:
			continue
		}
	}
}

func (p *Player) Cancel() {
	p.once.Do(func() {
		close(p.done)
	})
}

func (p *Player) SendMsg(msg *gamesvc.Message) error {
	return p.socket.Send(msg)
}
