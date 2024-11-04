package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

var ErrGameAlreadyExists = errors.New("game already exists")

type GamePlayService interface {
	JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	JoinWaitingPublicGame(ctx context.Context, playerID string) (*entity.Game, error)

	GetOrCreateGame(ctx context.Context, player *entity.Player, gameType string) (*entity.Game, error)
	CleanupGame(ctx context.Context, game *entity.Game)

	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type gamePlayService struct {
	logger *slog.Logger

	playerService PlayerService
	gameService   GameService
	botService    BotService
}

func NewGamePlayService(logger *slog.Logger, playerService PlayerService, gameService GameService, botService BotService) GamePlayService {
	return &gamePlayService{
		logger:        logger,
		playerService: playerService,
		gameService:   gameService,
		botService:    botService,
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

	if !game.IsOngoing() {
		return nil, ErrGameAlreadyExists
	}

	if err = game.MakeTurn(player.Mark, cell); err != nil {
		return game, fmt.Errorf("failed to make turn: %w", err)
	}

	if game.IsFinished() {
		if err = that.gameService.UpdateGame(ctx, game); err != nil {
			return nil, fmt.Errorf("failed to update game: %w", err)
		}

		return game, nil
	}

	if game.IsWithBot() {
		if err = that.botService.MakeTurn(game); err != nil {
			return nil, fmt.Errorf("bot failed to make turn: %w", err) // ToDo: Need Fix this
		}
	}

	if err = that.gameService.UpdateGame(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return game, nil
}

func (that *gamePlayService) JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
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

func (that *gamePlayService) JoinWaitingPublicGame(ctx context.Context, playerID string) (*entity.Game, error) {
	game, err := that.gameService.GetPublicGameByID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get game waiting public game: %w", err)
	}

	player, err := that.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player by id: %w", err)
	}

	if player.GameID == game.ID {
		return game, nil
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

func (that *gamePlayService) GetOrCreateGame(ctx context.Context, player *entity.Player, gameType string) (*entity.Game, error) {
	if player.GameID == "" {
		game, err := that.createGame(ctx, player, gameType)
		if err != nil {
			return nil, fmt.Errorf("failed to create new game: %w", err)
		}

		return game, nil
	}

	game, err := that.gameService.GetGameByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return game, nil
}

func (that *gamePlayService) createGame(ctx context.Context, player *entity.Player, gameType string) (*entity.Game, error) {
	game, updatedPlayer, err := that.gameService.CreateGame(ctx, player, gameType)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	if err = that.playerService.UpdatePlayer(ctx, updatedPlayer); err != nil {
		return nil, fmt.Errorf("failed to update player: %w", err)
	}

	if game.IsWithBot() {
		err = that.addBotToGame(ctx, game)
		if err != nil {
			return nil, fmt.Errorf("failed to add bot to game: %w", err)
		}
	}

	return game, nil
}

func (that *gamePlayService) addBotToGame(ctx context.Context, game *entity.Game) error {
	botPlayer := entity.NewBotPlayer(game.ID, "")

	game.Players = append(game.Players, botPlayer)
	game.Status = entity.StatusOngoing

	playerMark, botMark := game.GetRandomMarks()
	for _, player := range game.Players {
		if !player.IsBot() {
			player.Mark = playerMark
			if err := that.playerService.UpdatePlayer(ctx, player); err != nil {
				return fmt.Errorf("failed to update player: %w", err)
			}
		}
	}
	botPlayer.Mark = botMark

	if err := that.playerService.UpdatePlayer(ctx, botPlayer); err != nil {
		return fmt.Errorf("failed to update bot player: %w", err)
	}

	if botMark == entity.PlayerX {
		if err := that.botService.MakeTurn(game); err != nil {
			return fmt.Errorf("bot failed to make first turn: %w", err)
		}
	}

	if err := that.gameService.UpdateGame(ctx, game); err != nil {
		return fmt.Errorf("failed to update game with bot: %w", err)
	}

	return nil
}

func (that *gamePlayService) CleanupGame(ctx context.Context, game *entity.Game) {
	log := that.logger.With("method", "cleanupGame", "gameID", game.ID)

	if err := that.gameService.DeleteGame(ctx, game.ID); err != nil {
		log.Error("failed to delete game", "error", err)
	}

	for _, player := range game.Players {
		oldMark := player.Mark
		player.GameID = ""
		player.Mark = ""
		if err := that.playerService.UpdatePlayer(ctx, player); err != nil {
			log.Error("failed to update", "player", player.ID, "error", err)
		}
		player.Mark = oldMark
	}
}
