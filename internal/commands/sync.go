package commands

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// SyncGuildCommands synchronizes commands for a specific guild
func SyncGuildCommands(session *discordgo.Session, registry *Registry, guildID string) error {
	// Get all command definitions from registry
	definitions := registry.GetDefinitions()

	// Get existing commands from Discord
	existingCmds, err := session.ApplicationCommands(session.State.User.ID, guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing commands: %w", err)
	}

	// Delete all existing commands
	for _, cmd := range existingCmds {
		err := session.ApplicationCommandDelete(session.State.User.ID, guildID, cmd.ID)
		if err != nil {
			slog.Warn("Failed to delete command", "command", cmd.Name, "error", err)
		}
	}

	// Register new commands
	for _, def := range definitions {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, def)
		if err != nil {
			return fmt.Errorf("failed to create command '%s': %w", def.Name, err)
		}
		slog.Info("Synced command", "name", def.Name, "guild_id", guildID)
	}

	return nil
}
