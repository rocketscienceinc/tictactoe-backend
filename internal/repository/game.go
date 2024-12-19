package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

var ErrGameNotFound = errors.New("game not found")

type GameRepository interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error

	GetByID(ctx context.Context, id string) (*entity.Game, error)
	GetOpenPublicGame(ctx context.Context) (*entity.Game, error)

	DeleteByID(ctx context.Context, id string) error
}

type gameRepository struct {
	logger *slog.Logger

	client *redis.Client
}

func NewGameRepository(logger *slog.Logger, client *redis.Client) GameRepository {
	return &gameRepository{
		logger: logger,
		client: client,
	}
}

// CreateOrUpdate - creates or updates a game object.
// Note:
// If the game is public, it adds it to the setList of public games.
// This solution is used to be able to retrieve all public games that can be connected.
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

	if game.IsPublic() {
		if err = that.client.SAdd(ctx, entity.PublicType, game.ID).Err(); err != nil {
			return fmt.Errorf("failed to add game to public set: %w", err)
		}
	}

	return nil
}

func (that *gameRepository) GetByID(ctx context.Context, id string) (*entity.Game, error) {
	gameKey := "game:" + id

	response, err := that.client.Get(ctx, gameKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrGameNotFound
		}

		return nil, fmt.Errorf("failed to get game by ID: %w", err)
	}

	var game entity.Game
	if err = json.Unmarshal([]byte(response), &game); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game: %w", err)
	}

	return &game, nil
}

func (that *gameRepository) GetOpenPublicGame(ctx context.Context) (*entity.Game, error) {
	log := that.logger.With("method", "GetLastActivePublicGame")

	gameIDs, err := that.client.SMembers(ctx, entity.PublicType).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get game IDs from set %s: %w", entity.PublicType, err)
	}

	publicGames := make([]*entity.Game, 0, len(gameIDs))
	for _, id := range gameIDs {
		game, err := that.GetByID(ctx, id)
		if err != nil {
			continue
		}

		if !game.IsWaiting() {
			continue
		}

		publicGames = append(publicGames, game)
	}

	if len(publicGames) == 0 {
		return nil, apperror.ErrNoActiveGames
	}

	log.Info("found active games", "count", len(publicGames))

	return publicGames[len(publicGames)-1], nil
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
