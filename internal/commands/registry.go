package commands

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"hiei-discord-bot/internal/i18n"
)

// Registry manages all bot commands
type Registry struct {
	commands map[string]Command
	mu       sync.RWMutex
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := strings.ToLower(cmd.Definition().Name)
	r.commands[name] = cmd
	slog.Info("Registered command", "name", name)
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[strings.ToLower(name)]
	return cmd, exists
}

// All returns all registered commands
func (r *Registry) All() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmds := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// GetDefinitions returns all command definitions for Discord registration
func (r *Registry) GetDefinitions() []*discordgo.ApplicationCommand {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definitions := make([]*discordgo.ApplicationCommand, 0, len(r.commands))
	for _, cmd := range r.commands {
		definitions = append(definitions, cmd.Definition())
	}
	return definitions
}

// HandleInteraction processes incoming slash command interactions
func (r *Registry) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Only handle application commands (slash commands)
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	cmdName := strings.ToLower(data.Name)

	// Get and execute command
	cmd, exists := r.Get(cmdName)
	if !exists {
		slog.Warn("Unknown command received", "command", cmdName)
		return
	}

	if err := cmd.Execute(s, i); err != nil {
		slog.Error("Error executing command", "command", cmdName, "error", err)

		// Get user locale for error message
		locale := i18n.GetUserLocaleFromInteraction(i)

		// Try to respond with error message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: i18n.T(locale, "command.execution_error"),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}
