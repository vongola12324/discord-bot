package blame

import (
	"hiei-discord-bot/internal/interactions"
)

func init() {
	router := interactions.GetRouter()

	// Register modal handler for blame reason input
	router.RegisterModal("blame_modal_", HandleModalSubmit)
}
