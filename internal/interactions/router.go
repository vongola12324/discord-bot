package interactions

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// ComponentHandler handles button and select menu interactions
type ComponentHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error

// ModalHandler handles modal submission interactions
type ModalHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error

// Router manages interaction handlers for components and modals
type Router struct {
	components map[string]ComponentHandler // customID prefix -> handler
	modals     map[string]ModalHandler     // customID prefix -> handler
	mu         sync.RWMutex
}

var instance *Router
var once sync.Once

// GetRouter returns the singleton interaction router
func GetRouter() *Router {
	once.Do(func() {
		instance = &Router{
			components: make(map[string]ComponentHandler),
			modals:     make(map[string]ModalHandler),
		}
	})
	return instance
}

// RegisterComponent registers a component interaction handler
func (r *Router) RegisterComponent(prefix string, handler ComponentHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.components[prefix] = handler
	slog.Info("Registered component handler", "prefix", prefix)
}

// RegisterModal registers a modal interaction handler
func (r *Router) RegisterModal(prefix string, handler ModalHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.modals[prefix] = handler
	slog.Info("Registered modal handler", "prefix", prefix)
}

// HandleComponent handles a component interaction
func (r *Router) HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find matching handler by prefix
	for prefix, handler := range r.components {
		if strings.HasPrefix(customID, prefix) {
			if err := handler(s, i); err != nil {
				slog.Error("Error handling component interaction",
					"customID", customID,
					"prefix", prefix,
					"error", err)
			}
			return
		}
	}

	slog.Warn("No handler found for component interaction", "customID", customID)
}

// HandleModal handles a modal submission
func (r *Router) HandleModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.ModalSubmitData().CustomID

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find matching handler by prefix
	for prefix, handler := range r.modals {
		if strings.HasPrefix(customID, prefix) {
			if err := handler(s, i); err != nil {
				slog.Error("Error handling modal interaction",
					"customID", customID,
					"prefix", prefix,
					"error", err)
			}
			return
		}
	}

	slog.Warn("No handler found for modal interaction", "customID", customID)
}
