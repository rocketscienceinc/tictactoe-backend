package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
	mockedUseCase "github.com/rocketscienceinc/tictactoe-backend/mocks/usecase"
)

var (
	errSomeError      = errors.New("some error")
	errStorageIsFull  = errors.New("storage is full")
	errCantGetPlayer  = errors.New("can't get player")
	errRedisDown      = errors.New("redis down")
	errPlayerNotFound = errors.New("player not found")
	errGameNotFound   = errors.New("game not found")
)

func TestGameUseCase_GetOrCreatePlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("Creates a new player when playerID is empty", func(t *testing.T) {
		// Given: A mock player repository and a mock game repository
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		mockPlayerRepo.EXPECT().
			CreateOrUpdate(mock.Anything, mock.AnythingOfType("*entity.Player")).
			Return(nil).
			Once()

		// When: Calling GetOrCreatePlayer with an empty playerID
		player, err := useCaseInstance.GetOrCreatePlayer(ctx, "")

		// Then: A new player should be created, and no error should occur
		require.NoError(t, err)
		assert.NotEmpty(t, player.ID)
	})

	t.Run("Returns existing player when playerID is not empty", func(t *testing.T) {
		// Given: A mock player repository that returns an existing player
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		existingPlayer := &entity.Player{ID: "player123"}
		mockPlayerRepo.EXPECT().
			GetByID(mock.Anything, "player123").
			Return(existingPlayer, nil).
			Once()

		// When: Calling GetOrCreatePlayer with a known playerID
		player, err := useCaseInstance.GetOrCreatePlayer(ctx, "player123")

		// Then: The existing player should be returned, and no error should occur
		require.NoError(t, err)
		assert.Equal(t, existingPlayer, player)
	})

	t.Run("Returns error if playerRepo.GetByID fails", func(t *testing.T) {
		// Given: A mock player repository that fails to get the player
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		mockPlayerRepo.EXPECT().
			GetByID(mock.Anything, "playerErr").
			Return((*entity.Player)(nil), errSomeError).
			Once()

		// When: Calling GetOrCreatePlayer with a failing playerRepo
		player, err := useCaseInstance.GetOrCreatePlayer(ctx, "playerErr")

		// Then: An error should be returned, and the player should be nil
		require.Error(t, err)
		assert.Nil(t, player)
	})

	t.Run("Returns error if playerRepo.CreateOrUpdate fails for new player", func(t *testing.T) {
		// Given: A mock player repository that fails on CreateOrUpdate
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		mockPlayerRepo.EXPECT().
			CreateOrUpdate(mock.Anything, mock.AnythingOfType("*entity.Player")).
			Return(errStorageIsFull).
			Once()

		// When: Calling GetOrCreatePlayer with an empty playerID
		player, err := useCaseInstance.GetOrCreatePlayer(ctx, "")

		// Then: An error should be returned, and the player should be nil
		require.Error(t, err)
		assert.Nil(t, player)
	})
}

func TestGameUseCase_GetOrCreateGame(t *testing.T) {
	ctx := context.Background()

	t.Run("Creates new game when player has no GameID", func(t *testing.T) {
		// Given: A mock setup where the player has no GameID
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		playerID := "p1"
		player := &entity.Player{ID: playerID, GameID: ""}
		mockPlayerRepo.EXPECT().
			GetByID(ctx, playerID).
			Return(player, nil).
			Once()

		mockGameRepo.EXPECT().
			CreateOrUpdate(ctx, mock.AnythingOfType("*entity.Game")).
			Return(nil).
			Once()

		mockPlayerRepo.EXPECT().
			CreateOrUpdate(ctx, mock.AnythingOfType("*entity.Player")).
			Return(nil).
			Once()

		// When: Calling GetOrCreateGame with a player who has no GameID
		game, err := useCaseInstance.GetOrCreateGame(ctx, playerID, entity.PrivateType, entity.HardDifficulty)

		// Then: A new game should be created and returned without error
		require.NoError(t, err)
		assert.NotNil(t, game)
		assert.Equal(t, entity.PrivateType, game.Type)
	})

	t.Run("Returns existing game if player has GameID", func(t *testing.T) {
		// Given: A mock setup where the player already has a GameID
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		playerID := "p2"
		player := &entity.Player{ID: playerID, GameID: "g123"}
		existingGame := &entity.Game{ID: "g123", Status: entity.StatusOngoing}

		mockPlayerRepo.EXPECT().
			GetByID(ctx, playerID).
			Return(player, nil).
			Once()

		mockGameRepo.EXPECT().
			GetByID(ctx, "g123").
			Return(existingGame, nil).
			Once()

		// When: Calling GetOrCreateGame with a player who has an existing GameID
		game, err := useCaseInstance.GetOrCreateGame(ctx, playerID, entity.PublicType, entity.EasyDifficulty)

		// Then: The existing game should be returned without error
		require.NoError(t, err)
		assert.Equal(t, existingGame, game)
	})

	t.Run("Returns error if playerRepo.GetByID fails", func(t *testing.T) {
		// Given: A mock player repository that fails when getting the player
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		mockPlayerRepo.EXPECT().
			GetByID(ctx, "somePlayer").
			Return((*entity.Player)(nil), errCantGetPlayer).
			Once()

		// When: Calling GetOrCreateGame but GetByID fails
		game, err := useCaseInstance.GetOrCreateGame(ctx, "somePlayer", entity.WithBotType, entity.EasyDifficulty)

		// Then: An error should be returned, and the game should be nil
		require.Error(t, err)
		assert.Nil(t, game)
	})

	t.Run("Returns error if gameRepo.CreateOrUpdate fails", func(t *testing.T) {
		// Given: A mock game repository that fails on CreateOrUpdate
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		player := &entity.Player{ID: "p3", GameID: ""}

		mockPlayerRepo.EXPECT().
			GetByID(ctx, "p3").
			Return(player, nil).
			Once()

		mockPlayerRepo.EXPECT().
			CreateOrUpdate(ctx, mock.MatchedBy(func(p *entity.Player) bool {
				return p.ID == "p3" && p.Mark == "X"
			})).
			Return(nil).
			Once()

		mockGameRepo.EXPECT().
			CreateOrUpdate(ctx, mock.AnythingOfType("*entity.Game")).
			Return(errRedisDown).
			Once()

		// When: Calling GetOrCreateGame and CreateOrUpdate fails
		game, err := useCaseInstance.GetOrCreateGame(ctx, "p3", entity.WithBotType, entity.HardDifficulty)

		// Then: An error should be returned, and the game should be nil
		require.Error(t, err)
		assert.Nil(t, game)
	})
}

func TestGameUseCase_MakeTurn(t *testing.T) {
	ctx := context.Background()

	t.Run("Error if cannot get player", func(t *testing.T) {
		// Given: A mock setup where retrieving the player fails
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		mockPlayerRepo.EXPECT().
			GetByID(ctx, "p1").
			Return((*entity.Player)(nil), errPlayerNotFound).
			Once()

		// When: Calling MakeTurn and the player does not exist
		game, err := useCaseInstance.MakeTurn(ctx, "p1", 0)

		// Then: An error should be returned, and the game should be nil
		require.Error(t, err)
		assert.Nil(t, game)
	})

	t.Run("Error if game not found", func(t *testing.T) {
		// Given: A mock setup where the game cannot be found
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		mockPlayerRepo.EXPECT().
			GetByID(ctx, "p2").
			Return(&entity.Player{ID: "p2", GameID: "g2"}, nil).
			Once()

		mockGameRepo.EXPECT().
			GetByID(ctx, "g2").
			Return((*entity.Game)(nil), errGameNotFound).
			Once()

		// When: Calling MakeTurn but the game does not exist
		game, err := useCaseInstance.MakeTurn(ctx, "p2", 1)

		// Then: An error should be returned, and the game should be nil
		require.Error(t, err)
		assert.Nil(t, game)
	})

	t.Run("Error if game is not Ongoing", func(t *testing.T) {
		// Given: A mock setup where the game is finished
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		mockPlayerRepo.EXPECT().
			GetByID(ctx, "p3").
			Return(&entity.Player{ID: "p3", GameID: "g3", Mark: entity.PlayerX}, nil).
			Once()

		notOngoing := &entity.Game{ID: "g3", Status: entity.StatusFinished}
		mockGameRepo.EXPECT().
			GetByID(ctx, "g3").
			Return(notOngoing, nil).
			Once()

		// When: Calling MakeTurn on a finished game
		game, err := useCaseInstance.MakeTurn(ctx, "p3", 2)

		// Then: An error should be returned, and the game should be nil
		require.Error(t, err)
		assert.Nil(t, game)
	})

	t.Run("Successful turn in a PVP game", func(t *testing.T) {
		// Given: A mock setup for a valid ongoing game with two human players
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		playerX := &entity.Player{ID: "pX", GameID: "gX", Mark: entity.PlayerX}
		gameOngoing := &entity.Game{
			ID:     "gX",
			Status: entity.StatusOngoing,
			Board:  [9]string{"", "", "", "", "", "", "", "", ""},
			Turn:   entity.PlayerX,
			Type:   entity.PrivateType,
		}

		mockPlayerRepo.EXPECT().
			GetByID(ctx, "pX").
			Return(playerX, nil).
			Once()

		mockGameRepo.EXPECT().
			GetByID(ctx, "gX").
			Return(gameOngoing, nil).
			Once()

		mockGameRepo.EXPECT().
			CreateOrUpdate(ctx, gameOngoing).
			Return(nil).
			Once()

		// When: Player X makes a valid turn on cell 4
		game, err := useCaseInstance.MakeTurn(ctx, "pX", 4)

		// Then: The turn should succeed, and the game should update accordingly
		require.NoError(t, err)
		assert.Equal(t, entity.PlayerO, game.Turn)
		assert.Equal(t, entity.PlayerX, game.Board[4])
	})

	t.Run("Player moves in a Bot game => Bot moves next", func(t *testing.T) {
		// Given: A mock setup for a game with a bot and an ongoing status
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		playerX := &entity.Player{ID: "pX", GameID: "gBot", Mark: entity.PlayerX}
		botPlayer := &entity.Player{ID: "bot:gBot", GameID: "gBot", Mark: entity.PlayerO}
		gameWithBot := &entity.Game{
			ID:         "gBot",
			Status:     entity.StatusOngoing,
			Board:      [9]string{"", "", "", "", "", "", "", "", ""},
			Turn:       entity.PlayerX,
			Players:    []*entity.Player{playerX, botPlayer},
			Type:       entity.WithBotType,
			Difficulty: entity.EasyDifficulty,
		}

		mockPlayerRepo.EXPECT().
			GetByID(ctx, "pX").
			Return(playerX, nil).
			Once()

		mockGameRepo.EXPECT().
			GetByID(ctx, "gBot").
			Return(gameWithBot, nil).
			Once()

		// Expect a single CreateOrUpdate call for this test scenario
		mockGameRepo.EXPECT().
			CreateOrUpdate(ctx, gameWithBot).
			Return(nil).
			Once()

		// When: Player X makes a turn on cell 0, then the bot should move
		game, err := useCaseInstance.MakeTurn(ctx, "pX", 0)

		// Then: The player's move should succeed, and the bot should also move
		require.NoError(t, err)
		require.NotNil(t, game)
		assert.Equal(t, entity.PlayerX, game.Board[0])

		botHasMoved := false
		for i := 1; i < 9; i++ {
			if game.Board[i] == entity.PlayerO {
				botHasMoved = true
				break
			}
		}
		assert.True(t, botHasMoved)
		assert.Equal(t, entity.StatusOngoing, game.Status)

		mockPlayerRepo.AssertExpectations(t)
		mockGameRepo.AssertExpectations(t)
	})
}

func TestGameUseCase_EndGame(t *testing.T) {
	ctx := context.Background()

	t.Run("Successfully ends the game and clears players", func(t *testing.T) {
		// Given: A mock setup for an already finished game with two players
		mockPlayerRepo := mockedUseCase.NewMockplayerRepoDep(t)
		mockGameRepo := mockedUseCase.NewMockgameRepoDep(t)
		useCaseInstance := NewGameUseCase(mockPlayerRepo, mockGameRepo)

		players := []*entity.Player{
			{ID: "p1", GameID: "game123", Mark: entity.PlayerX},
			{ID: "p2", GameID: "game123", Mark: entity.PlayerO},
		}
		game := &entity.Game{
			ID:      "game123",
			Players: players,
			Status:  entity.StatusFinished,
		}

		mockGameRepo.EXPECT().
			DeleteByID(ctx, "game123").
			Return(nil).
			Once()

		mockPlayerRepo.EXPECT().
			CreateOrUpdate(ctx, &entity.Player{ID: "p1", GameID: "", Mark: ""}).
			Return(nil).
			Once()

		mockPlayerRepo.EXPECT().
			CreateOrUpdate(ctx, &entity.Player{ID: "p2", GameID: "", Mark: ""}).
			Return(nil).
			Once()

		// When: EndGame is called on a finished game
		err := useCaseInstance.EndGame(ctx, game)

		// Then: The game should be deleted and players should be cleared
		require.NoError(t, err)
	})
}
