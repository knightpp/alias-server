package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/storage"
	"google.golang.org/protobuf/proto"
)

var _ storage.PlayerDB = (*redisImpl)(nil)

type redisImpl struct {
	db *redis.Client
}

func New(addr string) storage.PlayerDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &redisImpl{
		db: rdb,
	}
}

func NewFromURL(url string) (storage.PlayerDB, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	rdb := redis.NewClient(opts)
	return &redisImpl{db: rdb}, nil
}

func (r *redisImpl) SetPlayer(ctx context.Context, p *gamesvc.Player) error {
	playerBytes, err := proto.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal player as protobuf: %w", err)
	}

	cmd := r.db.Set(ctx, p.Id, playerBytes, (24*time.Hour)*30)
	return cmd.Err()
}

func (r *redisImpl) GetPlayer(ctx context.Context, playerID string) (*gamesvc.Player, error) {
	if playerID == "" {
		return nil, errors.New("error: player id is empty")
	}

	cmd := r.db.Get(ctx, playerID)

	playerBytes, err := cmd.Bytes()
	if err != nil {
		return nil, fmt.Errorf("get redis bytes: %w", err)
	}

	var playerPb gamesvc.Player

	err = proto.Unmarshal(playerBytes, &playerPb)
	if err != nil {
		return nil, fmt.Errorf("unmarshal proto: %w", err)
	}

	return &playerPb, nil
}
