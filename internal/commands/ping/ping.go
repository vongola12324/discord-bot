package ping

import (
	"hiei-discord-bot/internal/commands"
	"hiei-discord-bot/internal/i18n"
	"hiei-discord-bot/internal/interactions"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.Register(New())
}

// Command implements the ping slash command
type Command struct{}

// New creates a new ping command instance
func New() *Command {
	return &Command{}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Check if the bot is responsive and shows latency",
	}
}

// Version returns the command version
func (c *Command) Version() string {
	return "1.0.0"
}

// Execute runs the ping command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get user locale
	locale := i18n.GetUserLocaleFromInteraction(i)

	latency := s.HeartbeatLatency()

	return interactions.RespondSuccess(s, i, locale, "command.ping.response", true, latency)
}
