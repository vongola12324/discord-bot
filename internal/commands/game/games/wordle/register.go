package wordle

import (
	"hiei-discord-bot/internal/commands/game"
	"hiei-discord-bot/internal/interactions"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	router := interactions.GetRouter()

	// Register component handler (buttons)
	router.RegisterComponent("game_wordle_", handleComponentInteraction)

	// Register modal handler
	router.RegisterModal("game_wordle_modal_", HandleModalSubmit)

	// Register as game subcommand
	game.RegisterSubCommand(&SubCommand{})
}

// SubCommand implements game.SubCommand interface
type SubCommand struct{}

func (s *SubCommand) Name() string {
	return "wordle"
}

func (s *SubCommand) Description() string {
	return "Play the Wordle word guessing game"
}

func (s *SubCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "length",
			Description: "Word length (3-10, default: 5)",
			Required:    false,
			MinValue:    float64Ptr(3),
			MaxValue:    10,
		},
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func (s *SubCommand) Handle(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return HandleStart(session, i)
}

// handleComponentInteraction routes button interactions to the appropriate handler
func handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID

	// Extract action from customID: game_wordle_{action}_{userID}
	parts := strings.Split(customID, "_")
	if len(parts) < 4 {
		return nil
	}

	action := parts[2] // "guess" or "giveup"
	return HandleButtonClick(s, i, action)
}
