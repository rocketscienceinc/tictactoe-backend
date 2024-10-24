package repository

import (
	"testing"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/testing/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlayerRepository_CreateOrUpdate(t *testing.T) {
	ctx, st := suite.New(t)

	playerRepo := NewPlayerRepository(st.Storage)

	// Given: a player with ID
	player := &entity.Player{
		ID: "123",
	}

	// When: CreateOrUpdate is called
	err := playerRepo.CreateOrUpdate(ctx, player)

	// Then: no error should be returned, and player is stored
	require.NoError(t, err)
}

func TestPlayerRepository_GetByID(t *testing.T) {
	t.Run("GetByID_Success", func(t *testing.T) {
		ctx, st := suite.New(t)

		playerRepo := NewPlayerRepository(st.Storage)

		// Given: a player with ID
		player := &entity.Player{
			ID: "123",
		}

		err := playerRepo.CreateOrUpdate(ctx, player)
		require.NoError(t, err)

		// When: GetByID is called with existing ID
		retrievedGame, err := playerRepo.GetByID(ctx, player.ID)

		// Then: the retrieved player should match the saved player
		require.NoError(t, err)
		require.Equal(t, player.ID, retrievedGame.ID)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		ctx, st := suite.New(t)

		playerRepo := NewPlayerRepository(st.Storage)

		nonExistentPlayerID := "9999999"

		// When: GetByID is called with non-existent ID
		retrievedPlayer, err := playerRepo.GetByID(ctx, nonExistentPlayerID)

		// Then: an ErrPlayerNotFound error should be returned
		require.Error(t, err)
		assert.Equal(t, ErrPlayerNotFound, err)
		assert.Nil(t, retrievedPlayer)
	})
}
