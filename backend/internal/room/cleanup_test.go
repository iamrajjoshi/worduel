package room

import (
	"testing"
	"time"

	"worduel-backend/internal/game"
)

func TestNewCleanupService(t *testing.T) {
	manager := NewRoomManager()
	service := NewCleanupService(manager)

	if service == nil {
		t.Fatal("Expected cleanup service to be created")
	}

	if service.manager != manager {
		t.Error("Expected cleanup service to have the correct manager")
	}

	if service.config.CleanupInterval != DefaultCleanupInterval {
		t.Errorf("Expected default cleanup interval %v, got %v", DefaultCleanupInterval, service.config.CleanupInterval)
	}
}

func TestNewCleanupServiceWithConfig(t *testing.T) {
	manager := NewRoomManager()
	config := CleanupConfig{
		CleanupInterval:     2 * time.Minute,
		InactiveRoomTimeout: 10 * time.Minute,
		EmptyRoomTimeout:    1 * time.Minute,
		FinishedGameTimeout: 5 * time.Minute,
		EnableLogging:       false,
	}

	service := NewCleanupServiceWithConfig(manager, config)

	if service == nil {
		t.Fatal("Expected cleanup service to be created")
	}

	if service.config.CleanupInterval != config.CleanupInterval {
		t.Errorf("Expected cleanup interval %v, got %v", config.CleanupInterval, service.config.CleanupInterval)
	}

	if service.config.EnableLogging != config.EnableLogging {
		t.Errorf("Expected logging %v, got %v", config.EnableLogging, service.config.EnableLogging)
	}
}

func TestCleanupServiceStartStop(t *testing.T) {
	manager := NewRoomManager()
	service := NewCleanupService(manager)

	// Test starting the service
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start cleanup service: %v", err)
	}

	stats := service.GetCleanupStats()
	if !stats.IsRunning {
		t.Error("Expected service to be running after start")
	}

	// Test starting already running service
	err = service.Start()
	if err != nil {
		t.Fatalf("Failed to start already running service: %v", err)
	}

	// Test stopping the service
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop cleanup service: %v", err)
	}

	stats = service.GetCleanupStats()
	if stats.IsRunning {
		t.Error("Expected service to be stopped after stop")
	}

	// Test stopping already stopped service
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop already stopped service: %v", err)
	}
}

func TestShouldCleanupRoom(t *testing.T) {
	manager := NewRoomManager()
	config := CleanupConfig{
		CleanupInterval:     1 * time.Minute,
		InactiveRoomTimeout: 5 * time.Minute,
		EmptyRoomTimeout:    2 * time.Minute,
		FinishedGameTimeout: 3 * time.Minute,
		EnableLogging:       false,
	}
	service := NewCleanupServiceWithConfig(manager, config)

	tests := []struct {
		name     string
		room     *game.Room
		expected bool
	}{
		{
			name: "Empty room past timeout",
			room: &game.Room{
				ID:        "TEST1",
				Players:   make(map[string]*game.Player),
				UpdatedAt: time.Now().Add(-3 * time.Minute), // Past empty timeout
				GameState: &game.GameState{Status: game.GameStatusWaiting},
			},
			expected: true,
		},
		{
			name: "Empty room within timeout",
			room: &game.Room{
				ID:        "TEST2",
				Players:   make(map[string]*game.Player),
				UpdatedAt: time.Now().Add(-1 * time.Minute), // Within empty timeout
				GameState: &game.GameState{Status: game.GameStatusWaiting},
			},
			expected: false,
		},
		{
			name: "Finished game past timeout",
			room: &game.Room{
				ID:      "TEST3",
				Players: map[string]*game.Player{"player1": &game.Player{ID: "player1", LastActivity: time.Now()}},
				GameState: &game.GameState{
					Status:     game.GameStatusFinished,
					FinishedAt: timePtr(time.Now().Add(-4 * time.Minute)), // Past finished timeout
				},
			},
			expected: true,
		},
		{
			name: "Active room with recent activity",
			room: &game.Room{
				ID: "TEST4",
				Players: map[string]*game.Player{
					"player1": &game.Player{
						ID:           "player1",
						Status:       game.PlayerStatusActive,
						LastActivity: time.Now(), // Recent activity
					},
				},
				GameState: &game.GameState{Status: game.GameStatusActive},
			},
			expected: false,
		},
		{
			name: "Inactive room with old player activity",
			room: &game.Room{
				ID: "TEST5",
				Players: map[string]*game.Player{
					"player1": &game.Player{
						ID:           "player1",
						Status:       game.PlayerStatusActive,
						LastActivity: time.Now().Add(-6 * time.Minute), // Past inactive timeout
					},
				},
				GameState: &game.GameState{Status: game.GameStatusActive},
			},
			expected: true,
		},
		{
			name: "All players disconnected and inactive",
			room: &game.Room{
				ID: "TEST6",
				Players: map[string]*game.Player{
					"player1": &game.Player{
						ID:           "player1",
						Status:       game.PlayerStatusDisconnected,
						LastActivity: time.Now().Add(-6 * time.Minute), // Past inactive timeout
					},
					"player2": &game.Player{
						ID:           "player2",
						Status:       game.PlayerStatusDisconnected,
						LastActivity: time.Now().Add(-6 * time.Minute), // Past inactive timeout
					},
				},
				GameState: &game.GameState{Status: game.GameStatusActive},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldCleanupRoom(tt.room)
			if result != tt.expected {
				t.Errorf("shouldCleanupRoom() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetLastPlayerActivity(t *testing.T) {
	manager := NewRoomManager()
	service := NewCleanupService(manager)

	now := time.Now()
	oldTime := now.Add(-1 * time.Hour)
	recentTime := now.Add(-10 * time.Minute)

	room := &game.Room{
		ID: "TEST",
		Players: map[string]*game.Player{
			"player1": &game.Player{ID: "player1", LastActivity: oldTime},
			"player2": &game.Player{ID: "player2", LastActivity: recentTime},
			"player3": &game.Player{ID: "player3", LastActivity: oldTime},
		},
		UpdatedAt: oldTime,
	}

	lastActivity := service.getLastPlayerActivity(room)

	// Should return the most recent activity time
	if !lastActivity.Equal(recentTime) {
		t.Errorf("Expected last activity %v, got %v", recentTime, lastActivity)
	}

	// Test room with no players
	emptyRoom := &game.Room{
		ID:        "EMPTY",
		Players:   make(map[string]*game.Player),
		UpdatedAt: oldTime,
	}

	lastActivity = service.getLastPlayerActivity(emptyRoom)
	if !lastActivity.Equal(oldTime) {
		t.Errorf("Expected room UpdatedAt %v, got %v", oldTime, lastActivity)
	}
}

func TestRequestCleanup(t *testing.T) {
	manager := NewRoomManager()
	service := NewCleanupService(manager)

	// Test normal cleanup request
	service.RequestCleanup("ROOM1")

	stats := service.GetCleanupStats()
	if stats.PendingRequests != 1 {
		t.Errorf("Expected 1 pending request, got %d", stats.PendingRequests)
	}

	// Test multiple cleanup requests
	service.RequestCleanup("ROOM2")
	service.RequestCleanup("ROOM3")

	stats = service.GetCleanupStats()
	if stats.PendingRequests != 3 {
		t.Errorf("Expected 3 pending requests, got %d", stats.PendingRequests)
	}
}

func TestForceCleanupExpiredRooms(t *testing.T) {
	manager := NewRoomManager()
	config := CleanupConfig{
		CleanupInterval:     1 * time.Minute,
		InactiveRoomTimeout: 5 * time.Minute,
		EmptyRoomTimeout:    1 * time.Minute,
		FinishedGameTimeout: 1 * time.Minute,
		EnableLogging:       false,
	}
	service := NewCleanupServiceWithConfig(manager, config)

	// Create some test rooms
	room1, _ := manager.CreateRoom("Room 1", 4)
	room2, _ := manager.CreateRoom("Room 2", 4)
	room3, _ := manager.CreateRoom("Room 3", 4)

	// Make rooms expire by setting old UpdatedAt times
	room1.UpdatedAt = time.Now().Add(-2 * time.Minute) // Should be cleaned (empty)
	room2.UpdatedAt = time.Now().Add(-2 * time.Minute) // Should be cleaned (empty)

	// Add a player to room3 so it shouldn't be cleaned
	room3.Players["player1"] = &game.Player{
		ID:           "player1",
		Status:       game.PlayerStatusActive,
		LastActivity: time.Now(),
	}

	initialCount := manager.GetRoomCount()
	if initialCount != 3 {
		t.Fatalf("Expected 3 rooms initially, got %d", initialCount)
	}

	cleanedCount, err := service.ForceCleanupExpiredRooms()
	if err != nil {
		t.Fatalf("ForceCleanupExpiredRooms failed: %v", err)
	}

	if cleanedCount != 2 {
		t.Errorf("Expected 2 rooms to be cleaned, got %d", cleanedCount)
	}

	finalCount := manager.GetRoomCount()
	if finalCount != 1 {
		t.Errorf("Expected 1 room remaining, got %d", finalCount)
	}
}

func TestUpdateConfig(t *testing.T) {
	manager := NewRoomManager()
	service := NewCleanupService(manager)

	newConfig := CleanupConfig{
		CleanupInterval:     10 * time.Second,
		InactiveRoomTimeout: 2 * time.Minute,
		EmptyRoomTimeout:    30 * time.Second,
		FinishedGameTimeout: 1 * time.Minute,
		EnableLogging:       false,
	}

	service.UpdateConfig(newConfig)

	if service.config.CleanupInterval != newConfig.CleanupInterval {
		t.Errorf("Expected cleanup interval %v, got %v", newConfig.CleanupInterval, service.config.CleanupInterval)
	}

	if service.config.InactiveRoomTimeout != newConfig.InactiveRoomTimeout {
		t.Errorf("Expected inactive timeout %v, got %v", newConfig.InactiveRoomTimeout, service.config.InactiveRoomTimeout)
	}
}

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// Benchmark tests
func BenchmarkShouldCleanupRoom(b *testing.B) {
	manager := NewRoomManager()
	service := NewCleanupService(manager)

	room := &game.Room{
		ID: "BENCHMARK",
		Players: map[string]*game.Player{
			"player1": &game.Player{
				ID:           "player1",
				Status:       game.PlayerStatusActive,
				LastActivity: time.Now().Add(-10 * time.Minute),
			},
		},
		GameState: &game.GameState{Status: game.GameStatusActive},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.shouldCleanupRoom(room)
	}
}

func BenchmarkGetLastPlayerActivity(b *testing.B) {
	manager := NewRoomManager()
	service := NewCleanupService(manager)

	room := &game.Room{
		ID: "BENCHMARK",
		Players: map[string]*game.Player{
			"player1": &game.Player{ID: "player1", LastActivity: time.Now().Add(-1 * time.Hour)},
			"player2": &game.Player{ID: "player2", LastActivity: time.Now().Add(-30 * time.Minute)},
			"player3": &game.Player{ID: "player3", LastActivity: time.Now().Add(-10 * time.Minute)},
			"player4": &game.Player{ID: "player4", LastActivity: time.Now().Add(-5 * time.Minute)},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.getLastPlayerActivity(room)
	}
}