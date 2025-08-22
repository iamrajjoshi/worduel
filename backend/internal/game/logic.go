package game

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"worduel-backend/internal/logging"
)

var (
	ErrInvalidWord     = errors.New("word is not in dictionary")
	ErrGameNotActive   = errors.New("game is not in active state")
	ErrPlayerNotFound  = errors.New("player not found in game")
	ErrTooManyGuesses  = errors.New("player has reached maximum number of guesses")
	ErrGameAlreadyWon  = errors.New("game has already been won")
	ErrInvalidWordLength = errors.New("word must be exactly 5 letters")
)

// GameLogic handles core game mechanics and state transitions
type GameLogic struct {
	dictionary *Dictionary
	logger     *logging.Logger
}

// NewGameLogic creates a new game logic instance
func NewGameLogic(dictionary *Dictionary, logger *logging.Logger) *GameLogic {
	return &GameLogic{
		dictionary: dictionary,
		logger:     logger,
	}
}

// GuessResult represents the result of processing a guess
type GuessResult struct {
	Word      string         `json:"word"`
	Results   []LetterResult `json:"results"`
	IsCorrect bool           `json:"is_correct"`
	Timestamp time.Time      `json:"timestamp"`
}

// ProcessGuess validates a guess and computes letter matching results
func (gl *GameLogic) ProcessGuess(room *Room, playerID, word string) (*GuessResult, error) {
	if room == nil {
		return nil, errors.New("room cannot be nil")
	}

	// Normalize the input word
	normalizedWord := strings.TrimSpace(strings.ToLower(word))
	
	// Validate word length
	if len(normalizedWord) != 5 {
		return nil, ErrInvalidWordLength
	}

	// Lock room for the entire operation
	room.Lock()
	defer room.Unlock()

	// Check game state
	if room.GameState.Status != GameStatusActive {
		return nil, ErrGameNotActive
	}

	// Check if game is already completed
	if room.GameState.Winner != "" {
		return nil, ErrGameAlreadyWon
	}

	// Find player
	player, exists := room.Players[playerID]
	if !exists {
		return nil, ErrPlayerNotFound
	}

	// Check if player has reached max guesses
	if len(player.Guesses) >= room.GameState.MaxGuesses {
		return nil, ErrTooManyGuesses
	}

	// Validate word against dictionary
	if !gl.dictionary.IsValidGuess(normalizedWord) {
		return nil, ErrInvalidWord
	}

	// Process the guess and compute letter results
	results := gl.computeLetterResults(normalizedWord, room.GameState.Word)
	isCorrect := normalizedWord == room.GameState.Word

	// Create guess result
	guessResult := &GuessResult{
		Word:      normalizedWord,
		Results:   results,
		IsCorrect: isCorrect,
		Timestamp: time.Now(),
	}

	// Create and add guess to player's history
	guess := Guess{
		Word:      normalizedWord,
		Results:   results,
		Timestamp: guessResult.Timestamp,
		IsCorrect: isCorrect,
	}
	
	player.Guesses = append(player.Guesses, guess)
	player.LastActivity = time.Now()

	// Update game state based on guess result
	gl.updateGameState(room, playerID, isCorrect)

	// Log the guess processing
	if gl.logger != nil {
		ctx := logging.WithCorrelationID(context.Background(), playerID)
		gl.logger.LogGameEvent(ctx, logging.GameEventFields{
			EventType: "guess_processed",
			RoomID:    room.ID,
			PlayerID:  playerID,
			GameState: string(room.GameState.Status),
		})
		
		if isCorrect {
			gl.logger.LogInfo(ctx, "Player guessed correctly", 
				"word", normalizedWord,
				"room_id", room.ID,
				"guess_count", len(player.Guesses))
		}
	}

	return guessResult, nil
}

// computeLetterResults implements the Wordle letter matching algorithm
func (gl *GameLogic) computeLetterResults(guess, target string) []LetterResult {
	results := make([]LetterResult, 5)
	targetLetters := []rune(target)
	guessLetters := []rune(guess)
	
	// Track which target letters have been matched
	targetUsed := make([]bool, 5)
	
	// First pass: mark correct positions (green)
	for i := 0; i < 5; i++ {
		if guessLetters[i] == targetLetters[i] {
			results[i] = LetterResultCorrect
			targetUsed[i] = true
		}
	}
	
	// Second pass: mark present letters in wrong positions (yellow)
	for i := 0; i < 5; i++ {
		if results[i] == LetterResultCorrect {
			continue // Already marked as correct
		}
		
		// Look for this letter in unused positions of target
		found := false
		for j := 0; j < 5; j++ {
			if !targetUsed[j] && guessLetters[i] == targetLetters[j] {
				results[i] = LetterResultPresent
				targetUsed[j] = true
				found = true
				break
			}
		}
		
		if !found {
			results[i] = LetterResultAbsent
		}
	}
	
	return results
}

// updateGameState handles game state transitions after a guess
func (gl *GameLogic) updateGameState(room *Room, playerID string, isCorrect bool) {
	gameState := room.GameState
	
	if isCorrect {
		// Player won the game
		gameState.Winner = playerID
		gameState.Status = GameStatusFinished
		now := time.Now()
		gameState.FinishedAt = &now
		
		// Update player status
		if player, exists := room.Players[playerID]; exists {
			player.Status = PlayerStatusFinished
			player.Score = gl.calculateScore(player.Guesses)
		}
		
		// Mark other players as finished (they lost)
		for id, player := range room.Players {
			if id != playerID && player.Status == PlayerStatusActive {
				player.Status = PlayerStatusFinished
			}
		}
	} else {
		// Check if player has used all guesses
		if player, exists := room.Players[playerID]; exists {
			if len(player.Guesses) >= gameState.MaxGuesses {
				player.Status = PlayerStatusFinished
			}
		}
		
		// Check if all active players have finished (no one won)
		allFinished := true
		for _, player := range room.Players {
			if player.Status == PlayerStatusActive {
				allFinished = false
				break
			}
		}
		
		if allFinished {
			gameState.Status = GameStatusFinished
			now := time.Now()
			gameState.FinishedAt = &now
		}
	}
	
	// Update room timestamp
	room.UpdatedAt = time.Now()
}

// calculateScore computes a score based on number of guesses (lower is better)
func (gl *GameLogic) calculateScore(guesses []Guess) int {
	if len(guesses) == 0 {
		return 0
	}
	
	// Find the winning guess
	for i, guess := range guesses {
		if guess.IsCorrect {
			// Score: 100 - (guess_number * 10), minimum 10
			score := 100 - (i * 10)
			if score < 10 {
				score = 10
			}
			return score
		}
	}
	
	// No winning guess found
	return 0
}

// IsComplete checks if the game has reached a completion state
func (gl *GameLogic) IsComplete(room *Room) (bool, string) {
	if room == nil {
		return false, ""
	}
	
	room.RLock()
	defer room.RUnlock()
	
	gameState := room.GameState
	if gameState.Status == GameStatusFinished {
		return true, gameState.Winner
	}
	
	return false, ""
}

// ValidateGameState performs comprehensive validation of game state consistency
func (gl *GameLogic) ValidateGameState(room *Room) error {
	if room == nil {
		return errors.New("room cannot be nil")
	}
	
	room.RLock()
	defer room.RUnlock()
	
	gameState := room.GameState
	if gameState == nil {
		return errors.New("game state cannot be nil")
	}
	
	// Validate word
	if gameState.Word == "" {
		return errors.New("target word cannot be empty")
	}
	
	if len(gameState.Word) != 5 {
		return errors.New("target word must be 5 letters")
	}
	
	// Validate max guesses
	if gameState.MaxGuesses <= 0 || gameState.MaxGuesses > 10 {
		return errors.New("max guesses must be between 1 and 10")
	}
	
	// Validate players
	if len(room.Players) == 0 {
		return errors.New("room must have at least one player")
	}
	
	// Validate each player's guesses
	for playerID, player := range room.Players {
		if player == nil {
			return fmt.Errorf("player %s cannot be nil", playerID)
		}
		
		if len(player.Guesses) > gameState.MaxGuesses {
			return fmt.Errorf("player %s has too many guesses (%d > %d)", 
				playerID, len(player.Guesses), gameState.MaxGuesses)
		}
		
		// Validate each guess
		for i, guess := range player.Guesses {
			if len(guess.Word) != 5 {
				return fmt.Errorf("player %s guess %d has invalid word length", playerID, i)
			}
			
			if len(guess.Results) != 5 {
				return fmt.Errorf("player %s guess %d has invalid results length", playerID, i)
			}
		}
	}
	
	// Validate game status consistency
	if gameState.Status == GameStatusFinished {
		if gameState.FinishedAt == nil {
			return errors.New("finished game must have finish time")
		}
		
		// If there's a winner, validate they actually won
		if gameState.Winner != "" {
			winner, exists := room.Players[gameState.Winner]
			if !exists {
				return fmt.Errorf("winner %s not found in players", gameState.Winner)
			}
			
			// Check if winner has a correct guess
			hasCorrectGuess := false
			for _, guess := range winner.Guesses {
				if guess.IsCorrect {
					hasCorrectGuess = true
					break
				}
			}
			
			if !hasCorrectGuess {
				return fmt.Errorf("winner %s has no correct guess", gameState.Winner)
			}
		}
	}
	
	return nil
}

// StartGame initializes a new game with the given target word
func (gl *GameLogic) StartGame(room *Room, targetWord string) error {
	if room == nil {
		return errors.New("room cannot be nil")
	}
	
	// Validate target word
	if len(targetWord) != 5 {
		return ErrInvalidWordLength
	}
	
	normalizedTarget := strings.ToLower(strings.TrimSpace(targetWord))
	
	room.Lock()
	defer room.Unlock()
	
	// Initialize game state
	gameState := room.GameState
	gameState.Status = GameStatusActive
	gameState.Word = normalizedTarget
	gameState.WordLength = 5
	gameState.MaxGuesses = 6
	gameState.CurrentRound = 1
	now := time.Now()
	gameState.StartedAt = &now
	gameState.RoundDuration = 1800 // 30 minutes
	
	// Reset all players to active status
	for _, player := range room.Players {
		player.Status = PlayerStatusActive
		player.Guesses = make([]Guess, 0)
		player.Score = 0
		player.LastActivity = now
	}
	
	room.UpdatedAt = now
	
	// Log game start
	if gl.logger != nil {
		ctx := context.Background()
		gl.logger.LogGameEvent(ctx, logging.GameEventFields{
			EventType: "game_started",
			RoomID:    room.ID,
			PlayerID:  "",
			GameState: string(GameStatusActive),
		})
		gl.logger.LogInfo(ctx, "Game started",
			"room_id", room.ID,
			"player_count", len(room.Players),
			"target_word", normalizedTarget)
	}
	
	return nil
}

// GetGameSummary returns a summary of the current game state for clients
func (gl *GameLogic) GetGameSummary(room *Room, playerID string) map[string]interface{} {
	if room == nil {
		return nil
	}
	
	room.RLock()
	defer room.RUnlock()
	
	summary := make(map[string]interface{})
	gameState := room.GameState
	
	// Basic game info (never include the actual target word)
	summary["status"] = gameState.Status
	summary["word_length"] = gameState.WordLength
	summary["max_guesses"] = gameState.MaxGuesses
	summary["current_round"] = gameState.CurrentRound
	summary["started_at"] = gameState.StartedAt
	summary["finished_at"] = gameState.FinishedAt
	summary["winner"] = gameState.Winner
	
	// Player information
	players := make(map[string]interface{})
	for id, player := range room.Players {
		playerInfo := map[string]interface{}{
			"id":            id,
			"name":          player.Name,
			"status":        player.Status,
			"score":         player.Score,
			"guess_count":   len(player.Guesses),
			"last_activity": player.LastActivity,
		}
		
		// Include full guess details for the requesting player only
		// For other players, only show guess patterns for competitive visibility
		if id == playerID {
			playerInfo["guesses"] = player.Guesses
		} else {
			// Show only the letter results (patterns) without revealing words
			patterns := make([][]LetterResult, len(player.Guesses))
			for i, guess := range player.Guesses {
				patterns[i] = guess.Results
			}
			playerInfo["guess_patterns"] = patterns
		}
		
		players[id] = playerInfo
	}
	summary["players"] = players
	
	return summary
}