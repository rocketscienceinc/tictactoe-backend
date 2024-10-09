package socket

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrNotYourTurn  = errors.New("not your turn")
	ErrCellOccupied = errors.New("cell occupied")
)

// handleConnect - handles create or connect player.
func (that *Server) handleConnect(msg *Message, bufrw *bufio.ReadWriter) error {
	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal player info: %w", err)
	}

	var responsePayload ResponsePayload

	if payload.Player.ID == "" {
		newPlayerID := GenerateNewSessionID()
		responsePayload = ResponsePayload{
			Player: &Player{
				ID: newPlayerID,
			},
			Game: nil,
		}

		that.logger.Info("registering new player", "player_id", newPlayerID)
	} else {
		responsePayload = ResponsePayload{
			Player: &Player{
				ID: payload.Player.ID,
			},
			Game: nil,
		}
	}

	if err := that.sendMessage(*bufrw, msg.Action, responsePayload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}

// handleNewGame - handles creation of a new game.
func (that *Server) handleNewGame(msg *Message, bufrw *bufio.ReadWriter) error {
	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal player info: %w", err)
	}

	newGame := NewGame(GenerateGameID())

	games[newGame.ID] = newGame

	player := Player{
		ID:   payload.Player.ID,
		Mark: "X",
	}
	if err := newGame.JoinPlayer(player); err != nil {
		return fmt.Errorf("failed to join player: %w", err)
	}

	responsePayload := ResponsePayload{
		Player: nil,
		Game: &Game{
			ID:      newGame.ID,
			Board:   newGame.Board,
			Turn:    newGame.Turn,
			Winner:  newGame.Winner,
			Status:  newGame.Status,
			Players: newGame.Players,
		},
	}

	if err := that.sendMessage(*bufrw, msg.Action, responsePayload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}

// handleJoinGame - handles join of game.
func (that *Server) handleJoinGame(msg *Message, bufrw *bufio.ReadWriter) error {
	var payload JoinGamePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal join game payload: %w", err)
	}

	game, err := FindGameByID(payload.Room.ID)
	if err != nil {
		return fmt.Errorf("failed to find game: %w", err)
	}

	player := Player{
		ID:   payload.Player.ID,
		Mark: "O",
	}
	if err := game.JoinPlayer(player); err != nil {
		return fmt.Errorf("failed to join player: %w", err)
	}

	responsePayload := ResponsePayload{
		Player: nil,
		Game: &Game{
			ID:      game.ID,
			Board:   game.Board,
			Turn:    game.Turn,
			Winner:  game.Winner,
			Status:  game.Status,
			Players: game.Players,
		},
	}

	if err := that.sendMessage(*bufrw, msg.Action, responsePayload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}

// handleTurn - handles player's turn and calculates the winner.
func (that *Server) handleTurn(msg *Message, bufrw *bufio.ReadWriter) error {
	var payload TurnPayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal turn payload: %w", err)
	}

	game, err := FindGameByID(payload.Game.ID)
	if err != nil {
		return fmt.Errorf("failed to find game: %w", err)
	}

	var currentPlayer Player
	for _, player := range game.Players {
		if player.ID == payload.Player.ID {
			currentPlayer = player
			break
		}
	}

	if game.Turn != currentPlayer.Mark {
		return fmt.Errorf("%w: %s", ErrNotYourTurn, currentPlayer.Mark)
	}

	if game.Board[payload.Cell] != "" {
		return ErrCellOccupied
	}

	game.Board[payload.Cell] = currentPlayer.Mark

	that.updateGameStatus(game)

	responsePayload := ResponsePayload{
		Player: nil,
		Game: &Game{
			ID:     game.ID,
			Board:  game.Board,
			Turn:   game.Turn,
			Winner: game.Winner,
			Status: game.Status,
		},
	}

	if err := that.sendMessage(*bufrw, msg.Action, responsePayload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}

func (that *Server) updateGameStatus(game *Game) {
	winner, isFull := checkGameStatus(game.Board)
	switch {
	case winner != "":
		game.Winner = winner
		game.Status = "finished"
	case isFull:
		game.Winner = "-"
		game.Status = "finished"
	default:
		game.Turn = toggleMark(game.Turn)
	}
}
