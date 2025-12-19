package bot

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hiei-discord-bot/internal/commands"
	"hiei-discord-bot/internal/config"
	"hiei-discord-bot/internal/events"
	"hiei-discord-bot/internal/interactions"
	"hiei-discord-bot/internal/settings"
	"hiei-discord-bot/internal/settings/store"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot instance
type Bot struct {
	session  *discordgo.Session
	registry *commands.Registry
	config   *config.Config
}

// New creates a new bot instance
func New(cfg *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Configure HTTP client with timeout
	session.Client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Set intents - we need guilds for slash commands
	session.Identify.Intents = discordgo.IntentsGuilds

	bot := &Bot{
		session:  session,
		registry: commands.Global(),
		config:   cfg,
	}

	// Initialize settings store
	sqliteStore, err := store.NewSQLiteStore("database.db")
	if err != nil {
		slog.Error("Failed to initialize settings store", "error", err)
	} else {
		settings.GetManager().SetStore(sqliteStore)
		slog.Info("Settings store initialized")
	}

	// Register commands
	bot.registerCommands()

	// Add handlers
	session.AddHandler(bot.registry.HandleInteraction)
	session.AddHandler(bot.handleInteraction)

	// Register event handlers
	eventHandler := events.NewHandler(session, bot.registry)
	eventHandler.Register()

	return bot, nil
}

// registerCommands registers all available commands
func (bot *Bot) registerCommands() {
	// Commands are now auto-registered via init() in their respective packages
	// and blank-imported in internal/bot/commands.go
}

// handleInteraction handles component and modal interactions
func (bot *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	router := interactions.GetRouter()

	switch i.Type {
	case discordgo.InteractionMessageComponent:
		router.HandleComponent(s, i)
	case discordgo.InteractionModalSubmit:
		router.HandleModal(s, i)
	}
}

// Start starts the bot and blocks until interrupted
func (bot *Bot) Start() error {
	if err := bot.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	// Sync local command versions at startup
	if err := commands.SyncLocalCommandVersions(bot.registry); err != nil {
		slog.Error("Failed to sync local command versions", "error", err)
	}

	slog.Info("Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return bot.Stop()
}

// Stop gracefully stops the bot
func (bot *Bot) Stop() error {
	slog.Info("Shutting down bot...")
	return bot.session.Close()
}
