package usecase

import (
	"context"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

type GameUseCase interface {
	GetOrCreatePlayer(ctx context.Context, playerID string) (*entity.Player, error)

	GetOrCreateGame(ctx context.Context, playerID string) (*entity.Game, error)
	JoinGame(ctx context.Context, gameID, playerID string) (*entity.Game, error)

	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type playerService interface {
	CreatePlayer(ctx context.Context) (*entity.Player, error)
	GetPlayerByID(ctx context.Context, id string) (*entity.Player, error)
	UpdatePlayer(ctx context.Context, player *entity.Player) error
}

type gameService interface {
	CreateGame(ctx context.Context, player *entity.Player) (*entity.Game, *entity.Player, error)
	GetGameByID(ctx context.Context, id string) (*entity.Game, error)
	UpdateGame(ctx context.Context, game *entity.Game) error
	DeleteGame(ctx context.Context, gameID string) error
}

type gamePlayService interface {
	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
	ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	GetGameState(ctx context.Context, player *entity.Player) (*entity.Game, error)
	CleanupGame(ctx context.Context, game *entity.Game)
}

type gameUseCase struct {
	playerService   playerService
	gameService     gameService
	gamePlayService gamePlayService
}

func NewGameUseCase(playerService playerService, gameService gameService, gamePlayService gamePlayService) GameUseCase {
	return &gameUseCase{
		playerService:   playerService,
		gameService:     gameService,
		gamePlayService: gamePlayService,
	}
}

func (that *gameUseCase) GetOrCreatePlayer(ctx context.Context, playerID string) (*entity.Player, error) {
	if playerID == "" {
		player, err := that.playerService.CreatePlayer(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not create player: %w", err)
		}

		return player, nil
	}

	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player by id: %w", err)
	}

	return player, nil
}

func (that *gameUseCase) GetOrCreateGame(ctx context.Context, playerID string) (*entity.Game, error) {
	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	game, err := that.gamePlayService.GetGameState(ctx, player)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) JoinGame(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
	game, err := that.gamePlayService.ConnectToGame(ctx, gameID, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to game: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	game, err := that.gamePlayService.MakeTurn(ctx, playerID, cell)
	if err != nil {
		return nil, fmt.Errorf("failed to make turn: %w", err)
	}

	if game.IsFinished() {
		that.gamePlayService.CleanupGame(ctx, game)

		return game, apperror.ErrGameFinished
	}

	return game, nil
}
