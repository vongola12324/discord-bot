package i18n

import (
	"hiei-discord-bot/internal/settings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// SupportedLocale represents supported languages
type SupportedLocale string

const (
	LocaleAuto SupportedLocale = "auto"  // Auto detect
	LocaleZhTW SupportedLocale = "zh-TW" // Traditional Chinese
	LocaleEnUS SupportedLocale = "en-US" // English (US)
)

// LocaleStore manages user locale preferences
type LocaleStore struct {
	userLocales map[string]SupportedLocale // userID -> locale
}

var store *LocaleStore
var once sync.Once

// GetStore returns the singleton locale store
func GetStore() *LocaleStore {
	once.Do(func() {
		store = &LocaleStore{
			userLocales: make(map[string]SupportedLocale),
		}

		// Register language setting
		settings.GetManager().Register(settings.SettingDefinition{
			Key:      "language",
			Module:   "general",
			Scope:    settings.ScopeUser,
			Type:     settings.TypeSelect,
			Default:  string(LocaleAuto),
			Options:  []string{string(LocaleAuto), string(LocaleZhTW), string(LocaleEnUS)},
			LabelKey: "setting.general.language.label",
			DescKey:  "setting.general.language.desc",
		})
	})
	return store
}

// DetectAndStore detects user's locale and stores it if not already set
func (s *LocaleStore) DetectAndStore(userID string, discordLocale discordgo.Locale) SupportedLocale {
	mgr := settings.GetManager()
	val, err := mgr.GetSettingValue(settings.ScopeUser, userID, "language")
	if err == nil && val != nil {
		locale := SupportedLocale(val.(string))
		if locale != LocaleAuto {
			return locale
		}
	}

	// Detect and convert Discord locale to supported locale
	return convertToSupportedLocale(discordLocale)
}

// GetUserLocale returns the stored locale for a user
func (s *LocaleStore) GetUserLocale(userID string) (SupportedLocale, bool) {
	mgr := settings.GetManager()
	val, err := mgr.GetSettingValue(settings.ScopeUser, userID, "language")
	if err != nil || val == nil {
		return LocaleEnUS, false
	}
	return SupportedLocale(val.(string)), true
}

// SetUserLocale manually sets a user's locale preference
func (s *LocaleStore) SetUserLocale(userID string, locale SupportedLocale) {
	mgr := settings.GetManager()
	_ = mgr.SetSettingValue(settings.ScopeUser, userID, "language", string(locale))
}

// convertToSupportedLocale converts Discord locale to our supported locale
func convertToSupportedLocale(discordLocale discordgo.Locale) SupportedLocale {
	switch discordLocale {
	case discordgo.ChineseTW:
		return LocaleZhTW
	default:
		// Default to English for all other locales (including zh-CN)
		return LocaleEnUS
	}
}

// GetUserLocaleFromInteraction is a helper to get locale from interaction
func GetUserLocaleFromInteraction(i *discordgo.InteractionCreate) SupportedLocale {
	userID := ""
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	if userID == "" {
		return LocaleEnUS
	}

	store := GetStore()
	return store.DetectAndStore(userID, i.Locale)
}
