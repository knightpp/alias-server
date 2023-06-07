package accountserver

import (
	"context"

	"github.com/google/uuid"
	accountsvc "github.com/knightpp/alias-proto/go/account/service/v1"
	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	"github.com/knightpp/alias-server/internal/gravatar"
	"github.com/knightpp/alias-server/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ accountsvc.AccountServiceServer = (*AccountService)(nil)

type AccountService struct {
	accountsvc.UnimplementedAccountServiceServer

	db storage.Player
}

func New(db storage.Player) *AccountService {
	return &AccountService{
		db: db,
	}
}

func (svc *AccountService) RegisterGuest(ctx context.Context, req *accountsvc.RegisterGuestRequest) (*accountsvc.RegisterGuestResponse, error) {
	id := uuid.NewString()
	// TODO: save to DB
	auth := uuid.NewString()
	err := svc.db.SetPlayer(ctx, auth, &gamesvc.Player{
		Id:          id,
		Name:        req.Name,
		GravatarUrl: gravatar.GetUrlOrDefault(req.Email),
	})
	if err != nil {
		return nil, err
	}

	return &accountsvc.RegisterGuestResponse{
		Account: &accountsvc.Account{
			Id:        id,
			AuthToken: auth,
			Name:      req.Name,
			Email:     req.Email,
		},
	}, nil
}

func (svc *AccountService) UpdateAccount(context.Context, *accountsvc.UpdateAccountRequest) (*accountsvc.UpdateAccountResponse, error) {
	return nil, status.Error(codes.Internal, "NOT IMPLEMENTED")
}

func (svc *AccountService) VerifyToken(
	ctx context.Context,
	req *accountsvc.VerifyTokenRequest,
) (*accountsvc.VerifyTokenResponse, error) {
	p, _ := svc.db.GetPlayer(ctx, req.Token)
	if p == nil {
		return nil, status.Error(codes.NotFound, "token not found")
	}

	return &accountsvc.VerifyTokenResponse{}, nil
}
