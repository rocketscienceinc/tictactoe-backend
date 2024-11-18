package service

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
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
	availableCells := make([]int, 0, len(game.Board))
	for i, cell := range game.Board {
		if cell == entity.EmptyCell {
			availableCells = append(availableCells, i)
		}
	}

	if len(availableCells) == 0 {
		return ErrNoAvailableMoves
	}

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

	chosenCell := availableCells[rand.Intn(len(availableCells))] //nolint: gosec // it's ok

	if err := game.MakeTurn(botPlayer.Mark, chosenCell); err != nil {
		return fmt.Errorf("bot failed to make turn: %w", err)
	}

	return nil
}
