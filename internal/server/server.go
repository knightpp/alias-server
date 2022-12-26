package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/metadata"
)

var _ gamesvc.GameServiceServer = (*GameService)(nil)

type GameService struct {
	gamesvc.UnimplementedGameServiceServer

	log zerolog.Logger
	db  storage.PlayerDB

	roomsMu sync.Mutex
	rooms   map[string]*game.Room
}

func New(log zerolog.Logger, db storage.PlayerDB) *GameService {
	return &GameService{
		rooms: make(map[string]*game.Room),
		log:   log,
		db:    db,
	}
}

func (gs *GameService) ListRooms(_ context.Context, _ *gamesvc.ListRoomsRequest) (*gamesvc.ListRoomsResponse, error) {
	gs.roomsMu.Lock()
	defer gs.roomsMu.Unlock()

	roomsProto := make([]*gamesvc.Room, 0, len(gs.rooms))
	for _, room := range gs.rooms {
		roomsProto = append(roomsProto, room.GetProto())
	}

	return &gamesvc.ListRoomsResponse{
		Rooms: roomsProto,
	}, nil
}

func (gs *GameService) CreateRoom(ctx context.Context, req *gamesvc.CreateRoomRequest) (*gamesvc.CreateRoomResponse, error) {
	gs.roomsMu.Lock()
	defer gs.roomsMu.Unlock()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata in request")
	}

	tokenMd := md.Get("token")
	if len(tokenMd) != 1 {
		return nil, fmt.Errorf("unexpected metadata value: %#v", tokenMd)
	}

	token := tokenMd[0]
	// TODO: find player's id by token
	_ = token

	id := uuid.NewString()
	room := game.NewRoom(id, "", req)
	gs.rooms[id] = room

	return &gamesvc.CreateRoomResponse{
		Id: id,
	}, nil
}

func (gs *GameService) Join(stream gamesvc.GameService_JoinServer) error {
	ctx := stream.Context()
	roomIDMD := metadata.ValueFromIncomingContext(ctx, "room-id")
	if len(roomIDMD) != 1 {
		return fmt.Errorf("unexpected metadata room-id value: %#v", roomIDMD)
	}

	roomID := roomIDMD[0]
	gs.roomsMu.Lock()
	defer gs.roomsMu.Unlock()

	room, ok := gs.rooms[roomID]
	if !ok {
		return fmt.Errorf("could not find room with id=%q", roomID)
	}

	wg := room.AddPlayer(stream, nil)
	wg.Wait()

	return nil
}
