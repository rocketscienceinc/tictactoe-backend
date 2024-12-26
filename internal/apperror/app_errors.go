package apperror

import "errors"

var (
	ErrGameFinished      = errors.New("game is already finished")
	ErrGameIsNotStarted  = errors.New("game is not started")
	ErrNotYourTurn       = errors.New("it's not your turn")
	ErrNoActiveGames     = errors.New("no active games")
	ErrCellOccupied      = errors.New("cell is already occupied")
	ErrGameAlreadyExists = errors.New("game already exists")
)
