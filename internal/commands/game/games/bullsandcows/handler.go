package bullsandcows

import (
	"fmt"
	"strings"

	"hiei-discord-bot/internal/i18n"

	"github.com/bwmarrin/discordgo"
)

// HandleStart starts a new Bulls and Cows game or shows game info
func HandleStart(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	userID := i.Member.User.ID
	manager := GetManager()
	locale := i18n.GetUserLocaleFromInteraction(i)

	// Get subcommand options (bullsandcows options)
	subcommandOptions := i.ApplicationCommandData().Options
	if len(subcommandOptions) == 0 {
		return showGameInfo(s, i)
	}

	// Get difficulty parameter from subcommand
	options := subcommandOptions[0].Options

	// If no difficulty specified, show game info
	if len(options) == 0 {
		return showGameInfo(s, i)
	}

	difficulty := Difficulty(options[0].StringValue())

	// Check if user already has an active game
	if _, exists := manager.GetGame(userID); exists {
		return respondError(s, i, locale, i18n.T(locale, "game.bullsandcows.error.already_active"))
	}

	// Start new game
	game := manager.StartGame(userID, difficulty, locale)

	// Create game message with buttons
	message := buildGameMessage(game, false)
	message.Components = buildGameButtons(userID, game.Locale)

	// Send initial response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: message,
	})
}

// showGameInfo displays game rules and difficulty information
func showGameInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	locale := i18n.GetUserLocaleFromInteraction(i)
	content := buildGameInfoContent(locale)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// buildGameInfoContent builds the complete game info message
func buildGameInfoContent(locale i18n.SupportedLocale) string {
	var builder strings.Builder

	// Title
	builder.WriteString(i18n.T(locale, "game.bullsandcows.title") + "\n\n")

	// How to play section
	buildInfoSection(&builder, locale, []string{
		"game.bullsandcows.info.how_to_play",
		"game.bullsandcows.info.step1",
		"game.bullsandcows.info.step2",
		"game.bullsandcows.info.step3",
		"game.bullsandcows.info.step3_a",
		"game.bullsandcows.info.step3_b",
		"game.bullsandcows.info.step4",
	})
	builder.WriteString("\n")

	// Difficulty levels section
	buildInfoSection(&builder, locale, []string{
		"game.bullsandcows.info.difficulty_levels",
		"game.bullsandcows.info.easy_mode",
		"game.bullsandcows.info.easy_mode_rule1",
		"game.bullsandcows.info.easy_mode_rule2",
	})
	builder.WriteString("\n")

	buildInfoSection(&builder, locale, []string{
		"game.bullsandcows.info.hard_mode",
		"game.bullsandcows.info.hard_mode_rule1",
		"game.bullsandcows.info.hard_mode_rule2",
	})
	builder.WriteString("\n")

	// Example section
	buildInfoSection(&builder, locale, []string{
		"game.bullsandcows.info.example",
		"game.bullsandcows.info.example_secret",
		"game.bullsandcows.info.example_guess",
		"game.bullsandcows.info.example_explain",
	})
	builder.WriteString("\n")

	// Ready to play section
	buildInfoSection(&builder, locale, []string{
		"game.bullsandcows.info.ready",
		"game.bullsandcows.info.start_command",
	})

	return builder.String()
}

// buildInfoSection writes a group of translation keys to the builder
func buildInfoSection(builder *strings.Builder, locale i18n.SupportedLocale, keys []string) {
	for _, key := range keys {
		builder.WriteString(i18n.T(locale, key) + "\n")
	}
}

// HandleButtonClick handles button interactions (guess/give up)
func HandleButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate, action string) error {
	userID := i.Member.User.ID
	manager := GetManager()

	game, exists := manager.GetGame(userID)
	if !exists {
		locale := i18n.GetUserLocaleFromInteraction(i)
		return respondError(s, i, locale, i18n.T(locale, "game.bullsandcows.error.no_active_game"))
	}

	switch action {
	case "guess":
		// Show modal for guess input
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: fmt.Sprintf("game_bullsandcows_modal_%s", userID),
				Title:    i18n.T(game.Locale, "game.bullsandcows.modal.title"),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "guess_input",
								Label:       i18n.T(game.Locale, "game.bullsandcows.modal.input_label"),
								Style:       discordgo.TextInputShort,
								Placeholder: i18n.T(game.Locale, "game.bullsandcows.modal.input_placeholder"),
								Required:    true,
								MaxLength:   4,
								MinLength:   4,
							},
						},
					},
				},
			},
		})

	case "giveup":
		manager.EndGame(userID)

		// Update message to show game over
		message := buildGameMessage(game, true)
		message.Content += "\n\n" + i18n.T(game.Locale, "game.bullsandcows.result.giveup")
		message.Components = []discordgo.MessageComponent{} // Remove buttons

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: message,
		})

	default:
		locale := i18n.GetUserLocaleFromInteraction(i)
		return respondError(s, i, locale, i18n.T(locale, "command.unknown"))
	}
}

// HandleModalSubmit handles the guess modal submission
func HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	userID := i.Member.User.ID
	manager := GetManager()

	game, exists := manager.GetGame(userID)
	if !exists {
		locale := i18n.GetUserLocaleFromInteraction(i)
		return respondError(s, i, locale, i18n.T(locale, "game.bullsandcows.error.no_active_game"))
	}

	// Extract guess from modal
	data := i.ModalSubmitData()
	guess := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	// Validate guess
	if !IsValidGuess(guess, game.Difficulty) {
		errorMsg := i18n.T(game.Locale, "game.bullsandcows.error.invalid_guess_hard")
		if game.Difficulty == DifficultyEasy {
			errorMsg = i18n.T(game.Locale, "game.bullsandcows.error.invalid_guess_easy")
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errorMsg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Process guess
	game.Attempts++
	bulls, cows := CheckGuess(game.Answer, guess)
	game.History = append(game.History, GuessResult{
		Guess: guess,
		Bulls: bulls,
		Cows:  cows,
	})

	// Check if won
	if bulls == 4 {
		manager.EndGame(userID)
		message := buildGameMessage(game, true)
		message.Content += "\n\n" + i18n.Tf(game.Locale, "game.bullsandcows.result.won", game.Attempts)
		message.Components = []discordgo.MessageComponent{} // Remove buttons

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: message,
		})
	}

	// Check if lost
	if game.Attempts >= game.MaxAttempts {
		manager.EndGame(userID)
		message := buildGameMessage(game, true)
		message.Content += "\n\n" + i18n.Tf(game.Locale, "game.bullsandcows.result.lost", game.MaxAttempts)
		message.Components = []discordgo.MessageComponent{} // Remove buttons

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: message,
		})
	}

	// Continue game - update message with new result
	message := buildGameMessage(game, false)
	message.Components = buildGameButtons(userID, game.Locale)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: message,
	})
}

// buildGameButtons creates the standard game action buttons
func buildGameButtons(userID string, locale i18n.SupportedLocale) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    i18n.T(locale, "game.bullsandcows.button.guess"),
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("game_bullsandcows_guess_%s", userID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎯",
					},
				},
				discordgo.Button{
					Label:    i18n.T(locale, "game.bullsandcows.button.giveup"),
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("game_bullsandcows_giveup_%s", userID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🏳️",
					},
				},
			},
		},
	}
}

// buildGameMessage builds the game status message
func buildGameMessage(game *GameState, gameOver bool) *discordgo.InteractionResponseData {
	var builder strings.Builder

	buildGameTitle(&builder, game)
	buildGameRules(&builder, game)
	buildGameProgress(&builder, game)
	buildGameHistory(&builder, game)

	if gameOver {
		buildGameAnswer(&builder, game)
	}

	return &discordgo.InteractionResponseData{
		Content: builder.String(),
		Flags:   discordgo.MessageFlagsEphemeral,
	}
}

// buildGameTitle writes the game title with difficulty indicator
func buildGameTitle(builder *strings.Builder, game *GameState) {
	locale := game.Locale
	difficultyEmoji := "🟢"
	difficultyName := i18n.T(locale, "game.bullsandcows.difficulty.easy")

	if game.Difficulty == DifficultyHard {
		difficultyEmoji = "🔴"
		difficultyName = i18n.T(locale, "game.bullsandcows.difficulty.hard")
	}

	builder.WriteString(i18n.Tf(locale, "game.bullsandcows.title_with_difficulty", difficultyEmoji, difficultyName))
	builder.WriteString("\n\n")
}

// buildGameRules writes the game rules section
func buildGameRules(builder *strings.Builder, game *GameState) {
	locale := game.Locale

	builder.WriteString(i18n.T(locale, "game.bullsandcows.rules.title") + "\n")

	if game.Difficulty == DifficultyEasy {
		builder.WriteString(i18n.T(locale, "game.bullsandcows.rules.unique_digits") + "\n")
	} else {
		builder.WriteString(i18n.T(locale, "game.bullsandcows.rules.repeating_digits") + "\n")
	}

	builder.WriteString(i18n.T(locale, "game.bullsandcows.rules.a") + "\n")
	builder.WriteString(i18n.T(locale, "game.bullsandcows.rules.b") + "\n\n")
}

// buildGameProgress writes the attempts counter
func buildGameProgress(builder *strings.Builder, game *GameState) {
	builder.WriteString(i18n.Tf(game.Locale, "game.bullsandcows.attempts", game.Attempts, game.MaxAttempts))
	builder.WriteString("\n\n")
}

// buildGameHistory writes the guess history
func buildGameHistory(builder *strings.Builder, game *GameState) {
	if len(game.History) > 0 {
		builder.WriteString(i18n.T(game.Locale, "game.bullsandcows.history") + "\n")
		for _, result := range game.History {
			builder.WriteString(fmt.Sprintf("`%s` → %dA%dB\n", result.Guess, result.Bulls, result.Cows))
		}
	} else {
		builder.WriteString(i18n.T(game.Locale, "game.bullsandcows.no_guesses") + "\n")
	}
}

// buildGameAnswer writes the answer (for game over)
func buildGameAnswer(builder *strings.Builder, game *GameState) {
	builder.WriteString("\n" + i18n.Tf(game.Locale, "game.bullsandcows.answer", game.Answer) + "\n")
}

// respondError sends an error response
func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, locale i18n.SupportedLocale, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: i18n.T(locale, "common.error_prefix") + " " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
