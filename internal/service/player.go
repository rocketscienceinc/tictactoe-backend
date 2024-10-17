package service

import (
	"context"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

type PlayerService interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

type playerService struct {
	playerRepo playerRepo
}

type playerRepo interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

func NewPlayerService(playerRepo playerRepo) PlayerService {
	return &playerService{
		playerRepo: playerRepo,
	}
}

func (that *playerService) CreateOrUpdate(ctx context.Context, player *entity.Player) error {
	if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return fmt.Errorf("create player %w", err)
	}

	return nil
}

func (that *playerService) GetByID(ctx context.Context, id string) (*entity.Player, error) {
	existingPlayer, err := that.playerRepo.GetByID(ctx, id)
	if err != nil {
		return &entity.Player{}, fmt.Errorf("get player by id %w", err)
	}

	return existingPlayer, nil
}
