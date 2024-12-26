package repository

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tictactoe-backend/testing/suite"
)

func getLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))
}

func TestGameRepository_CreateOrUpdate(t *testing.T) {
	ctx, st := suite.New(t)

	logger := getLogger()

	gameRepo := NewGameRepository(logger, st.Storage)

	// Given: a game with ID and status
	game := &entity.Game{
		ID:     "123",
		Status: entity.StatusWaiting,
	}

	// When: CreateOrUpdate is called
	err := gameRepo.CreateOrUpdate(ctx, game)

	// Then: no error should be returned, and game is stored
	require.NoError(t, err)
}

func TestGameRepository_GetByID(t *testing.T) {
	t.Run("GetByID_Success", func(t *testing.T) {
		ctx, st := suite.New(t)

		logger := getLogger()

		gameRepo := NewGameRepository(logger, st.Storage)

		// Given: a game with ID and status
		game := &entity.Game{
			ID:     "123",
			Status: entity.StatusWaiting,
		}

		err := gameRepo.CreateOrUpdate(ctx, game)
		require.NoError(t, err)

		// When: GetByID is called with existing ID
		retrievedGame, err := gameRepo.GetByID(ctx, game.ID)

		// Then: the retrieved game should match the saved game
		require.NoError(t, err)
		require.Equal(t, game.ID, retrievedGame.ID)
		require.Equal(t, game.Status, retrievedGame.Status)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		ctx, st := suite.New(t)

		logger := getLogger()

		gameRepo := NewGameRepository(logger, st.Storage)

		nonExistentGameID := "9999999"

		// When: GetByID is called with non-existent ID
		_, err := gameRepo.GetByID(ctx, nonExistentGameID)

		// Then: an ErrGameNotFound error should be returned
		require.Error(t, err)
		assert.Equal(t, ErrGameNotFound, err)
	})
}

func TestGameRepository_DeleteByID(t *testing.T) {
	t.Run("DeleteByID_Success", func(t *testing.T) {
		ctx, st := suite.New(t)

		logger := getLogger()

		gameRepo := NewGameRepository(logger, st.Storage)

		// Given: a game with ID and status
		game := &entity.Game{
			ID:     "123",
			Status: entity.StatusFinished,
		}

		err := gameRepo.CreateOrUpdate(ctx, game)
		require.NoError(t, err)

		// When: DeleteByID is called with existing ID
		err = gameRepo.DeleteByID(ctx, game.ID)

		// Then: no error should be returned
		require.NoError(t, err)

		_, err = gameRepo.GetByID(ctx, game.ID)
		require.Error(t, err)
		assert.Equal(t, ErrGameNotFound, err)
	})
	t.Run("DeleteByID_NotFound", func(t *testing.T) {
		ctx, st := suite.New(t)

		logger := getLogger()

		gameRepo := NewGameRepository(logger, st.Storage)

		// Given: a non-existent game ID
		nonExistentGameID := "9999999"

		// When: DeleteByID is called with non-existent ID
		err := gameRepo.DeleteByID(ctx, nonExistentGameID)

		// Then: an ErrGameNotFound error should be returned
		require.Error(t, err)
		require.Equal(t, ErrGameNotFound, err)
	})
}

func TestGameRepository_GetWaitingPublicGame(t *testing.T) {
	t.Run("GetWaitingPublicGame_Success", func(t *testing.T) {
		ctx, st := suite.New(t)
		logger := getLogger()

		gameRepo := NewGameRepository(logger, st.Storage)

		// Given: an existing public game with status waiting
		existingGame := &entity.Game{
			ID:     "123",
			Status: entity.StatusWaiting,
			Type:   entity.PublicType,
		}

		err := gameRepo.CreateOrUpdate(ctx, existingGame)
		require.NoError(t, err)

		// When: GetWaitingPublicGame is called
		game, err := gameRepo.GetWaitingPublicGame(ctx)

		// Then: the retrieved game should match the existing game
		require.NoError(t, err)
		require.Equal(t, existingGame.ID, game.ID)
	})
	t.Run("GetWaitingPublicGame_NotFound", func(t *testing.T) {
		ctx, st := suite.New(t)
		logger := getLogger()

		gameRepo := NewGameRepository(logger, st.Storage)

		// Given: an existing private game with status waiting
		existingGame := &entity.Game{
			ID:     "9999999",
			Type:   entity.PrivateType,
			Status: entity.StatusWaiting,
		}

		err := gameRepo.CreateOrUpdate(ctx, existingGame)
		require.NoError(t, err)

		game, err := gameRepo.GetWaitingPublicGame(ctx)

		// Then: an ErrNoActiveGames error should be returned and no game should be retrieved
		require.Error(t, err)
		require.Equal(t, apperror.ErrNoActiveGames, err)
		require.Nil(t, game)
	})
}
