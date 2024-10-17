package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/rocketscienceinc/tittactoe-backend/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/app_error"
	"github.com/rocketscienceinc/tittactoe-backend/internal/utils"
)

type playerRepo interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

type gameRepo interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error
	GetByID(ctx context.Context, id string) (*entity.Game, error)
	DeleteByID(ctx context.Context, id string) error
}

type iGame interface {
	Create(id string) *entity.Game
	MakeTurn(player string, cell int) error
}

type gameFactory func(game *entity.Game) iGame

type GameUseCase struct {
	playerRepo  playerRepo
	gameRepo    gameRepo
	gameFactory gameFactory
}

func NewGameUseCase(playerRepo playerRepo, gameRepo gameRepo, factory gameFactory) *GameUseCase {
	return &GameUseCase{
		playerRepo:  playerRepo,
		gameRepo:    gameRepo,
		gameFactory: factory,
	}
}

func (that *GameUseCase) MakeMove(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed get player by id: %w", err)
	}

	gameInstance, err := that.getGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed get game by id: %w", err)
	}

	if !gameInstance.IsOngoing() {
		return nil, app_error.ErrGameIsNotStarted
	}

	gameBoard := that.gameFactory(gameInstance)
	if err = gameBoard.MakeTurn(player.Mark, cell); err != nil {
		if errors.Is(err, app_error.ErrGameFinished) {

		}
	}

	actualGame, err := that.game.MakeMove(existingGame, player.Mark, cell)
	if errors.As(err, &ErrGameFinished) {
		if err = that.deleteGame(ctx, actualGame); err != nil {
			return nil, fmt.Errorf("failed to delete game: %w", err)
		}
		return existingGame, ErrGameFinished
	}

	if err != nil {
		return nil, fmt.Errorf("failed make move: %w", err)
	}

	if err = that.updateGame(ctx, actualGame); err != nil {
		return nil, fmt.Errorf("failed update game: %w", err)
	}

	if actualGame.Status == statusFinished {
		if err = that.deleteGame(ctx, actualGame); err != nil {
			return nil, fmt.Errorf("failed to delete game: %w", err)
		}
		return existingGame, app_error.ErrGameFinished
	}

	return existingGame, nil
}

func (that *GameUseCase) ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
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
		return nil, fmt.Errorf("game with id %s already exists", gameID)
	}

	player.GameID = existingGame.ID
	player.Mark = playerO
	if err = that.updatePlayer(ctx, player); err != nil {
		return nil, fmt.Errorf("failed update player by id: %w", err)
	}

	existingGame.Status = statusOngoing
	existingGame.Players = append(existingGame.Players, player)
	if err = that.updateGame(ctx, existingGame); err != nil {
		return nil, fmt.Errorf("failed update game by id: %w", err)
	}

	return existingGame, nil
}

func (that *GameUseCase) GetOrCreateGame(ctx context.Context, id string) (*entity.Game, error) {
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

func (that *GameUseCase) createGame(ctx context.Context, player *entity.Player) (*entity.Game, error) {
	gameID := utils.GenerateGameID()
	player.GameID = gameID
	player.Mark = playerX

	if err := that.updatePlayer(ctx, player); err != nil {
		return nil, fmt.Errorf("failed update player: %w", err)
	}

	existingGame := that.game.Create(gameID)
	existingGame.Status = statusWaiting
	existingGame.Players = []*entity.Player{
		player,
	}

	if err := that.gameRepo.CreateOrUpdate(ctx, existingGame); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return existingGame, nil
}

func (that *GameUseCase) getGameByID(ctx context.Context, id string) (*entity.Game, error) {
	existingGame, err := that.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return existingGame, nil
}

func (that *GameUseCase) updateGame(ctx context.Context, game *entity.Game) error {
	if err := that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

func (that *GameUseCase) deleteGame(ctx context.Context, game *entity.Game) {
	if err := that.gameRepo.DeleteByID(ctx, game.ID); err != nil {
		// ToDo:
		return fmt.Errorf("failed to delete game: %w", err)
	}

	for _, player := range game.Players {
		player.Mark = ""
		player.GameID = ""

		if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
			return fmt.Errorf("failed to update player: %w", err)
		}
	}

	return nil
}

func (that *GameUseCase) GetOrCreatePlayer(ctx context.Context, id string) (*entity.Player, error) {
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

func (that *GameUseCase) createPlayer(ctx context.Context) (*entity.Player, error) {
	playerID := utils.GenerateNewSessionID()

	player := &entity.Player{
		ID: playerID,
	}

	if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	return player, nil
}

func (that *GameUseCase) getPlayerByID(ctx context.Context, id string) (*entity.Player, error) {
	player, err := that.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

func (that *GameUseCase) updatePlayer(ctx context.Context, player *entity.Player) error {
	if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	return nil
}
