package server

import (
	"context"
	"errors"
	"fmt"

	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ gamesvc.GameServiceServer = (*GameService)(nil)

var ErrUnauthenticated = status.Error(codes.Unauthenticated, "Unauthenticated")

type AuthKey struct{}

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

func (svc *GameService) ListRooms(ctx context.Context, req *gamesvc.ListRoomsRequest) (*gamesvc.ListRoomsResponse, error) {
	return &gamesvc.ListRoomsResponse{
		Rooms: svc.game.ListRooms(),
	}, nil
}

func (svc *GameService) CreateRoom(ctx context.Context, req *gamesvc.CreateRoomRequest) (*gamesvc.CreateRoomResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata in request")
	}

	tokenMd := md.Get("token")
	if len(tokenMd) != 1 {
		return nil, fmt.Errorf("unexpected metadata value: %#v", tokenMd)
	}

	token := tokenMd[0]

	player, err := svc.db.GetPlayer(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get player: %w", err)
	}

	id := svc.game.CreateRoom(player, req)

	return &gamesvc.CreateRoomResponse{
		Id: id,
	}, nil
}

// func (svc *GameService) UpdateRoom(ctx context.Context, req *gamesvc.UpdateRoomRequest) (*gamesvc.UpdateRoomResponse, error)

func (svc *GameService) JoinRoom(req *gamesvc.JoinRoomRequest, srv gamesvc.GameService_JoinRoomServer) error {
	player, err := getPlayer(srv.Context())
	if err != nil {
		return err
	}

	return svc.game.JoinRoom(player, req, srv)
}

func (svc *GameService) TransferLeadership(ctx context.Context, req *gamesvc.TransferLeadershipRequest) (*gamesvc.TransferLeadershipResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.game.TransferLeadership(player, req)
}

func (svc *GameService) CreateTeam(ctx context.Context, req *gamesvc.CreateTeamRequest) (*gamesvc.CreateTeamResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.game.CreateTeam(player, req)
}

func (svc *GameService) UpdateTeam(ctx context.Context, req *gamesvc.UpdateTeamRequest) (*gamesvc.UpdateTeamResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	_ = player
	panic("TODO")
	// return svc.game.UpdateTeam(player, req)
}
func (svc *GameService) JoinTeam(ctx context.Context, req *gamesvc.JoinTeamRequest) (*gamesvc.JoinTeamResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.game.JoinTeam(player, req)
}
func (svc *GameService) StartGame(ctx context.Context, req *gamesvc.StartGameRequest) (*gamesvc.StartGameResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.game.StartGame(player, req)
}
func (svc *GameService) StopGame(ctx context.Context, req *gamesvc.StopGameRequest) (*gamesvc.StopGameResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.game.StopGame(player, req)
}
func (svc *GameService) StartTurn(ctx context.Context, req *gamesvc.StartTurnRequest) (*gamesvc.StartTurnResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.game.StartTurn(player, req)
}
func (svc *GameService) StopTurn(ctx context.Context, req *gamesvc.StopTurnRequest) (*gamesvc.StopTurnResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.game.StopTurn(player, req)
}
func (svc *GameService) Score(ctx context.Context, req *gamesvc.ScoreRequest) (*gamesvc.ScoreResponse, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	_ = player
	panic("TODO")
	// return svc.game.Score(player, req)
}

// func (gs *GameService) Join(stream gamesvc.GameService_JoinServer) error {
// 	ctx := stream.Context()

// 	md, _ := metadata.FromIncomingContext(ctx)

// 	roomID, err := singleFieldMD(mdkey.RoomID, md)
// 	if err != nil {
// 		return status.Errorf(codes.InvalidArgument, "get room id from md: %s", err)
// 	}

// 	authToken, err := singleFieldMD(mdkey.Auth, md)
// 	if err != nil {
// 		return status.Errorf(codes.Unauthenticated, "get auth token from md: %s", err)
// 	}

// 	player, err := gs.db.GetPlayer(ctx, authToken)
// 	if err != nil {
// 		return status.Errorf(codes.Unauthenticated, "get player: %s", err)
// 	}

// 	return gs.game.StartPlayerInRoom(roomID, player, stream)
// }

func singleFieldMD(field string, md metadata.MD) (string, error) {
	values := md.Get(field)
	if len(values) != 1 {
		return "", fmt.Errorf("metadata entry should be of length 1, but was %d", len(values))
	}

	return values[0], nil
}

func getPlayer(ctx context.Context) (*gamesvc.Player, error) {
	player, ok := ctx.Value(AuthKey{}).(*gamesvc.Player)
	if !ok {
		return nil, ErrUnauthenticated
	}

	return player, nil
}
