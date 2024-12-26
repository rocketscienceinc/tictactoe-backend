package service

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
)

var (
	ErrBotNotFound      = errors.New("bot player not found")
	ErrNoAvailableMoves = errors.New("no available moves")
)

type BotService interface {
	MakeTurn(game *entity.Game) error
}

type botService struct{}

func NewBotService() BotService {
	return &botService{}
}

func (that *botService) MakeTurn(game *entity.Game) error {
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

func (that *botService) getAvailableCells(game *entity.Game) []int {
	availableCells := []int{}
	for i, cell := range game.Board {
		if cell == entity.EmptyCell {
			availableCells = append(availableCells, i)
		}
	}
	return availableCells
}

func (that *botService) easyStrategy(available []int) int {
	return available[rand.Intn(len(available))] //nolint:gosec // it's ok
}

func (that *botService) hardStrategy(game *entity.Game, mark string, available []int) int {
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

func (that *botService) invincibleStrategy(game *entity.Game, mark string) int {
	return that.hardStrategy(game, mark, that.getAvailableCells(game))
}

func (that *botService) findWinningMove(game *entity.Game, mark string) int {
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
