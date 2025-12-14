package ping

import (
	"github.com/bwmarrin/discordgo"
	"hiei-discord-bot/internal/i18n"
)

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

// Execute runs the ping command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get user locale
	locale := i18n.GetUserLocaleFromInteraction(i)

	latency := s.HeartbeatLatency()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: i18n.Tf(locale, "command.ping.response", latency),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
