package i18n

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"hiei-discord-bot/resources"
)

// translations stores all loaded translations
var translations map[SupportedLocale]map[string]interface{}

// LoadTranslations loads all translation files
func LoadTranslations() error {
	translations = make(map[SupportedLocale]map[string]interface{})

	locales := []SupportedLocale{LocaleZhTW, LocaleEnUS}

	for _, locale := range locales {
		filename := fmt.Sprintf("%s/%s.json", resources.I18nBasePath, locale)
		data, err := resources.I18n.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", locale, err)
		}

		var translationData map[string]interface{}
		if err := json.Unmarshal(data, &translationData); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", locale, err)
		}

		translations[locale] = translationData
		slog.Info("Loaded translations", "locale", locale, "path", filename)
	}

	return nil
}

// T translates a key to the target locale
// Key format: "section.subsection.key" (e.g., "game.bullsandcows.title")
func T(locale SupportedLocale, key string) string {
	value := getNestedValue(translations[locale], key)
	if value != "" {
		return value
	}

	// Fallback to English
	value = getNestedValue(translations[LocaleEnUS], key)
	if value != "" {
		return value
	}

	// Return key if no translation found
	return key
}

// Tf translates a key with format arguments
func Tf(locale SupportedLocale, key string, args ...interface{}) string {
	template := T(locale, key)
	return fmt.Sprintf(template, args...)
}

// getNestedValue retrieves a value from nested map using dot notation
func getNestedValue(data map[string]interface{}, key string) string {
	parts := strings.Split(key, ".")
	current := data

	for i, part := range parts {
		if current == nil {
			return ""
		}

		value, ok := current[part]
		if !ok {
			return ""
		}

		// If this is the last part, return the string value
		if i == len(parts)-1 {
			if str, ok := value.(string); ok {
				return str
			}
			return ""
		}

		// Otherwise, navigate deeper
		if nested, ok := value.(map[string]interface{}); ok {
			current = nested
		} else {
			return ""
		}
	}

	return ""
}
