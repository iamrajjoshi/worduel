package game

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStateManager_CreateRoom(t *testing.T) {
	sm := NewStateManager()

	room := sm.CreateRoom("room1", "ABC123", 2)

	if room == nil {
		t.Fatal("expected room to be created")
	}

	if room.ID != "room1" {
		t.Errorf("expected room ID 'room1', got '%s'", room.ID)
	}

	if room.Name != "ABC123" {
		t.Errorf("expected room name 'ABC123', got '%s'", room.Name)
	}

	if room.MaxPlayers != 2 {
		t.Errorf("expected max players 2, got %d", room.MaxPlayers)
	}

	if room.GameState.Status != GameStatusWaiting {
		t.Errorf("expected status %s, got %s", GameStatusWaiting, room.GameState.Status)
	}

	if len(room.Players) != 0 {
		t.Errorf("expected 0 players, got %d", len(room.Players))
	}

	// Verify room is stored in manager
	retrievedRoom, exists := sm.GetRoom("room1")
	if !exists {
		t.Error("expected room to exist in manager")
	}

	if retrievedRoom.ID != room.ID {
		t.Error("retrieved room does not match created room")
	}
}

func TestStateManager_GetRoom(t *testing.T) {
	sm := NewStateManager()

	// Test non-existent room
	_, exists := sm.GetRoom("nonexistent")
	if exists {
		t.Error("expected room to not exist")
	}

	// Create room and test retrieval
	originalRoom := sm.CreateRoom("room1", "ABC123", 2)
	retrievedRoom, exists := sm.GetRoom("room1")

	if !exists {
		t.Error("expected room to exist")
	}

	if retrievedRoom.ID != originalRoom.ID {
		t.Error("retrieved room ID does not match")
	}
}

func TestStateManager_RemoveRoom(t *testing.T) {
	sm := NewStateManager()

	// Create and then remove room
	sm.CreateRoom("room1", "ABC123", 2)
	sm.RemoveRoom("room1")

	// Verify room is removed
	_, exists := sm.GetRoom("room1")
	if exists {
		t.Error("expected room to be removed")
	}
}

func TestStateManager_Concurrency(t *testing.T) {
	sm := NewStateManager()
	var wg sync.WaitGroup
	roomCount := 100

	// Concurrently create rooms
	wg.Add(roomCount)
	for i := 0; i < roomCount; i++ {
		go func(id int) {
			defer wg.Done()
			roomID := fmt.Sprintf("room%d", id)
			sm.CreateRoom(roomID, fmt.Sprintf("CODE%d", id), 2)
		}(i)
	}
	wg.Wait()

	// Verify all rooms exist
	if sm.GetRoomCount() != roomCount {
		t.Errorf("expected %d rooms, got %d", roomCount, sm.GetRoomCount())
	}

	// Concurrently access rooms
	wg.Add(roomCount)
	for i := 0; i < roomCount; i++ {
		go func(id int) {
			defer wg.Done()
			roomID := fmt.Sprintf("room%d", id)
			_, exists := sm.GetRoom(roomID)
			if !exists {
				t.Errorf("room %s should exist", roomID)
			}
		}(i)
	}
	wg.Wait()
}

func TestRoom_AddPlayer(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 2)

	// Add first player
	err := room.AddPlayer("player1", "Alice")
	if err != nil {
		t.Errorf("unexpected error adding player: %v", err)
	}

	if room.GetPlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", room.GetPlayerCount())
	}

	player, exists := room.GetPlayer("player1")
	if !exists {
		t.Error("expected player to exist")
	}

	if player.Name != "Alice" {
		t.Errorf("expected player name 'Alice', got '%s'", player.Name)
	}

	if player.Status != PlayerStatusActive {
		t.Errorf("expected player status %s, got %s", PlayerStatusActive, player.Status)
	}

	// Add second player
	err = room.AddPlayer("player2", "Bob")
	if err != nil {
		t.Errorf("unexpected error adding second player: %v", err)
	}

	if room.GetPlayerCount() != 2 {
		t.Errorf("expected 2 players, got %d", room.GetPlayerCount())
	}

	// Try to add third player (should fail)
	err = room.AddPlayer("player3", "Charlie")
	if err != ErrRoomFull {
		t.Errorf("expected ErrRoomFull, got %v", err)
	}

	// Try to add duplicate player
	err = room.AddPlayer("player1", "Alice2")
	if err != ErrPlayerExists {
		t.Errorf("expected ErrPlayerExists, got %v", err)
	}
}

func TestRoom_UpdatePlayer(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 2)
	room.AddPlayer("player1", "Alice")

	// Update player score
	err := room.UpdatePlayer("player1", func(p *Player) error {
		p.Score = 100
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error updating player: %v", err)
	}

	player, _ := room.GetPlayer("player1")
	if player.Score != 100 {
		t.Errorf("expected score 100, got %d", player.Score)
	}

	// Try to update non-existent player
	err = room.UpdatePlayer("nonexistent", func(p *Player) error {
		return nil
	})

	if err != ErrPlayerNotFound {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestRoom_RemovePlayer(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 2)
	room.AddPlayer("player1", "Alice")
	room.AddPlayer("player2", "Bob")

	// Remove player
	err := room.RemovePlayer("player1")
	if err != nil {
		t.Errorf("unexpected error removing player: %v", err)
	}

	if room.GetPlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", room.GetPlayerCount())
	}

	_, exists := room.GetPlayer("player1")
	if exists {
		t.Error("expected player to be removed")
	}

	// Try to remove non-existent player
	err = room.RemovePlayer("nonexistent")
	if err != ErrPlayerNotFound {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestRoom_GameLifecycle(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 2)
	room.AddPlayer("player1", "Alice")
	room.AddPlayer("player2", "Bob")

	// Initial state should be waiting
	if room.GetGameStatus() != GameStatusWaiting {
		t.Errorf("expected status %s, got %s", GameStatusWaiting, room.GetGameStatus())
	}

	// Start game
	err := room.StartGame("about")
	if err != nil {
		t.Errorf("unexpected error starting game: %v", err)
	}

	if room.GetGameStatus() != GameStatusActive {
		t.Errorf("expected status %s, got %s", GameStatusActive, room.GetGameStatus())
	}

	if room.GetTargetWord() != "about" {
		t.Errorf("expected target word 'about', got '%s'", room.GetTargetWord())
	}

	// Verify players were reset
	player, _ := room.GetPlayer("player1")
	if len(player.Guesses) != 0 {
		t.Errorf("expected 0 guesses, got %d", len(player.Guesses))
	}

	if player.Score != 0 {
		t.Errorf("expected score 0, got %d", player.Score)
	}

	// End game with winner
	err = room.EndGame("player1")
	if err != nil {
		t.Errorf("unexpected error ending game: %v", err)
	}

	if room.GetGameStatus() != GameStatusFinished {
		t.Errorf("expected status %s, got %s", GameStatusFinished, room.GetGameStatus())
	}

	if room.GetGameWinner() != "player1" {
		t.Errorf("expected winner 'player1', got '%s'", room.GetGameWinner())
	}

	// Try to start game again (should fail)
	err = room.StartGame("world")
	if err == nil {
		t.Error("expected error when starting already finished game")
	}

	// Reset game
	err = room.ResetGame()
	if err != nil {
		t.Errorf("unexpected error resetting game: %v", err)
	}

	if room.GetGameStatus() != GameStatusWaiting {
		t.Errorf("expected status %s, got %s", GameStatusWaiting, room.GetGameStatus())
	}

	if room.GetTargetWord() != "" {
		t.Errorf("expected empty target word, got '%s'", room.GetTargetWord())
	}
}

func TestRoom_SerializeForClient(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 2)
	room.AddPlayer("player1", "Alice")
	room.AddPlayer("player2", "Bob")
	room.StartGame("about")

	// Add some guesses for testing
	room.UpdatePlayer("player1", func(p *Player) error {
		p.Guesses = append(p.Guesses, Guess{
			Word: "crane",
			Results: []LetterResult{
				LetterResultAbsent, LetterResultAbsent, LetterResultPresent,
				LetterResultAbsent, LetterResultAbsent,
			},
			Timestamp: time.Now(),
			IsCorrect: false,
		})
		return nil
	})

	// Serialize for player1
	data, err := room.SerializeForClient("player1")
	if err != nil {
		t.Errorf("unexpected error serializing: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Errorf("error unmarshaling JSON: %v", err)
	}

	// Verify basic room data
	if result["id"] != "room1" {
		t.Errorf("expected room ID 'room1', got '%v'", result["id"])
	}

	// Verify game state doesn't include target word
	gameState := result["game_state"].(map[string]interface{})
	if _, exists := gameState["word"]; exists {
		t.Error("game state should not include target word for client")
	}

	if gameState["status"] != string(GameStatusActive) {
		t.Errorf("expected status %s, got %v", GameStatusActive, gameState["status"])
	}

	// Verify player data
	players := result["players"].(map[string]interface{})
	player1Data := players["player1"].(map[string]interface{})
	player2Data := players["player2"].(map[string]interface{})

	// Player1 should have full guesses
	if _, exists := player1Data["guesses"]; !exists {
		t.Error("player1 should have full guess data")
	}

	// Player2 should only have guess patterns
	if _, exists := player2Data["guesses"]; exists {
		t.Error("player2 should not have full guess data")
	}

	if _, exists := player2Data["guess_patterns"]; !exists {
		t.Error("player2 should have guess patterns")
	}
}

func TestRoom_ValidateRoomState(t *testing.T) {
	sm := NewStateManager()

	tests := []struct {
		name        string
		setupRoom   func() *Room
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid room",
			setupRoom: func() *Room {
				room := sm.CreateRoom("room1", "ABC123", 2)
				room.AddPlayer("player1", "Alice")
				return room
			},
			expectError: false,
		},
		{
			name: "empty room ID",
			setupRoom: func() *Room {
				room := sm.CreateRoom("room1", "ABC123", 2)
				room.ID = ""
				return room
			},
			expectError: true,
			errorMsg:    "room ID cannot be empty",
		},
		{
			name: "invalid max players",
			setupRoom: func() *Room {
				room := sm.CreateRoom("room1", "ABC123", 2)
				room.MaxPlayers = 0
				return room
			},
			expectError: true,
			errorMsg:    "max players must be between 1 and 10",
		},
		{
			name: "active game without target word",
			setupRoom: func() *Room {
				room := sm.CreateRoom("room1", "ABC123", 2)
				room.AddPlayer("player1", "Alice")
				room.StartGame("about")
				room.GameState.Word = ""
				return room
			},
			expectError: true,
			errorMsg:    "active game must have a target word",
		},
		{
			name: "finished game without end time",
			setupRoom: func() *Room {
				room := sm.CreateRoom("room1", "ABC123", 2)
				room.AddPlayer("player1", "Alice")
				room.StartGame("about")
				room.EndGame("player1")
				room.GameState.FinishedAt = nil
				return room
			},
			expectError: true,
			errorMsg:    "finished game must have end time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := tt.setupRoom()
			err := room.ValidateRoomState()

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

func TestRoom_ConcurrentPlayerOperations(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 10) // Larger room for concurrency test

	var wg sync.WaitGroup
	playerCount := 8

	// Concurrently add players
	wg.Add(playerCount)
	for i := 0; i < playerCount; i++ {
		go func(id int) {
			defer wg.Done()
			playerID := fmt.Sprintf("player%d", id)
			playerName := fmt.Sprintf("Player%d", id)
			room.AddPlayer(playerID, playerName)
		}(i)
	}
	wg.Wait()

	if room.GetPlayerCount() != playerCount {
		t.Errorf("expected %d players, got %d", playerCount, room.GetPlayerCount())
	}

	// Concurrently update players
	wg.Add(playerCount)
	for i := 0; i < playerCount; i++ {
		go func(id int) {
			defer wg.Done()
			playerID := fmt.Sprintf("player%d", id)
			room.UpdatePlayer(playerID, func(p *Player) error {
				p.Score = id * 10
				return nil
			})
		}(i)
	}
	wg.Wait()

	// Verify updates
	for i := 0; i < playerCount; i++ {
		playerID := fmt.Sprintf("player%d", i)
		player, exists := room.GetPlayer(playerID)
		if !exists {
			t.Errorf("player %s should exist", playerID)
			continue
		}
		if player.Score != i*10 {
			t.Errorf("player %s expected score %d, got %d", playerID, i*10, player.Score)
		}
	}
}

func TestRoom_GetAllPlayers(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 2)
	room.AddPlayer("player1", "Alice")
	room.AddPlayer("player2", "Bob")

	// Add a guess to player1
	room.UpdatePlayer("player1", func(p *Player) error {
		p.Guesses = append(p.Guesses, Guess{
			Word: "test",
			Results: []LetterResult{
				LetterResultCorrect, LetterResultCorrect,
				LetterResultCorrect, LetterResultCorrect,
			},
			Timestamp: time.Now(),
			IsCorrect: false,
		})
		return nil
	})

	// Get all players (should be deep copies)
	allPlayers := room.GetAllPlayers()

	if len(allPlayers) != 2 {
		t.Errorf("expected 2 players, got %d", len(allPlayers))
	}

	// Verify deep copy by modifying the returned player
	if player, exists := allPlayers["player1"]; exists {
		player.Score = 999
		player.Guesses = append(player.Guesses, Guess{Word: "modified"})
	}

	// Original player should be unchanged
	originalPlayer, _ := room.GetPlayer("player1")
	if originalPlayer.Score == 999 {
		t.Error("original player was modified, copy was not deep enough")
	}

	if len(originalPlayer.Guesses) > 1 {
		t.Error("original player guesses were modified, copy was not deep enough")
	}
}

func TestRoom_ActivityTracking(t *testing.T) {
	sm := NewStateManager()
	room := sm.CreateRoom("room1", "ABC123", 2)

	initialActivity := room.GetLastActivity()

	// Wait a bit to ensure timestamps differ
	time.Sleep(10 * time.Millisecond)

	// Update activity
	room.UpdateActivity()
	newActivity := room.GetLastActivity()

	if !newActivity.After(initialActivity) {
		t.Error("activity timestamp should be updated")
	}

	// Adding a player should also update activity
	time.Sleep(10 * time.Millisecond)
	room.AddPlayer("player1", "Alice")
	playerAddActivity := room.GetLastActivity()

	if !playerAddActivity.After(newActivity) {
		t.Error("adding player should update activity timestamp")
	}
}