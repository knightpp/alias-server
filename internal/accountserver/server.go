package loginservice

import (
	"context"

	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	loginsvc "github.com/knightpp/alias-proto/go/login_service"
	"github.com/knightpp/alias-server/internal/gravatar"
	"github.com/knightpp/alias-server/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ loginsvc.LoginServiceServer = (*LoginService)(nil)

type LoginService struct {
	loginsvc.UnimplementedLoginServiceServer

	db storage.Player
}

func New(db storage.Player) *LoginService {
	return &LoginService{
		db: db,
	}
}

func (l *LoginService) LoginGuest(ctx context.Context, req *loginsvc.LoginGuestRequest) (*loginsvc.LoginGuestResponse, error) {
	id := uuid.NewString()
	// TODO: save to DB
	auth := uuid.NewString()
	err := l.db.SetPlayer(ctx, auth, &gamesvc.Player{
		Id:          id,
		Name:        req.Name,
		GravatarUrl: gravatar.GetUrlOrDefault(req.Email),
	})
	if err != nil {
		return nil, err
	}

	return &loginsvc.LoginGuestResponse{
		Account: &loginsvc.Account{
			Id:        id,
			AuthToken: auth,
			Name:      req.Name,
			Email:     req.Email,
		},
	}, nil
}

func (l *LoginService) VerifyToken(
	ctx context.Context,
	req *loginsvc.VerifyTokenRequest,
) (*loginsvc.VerifyTokenResponse, error) {
	p, _ := l.db.GetPlayer(ctx, req.Token)
	if p == nil {
		return nil, status.Error(codes.NotFound, "token not found")
	}

	return &loginsvc.VerifyTokenResponse{}, nil
}
