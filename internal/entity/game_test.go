package entity

import (
	"testing"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameStatusMethods(t *testing.T) {
	t.Run("IsFinished returns true when game status is finished", func(t *testing.T) {
		// Given: a game with StatusFinished
		game := &Game{Status: StatusFinished}

		// When: checking if the game is finished
		isFinished := game.IsFinished()

		// Then: it should return true
		assert.True(t, isFinished)
	})

	t.Run("IsOngoing returns true when game status is ongoing", func(t *testing.T) {
		// Given: a game with StatusOngoing
		game := &Game{Status: StatusOngoing}

		// When: checking if the game is ongoing
		isOngoing := game.IsOngoing()

		// Then: it should return true
		assert.True(t, isOngoing)
	})

	t.Run("IsWaiting returns true when game status is waiting", func(t *testing.T) {
		// Given: a game with StatusWaiting
		game := &Game{Status: StatusWaiting}

		// When: checking if the game is waiting
		isWaiting := game.IsWaiting()

		// Then: it should return true
		assert.True(t, isWaiting)
	})
}

func TestGame_ConfirmOngoingState(t *testing.T) {
	t.Run("Returns nil when game is ongoing", func(t *testing.T) {
		// Given: a game with StatusOngoing
		game := &Game{Status: StatusOngoing}

		// When: checking if the game is active
		err := game.ConfirmOngoingState()

		// Then: it should return nil error
		assert.NoError(t, err)
	})

	t.Run("Returns ErrGameIsNotStarted when game is waiting", func(t *testing.T) {
		// Given: a game with StatusWaiting
		game := &Game{Status: StatusWaiting}

		// When: checking if the game is active
		err := game.ConfirmOngoingState()

		// Then: it should return ErrGameIsNotStarted
		assert.ErrorIs(t, err, apperror.ErrGameIsNotStarted)
	})

	t.Run("Returns ErrGameFinished when game is finished", func(t *testing.T) {
		// Given: a game with StatusFinished
		game := &Game{Status: StatusFinished}

		// When: checking if the game is active
		err := game.ConfirmOngoingState()

		// Then: it should return ErrGameFinished
		assert.ErrorIs(t, err, apperror.ErrGameFinished)
	})

	t.Run("Returns error for unknown game status", func(t *testing.T) {
		// Given: a game with unknown status
		game := &Game{Status: "unknown"}

		// When: checking if the game is active
		err := game.ConfirmOngoingState()

		// Then: it should return an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown game status")
	})
}

func TestGame_DetermineGameResult(t *testing.T) {
	t.Run("Returns PlayerX when Player X wins", func(t *testing.T) {
		// Given: a game where Player X has a winning combination
		game := &Game{
			Board: [9]string{
				PlayerX, PlayerX, PlayerX,
				EmptyCell, EmptyCell, EmptyCell,
				EmptyCell, EmptyCell, EmptyCell,
			},
		}

		// When: determining the game result
		result := game.DetermineGameResult()

		// Then: it should return PlayerX as the winner
		assert.Equal(t, PlayerX, result)
	})

	t.Run("Returns PlayerO when Player O wins", func(t *testing.T) {
		// Given: a game where Player O has a winning combination
		game := &Game{
			Board: [9]string{
				PlayerO, PlayerO, PlayerO,
				EmptyCell, EmptyCell, EmptyCell,
				EmptyCell, EmptyCell, EmptyCell,
			},
		}

		// When: determining the game result
		result := game.DetermineGameResult()

		// Then: it should return PlayerO as the winner
		assert.Equal(t, PlayerO, result)
	})

	t.Run("Returns PlayerTie when the game is a tie", func(t *testing.T) {
		// Given: a game that ended in a tie
		game := &Game{
			Board: [9]string{
				PlayerX, PlayerO, PlayerX,
				PlayerO, PlayerX, PlayerO,
				PlayerO, PlayerX, PlayerO,
			},
		}

		// When: determining the game result
		result := game.DetermineGameResult()

		// Then: it should return PlayerTie
		assert.Equal(t, PlayerTie, result)
	})

	t.Run("Returns EmptyCell when the game is ongoing", func(t *testing.T) {
		// Given: a game that is still ongoing
		game := &Game{
			Board: [9]string{
				PlayerX, PlayerO, EmptyCell,
				EmptyCell, PlayerX, EmptyCell,
				EmptyCell, EmptyCell, PlayerO,
			},
		}

		// When: determining the game result
		result := game.DetermineGameResult()

		// Then: it should return EmptyCell (game continues)
		assert.Equal(t, EmptyCell, result)
	})
}

func TestGame_UpdateGameState(t *testing.T) {
	t.Run("Updates game state when Player X wins", func(t *testing.T) {
		// Given: a game where Player X has a winning combination
		game := &Game{
			Board: [9]string{
				PlayerX, PlayerX, PlayerX,
				EmptyCell, EmptyCell, EmptyCell,
				EmptyCell, EmptyCell, EmptyCell,
			},
			Status: StatusOngoing,
			Turn:   PlayerO,
		}

		// When: updating the game state
		game.UpdateGameState()

		// Then: the game should be finished with Player X as the winner
		assert.Equal(t, StatusFinished, game.Status)
		assert.Equal(t, PlayerX, game.Winner)
		assert.Equal(t, EmptyCell, game.Turn)
	})

	t.Run("Updates game state when the game is a tie", func(t *testing.T) {
		// Given: a game that ended in a tie
		game := &Game{
			Board: [9]string{
				PlayerX, PlayerO, PlayerX,
				PlayerO, PlayerX, PlayerO,
				PlayerO, PlayerX, PlayerO,
			},
			Status: StatusOngoing,
			Turn:   PlayerX,
		}

		// When: updating the game state
		game.UpdateGameState()

		// Then: the game should be finished with a tie
		assert.Equal(t, StatusFinished, game.Status)
		assert.Equal(t, PlayerTie, game.Winner)
		assert.Equal(t, EmptyCell, game.Turn)
	})

	t.Run("Game remains ongoing when there is no winner or tie", func(t *testing.T) {
		// Given: a game that is still ongoing
		game := &Game{
			Board: [9]string{
				PlayerX, PlayerO, EmptyCell,
				EmptyCell, PlayerX, EmptyCell,
				EmptyCell, EmptyCell, PlayerO,
			},
			Status: StatusOngoing,
			Turn:   PlayerO,
		}

		// When: updating the game state
		game.UpdateGameState()

		// Then: the game should remain ongoing
		assert.Equal(t, StatusOngoing, game.Status)
		assert.Equal(t, "", game.Winner)
		assert.Equal(t, PlayerO, game.Turn)
	})
}

func TestGame_MakeTurn(t *testing.T) {
	t.Run("Successful Turn", func(t *testing.T) {
		// Given: A new game
		game := NewGame("123", PrivateType)
		game.Status = StatusOngoing

		// When: Player X makes a valid turn
		err := game.MakeTurn(PlayerX, 0)
		require.NoError(t, err)

		// Then: The game state should reflect the turn and player turn should switch
		expectedGame := &Game{
			ID:      "123",
			Board:   [9]string{PlayerX, "", "", "", "", "", "", "", ""},
			Turn:    PlayerO,
			Winner:  "",
			Status:  StatusOngoing,
			Players: nil,
			Type:    PrivateType,
		}

		require.Equal(t, expectedGame, game)
	})

	t.Run("Error on Cell Already Occupied", func(t *testing.T) {
		// Given: A game where cell 0 is occupied by Player X
		game := NewGame("123", PrivateType)
		game.Status = StatusOngoing
		err := game.MakeTurn(PlayerX, 0)
		require.NoError(t, err)

		// When: Player O tries to make a move to the same cell
		err = game.MakeTurn(PlayerO, 0)

		// Then: An ErrCellOccupied error should be returned
		require.ErrorIs(t, err, apperror.ErrCellOccupied)

		// And: The game state should remain unchanged
		expectedGame := &Game{
			ID:      "123",
			Board:   [9]string{PlayerX, "", "", "", "", "", "", "", ""},
			Turn:    PlayerO,
			Winner:  "",
			Status:  StatusOngoing,
			Players: nil,
			Type:    PrivateType,
		}

		require.Equal(t, expectedGame, game)
	})

	t.Run("Error on Playing Out of Turn", func(t *testing.T) {
		// Given: A new game where it's Player X's turn
		game := NewGame("123", PrivateType)
		game.Status = StatusOngoing

		// When: Player O tries to make a move
		err := game.MakeTurn(PlayerO, 1)

		// Then: An ErrNotYourTurn error should be returned
		require.ErrorIs(t, err, apperror.ErrNotYourTurn)

		// And: The game state should remain unchanged
		expectedGame := &Game{
			ID:      "123",
			Board:   [9]string{"", "", "", "", "", "", "", "", ""},
			Turn:    PlayerX,
			Winner:  "",
			Status:  StatusOngoing,
			Players: nil,
			Type:    PrivateType,
		}

		require.Equal(t, expectedGame, game)
	})

	t.Run("Error on Invalid Cell Index (Greater than Range)", func(t *testing.T) {
		// Given: A new game
		game := NewGame("123", PrivateType)
		game.Status = StatusOngoing

		// When: An invalid cell index is passed (greater than the range)
		err := game.MakeTurn(PlayerX, 20)

		// Then: An ErrInvalidCell error should be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})

	t.Run("Error on Invalid Cell Index (Negative)", func(t *testing.T) {
		// Given: A new game
		game := NewGame("123", PrivateType)
		game.Status = StatusOngoing

		// When: A negative cell index is passed
		err := game.MakeTurn(PlayerX, -1)

		// Then: An ErrInvalidCell error should be returned
		assert.ErrorIs(t, err, ErrInvalidCell)
	})
}
