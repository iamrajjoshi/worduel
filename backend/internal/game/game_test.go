package game_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"worduel-backend/internal/game"
)

func TestGameLogic_ProcessGuess_WordValidation(t *testing.T) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)

	tests := []struct {
		name          string
		targetWord    string
		guess         string
		expectedError error
		expectedResults []game.LetterResult
		expectedCorrect bool
	}{
		{
			name:          "valid correct guess",
			targetWord:    "about",
			guess:         "about",
			expectedError: nil,
			expectedResults: []game.LetterResult{
				game.LetterResultCorrect, game.LetterResultCorrect, game.LetterResultCorrect,
				game.LetterResultCorrect, game.LetterResultCorrect,
			},
			expectedCorrect: true,
		},
		{
			name:          "valid incorrect guess",
			targetWord:    "about",
			guess:         "above",
			expectedError: nil,
			expectedResults: []game.LetterResult{
				game.LetterResultCorrect, // a
				game.LetterResultCorrect, // b
				game.LetterResultCorrect, // o
				game.LetterResultAbsent,  // v
				game.LetterResultAbsent,  // e
			},
			expectedCorrect: false,
		},
		{
			name:          "invalid word length - too short",
			targetWord:    "about",
			guess:         "hi",
			expectedError: game.ErrInvalidWordLength,
		},
		{
			name:          "invalid word length - too long",
			targetWord:    "about",
			guess:         "testing",
			expectedError: game.ErrInvalidWordLength,
		},
		{
			name:          "word not in dictionary",
			targetWord:    "about",
			guess:         "zzzzz",
			expectedError: game.ErrInvalidWord,
		},
		{
			name:          "empty guess",
			targetWord:    "about",
			guess:         "",
			expectedError: game.ErrInvalidWordLength,
		},
		{
			name:          "whitespace handling",
			targetWord:    "about",
			guess:         " about ",
			expectedError: nil,
			expectedResults: []game.LetterResult{
				game.LetterResultCorrect, game.LetterResultCorrect, game.LetterResultCorrect,
				game.LetterResultCorrect, game.LetterResultCorrect,
			},
			expectedCorrect: true,
		},
		{
			name:          "case insensitive",
			targetWord:    "about",
			guess:         "ABOUT",
			expectedError: nil,
			expectedResults: []game.LetterResult{
				game.LetterResultCorrect, game.LetterResultCorrect, game.LetterResultCorrect,
				game.LetterResultCorrect, game.LetterResultCorrect,
			},
			expectedCorrect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := createTestRoom(tt.targetWord, "player1")
			
			result, err := logic.ProcessGuess(room, "player1", tt.guess)

			if tt.expectedError != nil {
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.IsCorrect != tt.expectedCorrect {
				t.Errorf("expected IsCorrect %v, got %v", tt.expectedCorrect, result.IsCorrect)
			}

			if len(result.Results) != len(tt.expectedResults) {
				t.Errorf("expected %d results, got %d", len(tt.expectedResults), len(result.Results))
				return
			}

			for i, expected := range tt.expectedResults {
				if result.Results[i] != expected {
					t.Errorf("position %d: expected %s, got %s", i, expected, result.Results[i])
				}
			}
		})
	}
}

func TestGameLogic_ScoringAlgorithm(t *testing.T) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)

	tests := []struct {
		name           string
		targetWord     string
		guesses        []string
		expectedScores []int
		finalScore     int
	}{
		{
			name:           "win on first guess",
			targetWord:     "about",
			guesses:        []string{"about"},
			expectedScores: []int{100},
			finalScore:     100,
		},
		{
			name:           "win on second guess",
			targetWord:     "about",
			guesses:        []string{"above", "about"},
			expectedScores: []int{0, 90},
			finalScore:     90,
		},
		{
			name:           "win on last guess",
			targetWord:     "about",
			guesses:        []string{"above", "abuse", "actor", "acute", "admit", "about"},
			expectedScores: []int{0, 0, 0, 0, 0, 50},
			finalScore:     50,
		},
		{
			name:           "no win - all guesses exhausted",
			targetWord:     "about",
			guesses:        []string{"above", "abuse", "actor", "acute", "admit", "adopt"},
			expectedScores: []int{0, 0, 0, 0, 0, 0},
			finalScore:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := createTestRoom(tt.targetWord, "player1")
			
			for i, guess := range tt.guesses {
				result, err := logic.ProcessGuess(room, "player1", guess)
				if err != nil {
					t.Fatalf("unexpected error on guess %d: %v", i+1, err)
				}
				
				player, _ := room.GetPlayer("player1")
				if result.IsCorrect {
					expectedScore := tt.expectedScores[i]
					if player.Score != expectedScore {
						t.Errorf("guess %d: expected score %d, got %d", i+1, expectedScore, player.Score)
					}
					break
				}
			}
			
			player, _ := room.GetPlayer("player1")
			if player.Score != tt.finalScore {
				t.Errorf("expected final score %d, got %d", tt.finalScore, player.Score)
			}
		})
	}
}

func TestGameLogic_StateTransitions(t *testing.T) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)

	t.Run("winning transitions", func(t *testing.T) {
		room := createTestRoomWithTwoPlayers("about", "player1", "player2")
		
		// Player 1 makes winning guess
		result, err := logic.ProcessGuess(room, "player1", "about")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if !result.IsCorrect {
			t.Error("expected correct guess")
		}
		
		// Check game state
		if room.GetGameStatus() != game.GameStatusFinished {
			t.Errorf("expected game status %s, got %s", game.GameStatusFinished, room.GetGameStatus())
		}
		
		if room.GetGameWinner() != "player1" {
			t.Errorf("expected winner 'player1', got '%s'", room.GetGameWinner())
		}
		
		// Check player statuses
		player1, _ := room.GetPlayer("player1")
		if player1.Status != game.PlayerStatusFinished {
			t.Errorf("expected winner status %s, got %s", game.PlayerStatusFinished, player1.Status)
		}
		
		player2, _ := room.GetPlayer("player2")
		if player2.Status != game.PlayerStatusFinished {
			t.Errorf("expected loser status %s, got %s", game.PlayerStatusFinished, player2.Status)
		}
	})

	t.Run("max guesses exhausted", func(t *testing.T) {
		room := createTestRoomWithTwoPlayers("about", "player1", "player2")
		
		// Player 1 exhausts guesses without winning
		incorrectGuesses := []string{"above", "abuse", "actor", "acute", "admit", "adopt"}
		for _, guess := range incorrectGuesses {
			_, err := logic.ProcessGuess(room, "player1", guess)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}
		
		player1, _ := room.GetPlayer("player1")
		if player1.Status != game.PlayerStatusFinished {
			t.Errorf("expected player1 status %s, got %s", game.PlayerStatusFinished, player1.Status)
		}
		
		if len(player1.Guesses) != 6 {
			t.Errorf("expected 6 guesses, got %d", len(player1.Guesses))
		}
	})

	t.Run("all players exhausted - no winner", func(t *testing.T) {
		room := createTestRoomWithTwoPlayers("about", "player1", "player2")
		
		// Both players exhaust guesses without winning
		incorrectGuesses := []string{"above", "abuse", "actor", "acute", "admit", "adopt"}
		
		for _, guess := range incorrectGuesses {
			_, err := logic.ProcessGuess(room, "player1", guess)
			if err != nil {
				t.Errorf("unexpected error for player1: %v", err)
			}
		}
		
		for _, guess := range incorrectGuesses {
			_, err := logic.ProcessGuess(room, "player2", guess)
			if err != nil {
				t.Errorf("unexpected error for player2: %v", err)
			}
		}
		
		// Game should be finished with no winner
		if room.GetGameStatus() != game.GameStatusFinished {
			t.Errorf("expected game status %s, got %s", game.GameStatusFinished, room.GetGameStatus())
		}
		
		if room.GetGameWinner() != "" {
			t.Errorf("expected no winner, got '%s'", room.GetGameWinner())
		}
	})
}

func TestGameLogic_DuplicateLetters(t *testing.T) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)

	tests := []struct {
		name     string
		target   string
		guess    string
		expected []game.LetterResult
	}{
		{
			name:   "duplicate letters in guess - both correct",
			target: "wheel",
			guess:  "wheel", 
			expected: []game.LetterResult{
				game.LetterResultCorrect, game.LetterResultCorrect, game.LetterResultCorrect,
				game.LetterResultCorrect, game.LetterResultCorrect,
			},
		},
		{
			name:   "duplicate letters in target",
			target: "allow", // has two 'l's at positions 1,2
			guess:  "alarm", // has one 'l' and two 'a's
			expected: []game.LetterResult{
				game.LetterResultCorrect, // a is correct at position 0
				game.LetterResultCorrect, // l is correct at position 1
				game.LetterResultAbsent,  // a - target only has one 'a' already used at position 0
				game.LetterResultAbsent,  // r not in target
				game.LetterResultAbsent,  // m not in target
			},
		},
		{
			name:   "complex duplicate scenario",
			target: "apple", // has two 'p's
			guess:  "about", // no duplicates
			expected: []game.LetterResult{
				game.LetterResultCorrect, // a is correct
				game.LetterResultAbsent,  // b not in target
				game.LetterResultAbsent,  // o not in target
				game.LetterResultAbsent,  // u not in target
				game.LetterResultAbsent,  // t not in target
			},
		},
		{
			name:   "both have duplicates",
			target: "apple", // a(0)p(1)p(2)l(3)e(4)
			guess:  "allow", // a(0)l(1)l(2)o(3)w(4)
			expected: []game.LetterResult{
				game.LetterResultCorrect, // a is correct at position 0
				game.LetterResultPresent, // l is in target at pos 3 but wrong position
				game.LetterResultAbsent,  // second l - no more l's available in target
				game.LetterResultAbsent,  // o not in target
				game.LetterResultAbsent,  // w not in target
			},
		},
		{
			name:   "edge case - same letter repeated",
			target: "apple",
			guess:  "apple",
			expected: []game.LetterResult{
				game.LetterResultCorrect, game.LetterResultCorrect, game.LetterResultCorrect,
				game.LetterResultCorrect, game.LetterResultCorrect,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := createTestRoom(tt.target, "player1")
			
			result, err := logic.ProcessGuess(room, "player1", tt.guess)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result.Results) != len(tt.expected) {
				t.Errorf("expected %d results, got %d", len(tt.expected), len(result.Results))
				return
			}

			for i, expected := range tt.expected {
				if result.Results[i] != expected {
					t.Errorf("position %d: expected %s, got %s", i, expected, result.Results[i])
				}
			}
		})
	}
}

func TestGameLogic_InvalidGameStates(t *testing.T) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)

	t.Run("game not active", func(t *testing.T) {
		room := createTestRoom("about", "player1")
		room.GameState.Status = game.GameStatusWaiting
		
		_, err := logic.ProcessGuess(room, "player1", "crane")
		if err != game.ErrGameNotActive {
			t.Errorf("expected ErrGameNotActive, got %v", err)
		}
	})

	t.Run("player not found", func(t *testing.T) {
		room := createTestRoom("about", "player1")
		
		_, err := logic.ProcessGuess(room, "nonexistent", "crane")
		if err != game.ErrPlayerNotFound {
			t.Errorf("expected ErrPlayerNotFound, got %v", err)
		}
	})

	t.Run("too many guesses", func(t *testing.T) {
		room := createTestRoom("about", "player1")
		
		// Add 6 guesses to exhaust limit
		player, _ := room.GetPlayer("player1")
		for i := 0; i < 6; i++ {
			player.Guesses = append(player.Guesses, game.Guess{
				Word: fmt.Sprintf("word%d", i),
				Results: []game.LetterResult{
					game.LetterResultAbsent, game.LetterResultAbsent,
					game.LetterResultAbsent, game.LetterResultAbsent,
					game.LetterResultAbsent,
				},
				IsCorrect: false,
			})
		}
		
		_, err := logic.ProcessGuess(room, "player1", "crane")
		if err != game.ErrTooManyGuesses {
			t.Errorf("expected ErrTooManyGuesses, got %v", err)
		}
	})

	t.Run("game already won", func(t *testing.T) {
		room := createTestRoom("about", "player1")
		room.GameState.Winner = "player1"
		
		_, err := logic.ProcessGuess(room, "player1", "crane")
		if err != game.ErrGameAlreadyWon {
			t.Errorf("expected ErrGameAlreadyWon, got %v", err)
		}
	})

	t.Run("nil room", func(t *testing.T) {
		_, err := logic.ProcessGuess(nil, "player1", "crane")
		if err == nil {
			t.Error("expected error for nil room")
		}
	})
}

func TestGameLogic_ConcurrentAccess(t *testing.T) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)

	t.Run("concurrent guesses from different players", func(t *testing.T) {
		room := createTestRoom("about", "")
		playerCount := 10
		
		// Add multiple players
		for i := 0; i < playerCount; i++ {
			playerID := fmt.Sprintf("player%d", i)
			room.AddPlayer(playerID, fmt.Sprintf("Player%d", i))
		}
		
		var wg sync.WaitGroup
		results := make([]error, playerCount)
		
		// Concurrent guesses
		wg.Add(playerCount)
		for i := 0; i < playerCount; i++ {
			go func(playerIndex int) {
				defer wg.Done()
				playerID := fmt.Sprintf("player%d", playerIndex)
				_, err := logic.ProcessGuess(room, playerID, "above")
				results[playerIndex] = err
			}(i)
		}
		wg.Wait()
		
		// All should succeed
		for i, err := range results {
			if err != nil {
				t.Errorf("player %d got error: %v", i, err)
			}
		}
		
		// Verify all players have one guess
		for i := 0; i < playerCount; i++ {
			playerID := fmt.Sprintf("player%d", i)
			player, exists := room.GetPlayer(playerID)
			if !exists {
				t.Errorf("player %d should exist", i)
				continue
			}
			if len(player.Guesses) != 1 {
				t.Errorf("player %d should have 1 guess, got %d", i, len(player.Guesses))
			}
		}
	})

	t.Run("concurrent guesses from same player", func(t *testing.T) {
		room := createTestRoom("about", "player1")
		
		var wg sync.WaitGroup
		goroutineCount := 5
		results := make([]error, goroutineCount)
		
		validGuesses := []string{"above", "abuse", "actor", "acute", "admit"}
		
		// Concurrent guesses from same player (should be serialized)
		wg.Add(goroutineCount)
		for i := 0; i < goroutineCount; i++ {
			go func(index int) {
				defer wg.Done()
				guess := validGuesses[index%len(validGuesses)]
				_, err := logic.ProcessGuess(room, "player1", guess)
				results[index] = err
			}(i)
		}
		wg.Wait()
		
		// Some should fail due to invalid words, but no race conditions
		player, _ := room.GetPlayer("player1")
		if len(player.Guesses) == 0 {
			t.Error("player should have at least some guesses recorded")
		}
		
		// Room state should be consistent
		if err := room.ValidateRoomState(); err != nil {
			t.Errorf("room state validation failed: %v", err)
		}
	})
}

func TestGameLogic_ThreadSafety(t *testing.T) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)

	t.Run("concurrent room operations", func(t *testing.T) {
		room := createTestRoom("about", "")
		
		// Add initial players
		for i := 0; i < 5; i++ {
			playerID := fmt.Sprintf("player%d", i)
			room.AddPlayer(playerID, fmt.Sprintf("Player%d", i))
		}
		
		var wg sync.WaitGroup
		operationCount := 100
		
		// Mix of operations running concurrently
		wg.Add(operationCount)
		for i := 0; i < operationCount; i++ {
			go func(index int) {
				defer wg.Done()
				
				playerID := fmt.Sprintf("player%d", index%5)
				
				switch index % 4 {
				case 0:
					// Make guess
					logic.ProcessGuess(room, playerID, "above")
				case 1:
					// Check game completion
					logic.IsComplete(room)
				case 2:
					// Get game summary
					logic.GetGameSummary(room, playerID)
				case 3:
					// Validate game state
					logic.ValidateGameState(room)
				}
			}(i)
		}
		wg.Wait()
		
		// Verify final state is consistent
		if err := room.ValidateRoomState(); err != nil {
			t.Errorf("final room state validation failed: %v", err)
		}
		
		if err := logic.ValidateGameState(room); err != nil {
			t.Errorf("final game state validation failed: %v", err)
		}
	})

	t.Run("stress test - high concurrency", func(t *testing.T) {
		room := createTestRoom("about", "")
		
		// Add many players
		playerCount := 20
		for i := 0; i < playerCount; i++ {
			playerID := fmt.Sprintf("player%d", i)
			room.AddPlayer(playerID, fmt.Sprintf("Player%d", i))
		}
		
		var wg sync.WaitGroup
		goroutineCount := 200
		
		wg.Add(goroutineCount)
		for i := 0; i < goroutineCount; i++ {
			go func(index int) {
				defer wg.Done()
				
				playerID := fmt.Sprintf("player%d", index%playerCount)
				
				// Rapid-fire operations
				validWords := []string{"above", "abuse", "actor"}
				for j := 0; j < 10; j++ {
					switch j % 3 {
					case 0:
						logic.ProcessGuess(room, playerID, validWords[j%len(validWords)])
					case 1:
						logic.IsComplete(room)
					case 2:
						logic.GetGameSummary(room, playerID)
					}
				}
			}(i)
		}
		wg.Wait()
		
		// Final validation
		if err := room.ValidateRoomState(); err != nil {
			t.Errorf("stress test room validation failed: %v", err)
		}
	})
}

// Benchmark tests for performance-critical methods
func BenchmarkGameLogic_ProcessGuess(b *testing.B) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		room := createTestRoom("about", "player1")
		logic.ProcessGuess(room, "player1", "above")
	}
}

func BenchmarkGameLogic_ComputeLetterResults(b *testing.B) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use reflection to access private method for benchmarking
		// In actual implementation, this would need to be made public for testing
		// For now, we'll test through ProcessGuess
		room := createTestRoom("about", "player1")
		logic.ProcessGuess(room, "player1", "above")
	}
}

func BenchmarkGameLogic_IsComplete(b *testing.B) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)
	room := createTestRoom("about", "player1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logic.IsComplete(room)
	}
}

func BenchmarkGameLogic_ValidateGameState(b *testing.B) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)
	room := createTestRoom("about", "player1")
	
	// Add some guesses to make validation more realistic
	player, _ := room.GetPlayer("player1")
	for i := 0; i < 3; i++ {
		player.Guesses = append(player.Guesses, game.Guess{
			Word: fmt.Sprintf("word%d", i),
			Results: []game.LetterResult{
				game.LetterResultAbsent, game.LetterResultAbsent,
				game.LetterResultAbsent, game.LetterResultAbsent,
				game.LetterResultAbsent,
			},
			IsCorrect: false,
		})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logic.ValidateGameState(room)
	}
}

func BenchmarkGameLogic_ConcurrentAccess(b *testing.B) {
	dict := game.NewDictionary()
	logic := game.NewGameLogic(dict, nil)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			room := createTestRoom("about", "player1")
			logic.ProcessGuess(room, "player1", "above")
		}
	})
}

// Helper functions
func createTestRoom(targetWord, playerID string) *game.Room {
	now := time.Now()
	room := &game.Room{
		ID:         "test-room",
		Name:       "TEST123",
		Players:    make(map[string]*game.Player),
		MaxPlayers: 10,
		CreatedAt:  now,
		UpdatedAt:  now,
		GameState: &game.GameState{
			Status:        game.GameStatusActive,
			Word:          targetWord,
			WordLength:    5,
			MaxGuesses:    6,
			CurrentRound:  1,
			RoundDuration: 1800,
			StartedAt:     &now,
		},
	}

	if playerID != "" {
		room.Players[playerID] = &game.Player{
			ID:           playerID,
			Name:         "Test Player",
			Status:       game.PlayerStatusActive,
			Guesses:      []game.Guess{},
			Score:        0,
			ConnectedAt:  time.Now(),
			LastActivity: time.Now(),
		}
	}

	return room
}

func createTestRoomWithTwoPlayers(targetWord, player1ID, player2ID string) *game.Room {
	room := createTestRoom(targetWord, player1ID)
	
	room.Players[player2ID] = &game.Player{
		ID:           player2ID,
		Name:         "Test Player 2",
		Status:       game.PlayerStatusActive,
		Guesses:      []game.Guess{},
		Score:        0,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
	}
	
	return room
}