package blame

import (
	"bytes"
	"fmt"
	"io"

	"hiei-discord-bot/internal/i18n"
	"hiei-discord-bot/resources"

	"github.com/bwmarrin/discordgo"
)

// Command implements the blame command
type Command struct{}

// New creates a new blame command instance
func New() *Command {
	return &Command{}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "blame",
		Description: "Severely condemn someone",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "target",
				Description: "The user to blame",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "The reason for blaming",
				Required:    true,
			},
		},
	}
}

// Execute runs the blame command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	locale := i18n.GetUserLocaleFromInteraction(i)
	options := i.ApplicationCommandData().Options

	// Extract parameters
	var targetUser *discordgo.User
	var reason string

	for _, opt := range options {
		switch opt.Name {
		case "target":
			targetUser = opt.UserValue(s)
		case "reason":
			reason = opt.StringValue()
		}
	}

	if targetUser == nil {
		return respondError(s, i, locale, "blame.error.no_target")
	}

	// Defer response to buy time for processing
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return err
	}

	// Read the blame image from embedded resources
	imageData, err := resources.Images.Open(resources.ImagesBasePath + "/blame.jpg")
	if err != nil {
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(i18n.T(locale, "blame.error.image_load_failed")),
		})
		return err
	}
	defer imageData.Close()

	// Read image data into memory
	imageBytes, err := io.ReadAll(imageData)
	if err != nil {
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(i18n.T(locale, "blame.error.image_load_failed")),
		})
		return err
	}

	// Build the message content
	content := buildBlameMessage(locale, targetUser, reason)

	// Send the full blame message (with image)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Files: []*discordgo.File{
			{
				Name:        "blame.jpg",
				ContentType: "image/jpeg",
				Reader:      bytes.NewReader(imageBytes),
			},
		},
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       i18n.T(locale, "blame.title"),
				Description: content,
				Color:       0xFF0000, // Red color for condemnation
			},
		},
	})

	return err
}

// buildBlameMessage constructs the blame message
func buildBlameMessage(locale i18n.SupportedLocale, target *discordgo.User, reason string) string {
	return fmt.Sprintf(
		"%s %s\n%s\n#譴責 #責任全在你方",
		target.Mention(),
		reason,
		i18n.T(locale, "blame.message"),
	)
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

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}
