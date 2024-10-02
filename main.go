package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	app "github.com/rocketscienceinc/tittactoe-backend/internal"
	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
)

// main - is the entry point of the application. It initializes the configuration, logger, and runs the application.
func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "recovered from panic: %v\n", err)
			os.Exit(1)
		}
	}()

	conf := initConfig()
	logger := initLogger(conf)

	if err := app.RunApp(logger, conf); err != nil {
		panic(fmt.Errorf("app run failed: %w", err))
	}
}

// initialize config.
func initConfig() *config.Config {
	baseDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get current directory: %w", err))
	}

	return config.MustLoad(filepath.Join(baseDir, "./config.yml"))
}

// initialize logger.
func initLogger(conf *config.Config) *slog.Logger {
	var level slog.Level

	switch conf.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
