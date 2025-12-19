package preferences

import (
	"fmt"
	"hiei-discord-bot/internal/commands"
	"hiei-discord-bot/internal/i18n"
	"hiei-discord-bot/internal/interactions"
	"hiei-discord-bot/internal/settings"
	"sort"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.Register(New())
}

type Command struct{}

func New() *Command {
	return &Command{}
}

func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "settings",
		Description: "Adjust bot settings for this server or yourself",
	}
}

func (c *Command) Version() string {
	return "1.0.2"
}

func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := GetMainPageData(i)
	if data == nil {
		locale := i18n.GetUserLocaleFromInteraction(i)
		return interactions.RespondError(s, i, locale, "setting.no_available_settings", true)
	}

	return interactions.RespondCustom(s, i, data)
}

func GetMainPageData(i *discordgo.InteractionCreate) *discordgo.InteractionResponseData {
	locale := i18n.GetUserLocaleFromInteraction(i)
	mgr := settings.GetManager()
	defs := mgr.GetDefinitions()

	// Check permissions
	isAdmin := false
	if i.Member != nil {
		isAdmin = i.Member.Permissions&discordgo.PermissionAdministrator != 0
	}

	// Group definitions by module
	groups := make(map[string][]settings.SettingDefinition)
	for _, def := range defs {
		if def.Scope == settings.ScopeGuild && !isAdmin {
			continue
		}
		groups[def.Module] = append(groups[def.Module], def)
	}

	if len(groups) == 0 {
		return nil
	}

	// Create group select menu
	var options []discordgo.SelectMenuOption
	var moduleNames []string
	for name := range groups {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)

	for _, name := range moduleNames {
		options = append(options, discordgo.SelectMenuOption{
			Label: i18n.T(locale, fmt.Sprintf("setting.module.%s", name)),
			Value: name,
		})
	}

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       i18n.T(locale, "setting.title"),
				Description: i18n.T(locale, "setting.select_group_desc"),
				Color:       0x00ff00,
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "setting_group_select",
						Options:     options,
						Placeholder: i18n.T(locale, "setting.select_group_placeholder"),
					},
				},
			},
		},
		Flags: discordgo.MessageFlagsEphemeral,
	}
}
