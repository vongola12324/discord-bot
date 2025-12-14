package wordle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"hiei-discord-bot/internal/i18n"
)

const (
	randomWordAPIURL = "https://random-word-api.herokuapp.com/word?length=%d"
)

// HandleStart handles the initial /game wordle command
func HandleStart(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	locale := i18n.GetUserLocaleFromInteraction(i)
	store := GetStore()

	// Extract user ID
	userID := getUserID(i)
	if userID == "" {
		return fmt.Errorf("could not get user ID")
	}

	// Check if user already has an active game
	if store.HasActiveGame(userID) {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: i18n.T(locale, "game.wordle.error.already_active"),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get word length from options (default: 5)
	wordLength := 5
	options := i.ApplicationCommandData().Options
	if len(options) > 0 {
		for _, opt := range options[0].Options {
			if opt.Name == "length" {
				wordLength = int(opt.IntValue())
			}
		}
	}

	// Validate word length (3-10)
	if wordLength < 3 || wordLength > 10 {
		wordLength = 5
	}

	// Defer the response to buy time for API call
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}

	// Fetch random word from API
	answer, err := fetchRandomWord(wordLength)
	if err != nil {
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("❌ Failed to fetch word: %v", err)),
		})
		return err
	}

	// Create initial message
	content := buildGameMessage(locale, wordLength, []Guess{})

	// Send the initial message
	resp, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    i18n.T(locale, "game.wordle.button.guess"),
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("game_wordle_guess_%s", userID),
					},
					discordgo.Button{
						Label:    i18n.T(locale, "game.wordle.button.giveup"),
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("game_wordle_giveup_%s", userID),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// Start the game
	store.StartGame(userID, resp.ID, strings.ToUpper(answer), wordLength)

	return nil
}

// HandleButtonClick handles button interactions (guess/giveup)
func HandleButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate, action string) error {
	locale := i18n.GetUserLocaleFromInteraction(i)
	store := GetStore()

	userID := getUserID(i)
	if userID == "" {
		return fmt.Errorf("could not get user ID")
	}

	game, exists := store.GetGame(userID)
	if !exists {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: i18n.T(locale, "game.wordle.error.no_active_game"),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	switch action {
	case "guess":
		// Show modal for user to enter their guess
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: fmt.Sprintf("game_wordle_modal_%s", userID),
				Title:    i18n.T(locale, "game.wordle.modal.title"),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "guess_input",
								Label:       i18n.Tf(locale, "game.wordle.modal.input_label", game.WordLength),
								Style:       discordgo.TextInputShort,
								Placeholder: i18n.T(locale, "game.wordle.modal.input_placeholder"),
								Required:    true,
								MaxLength:   game.WordLength,
								MinLength:   game.WordLength,
							},
						},
					},
				},
			},
		})

	case "giveup":
		store.EndGame(userID)

		// Build final message with answer revealed
		content := buildGameMessage(locale, game.WordLength, game.Guesses)
		content += "\n\n" + i18n.T(locale, "game.wordle.result.giveup")
		content += "\n" + i18n.Tf(locale, "game.wordle.answer", game.Answer)

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    content,
				Components: []discordgo.MessageComponent{}, // Remove buttons
			},
		})
	}

	return nil
}

// HandleModalSubmit handles modal submissions (user's guess)
func HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	locale := i18n.GetUserLocaleFromInteraction(i)
	store := GetStore()

	userID := getUserID(i)
	if userID == "" {
		return fmt.Errorf("could not get user ID")
	}

	game, exists := store.GetGame(userID)
	if !exists {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: i18n.T(locale, "game.wordle.error.no_active_game"),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get the guess from modal
	data := i.ModalSubmitData()
	guess := strings.ToUpper(strings.TrimSpace(data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value))

	// Validate guess
	if len(guess) != game.WordLength {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: i18n.Tf(locale, "game.wordle.error.invalid_length", game.WordLength),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if all characters are letters
	for _, char := range guess {
		if !unicode.IsLetter(char) {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: i18n.T(locale, "game.wordle.error.not_alpha"),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	}

	// Add guess to game
	game.AddGuess(guess)

	// Check if game is won or lost
	var content string
	var components []discordgo.MessageComponent

	if game.IsWon() {
		store.EndGame(userID)
		content = buildGameMessage(locale, game.WordLength, game.Guesses)
		content += "\n\n" + i18n.Tf(locale, "game.wordle.result.won", len(game.Guesses))
		content += "\n" + i18n.Tf(locale, "game.wordle.answer", game.Answer)
		components = []discordgo.MessageComponent{} // Remove buttons
	} else if game.IsLost() {
		store.EndGame(userID)
		content = buildGameMessage(locale, game.WordLength, game.Guesses)
		content += "\n\n" + i18n.T(locale, "game.wordle.result.lost")
		content += "\n" + i18n.Tf(locale, "game.wordle.answer", game.Answer)
		components = []discordgo.MessageComponent{} // Remove buttons
	} else {
		// Game continues
		content = buildGameMessage(locale, game.WordLength, game.Guesses)
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    i18n.T(locale, "game.wordle.button.guess"),
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("game_wordle_guess_%s", userID),
					},
					discordgo.Button{
						Label:    i18n.T(locale, "game.wordle.button.giveup"),
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("game_wordle_giveup_%s", userID),
					},
				},
			},
		}
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: components,
		},
	})
}

// buildGameMessage builds the game state message
func buildGameMessage(locale i18n.SupportedLocale, wordLength int, guesses []Guess) string {
	var builder strings.Builder

	// Title
	builder.WriteString(i18n.Tf(locale, "game.wordle.title_with_length", wordLength))
	builder.WriteString("\n\n")

	// Attempts
	builder.WriteString(i18n.Tf(locale, "game.wordle.attempts", len(guesses)))
	builder.WriteString("\n\n")

	// History
	builder.WriteString(i18n.T(locale, "game.wordle.history"))
	builder.WriteString("\n")

	if len(guesses) == 0 {
		builder.WriteString(i18n.T(locale, "game.wordle.no_guesses"))
	} else {
		for _, guess := range guesses {
			builder.WriteString(formatGuess(guess))
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// formatGuess formats a guess with colored blocks and the word
func formatGuess(guess Guess) string {
	var builder strings.Builder

	// Add colored blocks
	for _, result := range guess.Results {
		switch result {
		case CorrectPosition:
			builder.WriteString("🟩")
		case WrongPosition:
			builder.WriteString("🟨")
		case NotInWord:
			builder.WriteString("🟥")
		}
	}

	builder.WriteString(" `")
	builder.WriteString(guess.Word)
	builder.WriteString("`")

	return builder.String()
}

// fetchRandomWord fetches a random word from the API
func fetchRandomWord(length int) (string, error) {
	url := fmt.Sprintf(randomWordAPIURL, length)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch word: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var words []string
	if err := json.Unmarshal(body, &words); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(words) == 0 {
		return "", fmt.Errorf("no words returned from API")
	}

	return words[0], nil
}

// getUserID extracts user ID from interaction
func getUserID(i *discordgo.InteractionCreate) string {
	if i == nil {
		return ""
	}

	// Guild interaction (Member takes priority)
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}

	// DM interaction
	if i.User != nil {
		return i.User.ID
	}

	return ""
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}
