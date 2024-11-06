package apperror

import "errors"

var (
	ErrGameFinished     = errors.New("game is already finished")
	ErrGameIsNotStarted = errors.New("game is not started")
	ErrNotYourTurn      = errors.New("it's not your turn")
	ErrNoActiveGames    = errors.New("no active games")
	ErrCellOccupied     = errors.New("cell is already occupied")

	ErrNotFound = errors.New("not found")
)
