package events

import (
	"hiei-discord-bot/internal/commands"

	"github.com/bwmarrin/discordgo"
)

// Handler manages all Discord event handlers
type Handler struct {
	session  *discordgo.Session
	registry *commands.Registry
}

// NewHandler creates a new event handler
func NewHandler(session *discordgo.Session, registry *commands.Registry) *Handler {
	return &Handler{
		session:  session,
		registry: registry,
	}
}

// Register registers all event handlers
func (h *Handler) Register() {
	h.session.AddHandler(h.OnGuildCreate)
	h.session.AddHandler(h.OnGuildDelete)
}
