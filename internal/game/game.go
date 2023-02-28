package game

import (
	"errors"
	"fmt"
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game/player"
	"github.com/knightpp/alias-server/internal/game/room"
	"github.com/knightpp/alias-server/internal/game/statemachine"
	"github.com/knightpp/alias-server/internal/tuple"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/status"
)

var (
	ErrRoomNoTeams  = room.ErrStartNoTeams
	ErrRoomNotFound = errors.New("room not found")
	ErrPlayerInRoom = errors.New("player already in the room")
)

type Game struct {
	log zerolog.Logger

	roomsMu sync.Mutex
	rooms   map[string]*room.Room
}

func New(log zerolog.Logger) *Game {
	return &Game{
		log:   log,
		rooms: make(map[string]*room.Room),
	}
}

func (g *Game) CreateRoom(
	leader *gamesvc.Player,
	req *gamesvc.CreateRoomRequest,
) (roomID string) {
	roomID = uuidgen.NewString()
	r := room.NewRoom(g.log, roomID, leader.Id, req)

	sm := statemachine.Lobby{}
	go func() {
		for {
			tuple := <-r.AggregationChan()

			r.Do(func(r *room.Room) {
				err := sm.HandleMessage(tuple.A, tuple.B, r)
				if err != nil {
					_ = tuple.B.SendError(err.Error())
				}
			})
		}
	}()
	go r.Start()

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
		r.Do(func(r *room.Room) {
			roomsProto = append(roomsProto, r.GetProto())
		})
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
	if ok {
		return ErrRoomNotFound
	}

	player := player.New(g.log, socket, playerProto)

	err := runFn1(r, func(r *room.Room) error {
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

	go func() {
		for {
			select {
			case <-r.Done():
				player.Cancel()
				return

			case msg, ok := <-player.Chan():
				if !ok {
					return
				}

				select {
				case r.AggregationChan() <- tuple.NewT2(msg, player):
					continue
				case <-r.Done():
					return
				}
			}
		}
	}()

	err = player.Start()
	if err != nil {
		g.log.
			Err(err).
			Stringer("status_code", status.Code(err)).
			Interface("player", player).
			Msg("tried to send message and something went wrong")

		r.Do(func(r *room.Room) {
			ok := r.RemovePlayer(player.ID)
			if !ok {
				return
			}

			r.AnnounceChange()
		})
		return fmt.Errorf("player loop: %w", err)
	}

	return nil
}

func runFn1[R1 any](r *room.Room, fn func(r *room.Room) R1) R1 {
	var r1 R1
	wait := make(chan struct{})
	r.Do(func(r *room.Room) {
		defer close(wait)

		r1 = fn(r)
	})
	<-wait

	return r1
}
