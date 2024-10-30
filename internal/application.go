package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository/storage"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository/storage/sqlite"
	"github.com/rocketscienceinc/tittactoe-backend/internal/service"
	"github.com/rocketscienceinc/tittactoe-backend/internal/usecase"
	"github.com/rocketscienceinc/tittactoe-backend/transport/rest"
	"github.com/rocketscienceinc/tittactoe-backend/transport/websocket"
)

const sqliteStoragePath = "data/sqlite/storage.db"

var ErrAddrNotFound = errors.New("redis address string is empty")

// RunApp - runs the application.
func RunApp(logger *slog.Logger, conf *config.Config) error {
	log := logger.With("component", "app")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	redisAddrString := conf.Redis.GetRedisAddr()
	if redisAddrString == "" {
		return ErrAddrNotFound
	}

	redisStorage, err := storage.New(ctx, redisAddrString)
	if err != nil {
		return fmt.Errorf("could not connect to redis storage: %w", err)
	}

	defer func() {
		if err = redisStorage.Connection.Close(); err != nil {
			log.Error("could not close redis storage", "error", err)
		}
	}()

	// create a new sqlite storage
	sqliteStorage, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		panic(fmt.Errorf("can't connect to sqlite storage: %w", err))
	}

	// init sqlite storage
	if err = sqliteStorage.Init(ctx); err != nil {
		panic(fmt.Errorf("can't init sqlite storage: %w", err))
	}

	userRepo := repository.NewUserRepository(sqliteStorage.Connection)
	playerRepo := repository.NewPlayerRepository(redisStorage.Connection)
	gameRepo := repository.NewGameRepository(log, redisStorage.Connection)

	playerService := service.NewPlayerService(playerRepo)
	gameService := service.NewGameService(gameRepo)
	botService := service.NewBotService()
	gamePlayService := service.NewGamePlayService(log, playerService, gameService, botService)

	gameUseCase := usecase.NewGameUseCase(playerService, gameService, gamePlayService)

	wsHandler := websocket.New(log, gameUseCase)

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService("pingo")

	restHandlers := rest.NewHandlers(conf.GoogleOAuth.RedirectURL, conf.GoogleOAuth.ClientID, conf.GoogleOAuth.ClientSecret, userService, authService)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/ping", restHandlers.PingHandler)
	mux.HandleFunc("/api/auth/google/login", restHandlers.GoogleLogin)
	mux.HandleFunc("/api/auth/google/callback", restHandlers.GoogleCallback)

	mux.HandleFunc("/ws", wsHandler.ServeHTTP)

	srv := &http.Server{
		Addr:         ":" + conf.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Info("Starting HTTP and WebSocket server on", "port", conf.HTTPPort)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	sig := <-sigs
	log.Info("Received signal, shutting down", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		panic(fmt.Errorf("server shutdown error: %w", err))
	}

	log.Info("Server gracefully stopped")

	return nil
}
