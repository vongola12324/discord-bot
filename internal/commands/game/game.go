package game

import (
	"hiei-discord-bot/internal/commands/game/games/bullsandcows"
	"hiei-discord-bot/internal/i18n"

	"github.com/bwmarrin/discordgo"
)

// Command implements the game command with subcommands
type Command struct{}

// New creates a new game command instance
func New() *Command {
	return &Command{}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "game",
		Description: "Play various games",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "bullsandcows",
				Description: "Play the 1A2B number guessing game",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "difficulty",
						Description: "Game difficulty (easy: unique digits, hard: repeating digits allowed)",
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
				},
			},
		},
	}
}

// Execute runs the game command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get user locale
	locale := i18n.GetUserLocaleFromInteraction(i)

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return respondError(s, i, locale, "game.select_game")
	}

	subcommand := options[0]

	switch subcommand.Name {
	case "bullsandcows":
		return bullsandcows.HandleStart(s, i)
	default:
		return respondError(s, i, locale, "game.unknown_game")
	}
}

// respondError sends an error response
func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, locale i18n.SupportedLocale, messageKey string) error {
	content := i18n.T(locale, "common.error_prefix") + " " + i18n.T(locale, messageKey)
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
