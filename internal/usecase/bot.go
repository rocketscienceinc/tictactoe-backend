package usecase

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

const (
	difficultyEasy       = "easy"
	difficultyHard       = "hard"
	difficultyInvincible = "invincible"

	center = 4
)

var (
	ErrBotNotFound      = errors.New("bot player not found")
	ErrNoAvailableMoves = errors.New("no available moves")
)

type BotUseCase interface {
	MakeTurn(game *entity.Game) error
}

type botUseCase struct{}

func NewBotUseCase() BotUseCase {
	return &botUseCase{}
}

func (that *botUseCase) MakeTurn(game *entity.Game) error {
	var botPlayer *entity.Player
	for _, player := range game.Players {
		if player.IsBot() {
			botPlayer = player
			break
		}
	}

	if botPlayer == nil {
		return ErrBotNotFound
	}

	availableCells := that.getAvailableCells(game)
	if len(availableCells) == 0 {
		return ErrNoAvailableMoves
	}

	difficulty := game.Difficulty
	if difficulty == "" {
		difficulty = difficultyEasy
	}

	var chosenCell int
	switch difficulty {
	case difficultyEasy:
		chosenCell = that.easyStrategy(availableCells)
	case difficultyHard:
		chosenCell = that.hardStrategy(game, botPlayer.Mark, availableCells)
	case difficultyInvincible:
		chosenCell = that.invincibleStrategy(game, botPlayer.Mark)
	default:
		chosenCell = that.easyStrategy(availableCells)
	}

	if err := game.MakeTurn(botPlayer.Mark, chosenCell); err != nil {
		return fmt.Errorf("bot failed to make turn: %w", err)
	}

	return nil
}

func (that *botUseCase) getAvailableCells(game *entity.Game) []int {
	availableCells := []int{}
	for i, cell := range game.Board {
		if cell == entity.EmptyCell {
			availableCells = append(availableCells, i)
		}
	}
	return availableCells
}

func (that *botUseCase) easyStrategy(available []int) int {
	return available[rand.Intn(len(available))] //nolint:gosec // it's ok
}

func (that *botUseCase) hardStrategy(game *entity.Game, mark string, available []int) int {
	if winMove := that.findWinningMove(game, mark); winMove != -1 {
		return winMove
	}

	oppMark := entity.PlayerO
	if mark == entity.PlayerO {
		oppMark = entity.PlayerX
	}
	if blockMove := that.findWinningMove(game, oppMark); blockMove != -1 {
		return blockMove
	}

	return that.easyStrategy(available)
}

func (that *botUseCase) invincibleStrategy(game *entity.Game, mark string) int {
	if winMove := that.findWinningMove(game, mark); winMove != -1 {
		return winMove
	}

	oppMark := entity.PlayerO
	if mark == entity.PlayerO {
		oppMark = entity.PlayerX
	}
	if blockMove := that.findWinningMove(game, oppMark); blockMove != -1 {
		return blockMove
	}

	if game.Board[center] == entity.EmptyCell {
		return center
	}

	for _, corner := range []int{0, 2, 6, 8} {
		if game.Board[corner] == entity.EmptyCell {
			return corner
		}
	}

	availableCells := that.getAvailableCells(game)
	return that.easyStrategy(availableCells)
}

func (that *botUseCase) findWinningMove(game *entity.Game, mark string) int {
	for _, combo := range entity.WinCombos {
		a, b, c := game.Board[combo[0]], game.Board[combo[1]], game.Board[combo[2]]
		if a == mark && b == mark && c == entity.EmptyCell {
			return combo[2]
		}
		if a == mark && c == mark && b == entity.EmptyCell {
			return combo[1]
		}
		if b == mark && c == mark && a == entity.EmptyCell {
			return combo[0]
		}
	}
	return -1
}
