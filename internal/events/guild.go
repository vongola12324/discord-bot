package events

import (
	"hiei-discord-bot/internal/commands"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// OnGuildCreate handles when the bot joins a guild or when the bot starts up
func (h *Handler) OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	slog.Info("Bot joined guild",
		"guild_id", g.ID,
		"guild_name", g.Name,
		"member_count", g.MemberCount,
		"owner_id", g.OwnerID,
	)

	if err := commands.SyncGuildCommands(s, h.registry, g.ID, false); err != nil {
		slog.Error("Failed to sync commands for new guild",
			"guild_id", g.ID,
			"guild_name", g.Name,
			"error", err,
		)
	} else {
		slog.Info("Synced commands for new guild",
			"guild_id", g.ID,
			"guild_name", g.Name,
		)
	}
}

// OnGuildDelete handles when the bot is removed from a guild
func (h *Handler) OnGuildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	slog.Info("Bot removed from guild",
		"guild_id", g.ID,
	)
}
