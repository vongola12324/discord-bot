package game

import (
	"hiei-discord-bot/internal/commands"
	"hiei-discord-bot/internal/i18n"
	"hiei-discord-bot/internal/interactions"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.Register(New())
}

// Command implements the game command with subcommands
type Command struct{}

// New creates a new game command instance
func New() *Command {
	return &Command{}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	options := []*discordgo.ApplicationCommandOption{}
	for _, sub := range GetSubCommands() {
		options = append(options, &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        sub.Name(),
			Description: sub.Description(),
			Options:     sub.Options(),
		})
	}

	return &discordgo.ApplicationCommand{
		Name:        "game",
		Description: "Play various games",
		Options:     options,
	}
}

// Version returns the command version
func (c *Command) Version() string {
	return "1.1.0"
}

// Execute runs the game command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get user locale
	locale := i18n.GetUserLocaleFromInteraction(i)

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return interactions.RespondError(s, i, locale, "game.select_game", true)
	}

	subcommandName := options[0].Name
	if sub, exists := GetSubCommand(subcommandName); exists {
		return sub.Handle(s, i)
	}

	return interactions.RespondError(s, i, locale, "game.unknown_game", true)
}
