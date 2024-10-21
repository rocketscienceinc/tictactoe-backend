package apperror

import "errors"

var ErrGameFinished = errors.New("game is already finished")

var ErrGameIsNotStarted = errors.New("game is not started")
