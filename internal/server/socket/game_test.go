package socket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckGameStatus(t *testing.T) {
	t.Run("Winning positions", func(t *testing.T) {
		board := [9]string{"X", "X", "X", "", "", "", "", "", ""}

		winner, isFull := checkGameStatus(board)

		assert.Equal(t, "X", winner)

		assert.False(t, isFull)
	})

	t.Run("Draw", func(t *testing.T) {
		board := [9]string{"X", "O", "X", "X", "X", "O", "O", "X", "O"}

		winner, isFull := checkGameStatus(board)

		assert.Equal(t, "", winner)

		assert.True(t, isFull)
	})

	t.Run("Game ongoing", func(t *testing.T) {
		board := [9]string{"X", "O", "X", "", "O", "", "", "X", ""}

		winner, isFull := checkGameStatus(board)

		assert.Equal(t, "", winner)

		assert.False(t, isFull)
	})
}
