package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	app "github.com/rocketscienceinc/tittactoe-backend/internal"
	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
)

var ErrUnknownLogLevel = errors.New("unknown log level")

// main - is the entry point of the application. It initializes the configuration, logger, and runs the application.
func main() {
	conf, err := initConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	logger, err := initLogger(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	if err = app.RunApp(logger, conf); err != nil {
		logger.Error("Application failed to run", "error", err)
		os.Exit(1)
	}
}

// initialize config.
func initConfig() (*config.Config, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return config.MustLoad(filepath.Join(baseDir, "./config.yml")), nil
}

// initialize logger.
func initLogger(conf *config.Config) (*slog.Logger, error) {
	var level slog.Level

	switch conf.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownLogLevel, conf.LogLevel)
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})), nil
}
