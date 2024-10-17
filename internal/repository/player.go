package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

var ErrPlayerNotFound = errors.New("player not found")

type PlayerRepository interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

type dbPlayer struct {
	client *redis.Client
}

func NewPlayerRepository(client *redis.Client) PlayerRepository {
	return &dbPlayer{
		client: client,
	}
}

func (that *dbPlayer) CreateOrUpdate(ctx context.Context, player *entity.Player) error {
	playerJSON, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("failed to marshal player: %w", err)
	}

	playerKey := "player:" + player.ID
	err = that.client.Set(ctx, playerKey, playerJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set player: %w", err)
	}

	return nil
}

func (that *dbPlayer) GetByID(ctx context.Context, id string) (*entity.Player, error) {
	playerKey := "player:" + id

	response, err := that.client.Get(ctx, playerKey).Result()

	if errors.Is(err, redis.Nil) {
		return &entity.Player{}, ErrPlayerNotFound
	}

	if err != nil {
		return &entity.Player{}, fmt.Errorf("failed to get player by ID: %w", err)
	}

	var existingPlayer entity.Player
	if err = json.Unmarshal([]byte(response), &existingPlayer); err != nil {
		return &entity.Player{}, fmt.Errorf("failed to unmarshal player: %w", err)
	}

	return &existingPlayer, nil
}
