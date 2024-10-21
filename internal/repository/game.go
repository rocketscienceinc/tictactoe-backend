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

type dbGame struct {
	client *redis.Client
}

func NewGameRepository(client *redis.Client) GameRepository {
	return &dbGame{
		client: client,
	}
}

func (that *dbGame) CreateOrUpdate(ctx context.Context, game *entity.Game) error {
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

func (that *dbGame) GetByID(ctx context.Context, id string) (*entity.Game, error) {
	gameKey := "game:" + id

	response, err := that.client.Get(ctx, gameKey).Result()

	if errors.Is(err, redis.Nil) {
		return &entity.Game{}, ErrGameNotFound
	}

	if err != nil {
		return &entity.Game{}, fmt.Errorf("%w by id", err)
	}

	var existingGame entity.Game
	if err = json.Unmarshal([]byte(response), &existingGame); err != nil {
		return &entity.Game{}, fmt.Errorf("failed to unmarshal player: %w", err)
	}

	return &existingGame, nil
}

func (that *dbGame) DeleteByID(ctx context.Context, id string) error {
	gameKey := "game:" + id

	err := that.client.Del(ctx, gameKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete player by ID: %w", err)
	}

	return nil
}
