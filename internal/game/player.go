package game

import (
	"fmt"

	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
)

type Player struct {
	ID          string
	Name        string
	GravatarUrl string

	socket  gamesvc.GameService_JoinServer
	msgChan chan func(gamesvc.GameService_JoinServer) error
	log     zerolog.Logger
}

func newPlayer(
	log zerolog.Logger,
	socket gamesvc.GameService_JoinServer,
	proto *gamesvc.Player,
) *Player {
	ch := make(chan func(gamesvc.GameService_JoinServer) error, 1)
	player := &Player{
		ID:          proto.Id,
		Name:        proto.Name,
		GravatarUrl: proto.GravatarUrl,

		log:     log,
		socket:  socket,
		msgChan: ch,
	}

	return player
}

func (p *Player) GetProto() *gamesvc.Player {
	if p == nil {
		return nil
	}

	return &gamesvc.Player{
		Id:          p.ID,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}

func (p *Player) Start(roomChan chan func(*Room)) error {
	var eg errgroup.Group

	eg.Go(func() error {
		for f := range p.msgChan {
			err := f(p.socket)
			if err != nil {
				return fmt.Errorf("execute func: %w", err)
			}
		}

		return nil
	})

	eg.Go(func() error {
		defer close(p.msgChan)

		for {
			msg, err := p.socket.Recv()
			if err != nil {
				return fmt.Errorf("socket recv: %w", err)
			}

			p.log.Debug().RawJSON("msg", []byte(protojson.Format(msg))).Msg("received a message")

			switch v := msg.Message.(type) {
			case *gamesvc.Message_CreateTeam:
				roomChan <- func(r *Room) {
					r.removePlayer(p.ID)

					team := &Team{
						ID:      uuid.NewString(),
						Name:    v.CreateTeam.Name,
						PlayerA: p,
						PlayerB: nil,
					}
					r.Teams = append(r.Teams, team)
				}
			default:
				p.log.Warn().Msgf("unhandled message: %T", msg.Message)
			}
		}
	})

	return eg.Wait()
}

// QueueMsg is a non-blocking send
func (p *Player) QueueMsg(msg *gamesvc.Message) {
	p.msgChan <- func(gs gamesvc.GameService_JoinServer) error {
		return gs.Send(msg)
	}
}
