package game

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// SubCommand represents a game sub-command
type SubCommand interface {
	Name() string
	Description() string
	Options() []*discordgo.ApplicationCommandOption
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

var (
	subCommands = make(map[string]SubCommand)
	mu          sync.RWMutex
)

// RegisterSubCommand registers a game sub-command
func RegisterSubCommand(cmd SubCommand) {
	mu.Lock()
	defer mu.Unlock()
	subCommands[cmd.Name()] = cmd
}

// GetSubCommands returns all registered sub-commands
func GetSubCommands() []SubCommand {
	mu.RLock()
	defer mu.RUnlock()
	cmds := make([]SubCommand, 0, len(subCommands))
	for _, cmd := range subCommands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// GetSubCommand retrieves a sub-command by name
func GetSubCommand(name string) (SubCommand, bool) {
	mu.RLock()
	defer mu.RUnlock()
	cmd, exists := subCommands[name]
	return cmd, exists
}
