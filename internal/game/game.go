package game

import "errors"

const (
	statusFinished = "finished"
	statusOngoing  = "ongoing"

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

type Game struct {
	Board  [9]string
	Turn   string
	Winner string
	Status string
}

func NewGame() *Game {
	return &Game{
		Turn:   playerX,
		Status: statusOngoing,
		Board:  [9]string{},
		Winner: "",
	}
}

func (that *Game) MakeMove(player string, cell int) error {
	if that.Status == statusFinished {
		return ErrGameFinished
	}

	if cell < 0 || cell >= len(that.Board) {
		return ErrInvalidCell
	}

	if that.Board[cell] != emptyCell {
		return ErrCellOccupied
	}

	if that.Turn != player {
		return ErrNotYourTurn
	}

	that.Board[cell] = player

	switch winner := checkGameStatus(that.Board); winner {
	case playerX, playerO:
		that.Winner = winner
		that.Status = statusFinished
	case playerTie:
		that.Winner = playerTie
		that.Status = statusFinished
	default:
		that.Turn = toggleMark(player)
	}

	return nil
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
