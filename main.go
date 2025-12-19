package main

import (
	"log/slog"
	"os"
	"strings"

	"hiei-discord-bot/internal/bot"
	"hiei-discord-bot/internal/config"
	"hiei-discord-bot/internal/i18n"
)

func main() {
	// Load configuration to get log level
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Parse log level from config and reconfigure logger
	var logLevel slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Reconfigure logger with the configured level
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Discord Bot...", "log_level", cfg.LogLevel)

	// Load translations
	if err := i18n.LoadTranslations(); err != nil {
		slog.Error("Failed to load translations", "error", err)
		os.Exit(1)
	}

	// Create bot instance
	b, err := bot.New(cfg)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	// Start bot (blocks until interrupted)
	if err := b.Start(); err != nil {
		slog.Error("Failed to start bot", "error", err)
		os.Exit(1)
	}

	slog.Info("Bot stopped gracefully")
}
