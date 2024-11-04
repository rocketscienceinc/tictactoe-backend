package service

import (
	"context"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/pkg"
)

type GameService interface {
	CreateGame(ctx context.Context, player *entity.Player, gameType string) (*entity.Game, *entity.Player, error)
	UpdateGame(ctx context.Context, game *entity.Game) error
	DeleteGame(ctx context.Context, gameID string) error

	GetGameByID(ctx context.Context, id string) (*entity.Game, error)
	GetPublicGameByID(ctx context.Context) (*entity.Game, error)
}

type gameRepo interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error

	GetByID(ctx context.Context, id string) (*entity.Game, error)
	GetWaitingPublicGame(ctx context.Context) (*entity.Game, error)

	DeleteByID(ctx context.Context, id string) error
}

type gameService struct {
	gameRepo gameRepo
}

func NewGameService(gameRepo gameRepo) GameService {
	return &gameService{
		gameRepo: gameRepo,
	}
}

func (that *gameService) CreateGame(ctx context.Context, player *entity.Player, gameType string) (*entity.Game, *entity.Player, error) {
	gameID, err := pkg.GenerateGameID()
	if err != nil {
		return nil, nil, fmt.Errorf("error generating game ID: %w", err)
	}

	game := entity.NewGame(gameID, gameType)

	player.GameID = gameID
	player.Mark = entity.PlayerX

	game.Players = []*entity.Player{player}
	if err := that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return nil, nil, fmt.Errorf("failed to create game from storage: %w", err)
	}
	return game, player, nil
}

func (that *gameService) GetGameByID(ctx context.Context, id string) (*entity.Game, error) {
	game, err := that.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve game from storage: %w", err)
	}
	return game, nil
}

func (that *gameService) GetPublicGameByID(ctx context.Context) (*entity.Game, error) {
	game, err := that.gameRepo.GetWaitingPublicGame(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active public game from storage: %w", err)
	}
	return game, nil
}

func (that *gameService) UpdateGame(ctx context.Context, game *entity.Game) error {
	if err := that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}
	return nil
}

func (that *gameService) DeleteGame(ctx context.Context, gameID string) error {
	if err := that.gameRepo.DeleteByID(ctx, gameID); err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}
	return nil
}
