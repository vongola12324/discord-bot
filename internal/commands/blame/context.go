package blame

import (
	"bytes"
	"fmt"
	"io"

	"hiei-discord-bot/internal/i18n"
	"hiei-discord-bot/resources"

	"github.com/bwmarrin/discordgo"
)

// ContextCommand implements the blame context menu command
type ContextCommand struct{}

// NewContext creates a new blame context menu command instance
func NewContext() *ContextCommand {
	return &ContextCommand{}
}

// Definition returns the context menu command definition
func (c *ContextCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:    "Blame",
		Type:    discordgo.MessageApplicationCommand,
		Version: "0.0.1",
	}
}

// Execute runs the context menu command
func (c *ContextCommand) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	locale := i18n.GetUserLocaleFromInteraction(i)

	// Get the target message from the context menu interaction
	var targetMessage *discordgo.Message
	if i.Type == discordgo.InteractionApplicationCommand {
		data := i.ApplicationCommandData()
		if data.TargetID != "" {
			// Fetch the target message
			msg, err := s.ChannelMessage(i.ChannelID, data.TargetID)
			if err != nil {
				return respondError(s, i, locale, "blame.error.message_not_found")
			}
			targetMessage = msg
		}
	}

	if targetMessage == nil || targetMessage.Author == nil {
		return respondError(s, i, locale, "blame.error.no_target")
	}
	// Show modal to get the reason
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("blame_modal_%s_%s", targetMessage.ID, targetMessage.Author.ID),
			Title:    i18n.T(locale, "blame.modal.title"),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "blame_reason",
							Label:       i18n.T(locale, "blame.modal.reason_label"),
							Style:       discordgo.TextInputShort,
							Placeholder: i18n.T(locale, "blame.modal.reason_placeholder"),
							Required:    true,
							MaxLength:   200,
							MinLength:   1,
						},
					},
				},
			},
		},
	})
}

// HandleModalSubmit handles the modal submission for blame reason
func HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	locale := i18n.GetUserLocaleFromInteraction(i)

	// Parse CustomID: blame_modal_{messageID}_{targetUserID}
	customID := i.ModalSubmitData().CustomID
	var messageID, targetUserID string
	fmt.Sscanf(customID, "blame_modal_%s_%s", &messageID, &targetUserID)

	if messageID == "" || targetUserID == "" {
		return respondError(s, i, locale, "blame.error.invalid_data")
	}

	// Get the reason from modal
	data := i.ModalSubmitData()
	var reason string
	for _, row := range data.Components {
		if actionRow, ok := row.(*discordgo.ActionsRow); ok {
			for _, component := range actionRow.Components {
				if textInput, ok := component.(*discordgo.TextInput); ok && textInput.CustomID == "blame_reason" {
					reason = textInput.Value
				}
			}
		}
	}

	if reason == "" {
		return respondError(s, i, locale, "blame.error.no_reason")
	}

	// Get target user
	targetUser, err := s.User(targetUserID)
	if err != nil {
		return respondError(s, i, locale, "blame.error.no_target")
	}

	// Get the user who issued the blame
	var blamer *discordgo.User
	if i.Member != nil && i.Member.User != nil {
		blamer = i.Member.User
	} else if i.User != nil {
		blamer = i.User
	}

	// Defer response
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}

	// Read the blame image
	imageData, err := resources.Images.Open(resources.ImagesBasePath + "/blame.jpg")
	if err != nil {
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(i18n.T(locale, "blame.error.image_load_failed")),
		})
		return err
	}
	defer imageData.Close()

	imageBytes, err := io.ReadAll(imageData)
	if err != nil {
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(i18n.T(locale, "blame.error.image_load_failed")),
		})
		return err
	}

	// Build the message content
	content := buildBlameMessageWithBlamer(locale, targetUser, reason, blamer)

	// Delete the deferred response
	err = s.InteractionResponseDelete(i.Interaction)
	if err != nil {
		return err
	}

	// Send the blame message as a reply to the target message
	_, err = s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Files: []*discordgo.File{
			{
				Name:        "blame.jpg",
				ContentType: "image/jpeg",
				Reader:      bytes.NewReader(imageBytes),
			},
		},
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       i18n.T(locale, "blame.title"),
				Description: content,
				Color:       0xFF0000, // Red color for condemnation
			},
		},
		Reference: &discordgo.MessageReference{
			MessageID: messageID,
			ChannelID: i.ChannelID,
		},
	})

	return err
}

// buildBlameMessageWithBlamer constructs the blame message with blamer info
func buildBlameMessageWithBlamer(locale i18n.SupportedLocale, target *discordgo.User, reason string, blamer *discordgo.User) string {
	blamedBy := ""
	if blamer != nil {
		blamedBy = i18n.Tf(locale, "blame.blamed_by", blamer.Mention())
	}

	return fmt.Sprintf(
		"%s %s\n%s\n#譴責 #責任全在你方\n\n%s",
		target.Mention(),
		reason,
		i18n.T(locale, "blame.message"),
		blamedBy,
	)
}
