package socket

import (
	"errors"
	"fmt"
)

var (
	ErrGameNotFound          = errors.New("game not found")
	ErrGameAlreadyHasPlayers = errors.New("game already has players")
)

func NewGame(gameID string) *Game {
	return &Game{
		ID:     gameID,
		Board:  [9]string{"", "", "", "", "", "", "", "", ""},
		Turn:   "X",
		Status: "waiting",
	}
}

func (that *Game) JoinPlayer(player Player) error {
	if len(that.Players) >= 2 {
		return fmt.Errorf("%w : %d players", ErrGameAlreadyHasPlayers, len(that.Players))
	}

	that.Players = append(that.Players, player)

	if len(that.Players) == 2 {
		that.Status = "ongoing"
	}

	return nil
}

var games = make(map[string]*Game)

func FindGameByID(gameID string) (*Game, error) {
	if game, exists := games[gameID]; exists {
		return game, nil
	}

	return nil, ErrGameNotFound
}

func checkGameStatus(board [9]string) (string, bool) {
	winConditions := [][3]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}

	for _, condition := range winConditions {
		a, b, c := board[condition[0]], board[condition[1]], board[condition[2]]

		if a != "" && a == b && b == c {
			return a, false
		}
	}

	isFull := true
	for _, cell := range board {
		if cell == "" {
			isFull = false
			break
		}
	}

	return "", isFull
}

func toggleMark(currentMark string) string {
	if currentMark == "X" {
		return "O"
	}
	return "X"
}
