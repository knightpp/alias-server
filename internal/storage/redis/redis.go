package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

var _ storage.Player = (*redisImpl)(nil)

type redisImpl struct {
	db *redis.Client
}

func New(addr string) storage.Player {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &redisImpl{
		db: rdb,
	}
}

func NewFromURL(url string) (storage.Player, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	rdb := redis.NewClient(opts)
	return &redisImpl{db: rdb}, nil
}

func (r *redisImpl) SetPlayer(ctx context.Context, token string, p *gamesvc.Player) error {
	playerBytes, err := proto.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal player as protobuf: %w", err)
	}

	cmd := r.db.Set(ctx, token, playerBytes, (24*time.Hour)*30)
	return cmd.Err()
}

func (r *redisImpl) GetPlayer(ctx context.Context, token string) (*gamesvc.Player, error) {
	if token == "" {
		return nil, errors.New("error: player id is empty")
	}

	cmd := r.db.Get(ctx, token)

	playerBytes, err := cmd.Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, storage.ErrNotFound
		default:
			return nil, fmt.Errorf("get redis bytes: %w", err)
		}
	}

	playerPb := &gamesvc.Player{}
	err = proto.Unmarshal(playerBytes, playerPb)
	if err != nil {
		return nil, fmt.Errorf("unmarshal proto: %w", err)
	}

	return playerPb, nil
}
