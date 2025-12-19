package commands

import (
	"fmt"
	"hiei-discord-bot/internal/models"
	"hiei-discord-bot/internal/settings"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/mod/semver"
)

// SyncLocalCommandVersions checks and updates local command versions in the database
func SyncLocalCommandVersions(registry *Registry) error {
	mgr := settings.GetManager()
	allCommands := registry.All()
	currentCommandNames := make(map[string]bool)

	for _, cmd := range allCommands {
		def := cmd.Definition()
		version := fmt.Sprintf("v%s", cmd.Version())
		currentCommandNames[def.Name] = true

		dbVer, err := mgr.GetLocalCommandVersionAndBuildTime(def.Name)
		if err != nil {
			slog.Error("Failed to get local command version", "name", def.Name, "error", err)
			continue
		}

		// If version is newer or not exists, update build_time
		if dbVer.Version == "" || isVersionNewer(version, dbVer.Version) {
			slog.Info("Updating local command version", "name", def.Name, "old", dbVer.Version, "new", version)
			err := mgr.UpdateLocalCommandVersionAndBuildTime(def.Name, models.CommandVersion{
				Version:   version,
				BuildTime: time.Now().UTC(),
			})
			if err != nil {
				slog.Error("Failed to update local command version", "name", def.Name, "error", err)
			}
		}
	}

	// Delete obsolete commands from local_command_versions
	dbCommandNames, err := mgr.GetAllLocalCommandNames()
	if err != nil {
		slog.Error("Failed to get all local command names from DB", "error", err)
	} else {
		for _, dbName := range dbCommandNames {
			if !currentCommandNames[dbName] {
				slog.Info("Deleting obsolete local command version", "name", dbName)
				if err := mgr.DeleteLocalCommandVersion(dbName); err != nil {
					slog.Error("Failed to delete obsolete local command version", "name", dbName, "error", err)
				}
			}
		}
	}

	return nil
}

// SyncGuildCommands synchronizes commands for a specific guild
func SyncGuildCommands(session *discordgo.Session, registry *Registry, guildID string, force bool) error {
	slog.Info("Starting command sync", "guild_id", guildID, "force", force)

	// Get all commands from registry
	allCommands := registry.All()
	slog.Debug("Commands to sync", "count", len(allCommands))

	mgr := settings.GetManager()

	// Get existing commands from Discord
	existingCmds, err := session.ApplicationCommands(session.State.User.ID, guildID)
	if err != nil {
		slog.Error("Failed to fetch existing commands", "guild_id", guildID, "error", err)
		return fmt.Errorf("failed to fetch existing commands: %w", err)
	}
	slog.Info("Existing commands fetched", "count", len(existingCmds), "guild_id", guildID)

	if force {
		// Delete all existing commands
		for _, cmd := range existingCmds {
			slog.Debug("Deleting command (force)", "name", cmd.Name, "id", cmd.ID)
			err := session.ApplicationCommandDelete(session.State.User.ID, guildID, cmd.ID)
			if err != nil {
				slog.Warn("Failed to delete command", "command", cmd.Name, "id", cmd.ID, "error", err)
			}
		}
		existingCmds = nil
	}

	existingMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existingCmds {
		existingMap[cmd.Name] = cmd
	}

	defMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range allCommands {
		defMap[cmd.Definition().Name] = cmd.Definition()
	}

	// Delete commands that are no longer in registry
	for name, cmd := range existingMap {
		if _, exists := defMap[name]; !exists {
			slog.Info("Deleting obsolete command", "name", name, "id", cmd.ID, "guild_id", guildID)
			err := session.ApplicationCommandDelete(session.State.User.ID, guildID, cmd.ID)
			if err != nil {
				slog.Warn("Failed to delete obsolete command", "name", name, "error", err)
			}

			// Also delete from guild_command_versions DB
			if err := mgr.DeleteGuildCommandVersion(guildID, name); err != nil {
				slog.Warn("Failed to delete obsolete guild command version from DB", "guild_id", guildID, "name", name, "error", err)
			}
		}
	}

	// Register or update commands
	var failedCommands []string
	successCount := 0
	skipCount := 0

	for _, cmd := range allCommands {
		def := cmd.Definition()
		localVer, err := mgr.GetLocalCommandVersionAndBuildTime(def.Name)
		if err != nil {
			slog.Error("Failed to get local command version", "name", def.Name, "error", err)
			continue
		}

		guildVerTime, err := mgr.GetGuildCommandLastVersionTime(guildID, def.Name)
		if err != nil {
			slog.Error("Failed to get guild command version time", "guild_id", guildID, "name", def.Name, "error", err)
			// Treat error as needing update
			guildVerTime = nil
		}

		needsUpdate := force || guildVerTime == nil || localVer.BuildTime.After(*guildVerTime)

		if !needsUpdate {
			slog.Debug("Skipping command update (already up to date)", "name", def.Name, "guild_id", guildID)
			skipCount++
			continue
		}

		existing, exists := existingMap[def.Name]
		var syncErr error
		if exists {
			slog.Info("Updating command", "name", def.Name, "guild_id", guildID)
			_, syncErr = session.ApplicationCommandEdit(session.State.User.ID, guildID, existing.ID, def)
		} else {
			slog.Info("Creating new command", "name", def.Name, "guild_id", guildID)
			_, syncErr = session.ApplicationCommandCreate(session.State.User.ID, guildID, def)
		}

		if syncErr != nil {
			slog.Error("Failed to sync command", "name", def.Name, "guild_id", guildID, "error", syncErr)
			failedCommands = append(failedCommands, fmt.Sprintf("%s: %v", def.Name, syncErr))
		} else {
			successCount++
			// Update guild version time
			if err := mgr.UpdateGuildCommandLastVersionTime(guildID, def.Name, time.Now().UTC()); err != nil {
				slog.Error("Failed to update guild command version time", "guild_id", guildID, "name", def.Name, "error", err)
			}
		}
	}

	slog.Info("Command sync completed",
		"guild_id", guildID,
		"total", len(allCommands),
		"succeeded", successCount,
		"failed", len(failedCommands),
		"skipped", skipCount)

	if len(failedCommands) > 0 {
		errMsg := fmt.Errorf("failed to sync %d command(s), succeeded: %d/%d. Failed commands: %v",
			len(failedCommands), successCount, len(allCommands), failedCommands)
		slog.Error("Command sync had failures", "guild_id", guildID, "error", errMsg)
		return errMsg
	}

	return nil
}

// isVersionNewer compares two semver strings (x.y.z)
func isVersionNewer(newVer string, oldVer string) bool {
	if oldVer == "" {
		return true
	}
	if !semver.IsValid(oldVer) {
		return true
	}

	return semver.Compare(newVer, oldVer) > 0
}
