package bullsandcows

import (
	"math/rand"
	"sync"
	"time"

	"hiei-discord-bot/internal/i18n"
)

// Difficulty represents the game difficulty level
type Difficulty string

const (
	DifficultyEasy Difficulty = "easy" // Unique digits only
	DifficultyHard Difficulty = "hard" // Repeating digits allowed
)

// GameState represents a Bulls and Cows game session
type GameState struct {
	Answer      string
	Attempts    int
	MaxAttempts int
	History     []GuessResult
	MessageID   string                // Store the game message ID for updates
	Difficulty  Difficulty            // Game difficulty level
	Locale      i18n.SupportedLocale  // User's language preference
}

// GuessResult stores a single guess result
type GuessResult struct {
	Guess string
	Bulls int
	Cows  int
}

// Manager manages active Bulls and Cows game sessions
type Manager struct {
	games map[string]*GameState // userID -> GameState
	mu    sync.RWMutex
}

var instance *Manager
var once sync.Once

// GetManager returns the singleton game manager
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			games: make(map[string]*GameState),
		}
	})
	return instance
}

// StartGame starts a new game for a user
func (m *Manager) StartGame(userID string, difficulty Difficulty, locale i18n.SupportedLocale) *GameState {
	m.mu.Lock()
	defer m.mu.Unlock()

	answer := generateAnswer(difficulty)
	state := &GameState{
		Answer:      answer,
		Attempts:    0,
		MaxAttempts: 10,
		History:     make([]GuessResult, 0),
		MessageID:   "",
		Difficulty:  difficulty,
		Locale:      locale,
	}

	m.games[userID] = state
	return state
}

// GetGame retrieves an active game for a user
func (m *Manager) GetGame(userID string) (*GameState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	game, exists := m.games[userID]
	return game, exists
}

// EndGame removes a game session
func (m *Manager) EndGame(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.games, userID)
}

// generateAnswer generates a random 4-digit number based on difficulty
func generateAnswer(difficulty Difficulty) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	if difficulty == DifficultyHard {
		// Hard mode: digits can repeat
		answer := make([]byte, 4)
		for i := 0; i < 4; i++ {
			answer[i] = byte('0' + rng.Intn(10))
		}
		return string(answer)
	}

	// Easy mode: unique digits only
	digits := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

	// Shuffle digits
	for i := len(digits) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		digits[i], digits[j] = digits[j], digits[i]
	}

	// Take first 4 digits
	return string(digits[:4])
}

// CheckGuess checks a guess against the answer and returns Bulls and Cows counts
func CheckGuess(answer, guess string) (int, int) {
	bulls := 0 // Correct digit in correct position
	cows := 0  // Correct digit in wrong position

	ansFreq := make(map[byte]int)
	guessFreq := make(map[byte]int)

	for i := 0; i < len(answer); i++ {
		if answer[i] == guess[i] {
			bulls++
		} else {
			ansFreq[answer[i]]++
			guessFreq[guess[i]]++
		}
	}
	for k, v := range guessFreq {
		if ansFreq[k] > 0 {
			if v < ansFreq[k] {
				cows += v
			} else {
				cows += ansFreq[k]
			}
		}
	}

	return bulls, cows
}

// IsValidGuess validates a guess based on difficulty
func IsValidGuess(guess string, difficulty Difficulty) bool {
	if len(guess) != 4 {
		return false
	}

	seen := make(map[rune]bool)
	for _, ch := range guess {
		if ch < '0' || ch > '9' {
			return false
		}
		// Only check for duplicates in easy mode
		if difficulty == DifficultyEasy {
			if seen[ch] {
				return false // Duplicate digit not allowed in easy mode
			}
			seen[ch] = true
		}
	}

	return true
}
