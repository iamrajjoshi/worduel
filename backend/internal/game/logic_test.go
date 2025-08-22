package game

import (
	"testing"
	"time"
)

func TestGameLogic_ProcessGuess(t *testing.T) {
	// Create dictionary and game logic
	dict := NewDictionary()
	logic := NewGameLogic(dict)

	tests := []struct {
		name          string
		guess         string
		expectedError error
		expectedResults []LetterResult
		expectedCorrect bool
	}{
		{
			name:          "correct guess",
			guess:         "about",
			expectedError: nil,
			expectedResults: []LetterResult{
				LetterResultCorrect, LetterResultCorrect, LetterResultCorrect,
				LetterResultCorrect, LetterResultCorrect,
			},
			expectedCorrect: true,
		},
		{
			name:          "partial match",
			guess:         "above", // this is in dictionary, target is "about"
			expectedError: nil,
			expectedResults: []LetterResult{
				LetterResultCorrect, // a matches
				LetterResultCorrect, // b matches 
				LetterResultCorrect, // o matches (both have o at position 2)
				LetterResultAbsent,  // v not in about
				LetterResultAbsent,  // e not in about
			},
			expectedCorrect: false,
		},
		{
			name:          "invalid word length",
			guess:         "hi",
			expectedError: ErrInvalidWordLength,
		},
		{
			name:          "word not in dictionary",
			guess:         "zzzzz",
			expectedError: ErrInvalidWord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh room and player for each test
			room := &Room{
				ID:       "test-room",
				Players:  make(map[string]*Player),
				GameState: &GameState{
					Status:     GameStatusActive,
					Word:       "about",
					WordLength: 5,
					MaxGuesses: 6,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Add test player
			playerID := "player1"
			room.Players[playerID] = &Player{
				ID:           playerID,
				Name:         "Test Player",
				Status:       PlayerStatusActive,
				Guesses:      []Guess{},
				LastActivity: time.Now(),
			}

			result, err := logic.ProcessGuess(room, playerID, tt.guess)

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

func TestGameLogic_computeLetterResults(t *testing.T) {
	logic := &GameLogic{}

	tests := []struct {
		name     string
		guess    string
		target   string
		expected []LetterResult
	}{
		{
			name:   "all correct",
			guess:  "about",
			target: "about",
			expected: []LetterResult{
				LetterResultCorrect, LetterResultCorrect, LetterResultCorrect,
				LetterResultCorrect, LetterResultCorrect,
			},
		},
		{
			name:   "all wrong",
			guess:  "zebra",
			target: "moist",
			expected: []LetterResult{
				LetterResultAbsent, LetterResultAbsent, LetterResultAbsent,
				LetterResultAbsent, LetterResultAbsent,
			},
		},
		{
			name:   "mixed results",
			guess:  "crane",
			target: "about",
			expected: []LetterResult{
				LetterResultAbsent,  // c not in about
				LetterResultAbsent,  // r not in about
				LetterResultPresent, // a is in about but wrong position
				LetterResultAbsent,  // n not in about
				LetterResultAbsent,  // e not in about
			},
		},
		{
			name:   "duplicate letters in guess",
			guess:  "erase",
			target: "lease",
			expected: []LetterResult{
				LetterResultPresent, // e is present but wrong position
				LetterResultAbsent,  // r not in lease
				LetterResultCorrect, // a is correct
				LetterResultCorrect, // s is correct
				LetterResultCorrect, // e is correct
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logic.computeLetterResults(tt.guess, tt.target)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d results, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("position %d: expected %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestGameLogic_IsComplete(t *testing.T) {
	logic := &GameLogic{}

	tests := []struct {
		name           string
		gameStatus     GameStatus
		winner         string
		expectedDone   bool
		expectedWinner string
	}{
		{
			name:           "game finished with winner",
			gameStatus:     GameStatusFinished,
			winner:         "player1",
			expectedDone:   true,
			expectedWinner: "player1",
		},
		{
			name:           "game finished no winner",
			gameStatus:     GameStatusFinished,
			winner:         "",
			expectedDone:   true,
			expectedWinner: "",
		},
		{
			name:           "game still active",
			gameStatus:     GameStatusActive,
			winner:         "",
			expectedDone:   false,
			expectedWinner: "",
		},
		{
			name:           "game waiting",
			gameStatus:     GameStatusWaiting,
			winner:         "",
			expectedDone:   false,
			expectedWinner: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := &Room{
				GameState: &GameState{
					Status: tt.gameStatus,
					Winner: tt.winner,
				},
			}

			done, winner := logic.IsComplete(room)

			if done != tt.expectedDone {
				t.Errorf("expected done %v, got %v", tt.expectedDone, done)
			}

			if winner != tt.expectedWinner {
				t.Errorf("expected winner %s, got %s", tt.expectedWinner, winner)
			}
		})
	}
}

func TestGameLogic_ValidateGameState(t *testing.T) {
	logic := &GameLogic{}

	tests := []struct {
		name        string
		room        *Room
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil room",
			room:        nil,
			expectError: true,
			errorMsg:    "room cannot be nil",
		},
		{
			name: "valid game state",
			room: &Room{
				Players: map[string]*Player{
					"player1": {
						ID:      "player1",
						Guesses: []Guess{},
					},
				},
				GameState: &GameState{
					Word:       "about",
					MaxGuesses: 6,
					Status:     GameStatusActive,
				},
			},
			expectError: false,
		},
		{
			name: "invalid word length",
			room: &Room{
				Players: map[string]*Player{
					"player1": {
						ID:      "player1",
						Guesses: []Guess{},
					},
				},
				GameState: &GameState{
					Word:       "hi",
					MaxGuesses: 6,
					Status:     GameStatusActive,
				},
			},
			expectError: true,
			errorMsg:    "target word must be 5 letters",
		},
		{
			name: "too many guesses for player",
			room: &Room{
				Players: map[string]*Player{
					"player1": {
						ID: "player1",
						Guesses: []Guess{
							{Word: "guess"}, {Word: "guess"}, {Word: "guess"},
							{Word: "guess"}, {Word: "guess"}, {Word: "guess"},
							{Word: "guess"}, // 7 guesses, max is 6
						},
					},
				},
				GameState: &GameState{
					Word:       "about",
					MaxGuesses: 6,
					Status:     GameStatusActive,
				},
			},
			expectError: true,
			errorMsg:    "player player1 has too many guesses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.ValidateGameState(tt.room)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGameLogic_StartGame(t *testing.T) {
	logic := &GameLogic{}

	// Create test room
	room := &Room{
		ID:      "test-room",
		Players: make(map[string]*Player),
		GameState: &GameState{
			Status: GameStatusWaiting,
		},
	}

	// Add players
	room.Players["player1"] = &Player{
		ID:     "player1",
		Status: PlayerStatusActive,
		Guesses: []Guess{{Word: "old"}, {Word: "guess"}}, // Some old guesses
	}

	err := logic.StartGame(room, "about")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify game state was properly initialized
	if room.GameState.Status != GameStatusActive {
		t.Errorf("expected status %s, got %s", GameStatusActive, room.GameState.Status)
	}

	if room.GameState.Word != "about" {
		t.Errorf("expected word 'about', got %s", room.GameState.Word)
	}

	if room.GameState.MaxGuesses != 6 {
		t.Errorf("expected max guesses 6, got %d", room.GameState.MaxGuesses)
	}

	// Verify player was reset
	player := room.Players["player1"]
	if len(player.Guesses) != 0 {
		t.Errorf("expected player guesses to be reset, got %d", len(player.Guesses))
	}

	if player.Status != PlayerStatusActive {
		t.Errorf("expected player status %s, got %s", PlayerStatusActive, player.Status)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}