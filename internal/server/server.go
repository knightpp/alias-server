package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-proto/go/mdkey"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/metadata"
)

var _ gamesvc.GameServiceServer = (*GameService)(nil)

type GameService struct {
	gamesvc.UnimplementedGameServiceServer

	log     zerolog.Logger
	uuidGen uuidgen.Generator
	db      storage.Player

	roomsMu sync.Mutex
	rooms   map[string]*game.Room
}

func New(log zerolog.Logger, db storage.Player, gen uuidgen.Generator) *GameService {
	return &GameService{
		rooms:   make(map[string]*game.Room),
		log:     log,
		db:      db,
		uuidGen: gen,
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

	player, err := gs.db.GetPlayer(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get player: %w", err)
	}

	id := gs.uuidGen.NewString()
	room := game.NewRoom(gs.log, id, player.Id, req, gs.uuidGen)
	go room.Start()
	gs.rooms[id] = room

	return &gamesvc.CreateRoomResponse{
		Id: id,
	}, nil
}

func (gs *GameService) Join(stream gamesvc.GameService_JoinServer) error {
	ctx := stream.Context()

	md, _ := metadata.FromIncomingContext(ctx)

	roomID, err := singleFieldMD(mdkey.RoomID, md)
	if err != nil {
		return fmt.Errorf("get room id from md: %w", err)
	}

	authToken, err := singleFieldMD(mdkey.Auth, md)
	if err != nil {
		return fmt.Errorf("get auth token from md: %w", err)
	}

	gs.roomsMu.Lock()

	room, ok := gs.rooms[roomID]
	if !ok {
		gs.roomsMu.Unlock()
		return fmt.Errorf("could not find room with id=%q", roomID)
	}

	gs.roomsMu.Unlock()

	player, err := gs.db.GetPlayer(ctx, authToken)
	if err != nil {
		return fmt.Errorf("get player: %w", err)
	}

	if room.HasPlayer(player.Id) {
		return fmt.Errorf("player %q already in the room", player.Id)
	}

	return room.AddAndStartPlayer(stream, player)
}

func singleFieldMD(field string, md metadata.MD) (string, error) {
	values := md.Get(field)
	if len(values) != 1 {
		return "", fmt.Errorf("metadata entry should be of length 1, but was %d", len(values))
	}

	return values[0], nil
}
