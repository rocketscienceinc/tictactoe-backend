package service

import (
	"context"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

type GameService interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error
	GetByID(ctx context.Context, id string) (*entity.Game, error)
	DeleteByID(ctx context.Context, id string) error
}

type gameService struct {
	gameRepo gameRepo
}

type gameRepo interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error
	GetByID(ctx context.Context, id string) (*entity.Game, error)
	DeleteByID(ctx context.Context, id string) error
}

func NewGameService(gameRepo gameRepo) GameService {
	return &gameService{
		gameRepo: gameRepo,
	}
}

func (that *gameService) CreateOrUpdate(ctx context.Context, game *entity.Game) error {
	if err := that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return fmt.Errorf("create game %w", err)
	}

	return nil
}

func (that *gameService) GetByID(ctx context.Context, id string) (*entity.Game, error) {
	existingGame, err := that.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get game %w", err)
	}

	return existingGame, nil
}

func (that *gameService) DeleteByID(ctx context.Context, id string) error {
	err := that.gameRepo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete game %w", err)
	}

	return nil
}
