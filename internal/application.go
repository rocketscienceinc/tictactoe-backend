package application

import (
	"fmt"
	"log/slog"

	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
	"github.com/rocketscienceinc/tittactoe-backend/internal/server"
	"github.com/rocketscienceinc/tittactoe-backend/internal/server/socket"
)

// RunApp - runs the application.
func RunApp(logger *slog.Logger, conf *config.Config) error {
	log := logger.With("component", "app")

	// run http server.
	go func() {
		log.Info("Starting HTTP server on ", "port:", conf.HTTPPort)
		if err := server.StartHTTPServer(conf.HTTPPort); err != nil {
			panic(fmt.Errorf("failed to start HTTP server: %w", err))
		}
	}()

	// run Websocket server.
	log.Info("Starting WebSocket server on  ", "port:", conf.SocketPort)

	wsServer := socket.New(logger)

	if err := wsServer.Start(conf.SocketPort); err != nil {
		return fmt.Errorf("failed to start socket server: %w", err)
	}

	return nil
}
