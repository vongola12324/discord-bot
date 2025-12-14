package main

import (
	"log/slog"
	"os"

	"hiei-discord-bot/internal/bot"
	"hiei-discord-bot/internal/config"
	"hiei-discord-bot/internal/i18n"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Discord Bot...")

	// Load translations
	if err := i18n.LoadTranslations(); err != nil {
		slog.Error("Failed to load translations", "error", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
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
