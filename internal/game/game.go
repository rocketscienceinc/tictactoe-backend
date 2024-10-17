package game

import (
	"errors"
	"fmt"
	"github.com/rocketscienceinc/tittactoe-backend/internal/app_error"

	"github.com/rocketscienceinc/tittactoe-backend/entity"
)

const (
	playerTie = "-"
	playerX   = "X"
	playerO   = "O"

	emptyCell = ""
)

var (
	ErrCellOccupied = errors.New("cell is already occupied")
	ErrNotYourTurn  = errors.New("it's not your turn")
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

type Game interface {
	Create(id string) *entity.Game
	MakeTurn(player string, cell int) error
}

type game struct {
	gameInstance *entity.Game
}

func New(gameInstance *entity.Game) Game {
	return &game{
		gameInstance: gameInstance,
	}
}

func (that *game) Create(id string) *entity.Game {
	that.gameInstance.ID = id
	that.gameInstance.Board = [9]string{emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell}
	that.gameInstance.PlayerTurn = playerX
	that.gameInstance.Status = entity.StatusWaiting

	return that.gameInstance
}

func (that *game) MakeTurn(player string, cell int) error {
	if that.gameInstance.IsFinished() {
		return app_error.ErrGameFinished
	}

	if err := validateMove(that.gameInstance, player, cell); err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}

	that.gameInstance.Board[cell] = player
	that.updateGameStatus(player)

	return nil
}

// validateMove - checks if the move is valid.
func validateMove(game *entity.Game, playerTurn string, cell int) error {
	if cell < 0 || cell >= len(game.Board) {
		return ErrInvalidCell
	}

	if game.Board[cell] != emptyCell {
		return ErrCellOccupied
	}

	if game.PlayerTurn != playerTurn {
		return ErrNotYourTurn
	}

	return nil
}

// updateGameStatus - checks the game status after a move.
func (that *game) updateGameStatus(player string) {
	switch winner := checkGameStatus(that.gameInstance.Board); winner {
	case playerX, playerO:
		that.gameInstance.Winner = winner
		that.gameInstance.Status = entity.StatusFinished
	case playerTie:
		that.gameInstance.Winner = playerTie
		that.gameInstance.Status = entity.StatusFinished
	default:
		that.gameInstance.PlayerTurn = toggleMark(player)
	}
}

func toggleMark(currentMark string) string {
	if currentMark == playerX {
		return playerO
	}
	return playerX
}

func checkGameStatus(board [9]string) string {
	for _, combo := range WinCombos {
		a, b, c := board[combo[0]], board[combo[1]], board[combo[2]]
		if a != emptyCell && a == b && b == c {
			return a
		}
	}

	for _, cell := range board {
		if cell == emptyCell {
			return ""
		}
	}

	return playerTie
}
