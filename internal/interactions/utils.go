package interactions

import (
	"hiei-discord-bot/internal/i18n"

	"github.com/bwmarrin/discordgo"
)

// RespondError sends an error response to an interaction
func RespondError(s *discordgo.Session, i *discordgo.InteractionCreate, locale i18n.SupportedLocale, messageKey string, ephemeral bool, args ...interface{}) error {
	data := &discordgo.InteractionResponseData{
		Content: i18n.T(locale, "common.error_prefix") + " " + i18n.Tf(locale, messageKey, args...),
	}
	if ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

// RespondSuccess sends a success response to an interaction
func RespondSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, locale i18n.SupportedLocale, messageKey string, ephemeral bool, args ...interface{}) error {
	data := &discordgo.InteractionResponseData{
		Content: i18n.T(locale, "common.success_prefix") + " " + i18n.Tf(locale, messageKey, args...),
	}
	if ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

// RespondCustom sends a custom response to an interaction
func RespondCustom(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}
