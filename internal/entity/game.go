package entity

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
)

const (
	StatusFinished = "finished"
	StatusOngoing  = "ongoing"
	StatusWaiting  = "waiting"

	PlayerX   = "X"
	PlayerO   = "O"
	PlayerTie = "-"

	EmptyCell = ""
)

const (
	PublicType  = "public"
	PrivateType = "private"
	WithBotType = "bot"
)

var (
	ErrInvalidCell       = errors.New("invalid cell index")
	ErrUnknownGameStatus = errors.New("unknown game status")

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
	ID      string    `json:"id"`
	Board   [9]string `json:"board"`
	Winner  string    `json:"winner"`
	Status  string    `json:"status"`
	Turn    string    `json:"player_turn"`
	Players []*Player `json:"players,omitempty"`
	Type    string    `json:"type,omitempty"`
}

func NewGame(id, gameType string) *Game {
	return &Game{
		ID:     id,
		Board:  [9]string{EmptyCell, EmptyCell, EmptyCell, EmptyCell, EmptyCell, EmptyCell, EmptyCell, EmptyCell, EmptyCell},
		Turn:   PlayerX,
		Status: StatusWaiting,
		Type:   gameType,
	}
}

func (that *Game) DetermineGameResult() string {
	for _, combo := range WinCombos {
		a, b, c := that.Board[combo[0]], that.Board[combo[1]], that.Board[combo[2]]
		if a != EmptyCell && a == b && b == c {
			return a
		}
	}

	// the game will continue until all the squares are full
	for _, cell := range that.Board {
		if cell == EmptyCell {
			return ""
		}
	}

	return PlayerTie
}

func (that *Game) UpdateGameState() {
	switch winner := that.DetermineGameResult(); winner {
	// one player wins
	case PlayerX, PlayerO:
		that.Winner = winner
		that.Status = StatusFinished
		that.Turn = ""
	// tie
	case PlayerTie:
		that.Winner = PlayerTie
		that.Status = StatusFinished
		that.Turn = ""
	// game continue
	default:
		that.Status = StatusOngoing
	}
}

func (that *Game) MakeTurn(playerMark string, cell int) error {
	if cell < 0 || cell >= len(that.Board) {
		return fmt.Errorf("%w: cell %d", ErrInvalidCell, cell)
	}

	if that.Turn != playerMark {
		return apperror.ErrNotYourTurn
	}

	if that.Board[cell] != EmptyCell {
		return apperror.ErrCellOccupied
	}

	that.Board[cell] = playerMark

	// It's simple logic for a game changing move
	if that.Turn == PlayerX {
		that.Turn = PlayerO
	} else {
		that.Turn = PlayerX
	}

	that.UpdateGameState()

	return nil
}

func (that *Game) IsFinished() bool {
	return that.Status == StatusFinished
}

func (that *Game) IsOngoing() bool {
	return that.Status == StatusOngoing
}

func (that *Game) IsWaiting() bool {
	return that.Status == StatusWaiting
}

func (that *Game) ConfirmOngoingState() error {
	switch {
	case that.IsWaiting():
		return apperror.ErrGameIsNotStarted
	case that.IsFinished():
		return apperror.ErrGameFinished
	case that.IsOngoing():
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownGameStatus, that.Status)
	}
}

func (that *Game) IsPublic() bool {
	return that.Type == PublicType
}

func (that *Game) IsWithBot() bool {
	return that.Type == WithBotType
}

func (that *Game) GetRandomMarks() (string, string) {
	if rand.Intn(2) == 0 { //nolint: gosec // it's ok
		return PlayerX, PlayerO
	}
	return PlayerO, PlayerX
}
