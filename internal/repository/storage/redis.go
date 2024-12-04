package storage

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	Connection *redis.Client
}

func NewRedisStorage(ctx context.Context, addr string) (*RedisStorage, error) {
	conn := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := conn.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{Connection: conn}, nil
}
