package tictactoe

import (
	"testing"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	// Given: create a new game
	gameInstance := &entity.Game{}
	actualGame := gameInstance.Create("123")

	// Then: the game state should correspond to the expected initial state
	expectedGame := &entity.Game{
		ID:         "123",
		Board:      [9]string{"", "", "", "", "", "", "", "", ""},
		PlayerTurn: entity.PlayerX,
		Winner:     "",
		Status:     entity.StatusWaiting,
	}

	require.Equal(t, expectedGame, actualGame)
}

func TestGame_MakeTurn(t *testing.T) {
	t.Run("MakeTurn", func(t *testing.T) {
		// Given: create a new game
		gameInstance := &entity.Game{}
		game := gameInstance.Create("123")
		game.Status = entity.StatusOngoing

		// When: player X makes a turn
		err := MakeTurn(game, entity.PlayerX, 0)
		require.NoError(t, err)

		// Then: the game state should reflect the turn and queue change
		expectedGame := &entity.Game{
			ID:         "123",
			Board:      [9]string{entity.PlayerX, "", "", "", "", "", "", "", ""},
			PlayerTurn: entity.PlayerO,
			Winner:     "",
			Status:     entity.StatusOngoing,
		}

		require.Equal(t, expectedGame, game)
	})

	t.Run("Error on cell already occupied", func(t *testing.T) {
		// Given: new game with player X's queue
		gameInstance := &entity.Game{}
		game := gameInstance.Create("123")
		game.Status = entity.StatusOngoing

		// When: player X moves to cell 0
		err := MakeTurn(game, entity.PlayerX, 0)
		require.NoError(t, err)

		// When: player O tries to make a move to the same square
		err = MakeTurn(game, entity.PlayerO, 0)

		// Then: an error ErrCellOccupied must be returned
		require.ErrorIs(t, err, ErrCellOccupied)

		// Then: game state updated
		expectedGame := &entity.Game{
			ID:         "123",
			Board:      [9]string{entity.PlayerX, "", "", "", "", "", "", "", ""},
			PlayerTurn: entity.PlayerO,
			Winner:     "",
			Status:     entity.StatusOngoing,
		}

		// Then: the game state remains unchanged
		require.Equal(t, expectedGame, game)
	})

	t.Run("Error on playing out of turn", func(t *testing.T) {
		// Given: a new game
		gameInstance := &entity.Game{}
		game := gameInstance.Create("123")
		game.Status = entity.StatusOngoing

		// When: player O tries to make a move when it is player X's turn
		err := MakeTurn(game, entity.PlayerO, 1)

		// Then: an error ErrNotYourTurn must be returned
		require.ErrorIs(t, err, apperror.ErrNotYourTurn)

		// Then: the game state remains unchanged
		expectedGame := &entity.Game{
			ID:         "123",
			Board:      [9]string{"", "", "", "", "", "", "", "", ""},
			PlayerTurn: entity.PlayerX,
			Winner:     "",
			Status:     entity.StatusOngoing,
		}

		require.Equal(t, expectedGame, game)
	})

	t.Run("Invalid Cell", func(t *testing.T) {
		// Given: a new game
		gameInstance := &entity.Game{}
		game := gameInstance.Create("123")
		game.Status = entity.StatusOngoing

		// When: an invalid cell index is passed (greater than the range)
		err := MakeTurn(game, entity.PlayerX, 20)

		// Then: an error ErrInvalidCell must be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})

	t.Run("Invalid Negative Cell", func(t *testing.T) {
		// Given: a new game
		gameInstance := &entity.Game{}
		game := gameInstance.Create("123")
		game.Status = entity.StatusOngoing

		// When: negative cell index is transmitted
		err := MakeTurn(game, entity.PlayerX, -1)

		// Then: an error ErrInvalidCell must be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})

	t.Run("Move After Game Finished", func(t *testing.T) {
		// Given: a game where player X has already won
		gameInstance := &entity.Game{
			Board:      [9]string{entity.PlayerX, entity.PlayerX, entity.PlayerX, "", entity.PlayerO, "", "", entity.PlayerO, ""},
			Status:     entity.StatusFinished,
			PlayerTurn: entity.PlayerO,
		}

		// When: player O tries to make a move after the game is over
		err := MakeTurn(gameInstance, entity.PlayerO, 3)

		// Then: an error app_error.ErrGameFinished should be returned.
		assert.ErrorIs(t, err, apperror.ErrGameFinished)
	})

	t.Run("Move After Tie", func(t *testing.T) {
		// Given: a game that ended in a draw
		gameInstance := &entity.Game{
			Board:  [9]string{entity.PlayerO, entity.PlayerX, entity.PlayerO, entity.PlayerO, entity.PlayerX, entity.PlayerX, entity.PlayerX, entity.PlayerO, entity.PlayerO},
			Status: entity.StatusFinished,
			Winner: "-",
		}

		// When: player X tries to make a move after a draw
		err := MakeTurn(gameInstance, entity.PlayerX, 3)

		// Then: an error app_error.ErrGameFinished should be returned.
		assert.ErrorIs(t, err, apperror.ErrGameFinished)
	})
}

func TestGame_checkGameStatus(t *testing.T) {
	t.Run("Winner X", func(t *testing.T) {
		// Given: a game where player X has a winning combination
		board := [9]string{entity.PlayerX, entity.PlayerO, "", entity.PlayerX, entity.PlayerO, "", entity.PlayerX, "", ""}

		// When: check the game status
		status := checkGameStatus(board)

		// Then: player X should be declared the winner
		require.Equal(t, entity.PlayerX, status)
	})

	t.Run("Ongoing Game", func(t *testing.T) {
		// Given: a game where there is no winner yet
		board := [9]string{entity.PlayerX, entity.PlayerO, entity.PlayerX, "", entity.PlayerO, "", entity.PlayerX, "", ""}

		// When: check the game status
		status := checkGameStatus(board)

		// Then: the game should continue (no winner)
		require.Equal(t, "", status)
	})

	t.Run("Tie", func(t *testing.T) {
		// Given: a game that ended in a tie
		board := [9]string{entity.PlayerO, entity.PlayerX, entity.PlayerO, entity.PlayerO, entity.PlayerX, entity.PlayerX, entity.PlayerX, entity.PlayerO, entity.PlayerX}

		// When: check the game status
		status := checkGameStatus(board)

		// Then: the game should be declared a tie
		assert.Equal(t, entity.PlayerTie, status)
	})
}
