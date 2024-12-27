package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

const lettersAndNumbers = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type playerRepoDep interface {
	CreateOrUpdate(ctx context.Context, player *entity.Player) error
	GetByID(ctx context.Context, id string) (*entity.Player, error)
}

type gameRepoDep interface {
	CreateOrUpdate(ctx context.Context, game *entity.Game) error

	GetByID(ctx context.Context, id string) (*entity.Game, error)
	GetOpenPublicGame(ctx context.Context) (*entity.Game, error)

	DeleteByID(ctx context.Context, id string) error
}

type gameUseCase struct {
	playerRepo playerRepoDep
	gameRepo   gameRepoDep
}

func NewGameUseCase(playerRepo playerRepoDep, gameRepo gameRepoDep) *gameUseCase { //nolint: revive // it's ok
	return &gameUseCase{
		playerRepo: playerRepo,
		gameRepo:   gameRepo,
	}
}

func (that *gameUseCase) GetOrCreatePlayer(ctx context.Context, playerID string) (*entity.Player, error) {
	if playerID == "" {
		playerID = that.generateNewPlayerID()
		player := &entity.Player{ID: playerID}

		if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
			return nil, fmt.Errorf("failed to create player from storage: %w", err)
		}

		return player, nil
	}

	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}

	return player, nil
}

func (that *gameUseCase) GetOrCreateGame(ctx context.Context, playerID, gameType, difficulty string) (*entity.Game, error) {
	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}

	if player.GameID == "" {
		game, err := that.createGame(ctx, gameType, difficulty, player)
		if err != nil {
			return nil, fmt.Errorf("failed to create game: %w", err)
		}

		return game, nil
	}

	game, err := that.gameRepo.GetByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game by ID: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) addBotToGame(ctx context.Context, game *entity.Game) error {
	botPlayer := entity.NewBotPlayer(game.ID, "")

	game.Players = append(game.Players, botPlayer)
	game.Status = entity.StatusOngoing

	playerMark, botMark := game.GetRandomMarks()
	for _, player := range game.Players {
		if !player.IsBot() {
			player.Mark = playerMark
			if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
				return fmt.Errorf("failed to update player: %w", err)
			}
		}
	}
	botPlayer.Mark = botMark

	if err := that.playerRepo.CreateOrUpdate(ctx, botPlayer); err != nil {
		return fmt.Errorf("failed to update bot player: %w", err)
	}

	if botMark == entity.PlayerX {
		if err := game.BotMakeTurn(); err != nil {
			return fmt.Errorf("bot failed to make first turn: %w", err)
		}
	}

	if err := that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return fmt.Errorf("failed to update game with bot: %w", err)
	}

	return nil
}

func (that *gameUseCase) GetGameByPlayerID(ctx context.Context, playerID string) (*entity.Game, error) {
	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}

	game, err := that.gameRepo.GetByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error) {
	gameID = strings.ToUpper(gameID)

	game, err := that.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve game from storage: %w", err)
	}

	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}

	if player.GameID == game.ID {
		return game, nil
	}

	if len(game.Players) >= 2 {
		return nil, fmt.Errorf("%w: game id %s", apperror.ErrGameAlreadyExists, gameID)
	}

	player.GameID = game.ID
	player.Mark = entity.PlayerO

	if err = that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to update player from storage: %w", err)
	}

	game.Status = entity.StatusOngoing
	game.Players = append(game.Players, player)

	if err = that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game with player: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) CreateOrJoinToPublicGame(ctx context.Context, playerID, gameType string) (*entity.Game, error) {
	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}

	game, err := that.gameRepo.GetOpenPublicGame(ctx)
	if err != nil {
		if errors.Is(err, apperror.ErrNoActiveGames) {
			game, err = that.createGame(ctx, gameType, "", player)
			if err != nil {
				return nil, fmt.Errorf("failed to create game: %w", err)
			}

			return game, nil
		}
		return nil, fmt.Errorf("failed to get game open public game: %w", err)
	}

	if player.GameID == game.ID {
		return game, nil
	}

	if len(game.Players) >= 2 {
		return nil, fmt.Errorf("%w: game id %s", apperror.ErrGameAlreadyExists, game.ID)
	}

	player.GameID = game.ID
	player.Mark = entity.PlayerO

	if err = that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to update player from storage: %w", err)
	}

	game.Status = entity.StatusOngoing
	game.Players = append(game.Players, player)

	if err = that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game with player: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) createGame(ctx context.Context, gameType, difficulty string, player *entity.Player) (*entity.Game, error) {
	gameID, err := that.generateGameID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate game ID: %w", err)
	}

	game := entity.NewGame(gameID, gameType)
	if game.IsWithBot() {
		game.Difficulty = difficulty
	}

	player.GameID = gameID
	player.Mark = entity.PlayerX
	if err = that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to update player from storage: %w", err)
	}

	game.Players = []*entity.Player{player}
	if err = that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	if game.IsWithBot() {
		err = that.addBotToGame(ctx, game)
		if err != nil {
			return nil, fmt.Errorf("failed to add bot to game: %w", err)
		}
	}

	return game, nil
}

func (that *gameUseCase) MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error) {
	player, err := that.getPlayerByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}

	game, err := that.gameRepo.GetByID(ctx, player.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game by id: %w", err)
	}

	if !game.IsOngoing() {
		return nil, apperror.ErrGameAlreadyExists
	}

	if err = game.MakeTurn(player.Mark, cell); err != nil {
		return game, fmt.Errorf("failed to make turn: %w", err)
	}

	if game.IsFinished() {
		if err = that.EndGame(ctx, game); err != nil {
			return game, fmt.Errorf("failed to end game: %w", err)
		}

		return game, apperror.ErrGameFinished
	}

	if !game.IsWithBot() {
		if err = that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
			return nil, fmt.Errorf("failed to update bot to game: %w", err)
		}

		return game, nil
	}

	// in this case if we are playing vs bot we should make a turn after players turn.
	if err = game.BotMakeTurn(); err != nil {
		return nil, fmt.Errorf("failed to make bot turn: %w", err)
	}

	// we update this for redis storage
	if err = that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	if game.IsFinished() {
		if err = that.EndGame(ctx, game); err != nil {
			return game, fmt.Errorf("failed to end game: %w", err)
		}

		return game, apperror.ErrGameFinished
	}

	return game, nil
}

func (that *gameUseCase) CreatePrivateGameWithTwoPlayers(ctx context.Context, player1, player2 *entity.Player) (*entity.Game, error) {
	gameID, err := that.generateGameID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate game ID: %w", err)
	}
	game := entity.NewGame(gameID, entity.PrivateType)

	player1.GameID = game.ID
	player2.GameID = game.ID

	player1.Mark = entity.PlayerX
	player2.Mark = entity.PlayerO

	if err = that.playerRepo.CreateOrUpdate(ctx, player1); err != nil {
		return nil, fmt.Errorf("failed to update player from storage: %w", err)
	}
	if err = that.playerRepo.CreateOrUpdate(ctx, player2); err != nil {
		return nil, fmt.Errorf("failed to update player with two players: %w", err)
	}

	game.Players = []*entity.Player{player1, player2}

	game.Status = entity.StatusOngoing

	if err = that.gameRepo.CreateOrUpdate(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game with player: %w", err)
	}

	return game, nil
}

func (that *gameUseCase) EndGame(ctx context.Context, game *entity.Game) error {
	if err := that.gameRepo.DeleteByID(ctx, game.ID); err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	if len(game.Players) >= 2 {
		player1 := game.Players[0]
		player2 := game.Players[1]

		player1.LastOpponentID = player2.ID
		player2.LastOpponentID = player1.ID
	}

	for _, player := range game.Players {
		oldMark := player.Mark
		player.GameID = ""
		player.Mark = ""
		if err := that.playerRepo.CreateOrUpdate(ctx, player); err != nil {
			return fmt.Errorf("failed to update player: %w", err)
		}
		player.Mark = oldMark
	}

	return nil
}

func (that *gameUseCase) getPlayerByID(ctx context.Context, playerID string) (*entity.Player, error) {
	player, err := that.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve player from storage: %w", err)
	}

	return player, nil
}

// GenerateGameID - generates a unique identifier for the room.
func (that *gameUseCase) generateGameID() (string, error) {
	length := 10

	gameID := make([]byte, length)
	for i := range gameID {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(lettersAndNumbers))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random index: %w", err)
		}
		gameID[i] = lettersAndNumbers[index.Int64()]
	}

	return string(gameID), nil
}

// GenerateNewPlayerID - generates a new unique playerID.
func (that *gameUseCase) generateNewPlayerID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "error-generating-player-id"
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
