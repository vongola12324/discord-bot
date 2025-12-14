package commands

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// SyncGuildCommands synchronizes commands for a specific guild
func SyncGuildCommands(session *discordgo.Session, registry *Registry, guildID string) error {
	slog.Info("Starting command sync", "guild_id", guildID)

	// Get all command definitions from registry
	definitions := registry.GetDefinitions()
	slog.Debug("Command definitions to sync", "count", len(definitions))

	// Log all definitions
	for i, def := range definitions {
		slog.Debug("Definition",
			"index", i,
			"name", def.Name,
			"type", def.Type,
			"description", def.Description)
	}

	// Get existing commands from Discord
	existingCmds, err := session.ApplicationCommands(session.State.User.ID, guildID)
	if err != nil {
		slog.Error("Failed to fetch existing commands", "guild_id", guildID, "error", err)
		return fmt.Errorf("failed to fetch existing commands: %w", err)
	}
	slog.Info("Existing commands fetched", "count", len(existingCmds), "guild_id", guildID)

	// Delete all existing commands
	for _, cmd := range existingCmds {
		slog.Debug("Deleting command", "name", cmd.Name, "type", cmd.Type, "id", cmd.ID)
		err := session.ApplicationCommandDelete(session.State.User.ID, guildID, cmd.ID)
		if err != nil {
			slog.Warn("Failed to delete command", "command", cmd.Name, "id", cmd.ID, "error", err)
		} else {
			slog.Debug("Deleted command successfully", "name", cmd.Name)
		}
	}

	// Register new commands
	var failedCommands []string
	successCount := 0

	slog.Info("Registering new commands", "count", len(definitions), "guild_id", guildID)

	for i, def := range definitions {
		slog.Debug("Creating command",
			"index", i,
			"name", def.Name,
			"type", def.Type,
			"description", def.Description,
			"guild_id", guildID)

		created, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, def)
		if err != nil {
			slog.Error("Failed to create command",
				"name", def.Name,
				"type", def.Type,
				"guild_id", guildID,
				"error", err,
				"error_type", fmt.Sprintf("%T", err))
			failedCommands = append(failedCommands, fmt.Sprintf("%s (type: %d): %v", def.Name, def.Type, err))
		} else {
			slog.Info("Synced command",
				"name", def.Name,
				"type", def.Type,
				"command_id", created.ID,
				"guild_id", guildID)
			successCount++
		}
	}

	slog.Info("Command sync completed",
		"guild_id", guildID,
		"total", len(definitions),
		"succeeded", successCount,
		"failed", len(failedCommands))

	if len(failedCommands) > 0 {
		errMsg := fmt.Errorf("failed to sync %d command(s), succeeded: %d/%d. Failed commands: %v",
			len(failedCommands), successCount, len(definitions), failedCommands)
		slog.Error("Command sync had failures", "guild_id", guildID, "error", errMsg)
		return errMsg
	}

	return nil
}
