package game

import (
	"testing"

	"github.com/rocketscienceinc/tittactoe-backend/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	// When: create a new game instance
	newGame := New(entity.Game{})

	actualGame := newGame.Create("123")
	actualGame.Status = statusOngoing

	// Then: the game should have the expected initial state
	expectedGame := entity.Game{
		ID:         "123",
		Board:      [9]string{emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
		PlayerTurn: playerX,
		Winner:     "",
		Status:     statusOngoing,
	}

	// Then: the game instance should not be nil
	require.NotNil(t, newGame)

	// Then: the game state should match the expected state
	require.Equal(t, &expectedGame, actualGame)
}

func TestGame_MakeMove(t *testing.T) {
	t.Run("MakeMove", func(t *testing.T) {
		// Given: We have a new game
		newGame := New(entity.Game{})

		actualGame := newGame.Create("123")
		actualGame.Status = statusOngoing

		// When: Players make their moves
		err := newGame.MakeMove(playerX, 0)
		require.NoError(t, err)

		// Then: the game state should reflect the move and turn change
		expectedGame := entity.Game{
			ID:         "123",
			Board:      [9]string{playerX, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
			PlayerTurn: playerO,
			Winner:     "",
			Status:     statusOngoing,
		}

		// Then: the game instance should not be nil
		require.NotNil(t, newGame)

		// Then: the game state should match the expected state
		require.Equal(t, &expectedGame, actualGame)
	})

	t.Run("Error on cell already occupied", func(t *testing.T) {
		// Given: A new game instance
		newGame := New(entity.Game{
			PlayerTurn: playerX,
		})

		actualGame := newGame.Create("123")
		actualGame.Status = statusOngoing

		err := newGame.MakeMove(playerX, 0)
		require.NoError(t, err)

		// Then: the game state should be as expected
		expectedGame := entity.Game{
			ID:         "123",
			Board:      [9]string{playerX, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
			PlayerTurn: playerO,
			Winner:     "",
			Status:     statusOngoing,
		}

		// When: player O tries to move to the same occupied cell
		err = newGame.MakeMove(playerO, 0)

		// Then: An ErrCellOccupied error should be returned
		require.ErrorIs(t, err, ErrCellOccupied)

		// Then: the game state should remain unchanged
		require.Equal(t, &expectedGame, actualGame)
	})

	t.Run("Error on playing out of turn", func(t *testing.T) {
		// Given: A new game instance
		newGame := New(entity.Game{})

		actualGame := newGame.Create("123")
		actualGame.Status = statusOngoing

		// When: player O tries to make a move before player X
		err := newGame.MakeMove("O", 1)

		// Then: An ErrNotYourTurn error should be returned
		require.ErrorIs(t, err, ErrNotYourTurn)

		// Then: the game state should remain as it was
		expectedGame := entity.Game{
			ID:         "123",
			Board:      [9]string{emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
			PlayerTurn: playerX,
			Winner:     "",
			Status:     statusOngoing,
		}

		require.Equal(t, &expectedGame, actualGame)
	})

	t.Run("Invalid Cell", func(t *testing.T) {
		// Given: A new game instance
		newGame := New(entity.Game{})

		// When: an invalid cell index is passed (outside the board range)
		err := newGame.MakeMove(playerX, 20)

		// Then: ErrInvalidCell should be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})

	t.Run("Invalid Negative Cell", func(t *testing.T) {
		// Given: A new game instance
		newGame := New(entity.Game{})

		// When: A negative cell index is passed
		err := newGame.MakeMove(playerX, -1)

		// Then: ErrInvalidCell should be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})

	t.Run("Move After Game Finished", func(t *testing.T) {
		// Given: A game where X has already won
		newGame := New(entity.Game{
			Board:      [9]string{playerX, playerX, playerX, emptyCell, playerO, emptyCell, emptyCell, playerO, emptyCell},
			Status:     statusFinished,
			PlayerTurn: playerO,
		})

		// When: player O tries to make a move after the game has finished
		err := newGame.MakeMove(playerO, 3)

		// Then: an ErrGameFinished error should be returned
		assert.ErrorIs(t, err, ErrGameFinished)
	})

	t.Run("Move After Tie", func(t *testing.T) {
		// Given: A game that ended in a tie
		newGame := New(entity.Game{
			Board:  [9]string{playerO, playerX, playerO, playerO, playerX, playerX, playerX, playerO, playerO},
			Status: statusFinished,
			Winner: playerTie,
		})

		// When: player X tries to make a move after the tie
		err := newGame.MakeMove(newGame, playerX, 3)

		// Then: an ErrGameFinished error should be returned
		assert.ErrorIs(t, err, ErrGameFinished)
	})
}

func TestGame_checkGameStatus(t *testing.T) {
	t.Run("Winner X", func(t *testing.T) {
		// Given: a game where player X has a winning combination
		newGame := New(entity.Game{})

		actualGame := newGame.Create("123")
		actualGame.Status = statusOngoing
		actualGame.Board = [9]string{playerX, playerO, emptyCell, playerX, playerO, emptyCell, playerX, emptyCell, emptyCell}

		// When: checking the game status
		status := checkGameStatus(actualGame.Board)

		// Then: player X should be declared the winner
		require.Equal(t, playerX, status)
	})

	t.Run("Turn", func(t *testing.T) {
		// Given: a game where no player has won yet
		newGame := New(entity.Game{})

		actualGame := newGame.Create("123")
		actualGame.Status = statusOngoing
		actualGame.Board = [9]string{playerX, playerO, playerX, emptyCell, playerO, emptyCell, playerX, emptyCell, emptyCell}

		// When: checking the game status
		status := checkGameStatus(actualGame.Board)

		// Then: the game should still be ongoing (no winner)
		require.Equal(t, "", status)
	})

	t.Run("Tie", func(t *testing.T) {
		// Given: a game that ended in a tie
		newGame := New(entity.Game{})

		actualGame := newGame.Create("123")
		actualGame.Status = statusOngoing
		actualGame.Board = [9]string{playerO, playerX, playerO, playerO, playerX, playerX, playerX, playerO, playerX}

		// When: checking the game status
		status := checkGameStatus(actualGame.Board)

		// Then: the game should be declared a tie
		assert.Equal(t, playerTie, status)
	})
}
