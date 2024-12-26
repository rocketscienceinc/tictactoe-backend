package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

type GameUseCase interface {
	GetOrCreatePlayer(ctx context.Context, playerID string) (*entity.Player, error)

	GetOrCreateGame(ctx context.Context, playerID, gameType string) (*entity.Game, error)
	GetGameByPlayerID(ctx context.Context, playerID string) (*entity.Game, error)
	CreateOrJoinToPublicGame(ctx context.Context, playerID, gameType string) (*entity.Game, error)
	JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	EndGame(ctx context.Context, game *entity.Game) error

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

	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type gameUseCase struct {
	logger          *slog.Logger
	playerService   playerService
	gameService     gameService
	gamePlayService gamePlayService
}

func NewGameUseCase(logger *slog.Logger, playerService playerService, gameService gameService, gamePlayService gamePlayService) GameUseCase {
	return &gameUseCase{
		logger:          logger.With("module", "usecase/game"),
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

func (that *gameUseCase) CreateOrJoinToPublicGame(ctx context.Context, playerID, gameType string) (*entity.Game, error) {
	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player by id: %w", err)
	}

	game, err := that.gamePlayService.JoinWaitingPublicGame(ctx, player.ID)
	if err != nil {
		if errors.Is(err, apperror.ErrNoActiveGames) {
			game, err = that.gamePlayService.GetOrCreateGame(ctx, player, gameType)
			if err != nil {
				return nil, fmt.Errorf("failed to get game state: %w", err)
			}
			return game, nil
		}
		return nil, fmt.Errorf("failed to join public game: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	game, err := that.gamePlayService.MakeTurn(ctx, playerID, cell)
	if err != nil {
		return game, fmt.Errorf("failed to make turn: %w", err)
	}

	if game.IsFinished() {
		if err = that.EndGame(ctx, game); err != nil {
			return game, fmt.Errorf("failed to end game: %w", err)
		}

		return game, apperror.ErrGameFinished
	}

	return game, nil
}

func (that *gameUseCase) EndGame(ctx context.Context, game *entity.Game) error {
	log := that.logger.With("method", "EndGame")

	if err := that.gameService.DeleteGame(ctx, game.ID); err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	for _, player := range game.Players {
		oldMark := player.Mark
		player.GameID = ""
		player.Mark = ""
		if err := that.playerService.UpdatePlayer(ctx, player); err != nil {
			log.Error("failed to update", "player", player.ID)
			return fmt.Errorf("failed to update player: %w", err)
		}
		player.Mark = oldMark
	}

	return nil
}
