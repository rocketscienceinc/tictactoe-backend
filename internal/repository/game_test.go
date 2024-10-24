package repository

import (
	"testing"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/testing/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameRepository_CreateOrUpdate(t *testing.T) {
	ctx, st := suite.New(t)

	gameRepo := NewGameRepository(st.Storage)

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

		gameRepo := NewGameRepository(st.Storage)

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

		gameRepo := NewGameRepository(st.Storage)

		nonExistentGameID := "9999999"

		// When: GetByID is called with non-existent ID
		retrievedGame, err := gameRepo.GetByID(ctx, nonExistentGameID)

		// Then: an ErrGameNotFound error should be returned
		require.Error(t, err)
		assert.Equal(t, ErrGameNotFound, err)
		assert.Empty(t, retrievedGame.ID)
		assert.Empty(t, retrievedGame.Status)
	})
}

func TestGameRepository_DeleteByID(t *testing.T) {
	t.Run("DeleteByID_Success", func(t *testing.T) {
		ctx, st := suite.New(t)

		gameRepo := NewGameRepository(st.Storage)

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

		gameRepo := NewGameRepository(st.Storage)

		// Given: a non-existent game ID
		nonExistentGameID := "9999999"

		// When: DeleteByID is called with non-existent ID
		err := gameRepo.DeleteByID(ctx, nonExistentGameID)

		// Then: an ErrGameNotFound error should be returned
		require.Error(t, err)
		require.Equal(t, ErrGameNotFound, err)
	})
}
