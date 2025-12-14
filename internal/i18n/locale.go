package i18n

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// SupportedLocale represents supported languages
type SupportedLocale string

const (
	LocaleZhTW SupportedLocale = "zh-TW" // Traditional Chinese
	LocaleEnUS SupportedLocale = "en-US" // English (US)
)

// LocaleStore manages user locale preferences
type LocaleStore struct {
	userLocales map[string]SupportedLocale // userID -> locale
	mu          sync.RWMutex
}

var store *LocaleStore
var once sync.Once

// GetStore returns the singleton locale store
func GetStore() *LocaleStore {
	once.Do(func() {
		store = &LocaleStore{
			userLocales: make(map[string]SupportedLocale),
		}
	})
	return store
}

// DetectAndStore detects user's locale and stores it if not already set
func (s *LocaleStore) DetectAndStore(userID string, discordLocale discordgo.Locale) SupportedLocale {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already has a stored locale
	if locale, exists := s.userLocales[userID]; exists {
		return locale
	}

	// Detect and convert Discord locale to supported locale
	supportedLocale := convertToSupportedLocale(discordLocale)
	s.userLocales[userID] = supportedLocale

	return supportedLocale
}

// GetUserLocale returns the stored locale for a user
func (s *LocaleStore) GetUserLocale(userID string) (SupportedLocale, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	locale, exists := s.userLocales[userID]
	return locale, exists
}

// SetUserLocale manually sets a user's locale preference
func (s *LocaleStore) SetUserLocale(userID string, locale SupportedLocale) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userLocales[userID] = locale
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
