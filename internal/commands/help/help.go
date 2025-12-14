package help

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// CommandInfo represents basic command information
type CommandInfo struct {
	Name        string
	Description string
}

// Command implements the help slash command
type Command struct {
	getCommands func() []CommandInfo
}

// New creates a new help command instance
func New(getCommands func() []CommandInfo) *Command {
	return &Command{
		getCommands: getCommands,
	}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "help",
		Description: "Display all available commands",
	}
}

// Execute runs the help command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var builder strings.Builder
	builder.WriteString("📋 **Available Commands:**\n\n")

	for _, cmd := range c.getCommands() {
		builder.WriteString(fmt.Sprintf("`/%s` - %s\n", cmd.Name, cmd.Description))
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
