package service

import (
	"context"
	"fmt"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tictactoe-backend/internal/pkg"
)

type PlayerService interface {
	CreatePlayer(ctx context.Context) (*entity.Player, error)
	GetPlayerByID(ctx context.Context, id string) (*entity.Player, error)
	UpdatePlayer(ctx context.Context, player *entity.Player) error
}

type playerRepo interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

type playerService struct {
	playerRepo playerRepo
}

func NewPlayerService(playerRepo playerRepo) PlayerService {
	return &playerService{
		playerRepo: playerRepo,
	}
}

func (that *playerService) CreatePlayer(ctx context.Context) (*entity.Player, error) {
	playerID := pkg.GenerateNewSessionID()
	player := &entity.Player{ID: playerID}

	if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to create player from storage: %w", err)
	}
	return player, nil
}

func (that *playerService) GetPlayerByID(ctx context.Context, id string) (*entity.Player, error) {
	player, err := that.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}
	return player, nil
}

func (that *playerService) UpdatePlayer(ctx context.Context, player *entity.Player) error {
	if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}
	return nil
}
