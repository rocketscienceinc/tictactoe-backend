package usecase

import (
	"context"
	"fmt"
	"github.com/rocketscienceinc/tittactoe-backend/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/utils"
)

func (that *uGame) GetOrCreatePlayer(ctx context.Context, id string) (*entity.Player, error) {
	if id == "" {
		player, err := that.createPlayer(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create new player %w", err)
		}

		return player, nil
	}

	player, err := that.playerService.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get player by id %w", err)
	}

	return player, nil
}

func (that *uGame) createPlayer(ctx context.Context) (*entity.Player, error) {
	playerID := utils.GenerateNewSessionID()

	player := &entity.Player{
		ID: playerID,
	}

	if err := that.playerService.CreateOrUpdate(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	return player, nil
}

func (that *uGame) getPlayerByID(ctx context.Context, id string) (*entity.Player, error) {
	player, err := that.playerService.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

func (that *uGame) updatePlayer(ctx context.Context, player *entity.Player) error {
	if err := that.playerService.CreateOrUpdate(ctx, player); err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	return nil
}
