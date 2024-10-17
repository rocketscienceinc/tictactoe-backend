package storage

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, addr string) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := db.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return db, nil
}
