package game

import (
	"errors"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/entity"
)

const (
	statusFinished = "finished"
	statusOngoing  = "ongoing"
	statusWaiting  = "waiting"

	playerTie = "-"
	playerX   = "X"
	playerO   = "O"

	emptyCell = ""
)

var (
	ErrCellOccupied = errors.New("cell is already occupied")
	ErrNotYourTurn  = errors.New("it's not your turn")
	ErrGameFinished = errors.New("game is already finished")
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
	MakeMove(game *entity.Game, player string, cell int) (*entity.Game, error)
}

type game struct {
	gameInstance entity.Game
}

func New(gameInstance entity.Game) Game {
	return &game{
		gameInstance: gameInstance,
	}
}

func (that *game) Create(id string) *entity.Game {
	that.gameInstance.ID = id
	that.gameInstance.Board = [9]string{emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell, emptyCell}
	that.gameInstance.PlayerTurn = playerX
	that.gameInstance.Status = statusWaiting

	return &that.gameInstance
}

func (that *game) MakeMove(game *entity.Game, player string, cell int) (*entity.Game, error) {
	if game.Status == statusFinished {
		return nil, ErrGameFinished
	}

	if err := validateMove(game, player, cell); err != nil {
		return nil, fmt.Errorf("invalid move: %w", err)
	}

	game.Board[cell] = player
	updateGameStatus(game, player)

	return game, nil
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
func updateGameStatus(game *entity.Game, player string) {
	switch winner := checkGameStatus(game.Board); winner {
	case playerX, playerO:
		game.Winner = winner
		game.Status = statusFinished
	case playerTie:
		game.Winner = playerTie
		game.Status = statusFinished
	default:
		game.PlayerTurn = toggleMark(player)
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
