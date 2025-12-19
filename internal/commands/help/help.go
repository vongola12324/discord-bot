package help

import (
	"fmt"
	"hiei-discord-bot/internal/commands"
	"hiei-discord-bot/internal/interactions"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.Register(New())
}

// CommandInfo represents basic command information
type CommandInfo struct {
	Name        string
	Description string
}

// Command implements the help slash command
type Command struct{}

// New creates a new help command instance
func New() *Command {
	return &Command{}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "help",
		Description: "Display all available commands",
	}
}

// Version returns the command version
func (c *Command) Version() string {
	return "1.0.0"
}

// Execute runs the help command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var builder strings.Builder
	builder.WriteString("📋 **Available Commands:**\n\n")

	cmds := commands.Global().All()
	for _, cmd := range cmds {
		def := cmd.Definition()
		builder.WriteString(fmt.Sprintf("`/%s` - %s\n", def.Name, def.Description))
	}

	return interactions.RespondCustom(s, i, &discordgo.InteractionResponseData{
		Content: builder.String(),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}
