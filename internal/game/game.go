package game

import (
	"context"
	"errors"
	"fmt"
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/entity"
	"github.com/knightpp/alias-server/internal/game/statemachine"
	"github.com/knightpp/alias-server/internal/tuple"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrRoomNoTeams        = entity.ErrStartNoTeams
	ErrRoomIncompleteTeam = entity.ErrStartIncompleteTeam
	ErrRoomNotFound       = errors.New("room not found")
	ErrPlayerInRoom       = errors.New("player already in the room")
)

type Game struct {
	log zerolog.Logger

	roomsMu sync.Mutex
	rooms   map[string]*entity.Room
}

func New(log zerolog.Logger) *Game {
	return &Game{
		log:   log,
		rooms: make(map[string]*entity.Room),
	}
}

func (g *Game) CreateRoom(
	leader *gamesvc.Player,
	req *gamesvc.CreateRoomRequest,
) (roomID string) {
	roomID = uuidgen.NewString()
	r := entity.NewRoom(g.log, roomID, leader.Id, req)

	go func() {
		state := statemachine.Stater(statemachine.Lobby{})

		for {
			select {
			case <-r.Ctx().Done():
				return
			case tuple := <-r.AggregationChan():
				r.Do(func(r *entity.Room) {
					var err error
					state, err = state.HandleMessage(tuple.A, tuple.B, r)
					if err != nil {
						_ = tuple.B.SendError(err.Error())
					}
				})
			}
		}
	}()
	go func() {
		r.Start()

		g.roomsMu.Lock()
		delete(g.rooms, roomID)
		g.roomsMu.Unlock()
	}()

	g.roomsMu.Lock()
	g.rooms[roomID] = r
	g.roomsMu.Unlock()

	return roomID
}

func (g *Game) ListRooms() []*gamesvc.Room {
	g.roomsMu.Lock()
	defer g.roomsMu.Unlock()

	roomsProto := make([]*gamesvc.Room, 0, len(g.rooms))
	for _, r := range g.rooms {
		proto := runFn1(r, func(r *entity.Room) *gamesvc.Room {
			return r.GetProto()
		})
		roomsProto = append(roomsProto, proto)
	}

	return roomsProto
}

func (g *Game) StartPlayerInRoom(
	roomID string,
	playerProto *gamesvc.Player,
	socket gamesvc.GameService_JoinServer,
) error {
	g.roomsMu.Lock()
	r, ok := g.rooms[roomID]
	g.roomsMu.Unlock()
	if !ok {
		return ErrRoomNotFound
	}

	player := entity.NewPlayer(g.log, socket, playerProto, r)

	err := runFn1(r, func(r *entity.Room) error {
		if r.HasPlayer(player.ID) {
			return ErrPlayerInRoom
		}

		r.Lobby = append(r.Lobby, player)
		r.AnnounceChange()
		return nil
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(r.Ctx())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-player.Chan():
				if !ok {
					return
				}

				select {
				case r.AggregationChan() <- tuple.NewT2(msg, player):
					continue
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	err = player.Start(ctx)
	cancel()
	if err != nil {
		if status.Code(err) != codes.Canceled {
			g.log.
				Err(err).
				Stringer("status_code", status.Code(err)).
				Interface("player", player).
				Msg("tried to send message and something went wrong")
		}

		err = fmt.Errorf("player loop: %w", err)
	}

	r.Do(func(r *entity.Room) {
		needsAnnounce := r.RemovePlayer(player.ID)

		if r.IsEmpty() {
			r.Cancel()
			return
		}

		if needsAnnounce {
			r.AnnounceChange()
		}
	})

	return err
}

func runFn1[R1 any](r *entity.Room, fn func(r *entity.Room) R1) R1 {
	var r1 R1
	wait := make(chan struct{})
	r.Do(func(r *entity.Room) {
		defer close(wait)

		r1 = fn(r)
	})
	<-wait

	return r1
}
