package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
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

func TestGame_BotMakeTurn_EasyDifficulty(t *testing.T) {
	t.Run("Bot makes a move in an available cell", func(t *testing.T) {
		// Given: a new game with easy difficulty and a human player
		game := NewGame("test-game-easy", WithBotType)
		game.Difficulty = EasyDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Simulate player's move
		err = game.MakeTurn(PlayerX, 0) // Player X moves at cell 0
		require.NoError(t, err)

		// Bot makes a move
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify that the bot has made a move (any available cell except 0)
		board := game.Board
		moveMade := false
		for i, cell := range board {
			if i != 0 && cell == PlayerO {
				moveMade = true
				break
			}
		}
		assert.True(t, moveMade, "Bot should have made a move on an available cell")
	})
}

func TestGame_BotMakeTurn_HardDifficulty(t *testing.T) {
	t.Run("Bot makes a winning move when possible", func(t *testing.T) {
		// Given: a game where bot can win
		game := NewGame("test-game-hard-win", WithBotType)
		game.Difficulty = HardDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Set up board where bot can win by placing at cell 5
		game.Board = [9]string{
			PlayerX, PlayerX, EmptyCell,
			PlayerO, PlayerO, EmptyCell,
			EmptyCell, EmptyCell, EmptyCell,
		}
		game.Turn = PlayerO

		// Bot makes a move
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify bot made the winning move at cell 5
		assert.Equal(t, PlayerO, game.Board[5], "Bot should have placed at cell 5 to win")
		assert.Equal(t, StatusFinished, game.Status, "Game should be finished after bot wins")
		assert.Equal(t, PlayerO, game.Winner, "Bot should be the winner")
	})

	t.Run("Bot blocks the player's winning move", func(t *testing.T) {
		// Given: a game where the player can win in the next move
		game := NewGame("test-game-hard-block", WithBotType)
		game.Difficulty = HardDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Set up board where player can win by placing at cell 2
		game.Board = [9]string{
			PlayerX, PlayerX, EmptyCell,
			PlayerO, EmptyCell, EmptyCell,
			EmptyCell, EmptyCell, EmptyCell,
		}
		game.Turn = PlayerO

		// Bot makes a move
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify bot blocked at cell 2
		assert.Equal(t, PlayerO, game.Board[2], "Bot should have blocked at cell 2")
		assert.Equal(t, StatusOngoing, game.Status, "Game should continue after bot blocks")
	})
}

func TestGame_BotMakeTurn_InvincibleDifficulty(t *testing.T) {
	t.Run("Bot takes the center if available", func(t *testing.T) {
		// Given: a game where center is available
		game := NewGame("test-game-invincible-center", WithBotType)
		game.Difficulty = InvincibleDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Set up board with center available
		game.Board = [9]string{
			PlayerX, EmptyCell, EmptyCell,
			EmptyCell, EmptyCell, EmptyCell,
			EmptyCell, EmptyCell, EmptyCell,
		}
		game.Turn = PlayerO

		// Bot makes a move
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify bot took the center cell (4)
		assert.Equal(t, PlayerO, game.Board[4], "Bot should have taken the center cell (4)")
		assert.Equal(t, PlayerX, game.Turn, "Turn should switch back to PlayerX")
	})

	t.Run("Bot takes a corner if center is occupied", func(t *testing.T) {
		// Given: a game where center is occupied
		game := NewGame("test-game-invincible-corner", WithBotType)
		game.Difficulty = InvincibleDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Set up board with center occupied by bot
		game.Board = [9]string{
			PlayerX, EmptyCell, EmptyCell,
			EmptyCell, PlayerO, EmptyCell,
			EmptyCell, EmptyCell, EmptyCell,
		}
		game.Turn = PlayerO

		// Bot makes a move
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify bot took one of the corner cells (0, 2, 6, 8)
		cornerCells := []int{0, 2, 6, 8}
		moveMade := false
		for _, cell := range cornerCells {
			if game.Board[cell] == PlayerO {
				moveMade = true
				break
			}
		}
		assert.True(t, moveMade, "Bot should have taken one of the corner cells")
		assert.Equal(t, PlayerX, game.Turn, "Turn should switch back to PlayerX")
	})

	t.Run("Bot prevents player from creating a fork", func(t *testing.T) {
		// Given: a game where player is attempting to create a fork
		game := NewGame("test-game-invincible-prevent-fork", WithBotType)
		game.Difficulty = InvincibleDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Set up board where player is trying to create a fork
		game.Board = [9]string{
			PlayerX, EmptyCell, EmptyCell,
			EmptyCell, PlayerO, EmptyCell,
			EmptyCell, EmptyCell, PlayerX,
		}
		game.Turn = PlayerO

		// Bot makes a move to block the fork
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify bot took cell 2 to block the fork
		assert.Equal(t, PlayerO, game.Board[2], "Bot should have placed at cell 2 to prevent fork")
		assert.Equal(t, StatusOngoing, game.Status, "Game should continue after bot's move")
	})

	t.Run("Bot makes a winning move when possible", func(t *testing.T) {
		// Given: a game where bot can win
		game := NewGame("test-game-invincible-win", WithBotType)
		game.Difficulty = InvincibleDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Set up board where bot can win by placing at cell 2
		game.Board = [9]string{
			PlayerO, PlayerO, EmptyCell,
			PlayerX, PlayerX, EmptyCell,
			EmptyCell, EmptyCell, EmptyCell,
		}
		game.Turn = PlayerO

		// Bot makes a move
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify bot made the winning move at cell 2
		assert.Equal(t, PlayerO, game.Board[2], "Bot should have placed at cell 2 to win")
		assert.Equal(t, StatusFinished, game.Status, "Game should be finished after bot's winning move")
		assert.Equal(t, PlayerO, game.Winner, "Bot should be the winner")
	})
}

func TestGame_BotMakeTurn_EasyDifficulty_FullBoard(t *testing.T) {
	t.Run("Bot makes the last move and game ends in a tie", func(t *testing.T) {
		// Given: a game with one empty cell remaining
		game := NewGame("test-game-easy-full-board", WithBotType)
		game.Difficulty = EasyDifficulty

		player := &Player{ID: "player1", Mark: PlayerX, GameID: game.ID}
		err := game.AddPlayer(player)
		require.NoError(t, err)

		// Add bot player with mark PlayerO
		err = addBotPlayer(game)
		require.NoError(t, err)

		// Set up board with one empty cell at position 8
		game.Board = [9]string{
			PlayerX, PlayerO, PlayerX,
			PlayerO, PlayerX, PlayerO,
			PlayerO, PlayerX, EmptyCell,
		}
		game.Turn = PlayerO

		// Bot makes a move
		err = game.BotMakeTurn()
		require.NoError(t, err)

		// Verify bot placed at cell 8, game is finished with a tie
		assert.Equal(t, PlayerO, game.Board[8], "Bot should have placed at the last available cell (8)")
		assert.Equal(t, StatusFinished, game.Status, "Game should be finished after the last move")
		assert.Equal(t, PlayerTie, game.Winner, "Game should be a tie")
	})
}

func TestGame_BotMakeTurn_HardDifficulty_AlreadyFinished(t *testing.T) {
	t.Run("Bot does not make a move if the game is already finished", func(t *testing.T) {
		// Given: a finished game with Player X as the winner
		game := NewGame("test-game-hard-already-finished", WithBotType)
		game.Difficulty = HardDifficulty
		game.Status = StatusFinished
		game.Winner = PlayerX
		game.Turn = "" // Ensure Turn is empty when game is finished

		// Add bot player with mark PlayerO
		err := addBotPlayer(game)
		require.NoError(t, err)

		// Bot attempts to make a move
		err = game.BotMakeTurn()
		require.Error(t, err, "Bot should not make a move if the game is already finished")
		assert.Contains(t, err.Error(), "bot failed to make turn")
	})
}

func TestGame_BotMakeTurn_InvincibleDifficulty_CannotMoveAfterTie(t *testing.T) {
	t.Run("Bot does not make a move after the game ends in a tie", func(t *testing.T) {
		// Given: a game that ended in a tie
		game := NewGame("test-game-invincible-tie", WithBotType)
		game.Difficulty = InvincibleDifficulty
		game.Status = StatusFinished
		game.Winner = PlayerTie
		game.Turn = "" // Ensure Turn is empty when game is finished

		// Add bot player with mark PlayerO
		err := addBotPlayer(game)
		require.NoError(t, err)

		// Bot attempts to make a move
		err = game.BotMakeTurn()
		require.Error(t, err, "Bot should not make a move after the game ends in a tie")
		assert.Contains(t, err.Error(), "bot failed to make turn")
	})
}

// Helper function to add a bot player to the game.
func addBotPlayer(game *Game) error {
	botPlayer := NewBotPlayer(game.ID, PlayerO)
	return game.AddPlayer(botPlayer)
}
