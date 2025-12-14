package bullsandcows

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"hiei-discord-bot/internal/interactions"
)

func init() {
	router := interactions.GetRouter()

	// Register component handler (buttons)
	router.RegisterComponent("game_bullsandcows_", handleComponentInteraction)

	// Register modal handler
	router.RegisterModal("game_bullsandcows_modal_", HandleModalSubmit)
}

// handleComponentInteraction routes button interactions to the appropriate handler
func handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID

	// Extract action from customID: game_bullsandcows_{action}_{userID}
	parts := strings.Split(customID, "_")
	if len(parts) < 4 {
		return nil
	}

	action := parts[2] // "guess" or "giveup"
	return HandleButtonClick(s, i, action)
}
