package game

import (
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/life4/genesis/slices"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
)

type Player struct {
	ID          string
	Name        string
	GravatarUrl string

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

func (p *Player) Start(roomChan chan func(*Room)) error {
	var eg errgroup.Group

	eg.Go(func() error {
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
						ID:      p.uuidGen.NewString(),
						Name:    v.CreateTeam.Name,
						PlayerA: p,
						PlayerB: nil,
					}
					r.Teams = append(r.Teams, team)
					r.announceNewPlayer()
				}
			case *gamesvc.Message_JoinTeam:
				roomChan <- func(r *Room) {
					team, ok := slices.Find(r.Teams, func(t *Team) bool {
						return t.ID == v.JoinTeam.TeamId
					})
					if ok != nil {
						p.log.Fatal().Msg("TODO")
						return
					}

					r.removePlayer(p.ID)
					switch {
					case team.PlayerA == nil:
						team.PlayerA = p
					case team.PlayerB == nil:
						team.PlayerB = p
					default:
						p.log.Fatal().Msg("TODO")
						return
					}

					r.announceNewPlayer()
				}
			default:
				p.log.Fatal().Msgf("unhandled message: %T", msg.Message)
			}
		}
	})

	return eg.Wait()
}

// SendMsg is a non-blocking send
func (p *Player) SendMsg(msg *gamesvc.Message) error {
	return p.socket.Send(msg)
}
