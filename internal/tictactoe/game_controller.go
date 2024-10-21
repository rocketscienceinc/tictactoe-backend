package tictactoe

import (
	"errors"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

var (
	ErrCellOccupied = errors.New("cell is already occupied")
	ErrInvalidCell  = errors.New("invalid cell index")

	WinCombos = [][3]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}
)

func MakeTurn(gameInstance *entity.Game, player string, cell int) error {
	if gameInstance.IsFinished() {
		return apperror.ErrGameFinished
	}

	if err := validateMove(gameInstance, player, cell); err != nil {
		return fmt.Errorf("invalid turn: %w", err)
	}

	gameInstance.Board[cell] = player
	updateGameStatus(gameInstance, player)

	return nil
}

// validateMove - checks if the move is valid.
func validateMove(gameInstance *entity.Game, playerTurn string, cell int) error {
	if cell < 0 || cell >= len(gameInstance.Board) {
		return ErrInvalidCell
	}

	if gameInstance.PlayerTurn != playerTurn {
		return apperror.ErrNotYourTurn
	}

	if gameInstance.Board[cell] != entity.EmptyCell {
		return ErrCellOccupied
	}

	return nil
}

// updateGameStatus - checks the game status after a move.
func updateGameStatus(gameInstance *entity.Game, player string) {
	switch winner := checkGameStatus(gameInstance.Board); winner {
	case entity.PlayerX, entity.PlayerO:
		gameInstance.Winner = winner
		gameInstance.Status = entity.StatusFinished
	case entity.PlayerTie:
		gameInstance.Winner = entity.PlayerTie
		gameInstance.Status = entity.StatusFinished
	default:
		gameInstance.PlayerTurn = toggleMark(player)
	}
}

func toggleMark(currentMark string) string {
	if currentMark == entity.PlayerX {
		return entity.PlayerO
	}
	return entity.PlayerX
}

func checkGameStatus(board [9]string) string {
	for _, combo := range WinCombos {
		a, b, c := board[combo[0]], board[combo[1]], board[combo[2]]
		if a != entity.EmptyCell && a == b && b == c {
			return a
		}
	}

	for _, cell := range board {
		if cell == entity.EmptyCell {
			return ""
		}
	}

	return entity.PlayerTie
}
