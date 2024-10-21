package websocket

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

func (that *Server) handleConnect(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleConnect")

	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	player, err := that.uGame.GetOrCreatePlayer(ctx, payload.Player.ID)
	if err != nil {
		log.Info("failed to create or get", "player", err)

		return that.sendErrorResponse(bufrw, msg.Action, "failed to create a new player")
	}

	if player.GameID == "" {
		payload = ResponsePayload{
			Player: player,
		}

		log.Info("Player connected", "playerID:", payload.Player.ID)

		if err = that.sendMessage(*bufrw, msg.Action, payload); err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}

		return nil
	}

	payload, err = that.inGame(ctx, player.GameID)
	if err != nil {
		log.Info("failed to get", "game", err)

		return that.sendErrorResponse(bufrw, msg.Action, "failed to get game")
	}

	if err = that.sendMessage(*bufrw, msg.Action, payload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}

func (that *Server) handleNewGame(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleNewGame")

	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	game, err := that.uGame.GetOrCreateGame(ctx, payload.Player.ID)
	if err != nil {
		log.Info("failed to create or get", "player", err)

		return that.sendErrorResponse(bufrw, msg.Action, "failed to create a new game")
	}

	payload = ResponsePayload{
		Game: game,
	}

	if err = that.sendMessage(*bufrw, msg.Action, payload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Info("Player is already in game", "gameID:", payload.Game.ID)

	return nil
}

func (that *Server) handleJoinGame(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleJoinGame")

	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	existingGame, err := that.uGame.ConnectToGame(ctx, payload.Game.ID, payload.Player.ID)
	if err != nil {
		log.Error("failed to join game", "error", err)

		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("game %s is already full", payload.Game.ID))
	}

	payload = ResponsePayload{
		Game: existingGame,
	}

	if err = that.sendMessage(*bufrw, msg.Action, payload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Info("Player joined game", "gameID:", payload.Game.ID)

	return nil
}

func (that *Server) handleGameTurn(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleGameTurn")

	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	existingGame, err := that.uGame.MakeTurn(ctx, payload.Player.ID, payload.Cell)
	if errors.Is(err, apperror.ErrGameFinished) {
		if err = that.handleGameFinished(bufrw, existingGame); err != nil {
			return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("failed to finish game %s: %v", existingGame.ID, err))
		}
	}

	if errors.Is(err, apperror.ErrGameIsNotStarted) {
		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("game %s is not started", existingGame.ID))
	}

	if err != nil {
		log.Error("failed to make move", "error", err)

		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("failed to move in game %s: %v", existingGame.ID, err))
	}

	payload = ResponsePayload{
		Game: existingGame,
	}

	if err = that.sendMessage(*bufrw, msg.Action, payload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Info("Player turn in game", "gameID:", payload.Game.ID)

	return nil
}

func (that *Server) handleGameFinished(bufrw *bufio.ReadWriter, game *entity.Game) error {
	log := that.logger.With("method", "handleGameFinished")

	action := "game:finished"

	payload := ResponsePayload{
		Game: game,
	}

	if err := that.sendMessage(*bufrw, action, payload); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}

	log.Info("Game finished", "gameID:", payload.Game.ID)

	return nil
}

func (that *Server) inGame(ctx context.Context, gameID string) (ResponsePayload, error) {
	var payload ResponsePayload

	game, err := that.uGame.GetOrCreateGame(ctx, gameID)
	if err != nil {
		return ResponsePayload{}, fmt.Errorf("failed to create or get player: %w", err)
	}

	payload = ResponsePayload{Game: game}

	return payload, nil
}

func (that *Server) sendErrorResponse(bufrw *bufio.ReadWriter, action, errorMsg string) error {
	payload := ResponsePayload{Error: errorMsg}
	if err := that.sendMessage(*bufrw, action, payload); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}

	return nil
}
