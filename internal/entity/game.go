package entity

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
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

	EasyDifficulty       = "easy"
	HardDifficulty       = "hard"
	InvincibleDifficulty = "invincible"
)

var (
	ErrInvalidCell       = errors.New("invalid cell index")
	ErrUnknownGameStatus = errors.New("unknown game status")
	ErrBotNotFound       = errors.New("bot not found")
	ErrNoAvailableMoves  = errors.New("no available moves")

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
	ID         string    `json:"id"`
	Board      [9]string `json:"board"`
	Winner     string    `json:"winner"`
	Status     string    `json:"status"`
	Turn       string    `json:"player_turn"`
	Players    []*Player `json:"players,omitempty"`
	Type       string    `json:"type,omitempty"`
	Difficulty string    `json:"difficulty,omitempty"`
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

func (that *Game) BotMakeTurn() error {
	botPlayer := that.GetBotPlayer()
	if botPlayer == nil {
		return ErrBotNotFound
	}

	availableCells := that.getAvailableCells()
	if len(availableCells) == 0 {
		return ErrNoAvailableMoves
	}

	chosenCell := that.selectBotMove(botPlayer.Mark, availableCells)

	if err := that.MakeTurn(botPlayer.Mark, chosenCell); err != nil {
		return fmt.Errorf("bot failed to make turn: %w", err)
	}

	return nil
}

func (that *Game) GetBotPlayer() *Player {
	for _, player := range that.Players {
		if player.IsBot() {
			return player
		}
	}
	return nil
}

func (that *Game) getAvailableCells() []int {
	availableCells := []int{}
	for i, cell := range that.Board {
		if cell == EmptyCell {
			availableCells = append(availableCells, i)
		}
	}
	return availableCells
}

func (that *Game) selectBotMove(mark string, available []int) int {
	difficulty := that.Difficulty
	if difficulty == "" {
		difficulty = EasyDifficulty
	}

	switch difficulty {
	case EasyDifficulty:
		return that.easyStrategy(available)
	case HardDifficulty:
		if winMove := that.findWinningMove(mark); winMove != -1 {
			return winMove
		}

		oppMark := PlayerO
		if mark == PlayerO {
			oppMark = PlayerX
		}
		if blockMove := that.findWinningMove(oppMark); blockMove != -1 {
			return blockMove
		}

		return that.easyStrategy(available)
	case InvincibleDifficulty:
		if winMove := that.findWinningMove(mark); winMove != -1 {
			return winMove
		}

		oppMark := PlayerO
		if mark == PlayerO {
			oppMark = PlayerX
		}
		if blockMove := that.findWinningMove(oppMark); blockMove != -1 {
			return blockMove
		}

		if that.Board[4] == EmptyCell {
			return 4 // center
		}

		for _, corner := range []int{0, 2, 6, 8} {
			if that.Board[corner] == EmptyCell {
				return corner
			}
		}

		return that.easyStrategy(available)
	default:
		return that.easyStrategy(available)
	}
}

func (that *Game) easyStrategy(available []int) int {
	return available[rand.Intn(len(available))] //nolint:gosec // it`s ok
}

func (that *Game) findWinningMove(mark string) int {
	for _, combo := range WinCombos {
		a, b, c := that.Board[combo[0]], that.Board[combo[1]], that.Board[combo[2]]
		if a == mark && b == mark && c == EmptyCell {
			return combo[2]
		}
		if a == mark && c == mark && b == EmptyCell {
			return combo[1]
		}
		if b == mark && c == mark && a == EmptyCell {
			return combo[0]
		}
	}
	return -1
}
