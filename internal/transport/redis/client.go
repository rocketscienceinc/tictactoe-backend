package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rocketscienceinc/tittactoe-backend/internal/game"
)

var ErrNoActiveGame = errors.New("no active game found")

type Client struct {
	client *redis.Client
}

func New(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Client{rdb}
}

// GetOrCreatePlayer - checks if the player exists, if not, creates a new one.
func (that *Client) GetOrCreatePlayer(ctx context.Context, sessionID string) (*game.Player, error) {
	val, err := that.client.Get(ctx, sessionID).Result()
	if errors.Is(err, redis.Nil) {
		player := &game.Player{ID: sessionID}
		playerJSON, err := json.Marshal(player)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal player: %w", err)
		}

		playerKey := "player:" + sessionID
		err = that.client.Set(ctx, playerKey, playerJSON, 0).Err()
		if err != nil {
			return nil, fmt.Errorf("failed to create player in Redis: %w", err)
		}
		return &game.Player{ID: sessionID}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get player in Redis: %w", err)
	}

	var player game.Player
	if err := json.Unmarshal([]byte(val), &player); err != nil {
		return nil, fmt.Errorf("failed to unmarshal player from Redis: %w", err)
	}

	return &game.Player{ID: val}, nil
}

// GetActiveGame - checks if the player has an active game.
func (that *Client) GetActiveGame(ctx context.Context, playerID string) (*game.Game, error) {
	gameKey := "game:" + playerID

	val, err := that.client.Get(ctx, gameKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNoActiveGame
	} else if err != nil {
		return nil, fmt.Errorf("failed to get active game: %w", err)
	}

	var gameData game.Game
	if err := json.Unmarshal([]byte(val), &gameData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game data: %w", err)
	}

	return &gameData, nil
}

// SaveGame - saves a new game in Redis.
func (that *Client) SaveGame(ctx context.Context, playerID string, game *game.Game) error {
	gameKey := "game:" + playerID
	gameJSON, err := json.Marshal(game)
	if err != nil {
		return fmt.Errorf("failed to marshal game data: %w", err)
	}

	err = that.client.Set(ctx, gameKey, gameJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save game in Redis: %w", err)
	}

	return nil
}
