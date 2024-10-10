package websocket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/game"
)

const (
	statusFinished = "finished"
	statusWaiting  = "waiting"
)

func (that *Server) handleConnect(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal player info: %w", err)
	}

	var playerID string
	if payload.Player.ID == "" {
		playerID = GenerateNewSessionID()
	} else {
		playerID = payload.Player.ID
	}

	player, err := that.redis.GetOrCreatePlayer(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get or create player: %w", err)
	}

	responsePayload := ResponsePayload{
		Player: player,
		Game:   nil,
	}

	if playerID == payload.Player.ID {
		that.logger.Info("Player connected", "playerID:", player.ID)
	} else {
		that.logger.Info("Registered new player", "playerID:", player.ID)
	}

	if err := that.sendMessage(*bufrw, msg.Action, responsePayload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}

func (that *Server) handleNewGame(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal player info: %w", err)
	}

	existingGame, err := that.redis.GetActiveGame(ctx, payload.Player.ID)
	if err != nil {
		return fmt.Errorf("failed to check active game: %w", err)
	}

	if existingGame != nil {
		if existingGame.Status == statusFinished {
			that.logger.Info("Existing game finished, creating a new game", "playerID", payload.Player.ID)
		} else {
			responsePayload := ResponsePayload{
				Player: existingGame.Player,
				Game: &GameResponse{
					ID:     existingGame.ID,
					Board:  existingGame.Board,
					Turn:   existingGame.Turn,
					Winner: existingGame.Winner,
					Status: existingGame.Status,
				},
			}
			return that.sendMessage(*bufrw, msg.Action, responsePayload)
		}
	}

	newGame := game.NewGame(payload.Player.ID)
	gameID := GenerateGameID()
	newGame.ID = gameID
	newGame.Status = statusWaiting

	if err := that.redis.SaveGame(ctx, payload.Player.ID, newGame); err != nil {
		return fmt.Errorf("failed to save new game: %w", err)
	}

	responsePayload := ResponsePayload{
		Player: newGame.Player,
		Game: &GameResponse{
			ID:     newGame.ID,
			Board:  newGame.Board,
			Turn:   newGame.Turn,
			Winner: newGame.Winner,
			Status: newGame.Status,
		},
	}

	if err := that.sendMessage(*bufrw, msg.Action, responsePayload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}
