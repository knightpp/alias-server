package server

import (
	"context"
	"errors"
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-proto/go/mdkey"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ gamesvc.GameServiceServer = (*GameService)(nil)

type GameService struct {
	gamesvc.UnimplementedGameServiceServer

	log zerolog.Logger
	db  storage.Player

	game *game.Game
}

func New(log zerolog.Logger, db storage.Player) *GameService {
	return &GameService{
		game: game.New(log),
		log:  log,
		db:   db,
	}
}

func (gs *GameService) ListRooms(_ context.Context, _ *gamesvc.ListRoomsRequest) (*gamesvc.ListRoomsResponse, error) {
	return &gamesvc.ListRoomsResponse{
		Rooms: gs.game.ListRooms(),
	}, nil
}

func (gs *GameService) CreateRoom(ctx context.Context, req *gamesvc.CreateRoomRequest) (*gamesvc.CreateRoomResponse, error) {
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

	id := gs.game.CreateRoom(player, req)

	return &gamesvc.CreateRoomResponse{
		Id: id,
	}, nil
}

func (gs *GameService) Join(stream gamesvc.GameService_JoinServer) error {
	ctx := stream.Context()

	md, _ := metadata.FromIncomingContext(ctx)

	roomID, err := singleFieldMD(mdkey.RoomID, md)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "get room id from md: %s", err)
	}

	authToken, err := singleFieldMD(mdkey.Auth, md)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "get auth token from md: %s", err)
	}

	player, err := gs.db.GetPlayer(ctx, authToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "get player: %w", err)
	}

	return gs.game.StartPlayerInRoom(roomID, player, stream)
}

func singleFieldMD(field string, md metadata.MD) (string, error) {
	values := md.Get(field)
	if len(values) != 1 {
		return "", fmt.Errorf("metadata entry should be of length 1, but was %d", len(values))
	}

	return values[0], nil
}
