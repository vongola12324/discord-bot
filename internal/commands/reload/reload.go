package reload

import (
	"fmt"
	"hiei-discord-bot/internal/commands"
	"hiei-discord-bot/internal/i18n"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.Register(New())
}

// Command implements the reload slash command
type Command struct{}

// New creates a new reload command instance
func New() *Command {
	return &Command{}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "reload",
		Description: "Reload all slash commands (admin only)",
	}
}

// Version returns the command version
func (c *Command) Version() string {
	return "1.0.0"
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
	definitions := commands.Global().GetDefinitions()
	reloadErr := commands.SyncGuildCommands(s, commands.Global(), guildID, true)
	if reloadErr != nil {
		// Send error message as follow-up
		errorMsg := fmt.Sprintf("%s %s", i18n.T(locale, "common.error_prefix"), i18n.Tf(locale, "command.reload.failed", reloadErr))
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(errorMsg),
		})
		return reloadErr
	}

	cmdCount := len(definitions)

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
