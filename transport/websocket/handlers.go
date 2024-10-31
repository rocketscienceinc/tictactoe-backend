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

	var payloadReq ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	player, err := that.gameUseCase.GetOrCreatePlayer(ctx, payloadReq.Player.ID)
	if err != nil {
		log.Info("failed to create or get", "player", err)

		return that.sendErrorResponse(bufrw, msg.Action, "failed to create a new player")
	}

	that.connections[player.ID] = bufrw

	if player.GameID != "" {
		game, err := that.gameUseCase.GetOrCreateGame(ctx, player.ID)
		if err != nil {
			log.Info("failed to get the game", "game", player.GameID)
			return that.sendErrorResponse(bufrw, msg.Action, "failed to get the game")
		}

		payloadResp := ResponsePayload{
			Player: player,
			Game:   game,
		}
		payloadResp.Game.Players = nil

		if err = that.sendMessage(*bufrw, msg.Action, payloadResp); err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}

		return nil
	}

	payloadResp := ResponsePayload{
		Player: player,
	}

	if err = that.sendMessage(*bufrw, msg.Action, payloadResp); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Info("successfully connected player")

	return nil
}

func (that *Server) handleNewGame(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleNewGame")

	var payloadReq ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	that.connections[payloadReq.Player.ID] = bufrw

	game, err := that.gameUseCase.GetOrCreateGame(ctx, payloadReq.Player.ID)
	if err != nil {
		log.Info("failed to create or get", "player", err)
		return that.sendErrorResponse(bufrw, msg.Action, "failed to create a new game")
	}

	log = log.With("gameID", game.ID)

	payloadResp := ResponsePayload{
		Player: payloadReq.Player,
		Game:   game,
	}

	payloadResp.Game.Players = nil

	if err = that.sendMessage(*bufrw, msg.Action, payloadResp); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Info("Player is already in game")

	return nil
}

func (that *Server) handleJoinGame(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleJoinGame")

	var payloadReq ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	that.connections[payloadReq.Player.ID] = bufrw

	log = log.With("playerID", payloadReq.Player.ID)

	game, err := that.gameUseCase.JoinGame(ctx, payloadReq.Game.ID, payloadReq.Player.ID)
	if err != nil {
		log.Error("failed to join game", "error", err)

		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("game %s: %v", payloadReq.Game.ID, err))
	}

	log = log.With("gameID", game.ID)

	for _, player := range game.Players {
		conn, ok := that.connections[player.ID]
		if !ok {
			log.Info("failed to find connection")
			continue
		}

		payloadResp := ResponsePayload{
			Game: game,
		}

		payloadResp.Game.Players = nil

		if err = that.sendMessage(*conn, msg.Action, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player joined game")

	return nil
}

func (that *Server) handleGameTurn(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleGameTurn")

	var payloadReq ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	that.connections[payloadReq.Player.ID] = bufrw

	log = log.With("playerID", payloadReq.Player.ID)

	game, err := that.gameUseCase.MakeTurn(ctx, payloadReq.Player.ID, payloadReq.Cell)
	if errors.Is(err, apperror.ErrGameFinished) {
		if err = that.handleGameFinished(game); err != nil {
			return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("failed to finish game %s: %v", game.ID, err))
		}

		return nil
	}

	if errors.Is(err, apperror.ErrGameIsNotStarted) {
		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("game %s: %v", game.ID, err))
	}

	if err != nil {
		log.Error("failed to make turn", "error", err)
		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("failed to turn in game %v", err))
	}

	log = log.With("gameID", game.ID)

	for _, player := range game.Players {
		conn, ok := that.connections[player.ID]
		if !ok {
			log.Info("failed to find connection")
			continue
		}

		payloadResp := ResponsePayload{
			Player: player,
			Game:   game,
		}

		payloadResp.Game.Players = nil

		if err = that.sendMessage(*conn, "game:turn:update", payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player made a turn")

	return nil
}

func (that *Server) handleGameFinished(game *entity.Game) error {
	log := that.logger.With("method", "handleGameFinished")

	action := "game:finished"

	payloadResp := ResponsePayload{
		Game: game,
	}

	payloadResp.Game.Players = nil

	for _, player := range game.Players {
		conn, ok := that.connections[player.ID]
		if !ok {
			log.Info("failed to find connection", "player", player.ID)
			continue
		}

		if err := that.sendMessage(*conn, action, payloadResp); err != nil {
			return fmt.Errorf("failed to send game finished message %s: %w", player.ID, err)
		}
	}

	log.Info("Game finished", "gameID", game.ID)

	return nil
}

func (that *Server) sendErrorResponse(bufrw *bufio.ReadWriter, action, errorMsg string) error {
	payload := ResponsePayload{Error: errorMsg}
	if err := that.sendMessage(*bufrw, action, payload); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}

	return nil
}
