package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/rocketscienceinc/tittactoe-backend/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/utils"
)

const (
	statusOngoing  = "ongoing"
	statusWaiting  = "waiting"
	statusFinished = "finished"

	playerX = "X"
	playerO = "O"
)

var (
	ErrGameFinished     = errors.New("game is already finished")
	ErrGameIsNotStarted = errors.New("game is not started")
)

type UGame interface {
	GetOrCreatePlayer(ctx context.Context, id string) (*entity.Player, error)
	GetOrCreateGame(ctx context.Context, id string) (*entity.Game, error)

	ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error)

	MakeMove(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type playerService interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

type gameService interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error
	GetByID(ctx context.Context, id string) (*entity.Game, error)
	DeleteByID(ctx context.Context, id string) error
}

type game interface {
	Create(id string) *entity.Game
	MakeMove(game *entity.Game, player string, cell int) (*entity.Game, error)
}

type uGame struct {
	playerService playerService
	gameService   gameService
	game          game
}

func NewUGame(playerService playerService, gameService gameService, game game) UGame {
	return &uGame{
		playerService: playerService,
		gameService:   gameService,
		game:          game,
	}
}

func (that *uGame) MakeMove(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed get player by id: %w", err)
	}

	existingGame, err := that.getGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed get game by id: %w", err)
	}

	if existingGame.Status != statusOngoing {
		return nil, ErrGameIsNotStarted
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
		return existingGame, ErrGameFinished
	}

	return existingGame, nil
}

func (that *uGame) ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
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

func (that *uGame) GetOrCreateGame(ctx context.Context, id string) (*entity.Game, error) {
	player, err := that.playerService.GetByID(ctx, id)
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

func (that *uGame) createGame(ctx context.Context, player *entity.Player) (*entity.Game, error) {
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

	if err := that.gameService.CreateOrUpdate(ctx, existingGame); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return existingGame, nil
}

func (that *uGame) getGameByID(ctx context.Context, id string) (*entity.Game, error) {
	existingGame, err := that.gameService.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return existingGame, nil
}

func (that *uGame) updateGame(ctx context.Context, game *entity.Game) error {
	if err := that.gameService.CreateOrUpdate(ctx, game); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

func (that *uGame) deleteGame(ctx context.Context, game *entity.Game) error {
	if err := that.gameService.DeleteByID(ctx, game.ID); err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	for _, player := range game.Players {
		player.Mark = ""
		player.GameID = ""

		if err := that.playerService.CreateOrUpdate(ctx, player); err != nil {
			return fmt.Errorf("failed to update player: %w", err)
		}
	}

	return nil
}
