package websocket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/pkg"
)

type uGame interface {
	GetOrCreatePlayer(ctx context.Context, id string) (*entity.Player, error)
	GetOrCreateGame(ctx context.Context, id string) (*entity.Game, error)

	ConnectToGame(ctx context.Context, gameID, playerID string) (*entity.Game, error)

	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type Server struct {
	logger *slog.Logger
	uGame  uGame

	handlers map[string]func(ctx context.Context, message *Message, writer *bufio.ReadWriter) error
}

func New(logger *slog.Logger, uGame uGame) *Server {
	server := &Server{
		logger: logger,
		uGame:  uGame,

		handlers: make(map[string]func(context.Context, *Message, *bufio.ReadWriter) error),
	}

	server.handlers["connect"] = server.handleConnect
	server.handlers["game:new"] = server.handleNewGame
	server.handlers["game:join"] = server.handleJoinGame
	server.handlers["game:turn"] = server.handleGameTurn

	return server
}

// Start - starts WebSocket server.
func (that *Server) Start(ctx context.Context, port string) error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		that.upgradeToWebSocket(ctx, w, r)
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// upgradeToWebSocket - upgrades the connection to WebSocket.
func (that *Server) upgradeToWebSocket(ctx context.Context, writer http.ResponseWriter, req *http.Request) {
	log := that.logger.With("method", "upgradeConnection")

	if req.Header.Get("Upgrade") != "websocket" {
		http.Error(writer, "not a websocket upgrade", http.StatusBadRequest)
		return
	}

	that.setSessionCookie(writer, req)

	key := req.Header.Get("Sec-WebSocket-Key")
	acceptKey := pkg.GenerateAcceptKey(key)

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
			log.Error("error processing message", "error", err) // ToDo: need log this
			continue
		}

		if err = handler(ctx, &message, bufrw); err != nil {
			log.Error("error processing message", "error", err)
		}
	}
}

// setSessionCookie - set user session.
func (that *Server) setSessionCookie(writer http.ResponseWriter, req *http.Request) {
	log := that.logger.With("method", "setSessionCookie")

	cookie, err := req.Cookie("user_session")
	if err != nil {
		cookie = &http.Cookie{
			Name:    "user_session",
			Value:   pkg.GenerateNewSessionID(),
			Expires: time.Now().Add(24 * time.Hour),
			Path:    "/ws",
		}
		http.SetCookie(writer, cookie)
		log.Info("session cookie not found, new one created", "cookie", cookie.Value)
		return
	}

	log.Info("session cookie found", "cookie", cookie.Value)
}
