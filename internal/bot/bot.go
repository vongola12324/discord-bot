package bot

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"hiei-discord-bot/internal/commands"
	"hiei-discord-bot/internal/commands/game"
	"hiei-discord-bot/internal/commands/help"
	"hiei-discord-bot/internal/commands/ping"
	"hiei-discord-bot/internal/commands/reload"
	"hiei-discord-bot/internal/config"
	"hiei-discord-bot/internal/events"
	"hiei-discord-bot/internal/interactions"

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

	// Set intents - we need guilds for slash commands
	session.Identify.Intents = discordgo.IntentsGuilds

	bot := &Bot{
		session:  session,
		registry: commands.NewRegistry(),
		config:   cfg,
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
	// Register ping command
	bot.registry.Register(ping.New())

	// Register help command
	bot.registry.Register(help.New(func() []help.CommandInfo {
		cmds := bot.registry.All()
		infos := make([]help.CommandInfo, 0, len(cmds))
		for _, cmd := range cmds {
			def := cmd.Definition()
			infos = append(infos, help.CommandInfo{
				Name:        def.Name,
				Description: def.Description,
			})
		}
		return infos
	}))

	// Register game command
	bot.registry.Register(game.New())

	// Register reload command with callback
	bot.registry.Register(reload.New(func(guildID string) (int, error) {
		definitions := bot.registry.GetDefinitions()
		if err := commands.SyncGuildCommands(bot.session, bot.registry, guildID); err != nil {
			return 0, err
		}
		return len(definitions), nil
	}))

	// Add more commands here as needed
}

// syncCommands synchronizes slash commands with Discord for all guilds
func (bot *Bot) syncCommands() error {
	slog.Info("Syncing slash commands with Discord...")

	var failedGuilds []string

	// Sync commands to all guilds the bot is in
	for _, guild := range bot.session.State.Guilds {
		if err := commands.SyncGuildCommands(bot.session, bot.registry, guild.ID); err != nil {
			slog.Error("Failed to sync commands for guild", "guild_id", guild.ID, "guild_name", guild.Name, "error", err)
			failedGuilds = append(failedGuilds, guild.ID)
			continue
		}
	}

	if len(failedGuilds) > 0 {
		return fmt.Errorf("failed to sync commands for %d guild(s): %v", len(failedGuilds), failedGuilds)
	}

	slog.Info("Successfully synced all commands")
	return nil
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
