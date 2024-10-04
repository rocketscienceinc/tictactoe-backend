package socket

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	logger   *slog.Logger
	handlers map[string]func(message *Message, writer *bufio.ReadWriter) error
}

func New(logger *slog.Logger) *Server {
	server := &Server{
		logger:   logger,
		handlers: make(map[string]func(*Message, *bufio.ReadWriter) error),
	}

	server.handlers["connect"] = server.handleConnect

	return server
}

var ErrUnknownAction = errors.New("unknown action")

// Start - starts WebSocket server.
func (that *Server) Start(port string) error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		that.upgradeToWebSocket(w, r)
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
func (that *Server) upgradeToWebSocket(writer http.ResponseWriter, req *http.Request) {
	log := that.logger.With("method", "upgradeConnection")

	if req.Header.Get("Upgrade") != "websocket" {
		http.Error(writer, "not a websocket upgrade", http.StatusBadRequest)
		return
	}

	that.handleSessionCookie(writer, req, log)

	key := req.Header.Get("Sec-WebSocket-Key")
	acceptKey := GenerateAcceptKey(key)

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

	if err := that.HandleMessages(bufrw); err != nil {
		log.Error("error handling messages", "error", err)
	}
}

// processMessage - processes incoming messages from the client.
func (that *Server) processMessage(msg *Message, bufrw *bufio.ReadWriter) error {
	if handler, ok := that.handlers[msg.Action]; ok {
		return handler(msg, bufrw)
	}
	return fmt.Errorf("%w: %s", ErrUnknownAction, msg.Action)
}
