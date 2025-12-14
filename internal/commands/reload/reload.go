package reload

import (
	"fmt"

	"hiei-discord-bot/internal/i18n"

	"github.com/bwmarrin/discordgo"
)

// Command implements the reload slash command
type Command struct {
	reloadCallback func(guildID string) (int, error) // Returns command count and error
}

// New creates a new reload command instance
func New(reloadCallback func(guildID string) (int, error)) *Command {
	return &Command{
		reloadCallback: reloadCallback,
	}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "reload",
		Description: "Reload all slash commands (admin only)",
	}
}

// Execute runs the reload command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Extract locale from interaction
	locale := i18n.GetUserLocaleFromInteraction(i)

	// Get guild ID from interaction
	guildID := i.GuildID
	if guildID == "" {
		return fmt.Errorf("%s", i18n.T(locale, "command.reload.guild_only"))
	}

	// Respond immediately to acknowledge the interaction
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to defer response: %w", err)
	}

	// Perform the reload for this guild
	cmdCount, reloadErr := c.reloadCallback(guildID)
	if reloadErr != nil {
		// Send error message as follow-up
		errorMsg := fmt.Sprintf("%s %s", i18n.T(locale, "common.error_prefix"), i18n.Tf(locale, "command.reload.failed", reloadErr))
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(errorMsg),
		})
		return reloadErr
	}

	// Send success message as follow-up
	successMsg := fmt.Sprintf("%s %s", i18n.T(locale, "common.success_prefix"), i18n.Tf(locale, "command.reload.success", cmdCount))
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: strPtr(successMsg),
	})

	return err
}

// strPtr is a helper function to get a string pointer
func strPtr(s string) *string {
	return &s
}
