package apperror

import "errors"

var (
	ErrGameFinished     = errors.New("game is already finished")
	ErrGameIsNotStarted = errors.New("game is not started")
	ErrNotYourTurn      = errors.New("it's not your turn")
)
