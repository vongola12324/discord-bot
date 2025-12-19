package commands

import (
	"log/slog"
	"strings"
	"sync"

	"hiei-discord-bot/internal/i18n"
	"hiei-discord-bot/internal/interactions"
	"hiei-discord-bot/internal/settings"

	"github.com/bwmarrin/discordgo"
)

// Registry manages all bot commands
type Registry struct {
	commands map[string]Command
	mu       sync.RWMutex
}

var (
	globalRegistry = NewRegistry()
)

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Global returns the global command registry
func Global() *Registry {
	return globalRegistry
}

// Register adds a command to the global registry
func Register(cmd Command) {
	globalRegistry.Register(cmd)
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := strings.ToLower(cmd.Definition().Name)
	r.commands[name] = cmd
	slog.Info("Registered command", "name", name)

	// Auto-register settings if command is configurable
	if conf, ok := cmd.(settings.Configurable); ok {
		mgr := settings.GetManager()
		for _, def := range conf.Settings() {
			mgr.Register(def)
			slog.Info("Registered setting", "key", def.Key, "command", name)
		}
	}
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
		interactions.RespondError(s, i, locale, "command.execution_error", true)
	}
}
