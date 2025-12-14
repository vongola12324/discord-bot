package wordle

import (
	"sync"
)

const (
	maxAttempts = 6
)

// GuessResult represents the result of a single letter guess
type GuessResult int

const (
	CorrectPosition GuessResult = iota // Green
	WrongPosition                      // Yellow
	NotInWord                          // Red
)

// Guess represents a single guess with its results
type Guess struct {
	Word    string
	Results []GuessResult
}

// GameState represents an active Wordle game
type GameState struct {
	UserID      string
	MessageID   string
	Answer      string
	WordLength  int
	Guesses     []Guess
	IsCompleted bool
}

// GameStore manages all active games
type GameStore struct {
	games map[string]*GameState // userID -> GameState
	mu    sync.RWMutex
}

var store *GameStore
var once sync.Once

// GetStore returns the singleton game store
func GetStore() *GameStore {
	once.Do(func() {
		store = &GameStore{
			games: make(map[string]*GameState),
		}
	})
	return store
}

// StartGame creates a new game for a user
func (s *GameStore) StartGame(userID, messageID, answer string, wordLength int) *GameState {
	s.mu.Lock()
	defer s.mu.Unlock()

	game := &GameState{
		UserID:      userID,
		MessageID:   messageID,
		Answer:      answer,
		WordLength:  wordLength,
		Guesses:     make([]Guess, 0, maxAttempts),
		IsCompleted: false,
	}

	s.games[userID] = game
	return game
}

// GetGame retrieves a game by user ID
func (s *GameStore) GetGame(userID string) (*GameState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, exists := s.games[userID]
	return game, exists
}

// HasActiveGame checks if a user has an active game
func (s *GameStore) HasActiveGame(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, exists := s.games[userID]
	return exists && !game.IsCompleted
}

// EndGame marks a game as completed
func (s *GameStore) EndGame(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if game, exists := s.games[userID]; exists {
		game.IsCompleted = true
	}
}

// DeleteGame removes a game from the store
func (s *GameStore) DeleteGame(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.games, userID)
}

// AddGuess adds a guess to the game and returns the results
func (g *GameState) AddGuess(guess string) []GuessResult {
	results := make([]GuessResult, g.WordLength)
	answerRunes := []rune(g.Answer)
	guessRunes := []rune(guess)

	// Track which letters in the answer have been matched
	answerUsed := make([]bool, g.WordLength)
	guessMatched := make([]bool, g.WordLength)

	// First pass: mark correct positions (green)
	for i := 0; i < g.WordLength; i++ {
		if guessRunes[i] == answerRunes[i] {
			results[i] = CorrectPosition
			answerUsed[i] = true
			guessMatched[i] = true
		}
	}

	// Second pass: mark wrong positions (yellow) and not in word (red)
	for i := 0; i < g.WordLength; i++ {
		if guessMatched[i] {
			continue // Already marked as correct position
		}

		found := false
		for j := 0; j < g.WordLength; j++ {
			if !answerUsed[j] && guessRunes[i] == answerRunes[j] {
				results[i] = WrongPosition
				answerUsed[j] = true
				found = true
				break
			}
		}

		if !found {
			results[i] = NotInWord
		}
	}

	g.Guesses = append(g.Guesses, Guess{
		Word:    guess,
		Results: results,
	})

	return results
}

// IsWon checks if the game has been won
func (g *GameState) IsWon() bool {
	if len(g.Guesses) == 0 {
		return false
	}

	lastGuess := g.Guesses[len(g.Guesses)-1]
	for _, result := range lastGuess.Results {
		if result != CorrectPosition {
			return false
		}
	}

	return true
}

// IsLost checks if the game has been lost (max attempts reached without winning)
func (g *GameState) IsLost() bool {
	return len(g.Guesses) >= maxAttempts && !g.IsWon()
}

// AttemptsLeft returns the number of attempts remaining
func (g *GameState) AttemptsLeft() int {
	return maxAttempts - len(g.Guesses)
}
