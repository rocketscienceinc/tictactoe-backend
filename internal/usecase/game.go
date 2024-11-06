package usecase

import (
	"context"
	"fmt"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

type GameUseCase interface {
	GetOrCreatePlayer(ctx context.Context, playerID string) (*entity.Player, error)

	GetOrCreateGame(ctx context.Context, playerID, gameType string) (*entity.Game, error)
	GetGameByPlayerID(ctx context.Context, playerID string) (*entity.Game, error)
	JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	JoinWaitingPublicGame(ctx context.Context, playerID string) (*entity.Game, error)

	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type playerService interface {
	CreatePlayer(ctx context.Context) (*entity.Player, error)
	GetPlayerByID(ctx context.Context, id string) (*entity.Player, error)
	UpdatePlayer(ctx context.Context, player *entity.Player) error
}

type gameService interface {
	GetGameByID(ctx context.Context, id string) (*entity.Game, error)
	UpdateGame(ctx context.Context, game *entity.Game) error
	DeleteGame(ctx context.Context, gameID string) error
}

type gamePlayService interface {
	JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	JoinWaitingPublicGame(ctx context.Context, playerID string) (*entity.Game, error)

	GetOrCreateGame(ctx context.Context, player *entity.Player, gameType string) (*entity.Game, error)
	CleanupGame(ctx context.Context, game *entity.Game) // ToDo: Need refactor

	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
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

func (that *gameUseCase) GetOrCreateGame(ctx context.Context, playerID, gameType string) (*entity.Game, error) {
	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	game, err := that.gamePlayService.GetOrCreateGame(ctx, player, gameType)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) GetGameByPlayerID(ctx context.Context, playerID string) (*entity.Game, error) {
	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	game, err := that.gameService.GetGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
	game, err := that.gamePlayService.JoinGameByID(ctx, gameID, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to join game: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) JoinWaitingPublicGame(ctx context.Context, playerID string) (*entity.Game, error) {
	game, err := that.gamePlayService.JoinWaitingPublicGame(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to join waiting public game: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	game, err := that.gamePlayService.MakeTurn(ctx, playerID, cell)
	if err != nil {
		return game, fmt.Errorf("failed to make turn: %w", err)
	}

	if game.IsFinished() {
		that.gamePlayService.CleanupGame(ctx, game)

		return game, apperror.ErrGameFinished
	}

	return game, nil
}
