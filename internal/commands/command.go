package commands

import (
	"hiei-discord-bot/internal/settings"

	"github.com/bwmarrin/discordgo"
)

// Command represents a bot slash command interface
type Command interface {
	// Definition returns the slash command definition for Discord registration
	Definition() *discordgo.ApplicationCommand

	// Version returns the command version (semver)
	Version() string

	// Execute runs the command logic when invoked
	Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// Configurable is an alias for settings.Configurable to avoid circular imports if needed
// but here we just use it to mark commands that have settings
type Configurable interface {
	settings.Configurable
}
