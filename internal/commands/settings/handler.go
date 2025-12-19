package settings

import (
	"fmt"
	"hiei-discord-bot/internal/i18n"
	"hiei-discord-bot/internal/interactions"
	"hiei-discord-bot/internal/settings"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	router := interactions.GetRouter()
	router.RegisterComponent("setting_group_select", handleGroupSelect)
	router.RegisterComponent("setting_item_select", handleItemSelect)
	router.RegisterComponent("setting_value_select", handleValueSelect)
	router.RegisterComponent("setting_channel_select", handleChannelSelect)
	router.RegisterComponent("setting_back", handleBack)
	router.RegisterModal("setting_value_modal", handleModalSubmit)
}

func handleGroupSelect(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	module := data.Values[0]
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: GetGroupPageData(s, i, module),
	})
}

func handleBack(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	parts := strings.Split(customID, ":")

	// setting_back -> main page
	// setting_back:module -> group page
	if len(parts) == 1 {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: GetMainPageData(i),
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: GetGroupPageData(s, i, parts[1]),
	})
}

func GetGroupPageData(s *discordgo.Session, i *discordgo.InteractionCreate, module string) *discordgo.InteractionResponseData {
	locale := i18n.GetUserLocaleFromInteraction(i)
	mgr := settings.GetManager()
	defs := mgr.GetDefinitions()

	var moduleDefs []settings.SettingDefinition
	for _, def := range defs {
		if def.Module == module {
			moduleDefs = append(moduleDefs, def)
		}
	}

	sort.Slice(moduleDefs, func(i, j int) bool {
		return moduleDefs[i].Key < moduleDefs[j].Key
	})

	var options []discordgo.SelectMenuOption
	var embedFields []*discordgo.MessageEmbedField

	userID := ""
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	for _, def := range moduleDefs {
		id := userID
		if def.Scope == settings.ScopeGuild {
			id = i.GuildID
		}

		val, _ := mgr.GetSettingValue(def.Scope, id, def.Key)

		// If it's a channel, try to get the channel name
		displayVal := fmt.Sprintf("%v", val)
		if def.Type == settings.TypeChannel && displayVal != "" {
			if ch, err := s.Channel(displayVal); err == nil {
				displayVal = "#" + ch.Name
			}
		}

		options = append(options, discordgo.SelectMenuOption{
			Label: i18n.T(locale, def.LabelKey),
			Value: def.Key,
		})

		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name:   i18n.T(locale, def.LabelKey),
			Value:  fmt.Sprintf("`%v`", displayVal),
			Inline: true,
		})
	}

	title := i18n.T(locale, "setting.title") + " > " + i18n.T(locale, fmt.Sprintf("setting.module.%s", module))

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:  title,
				Fields: embedFields,
				Color:  0x00ff00,
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "setting_item_select",
						Options:     options,
						Placeholder: i18n.T(locale, "setting.select_item_placeholder"),
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    i18n.T(locale, "setting.back"),
						Style:    discordgo.SecondaryButton,
						CustomID: "setting_back",
					},
				},
			},
		},
	}
}

func handleItemSelect(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	key := i.MessageComponentData().Values[0]
	locale := i18n.GetUserLocaleFromInteraction(i)
	mgr := settings.GetManager()

	var targetDef settings.SettingDefinition
	for _, def := range mgr.GetDefinitions() {
		if def.Key == key {
			targetDef = def
			break
		}
	}

	breadcrumb := i18n.T(locale, "setting.title") + " > " + i18n.T(locale, fmt.Sprintf("setting.module.%s", targetDef.Module)) + " > " + i18n.T(locale, targetDef.LabelKey)

	if targetDef.Type == settings.TypeSelect {
		var options []discordgo.SelectMenuOption
		for _, opt := range targetDef.Options {
			options = append(options, discordgo.SelectMenuOption{
				Label: opt,
				Value: opt,
			})
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: breadcrumb + "\n\n" + i18n.T(locale, "setting.select_value_desc"),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID: "setting_value_select:" + key,
								Options:  options,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    i18n.T(locale, "setting.back"),
								Style:    discordgo.SecondaryButton,
								CustomID: "setting_back:" + targetDef.Module,
							},
						},
					},
				},
			},
		})
	}

	if targetDef.Type == settings.TypeChannel {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: breadcrumb + "\n\n" + i18n.T(locale, "setting.select_value_desc"),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:     "setting_channel_select:" + key,
								MenuType:     discordgo.ChannelSelectMenu,
								ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    i18n.T(locale, "setting.back"),
								Style:    discordgo.SecondaryButton,
								CustomID: "setting_back:" + targetDef.Module,
							},
						},
					},
				},
			},
		})
	}

	// For String/Int, show Modal
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "setting_value_modal:" + key,
			Title:    i18n.T(locale, targetDef.LabelKey),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "value",
							Label:       i18n.T(locale, "setting.input_value_label"),
							Style:       discordgo.TextInputShort,
							Placeholder: fmt.Sprintf("%v", targetDef.Default),
							Required:    true,
						},
					},
				},
			},
		},
	})
}

func handleValueSelect(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	key := strings.Split(customID, ":")[1]
	value := i.MessageComponentData().Values[0]
	return updateSetting(s, i, key, value)
}

func handleChannelSelect(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	key := strings.Split(customID, ":")[1]
	value := i.MessageComponentData().Values[0]
	return updateSetting(s, i, key, value)
}

func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.ModalSubmitData().CustomID
	key := strings.Split(customID, ":")[1]
	value := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	return updateSetting(s, i, key, value)
}

func updateSetting(s *discordgo.Session, i *discordgo.InteractionCreate, key, value string) error {
	mgr := settings.GetManager()

	var targetDef settings.SettingDefinition
	for _, def := range mgr.GetDefinitions() {
		if def.Key == key {
			targetDef = def
			break
		}
	}

	targetID := ""
	if i.Member != nil {
		targetID = i.Member.User.ID
	} else if i.User != nil {
		targetID = i.User.ID
	}

	if targetDef.Scope == settings.ScopeGuild {
		targetID = i.GuildID
	}

	if err := mgr.SetSettingValue(targetDef.Scope, targetID, key, value); err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// After update, return to group page
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: GetGroupPageData(s, i, targetDef.Module),
	})
}
