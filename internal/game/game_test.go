package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	// When: create a new game instance
	game := NewGame("000")

	// Then: the game should have the expected initial state
	expectedGame := Game{
		Board:  [9]string{emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
		Turn:   playerX,
		Winner: "",
		Status: statusOngoing,
		Player: &Player{
			ID:   "000",
			Mark: "X",
		},
	}

	// Then: the game instance should not be nil
	require.NotNil(t, game)

	// Then: the game state should match the expected state
	require.Equal(t, expectedGame, *game)
}

func TestGame_MakeMove(t *testing.T) {
	t.Run("MakeMove", func(t *testing.T) {
		// Given: We have a new game
		game := NewGame("000")

		// When: Players make their moves
		err := game.MakeMove(playerX, 0)
		require.NoError(t, err)

		// Then: the game state should reflect the move and turn change
		expectedGame := Game{
			Board:  [9]string{playerX, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
			Turn:   playerO,
			Winner: "",
			Status: statusOngoing,
			Player: &Player{
				ID:   "000",
				Mark: "X",
			},
		}

		// Then: the game instance should not be nil
		require.NotNil(t, game)

		// Then: the game state should match the expected state
		require.Equal(t, expectedGame, *game)
	})

	t.Run("Error on cell already occupied", func(t *testing.T) {
		// Given: A new game instance
		game := NewGame("000")

		err := game.MakeMove(playerX, 0)
		require.NoError(t, err)

		// Then: the game state should be as expected
		expectedGame := Game{
			Board:  [9]string{playerX, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
			Turn:   playerO,
			Winner: "",
			Status: statusOngoing,
			Player: &Player{
				ID:   "000",
				Mark: "X",
			},
		}

		// When: player O tries to move to the same occupied cell
		err = game.MakeMove(playerO, 0)

		// Then: An ErrCellOccupied error should be returned
		require.ErrorIs(t, err, ErrCellOccupied)

		// Then: the game state should remain unchanged
		require.Equal(t, expectedGame, *game)
	})

	t.Run("Error on playing out of turn", func(t *testing.T) {
		// Given: A new game instance
		game := NewGame("000")

		// When: player O tries to make a move before player X
		err := game.MakeMove("O", 1)

		// Then: An ErrNotYourTurn error should be returned
		require.ErrorIs(t, err, ErrNotYourTurn)

		// Then: the game state should remain as it was
		expectedGame := Game{
			Board:  [9]string{emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell},
			Turn:   playerX,
			Winner: "",
			Status: statusOngoing,
			Player: &Player{
				ID:   "000",
				Mark: "X",
			},
		}

		require.Equal(t, expectedGame, *game)
	})

	t.Run("Invalid Cell", func(t *testing.T) {
		// Given: A new game instance
		game := NewGame("000")

		// When: an invalid cell index is passed (outside the board range)
		err := game.MakeMove(playerX, 20)

		// Then: ErrInvalidCell should be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})

	t.Run("Invalid Negative Cell", func(t *testing.T) {
		// Given: A new game instance
		game := NewGame("000")

		// When: A negative cell index is passed
		err := game.MakeMove(playerX, -1)

		// Then: ErrInvalidCell should be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})

	t.Run("Move After Game Finished", func(t *testing.T) {
		// Given: A game where X has already won
		game := NewGame("000")
		game.Board = [9]string{playerX, playerX, playerX, emptyCell, playerO, emptyCell, emptyCell, playerO, emptyCell}
		game.Status = statusFinished

		// When: player O tries to make a move after the game has finished
		err := game.MakeMove(playerO, 3)

		// Then: an ErrGameFinished error should be returned
		assert.ErrorIs(t, err, ErrGameFinished)
	})

	t.Run("Move After Tie", func(t *testing.T) {
		// Given: A game that ended in a tie
		game := NewGame("000")
		game.Board = [9]string{playerO, playerX, playerO, playerO, playerX, playerX, playerX, playerO, playerO}
		game.Status = statusFinished
		game.Winner = playerTie

		// When: player X tries to make a move after the tie
		err := game.MakeMove(playerX, 3)

		// Then: an ErrGameFinished error should be returned
		assert.ErrorIs(t, err, ErrGameFinished)
	})
}

func TestGame_checkGameStatus(t *testing.T) {
	t.Run("Winner X", func(t *testing.T) {
		// Given: a game where player X has a winning combination
		game := NewGame("000")

		game.Board = [9]string{playerX, playerO, emptyCell, playerX, playerO, emptyCell, playerX, emptyCell, emptyCell}

		// When: checking the game status
		status := checkGameStatus(game.Board)

		// Then: player X should be declared the winner
		require.Equal(t, playerX, status)
	})

	t.Run("Turn", func(t *testing.T) {
		// Given: a game where no player has won yet
		game := NewGame("000")
		game.Board = [9]string{playerX, playerO, playerX, emptyCell, playerO, emptyCell, playerX, emptyCell, emptyCell}

		// When: checking the game status
		status := checkGameStatus(game.Board)

		// Then: the game should still be ongoing (no winner)
		require.Equal(t, "", status)
	})

	t.Run("Tie", func(t *testing.T) {
		// Given: a game that ended in a tie
		game := NewGame("000")
		game.Board = [9]string{playerO, playerX, playerO, playerO, playerX, playerX, playerX, playerO, playerX}

		// When: checking the game status
		status := checkGameStatus(game.Board)

		// Then: the game should be declared a tie
		assert.Equal(t, playerTie, status)
	})
}
