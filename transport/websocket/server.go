package websocket

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/pkg"
)

type gameUseCase interface {
	GetOrCreatePlayer(ctx context.Context, playerID string) (*entity.Player, error)

	GetOrCreateGame(ctx context.Context, playerID, gameType string) (*entity.Game, error)
	GetGameByPlayerID(ctx context.Context, playerID string) (*entity.Game, error)
	JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	JoinWaitingPublicGame(ctx context.Context, playerID string) (*entity.Game, error)

	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type Server struct {
	logger      *slog.Logger
	gameUseCase gameUseCase

	handlers map[string]func(ctx context.Context, message *Message, writer *bufio.ReadWriter) error

	connections map[string]*bufio.ReadWriter
}

func New(logger *slog.Logger, gameUseCase gameUseCase) *Server {
	server := &Server{
		logger:      logger,
		gameUseCase: gameUseCase,

		handlers:    make(map[string]func(context.Context, *Message, *bufio.ReadWriter) error),
		connections: make(map[string]*bufio.ReadWriter),
	}

	server.handlers["connect"] = server.handleConnect
	server.handlers["game:new"] = server.handleNewGame
	server.handlers["game:join"] = server.handleJoinGame
	server.handlers["game:turn"] = server.handleGameTurn

	return server
}

func (that *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	that.upgradeToWebSocket(r.Context(), w, r)
}

// upgradeToWebSocket - upgrades the connection to WebSocket.
func (that *Server) upgradeToWebSocket(ctx context.Context, writer http.ResponseWriter, req *http.Request) {
	log := that.logger.With("method", "upgradeConnection")

	if req.Header.Get("Upgrade") != "websocket" {
		http.Error(writer, "not a websocket upgrade", http.StatusBadRequest)
		return
	}

	wsKey := req.Header.Get("Sec-WebSocket-Key")
	acceptKey := pkg.GenerateAcceptKey(wsKey)

	writer.Header().Set("Upgrade", "websocket")
	writer.Header().Set("Connection", "Upgrade")
	writer.Header().Set("Sec-WebSocket-Accept", acceptKey)
	writer.WriteHeader(http.StatusSwitchingProtocols)

	hijacker, ok := writer.(http.Hijacker)
	if !ok {
		log.Error("web server does not support hijacking", "error", http.StatusText(http.StatusInternalServerError))
		return
	}

	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		log.Error("failed to hijack connection", "error", err)
		return
	}

	defer conn.Close()

	log.Info("WebSocket connection established")

	if err = that.handleMessages(ctx, bufrw); err != nil {
		log.Error("error handling messages", "error", err)
	}
}

// handleMessages - processes messages from the client.
func (that *Server) handleMessages(ctx context.Context, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "HandleMessages")

	for {
		reqBody, err := that.readRequest(bufrw)
		if err != nil {
			log.Error("error reading message", "error", err)
			return err
		}

		var message Message
		if err = json.Unmarshal(reqBody, &message); err != nil {
			log.Error("failed to unmarshal message", "error", err)
			continue
		}

		handler, ok := that.handlers[message.Action]
		if !ok {
			log.Error("action handler not found")

			payloadResp := ResponsePayload{
				Error: "action handler not found",
			}

			err = that.sendMessage(*bufrw, message.Action, payloadResp)
			if err != nil {
				log.Error("failed to send message", "error", err)
			}

			continue
		}

		if err = handler(ctx, &message, bufrw); err != nil {
			log.Error("invalid handle message", "error", err)
		}
	}
}
