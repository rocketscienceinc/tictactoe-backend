package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
	"github.com/rocketscienceinc/tittactoe-backend/internal/transport/redis"
	"github.com/rocketscienceinc/tittactoe-backend/internal/transport/rest"
	"github.com/rocketscienceinc/tittactoe-backend/internal/transport/websocket"
)

// RunApp - runs the application.
func RunApp(logger *slog.Logger, conf *config.Config) error {
	log := logger.With("component", "app")

	ctx := context.Background()

	// run http server
	go func() {
		log.Info("Starting HTTP server on ", "port:", conf.HTTPPort)
		if err := rest.Start(conf.HTTPPort); err != nil {
			panic(fmt.Errorf("failed to start HTTP server: %w", err))
		}
	}()

	// run Redis
	log.Info("Starting Redis server on ", "port", conf.Redis.Port)

	redisClient := redis.New(conf.RedisAddr)

	// run Websocket server
	log.Info("Starting WebSocket server on  ", "port:", conf.SocketPort)

	wsServer := websocket.New(logger, redisClient)

	if err := wsServer.Start(ctx, conf.SocketPort); err != nil {
		return fmt.Errorf("failed to start socket server: %w", err)
	}

	return nil
}
