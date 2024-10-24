package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

var ErrGameNotFound = errors.New("game not found")

type GameRepository interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error
	GetByID(ctx context.Context, id string) (*entity.Game, error)
	DeleteByID(ctx context.Context, id string) error
}

type gameRepository struct {
	client *redis.Client
}

func NewGameRepository(client *redis.Client) GameRepository {
	return &gameRepository{
		client: client,
	}
}

func (that *gameRepository) CreateOrUpdate(ctx context.Context, game *entity.Game) error {
	gameJSON, err := json.Marshal(game)
	if err != nil {
		return fmt.Errorf("could not marshal game: %w", err)
	}

	gameKey := "game:" + game.ID
	err = that.client.Set(ctx, gameKey, gameJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set game: %w", err)
	}

	return nil
}

func (that *gameRepository) GetByID(ctx context.Context, id string) (*entity.Game, error) {
	gameKey := "game:" + id

	response, err := that.client.Get(ctx, gameKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &entity.Game{}, ErrGameNotFound
		}

		return &entity.Game{}, fmt.Errorf("failed to get game by ID: %w", err)
	}

	var existingGame entity.Game
	if err = json.Unmarshal([]byte(response), &existingGame); err != nil {
		return &entity.Game{}, fmt.Errorf("failed to unmarshal game: %w", err)
	}

	return &existingGame, nil
}

func (that *gameRepository) DeleteByID(ctx context.Context, id string) error {
	gameKey := "game:" + id

	result, err := that.client.Del(ctx, gameKey).Result()
	if err != nil {
		return fmt.Errorf("failed to delete game by ID: %w", err)
	}

	if result == 0 {
		return ErrGameNotFound
	}

	return nil
}
