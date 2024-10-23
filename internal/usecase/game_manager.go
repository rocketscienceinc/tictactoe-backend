package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/pkg"
	"github.com/rocketscienceinc/tittactoe-backend/internal/tictactoe"
)

var ErrGameAlreadyExists = errors.New("game already exists")

type playerRepo interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

type gameRepo interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error
	GetByID(ctx context.Context, id string) (*entity.Game, error)
	DeleteByID(ctx context.Context, id string) error
}

type GameManager struct {
	logger     *slog.Logger
	playerRepo playerRepo
	gameRepo   gameRepo
}

func NewGameManager(logger *slog.Logger, playerRepo playerRepo, gameRepo gameRepo) *GameManager {
	return &GameManager{
		logger: logger,

		playerRepo: playerRepo,
		gameRepo:   gameRepo,
	}
}

func (that *GameManager) MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed get player by id: %w", err)
	}

	game, err := that.getGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("%w by id", err)
	}

	if game.IsWaiting() {
		return game, apperror.ErrGameIsNotStarted
	}

	if err = tictactoe.MakeTurn(game, player.Mark, cell); err != nil {
		if errors.Is(err, apperror.ErrGameFinished) {
			that.deleteGame(ctx, game)

			return game, apperror.ErrGameFinished
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed make turn: %w", err)
	}

	if err = that.updateGame(ctx, game); err != nil {
		return nil, fmt.Errorf("failed update game: %w", err)
	}

	if game.Status == entity.StatusFinished {
		that.deleteGame(ctx, game)

		return game, apperror.ErrGameFinished
	}

	return game, nil
}

func (that *GameManager) ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
	existingGame, err := that.getGameByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed get game by id: %w", err)
	}

	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed get player by id: %w", err)
	}

	if player.GameID == existingGame.ID {
		return existingGame, nil
	}

	if len(existingGame.Players) == 2 {
		return nil, fmt.Errorf("%w: game id %s", ErrGameAlreadyExists, gameID)
	}

	player.GameID = existingGame.ID
	player.Mark = entity.PlayerO
	if err = that.updatePlayer(ctx, player); err != nil {
		return nil, fmt.Errorf("failed update player by id: %w", err)
	}

	existingGame.Status = entity.StatusOngoing
	existingGame.Players = append(existingGame.Players, player)
	if err = that.updateGame(ctx, existingGame); err != nil {
		return nil, fmt.Errorf("failed update game by id: %w", err)
	}

	return existingGame, nil
}

func (that *GameManager) GetOrCreateGame(ctx context.Context, id string) (*entity.Game, error) {
	player, err := that.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed get player by id: %w", err)
	}

	if player.GameID == "" {
		existingGame, err := that.createGame(ctx, player)
		if err != nil {
			return nil, fmt.Errorf("failed create game: %w", err)
		}

		return existingGame, nil
	}

	existingGame, err := that.getGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed get game: %w", err)
	}

	return existingGame, nil
}

func (that *GameManager) InGame(ctx context.Context, playerID string) (*entity.Game, error) {
	player, err := that.GetOrCreatePlayer(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or get player by id: %w", err)
	}

	game, err := that.getGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed get game: %w", err)
	}

	return game, nil
}

func (that *GameManager) createGame(ctx context.Context, player *entity.Player) (*entity.Game, error) {
	gameID := pkg.GenerateGameID()
	player.GameID = gameID
	player.Mark = entity.PlayerX

	if err := that.updatePlayer(ctx, player); err != nil {
		return nil, fmt.Errorf("failed update player: %w", err)
	}

	newGame := entity.Game{}
	newGame.Create(gameID)

	newGame.Players = []*entity.Player{
		player,
	}

	if err := that.gameRepo.CreateOrUpdate(ctx, &newGame); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return &newGame, nil
}

func (that *GameManager) getGameByID(ctx context.Context, id string) (*entity.Game, error) {
	existingGame, err := that.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return existingGame, nil
}

func (that *GameManager) updateGame(ctx context.Context, game *entity.Game) error {
	if err := that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

func (that *GameManager) deleteGame(ctx context.Context, game *entity.Game) {
	log := that.logger.With("method", "deleteGame")

	if err := that.gameRepo.DeleteByID(ctx, game.ID); err != nil {
		log.Error("failed to delete game", "error", err)
	}

	for _, player := range game.Players {
		player.Mark = ""
		player.GameID = ""

		if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
			log.Error("failed to update player", "error", err)
		}
	}

	log.Info("game deleted")
}

func (that *GameManager) GetOrCreatePlayer(ctx context.Context, id string) (*entity.Player, error) {
	if id == "" {
		player, err := that.createPlayer(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create new player %w", err)
		}

		return player, nil
	}

	player, err := that.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get player by id %w", err)
	}

	return player, nil
}

func (that *GameManager) createPlayer(ctx context.Context) (*entity.Player, error) {
	playerID := pkg.GenerateNewSessionID()

	player := &entity.Player{
		ID: playerID,
	}

	if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	return player, nil
}

func (that *GameManager) getPlayerByID(ctx context.Context, id string) (*entity.Player, error) {
	player, err := that.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

func (that *GameManager) updatePlayer(ctx context.Context, player *entity.Player) error {
	if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	return nil
}
