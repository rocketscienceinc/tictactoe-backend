package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

var ErrGameAlreadyExists = errors.New("game already exists")

type GamePlayService interface {
	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
	ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	GetGameState(ctx context.Context, player *entity.Player) (*entity.Game, error)
}

type gamePlayService struct {
	logger *slog.Logger

	playerService PlayerService
	gameService   GameService
}

func NewGamePlayService(logger *slog.Logger, playerService PlayerService, gameService GameService) GamePlayService {
	return &gamePlayService{
		logger:        logger,
		playerService: playerService,
		gameService:   gameService,
	}
}

func (that *gamePlayService) MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player by id: %w", err)
	}

	game, err := that.gameService.GetGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game by id: %w", err)
	}

	if err = game.IsActive(); err != nil {
		if errors.Is(err, apperror.ErrGameFinished) {
			that.cleanupGame(ctx, game)

			return nil, err //nolint: wrapcheck // it`s ok
		}
		return game, fmt.Errorf("game is not active: %w", err)
	}

	if err = game.MakeTurn(player.Mark, cell); err != nil {
		return nil, fmt.Errorf("failed to make turn: %w", err)
	}

	if err = that.gameService.UpdateGame(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	if game.Status == entity.StatusFinished {
		that.cleanupGame(ctx, game)
	}

	return game, nil
}

func (that *gamePlayService) ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
	game, err := that.gameService.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game by id: %w", err)
	}

	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player by id: %w", err)
	}

	if player.GameID == game.ID {
		return game, nil
	}

	if len(game.Players) >= 2 {
		return nil, fmt.Errorf("%w: game id %s", ErrGameAlreadyExists, gameID)
	}

	player.GameID = game.ID
	player.Mark = entity.PlayerO
	if err = that.playerService.UpdatePlayer(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to update player: %w", err)
	}

	game.Status = entity.StatusOngoing
	game.Players = append(game.Players, player)
	if err = that.gameService.UpdateGame(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return game, nil
}

func (that *gamePlayService) GetGameState(ctx context.Context, player *entity.Player) (*entity.Game, error) {
	if player.GameID == "" {
		game, updatedPlayer, err := that.gameService.CreateGame(ctx, player)
		if err != nil {
			return nil, fmt.Errorf("failed to create game: %w", err)
		}

		if err = that.playerService.UpdatePlayer(ctx, updatedPlayer); err != nil {
			return nil, fmt.Errorf("failed to update player: %w", err)
		}

		return game, nil
	}

	game, err := that.gameService.GetGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return game, nil
}

func (that *gamePlayService) cleanupGame(ctx context.Context, game *entity.Game) {
	log := that.logger.With("method", "cleanupGame")

	if err := that.gameService.DeleteGame(ctx, game.ID); err != nil {
		log.Error("failed to delete game %s: %w", game.ID, err)
	}

	for _, player := range game.Players {
		player.GameID = ""
		player.Mark = ""
		if err := that.playerService.UpdatePlayer(ctx, player); err != nil {
			log.Error("failed to update player %s: %w", player.GameID, err)
		}
	}
}
