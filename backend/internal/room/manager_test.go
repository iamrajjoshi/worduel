package room

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"worduel-backend/internal/game"
)

func TestRoomManagerCreation(t *testing.T) {
	manager := NewRoomManager()
	
	if manager == nil {
		t.Fatal("Expected room manager to be created")
	}
	
	if manager.GetRoomCount() != 0 {
		t.Errorf("Expected 0 rooms initially, got %d", manager.GetRoomCount())
	}
}

func TestCreateRoom(t *testing.T) {
	manager := NewRoomManager()
	
	tests := []struct {
		name       string
		roomName   string
		maxPlayers int
		expectErr  bool
	}{
		{"Valid room creation", "Test Room", 4, false},
		{"Room with zero max players", "Zero Room", 0, false}, // Should use default
		{"Room with negative max players", "Negative Room", -1, false}, // Should use default
		{"Room with large max players", "Large Room", 100, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdRoom, err := manager.CreateRoom(tt.roomName, tt.maxPlayers)
			
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if err == nil {
				if createdRoom.Name != tt.roomName {
					t.Errorf("Expected room name %s, got %s", tt.roomName, createdRoom.Name)
				}
				
				if len(createdRoom.ID) != RoomCodeLength {
					t.Errorf("Expected room ID length %d, got %d", RoomCodeLength, len(createdRoom.ID))
				}
				
				expectedMaxPlayers := tt.maxPlayers
				if expectedMaxPlayers <= 0 {
					expectedMaxPlayers = DefaultMaxPlayers
				}
				
				if createdRoom.MaxPlayers != expectedMaxPlayers {
					t.Errorf("Expected max players %d, got %d", expectedMaxPlayers, createdRoom.MaxPlayers)
				}
				
				if createdRoom.GameState.Status != game.GameStatusWaiting {
					t.Errorf("Expected game status %s, got %s", game.GameStatusWaiting, createdRoom.GameState.Status)
				}
			}
		})
	}
}

func TestCreateMultipleRoomsUniqueIDs(t *testing.T) {
	manager := NewRoomManager()
	roomIDs := make(map[string]bool)
	numRooms := 100
	
	for i := 0; i < numRooms; i++ {
		createdRoom, err := manager.CreateRoom("Test Room", 4)
		if err != nil {
			t.Fatalf("Failed to create room %d: %v", i, err)
		}
		
		if roomIDs[createdRoom.ID] {
			t.Errorf("Duplicate room ID generated: %s", createdRoom.ID)
		}
		roomIDs[createdRoom.ID] = true
	}
	
	if manager.GetRoomCount() != numRooms {
		t.Errorf("Expected %d rooms, got %d", numRooms, manager.GetRoomCount())
	}
}

func TestJoinRoom(t *testing.T) {
	manager := NewRoomManager()
	testRoom, err := manager.CreateRoom("Test Room", 4)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	tests := []struct {
		name       string
		roomCode   string
		playerID   string
		playerName string
		expectErr  bool
		errType    error
	}{
		{"Valid join", testRoom.ID, "player1", "Player One", false, nil},
		{"Second valid join", testRoom.ID, "player2", "Player Two", false, nil},
		{"Player already exists", testRoom.ID, "player1", "Player One Again", true, ErrPlayerExists},
		{"Room full", testRoom.ID, "player3", "Player Three", false, nil}, // Should succeed with 4 max players
		{"Room full now", testRoom.ID, "player4", "Player Four", false, nil}, // Should succeed with 4 max players
		{"Room actually full", testRoom.ID, "player5", "Player Five", true, ErrRoomFull},
		{"Invalid room code", "INVALID", "player4", "Player Four", true, ErrInvalidRoomCode},
		{"Room not found", "ABC123", "player5", "Player Five", true, ErrRoomNotFound},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			joinedRoom, err := manager.JoinRoom(tt.roomCode, tt.playerID, tt.playerName)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errType != nil && err != tt.errType {
					t.Errorf("Expected error %v, got %v", tt.errType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				
				if joinedRoom == nil {
					t.Error("Expected room to be returned")
				} else {
					player, exists := joinedRoom.Players[tt.playerID]
					if !exists {
						t.Error("Expected player to be added to room")
					} else if player.Name != tt.playerName {
						t.Errorf("Expected player name %s, got %s", tt.playerName, player.Name)
					}
				}
			}
		})
	}
}

func TestGetRoom(t *testing.T) {
	manager := NewRoomManager()
	createdRoom, err := manager.CreateRoom("Test Room", 4)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	tests := []struct {
		name      string
		roomCode  string
		expectErr bool
		errType   error
	}{
		{"Valid room code", createdRoom.ID, false, nil},
		{"Invalid room code format", "abc", true, ErrInvalidRoomCode},
		{"Room not found", "ABC123", true, ErrRoomNotFound},
		{"Empty room code", "", true, ErrInvalidRoomCode},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedRoom, err := manager.GetRoom(tt.roomCode)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errType != nil && err != tt.errType {
					t.Errorf("Expected error %v, got %v", tt.errType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				
				if retrievedRoom == nil {
					t.Error("Expected room to be returned")
				} else if retrievedRoom.ID != tt.roomCode {
					t.Errorf("Expected room ID %s, got %s", tt.roomCode, retrievedRoom.ID)
				}
			}
		})
	}
}

func TestLeaveRoom(t *testing.T) {
	manager := NewRoomManager()
	testRoom, err := manager.CreateRoom("Test Room", 4)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	// Add players to room
	_, err = manager.JoinRoom(testRoom.ID, "player1", "Player One")
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}
	
	_, err = manager.JoinRoom(testRoom.ID, "player2", "Player Two")
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}
	
	tests := []struct {
		name      string
		roomCode  string
		playerID  string
		expectErr bool
	}{
		{"Valid leave", testRoom.ID, "player1", false},
		{"Player not in room", testRoom.ID, "nonexistent", true},
		{"Invalid room code", "INVALID", "player2", true},
		{"Room not found", "ABC123", "player2", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.LeaveRoom(tt.roomCode, tt.playerID)
			
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if err == nil {
				retrievedRoom, _ := manager.GetRoom(tt.roomCode)
				if _, exists := retrievedRoom.Players[tt.playerID]; exists {
					t.Error("Expected player to be removed from room")
				}
			}
		})
	}
}

func TestRemoveRoom(t *testing.T) {
	manager := NewRoomManager()
	testRoom, err := manager.CreateRoom("Test Room", 4)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	initialCount := manager.GetRoomCount()
	
	err = manager.RemoveRoom(testRoom.ID)
	if err != nil {
		t.Errorf("Failed to remove room: %v", err)
	}
	
	if manager.GetRoomCount() != initialCount-1 {
		t.Errorf("Expected room count to decrease by 1")
	}
	
	_, err = manager.GetRoom(testRoom.ID)
	if err != ErrRoomNotFound {
		t.Errorf("Expected room not found error, got %v", err)
	}
	
	// Test removing non-existent room
	err = manager.RemoveRoom("NONEXISTENT")
	if err != ErrRoomNotFound {
		t.Errorf("Expected room not found error for non-existent room, got %v", err)
	}
}

func TestConcurrentRoomCreation(t *testing.T) {
	manager := NewRoomManager()
	numGoroutines := 100
	numRoomsPerGoroutine := 10
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	createdRooms := make([]*game.Room, 0, numGoroutines*numRoomsPerGoroutine)
	errors := make([]error, 0)
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < numRoomsPerGoroutine; j++ {
				createdRoom, err := manager.CreateRoom("Concurrent Room", 4)
				
				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				} else {
					createdRooms = append(createdRooms, createdRoom)
				}
				mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d errors: %v", len(errors), errors[0])
	}
	
	expectedRooms := numGoroutines * numRoomsPerGoroutine
	if len(createdRooms) != expectedRooms {
		t.Errorf("Expected %d rooms, got %d", expectedRooms, len(createdRooms))
	}
	
	// Verify all room IDs are unique
	roomIDs := make(map[string]bool)
	for _, createdRoom := range createdRooms {
		if roomIDs[createdRoom.ID] {
			t.Errorf("Duplicate room ID found: %s", createdRoom.ID)
		}
		roomIDs[createdRoom.ID] = true
	}
}

func TestConcurrentJoinRoom(t *testing.T) {
	manager := NewRoomManager()
	testRoom, err := manager.CreateRoom("Concurrent Test", 50)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	numPlayers := 30
	var wg sync.WaitGroup
	var mu sync.Mutex
	successfulJoins := 0
	errors := make([]error, 0)
	
	for i := 0; i < numPlayers; i++ {
		wg.Add(1)
		go func(playerID int) {
			defer wg.Done()
			
			_, err := manager.JoinRoom(testRoom.ID, 
				fmt.Sprintf("player%d", playerID), 
				fmt.Sprintf("Player %d", playerID))
			
			mu.Lock()
			if err != nil {
				errors = append(errors, err)
			} else {
				successfulJoins++
			}
			mu.Unlock()
		}(i)
	}
	
	wg.Wait()
	
	if len(errors) > 0 {
		t.Errorf("Unexpected errors during concurrent join: %v", errors[0])
	}
	
	if successfulJoins != numPlayers {
		t.Errorf("Expected %d successful joins, got %d", numPlayers, successfulJoins)
	}
	
	retrievedRoom, err := manager.GetRoom(testRoom.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve room: %v", err)
	}
	
	if len(retrievedRoom.Players) != numPlayers {
		t.Errorf("Expected %d players in room, got %d", numPlayers, len(retrievedRoom.Players))
	}
}

func TestRoomCapacityLimits(t *testing.T) {
	manager := NewRoomManager()
	maxPlayers := 2
	
	testRoom, err := manager.CreateRoom("Capacity Test", maxPlayers)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	// Fill room to capacity
	for i := 0; i < maxPlayers; i++ {
		_, err := manager.JoinRoom(testRoom.ID, fmt.Sprintf("player%d", i), fmt.Sprintf("Player %d", i))
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}
	}
	
	// Try to add one more player (should fail)
	_, err = manager.JoinRoom(testRoom.ID, "overflow", "Overflow Player")
	if err != ErrRoomFull {
		t.Errorf("Expected room full error, got %v", err)
	}
	
	// Verify room still has correct number of players
	retrievedRoom, err := manager.GetRoom(testRoom.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve room: %v", err)
	}
	
	if len(retrievedRoom.Players) != maxPlayers {
		t.Errorf("Expected %d players, got %d", maxPlayers, len(retrievedRoom.Players))
	}
}

func TestEmptyRoomCleanup(t *testing.T) {
	manager := NewRoomManager()
	testRoom, err := manager.CreateRoom("Cleanup Test", 4)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	// Add and remove a player to trigger cleanup logic
	_, err = manager.JoinRoom(testRoom.ID, "player1", "Player One")
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}
	
	err = manager.LeaveRoom(testRoom.ID, "player1")
	if err != nil {
		t.Fatalf("Failed to leave room: %v", err)
	}
	
	// Room should still exist immediately after player leaves
	_, err = manager.GetRoom(testRoom.ID)
	if err != nil {
		t.Errorf("Room should still exist immediately after last player leaves: %v", err)
	}
	
	// Note: The actual cleanup happens in a goroutine with a 5-minute delay
	// We can't easily test the automatic cleanup here without mocking time
	// This test verifies the immediate behavior is correct
}

func TestRoomCodeValidation(t *testing.T) {
	manager := NewRoomManager()
	
	tests := []struct {
		name     string
		roomCode string
		valid    bool
	}{
		{"Valid 6-char alphanumeric", "ABC123", true},
		{"Valid all letters", "ABCDEF", true},
		{"Valid all numbers", "123456", true},
		{"Invalid too short", "ABC12", false},
		{"Invalid too long", "ABC1234", false},
		{"Valid lowercase", "abc123", true}, // Should be converted to uppercase internally
		{"Invalid special characters", "ABC-12", false},
		{"Empty string", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.GetRoom(tt.roomCode)
			
			if tt.valid {
				// Should get "room not found" error, not "invalid room code"
				if err != ErrRoomNotFound {
					t.Errorf("Expected room not found error for valid code, got %v", err)
				}
			} else {
				// Should get "invalid room code" error
				if err != ErrInvalidRoomCode {
					t.Errorf("Expected invalid room code error, got %v", err)
				}
			}
		})
	}
}

func TestPlayerActivityTracking(t *testing.T) {
	manager := NewRoomManager()
	testRoom, err := manager.CreateRoom("Activity Test", 4)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	beforeJoin := time.Now()
	
	_, err = manager.JoinRoom(testRoom.ID, "player1", "Player One")
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}
	
	afterJoin := time.Now()
	
	retrievedRoom, err := manager.GetRoom(testRoom.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve room: %v", err)
	}
	
	player := retrievedRoom.Players["player1"]
	if player == nil {
		t.Fatal("Expected player to exist")
	}
	
	// Check that timestamps are reasonable
	if player.ConnectedAt.Before(beforeJoin) || player.ConnectedAt.After(afterJoin) {
		t.Errorf("ConnectedAt timestamp not within expected range")
	}
	
	if player.LastActivity.Before(beforeJoin) || player.LastActivity.After(afterJoin) {
		t.Errorf("LastActivity timestamp not within expected range")
	}
	
	if player.Status != game.PlayerStatusActive {
		t.Errorf("Expected player status %s, got %s", game.PlayerStatusActive, player.Status)
	}
}

func TestGetAllRooms(t *testing.T) {
	manager := NewRoomManager()
	
	// Create multiple rooms
	numRooms := 5
	createdRooms := make([]*game.Room, 0, numRooms)
	
	for i := 0; i < numRooms; i++ {
		createdRoom, err := manager.CreateRoom(fmt.Sprintf("Room %d", i), 4)
		if err != nil {
			t.Fatalf("Failed to create room %d: %v", i, err)
		}
		createdRooms = append(createdRooms, createdRoom)
	}
	
	allRooms := manager.GetAllRooms()
	
	if len(allRooms) != numRooms {
		t.Errorf("Expected %d rooms, got %d", numRooms, len(allRooms))
	}
	
	// Verify all created rooms are in the returned map
	for _, createdRoom := range createdRooms {
		if _, exists := allRooms[createdRoom.ID]; !exists {
			t.Errorf("Created room %s not found in GetAllRooms result", createdRoom.ID)
		}
	}
	
	// Verify the returned map is a copy (not the original)
	allRooms["TEST"] = &game.Room{}
	allRoomsAgain := manager.GetAllRooms()
	
	if len(allRoomsAgain) != numRooms {
		t.Errorf("GetAllRooms should return a copy, but original was modified")
	}
}

