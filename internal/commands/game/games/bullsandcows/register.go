package bullsandcows

import (
	"hiei-discord-bot/internal/commands/game"
	"hiei-discord-bot/internal/interactions"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	router := interactions.GetRouter()

	// Register component handler (buttons)
	router.RegisterComponent("game_bullsandcows_", handleComponentInteraction)

	// Register modal handler
	router.RegisterModal("game_bullsandcows_modal_", HandleModalSubmit)

	// Register as game subcommand
	game.RegisterSubCommand(&SubCommand{})
}

// SubCommand implements game.SubCommand interface
type SubCommand struct{}

func (s *SubCommand) Name() string {
	return "bullsandcows"
}

func (s *SubCommand) Description() string {
	return "Play the 1A2B number guessing game"
}

func (s *SubCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "difficulty",
			Description: "Game difficulty",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Easy (Unique digits)",
					Value: "easy",
				},
				{
					Name:  "Hard (Repeating digits allowed)",
					Value: "hard",
				},
			},
		},
	}
}

func (s *SubCommand) Handle(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return HandleStart(session, i)
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
