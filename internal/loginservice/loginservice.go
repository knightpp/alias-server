package loginservice

import (
	"context"

	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	loginsvc "github.com/knightpp/alias-proto/go/login_service"
	"github.com/knightpp/alias-server/internal/gravatar"
	"github.com/knightpp/alias-server/internal/storage"
)

var _ loginsvc.LoginServiceServer = (*LoginService)(nil)

type LoginService struct {
	loginsvc.UnimplementedLoginServiceServer

	db storage.PlayerDB
}

func New(db storage.PlayerDB) *LoginService {
	return &LoginService{
		db: db,
	}
}

func (l *LoginService) LoginGuest(ctx context.Context, req *loginsvc.LoginGuestRequest) (*loginsvc.LoginGuestResponse, error) {
	id := uuid.NewString()
	// TODO: save to DB
	auth := uuid.NewString()
	err := l.db.SetPlayer(ctx, &gamesvc.Player{
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
