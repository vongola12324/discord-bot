package commands

import "github.com/bwmarrin/discordgo"

// Command represents a bot slash command interface
type Command interface {
	// Definition returns the slash command definition for Discord registration
	Definition() *discordgo.ApplicationCommand

	// Execute runs the command logic when invoked
	Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error
}
