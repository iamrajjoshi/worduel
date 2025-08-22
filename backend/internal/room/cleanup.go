package room

import (
	"context"
	"log"
	"sync"
	"time"

	"worduel-backend/internal/game"
)

const (
	// Default cleanup intervals and timeouts
	DefaultCleanupInterval     = 5 * time.Minute
	DefaultInactiveRoomTimeout = 30 * time.Minute
	DefaultEmptyRoomTimeout    = 5 * time.Minute
	DefaultFinishedGameTimeout = 15 * time.Minute
)

// CleanupConfig holds configuration for room cleanup
type CleanupConfig struct {
	CleanupInterval     time.Duration
	InactiveRoomTimeout time.Duration
	EmptyRoomTimeout    time.Duration
	FinishedGameTimeout time.Duration
	EnableLogging       bool
}

// CleanupService handles automatic room cleanup and expiration
type CleanupService struct {
	manager   *RoomManager
	config    CleanupConfig
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.RWMutex
	running   bool
	cleanupCh chan string // Channel for manual cleanup requests
}

// NewCleanupService creates a new cleanup service with default configuration
func NewCleanupService(manager *RoomManager) *CleanupService {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &CleanupService{
		manager: manager,
		config: CleanupConfig{
			CleanupInterval:     DefaultCleanupInterval,
			InactiveRoomTimeout: DefaultInactiveRoomTimeout,
			EmptyRoomTimeout:    DefaultEmptyRoomTimeout,
			FinishedGameTimeout: DefaultFinishedGameTimeout,
			EnableLogging:       true,
		},
		ctx:       ctx,
		cancel:    cancel,
		cleanupCh: make(chan string, 100), // Buffered channel for cleanup requests
	}
}

// NewCleanupServiceWithConfig creates a cleanup service with custom configuration
func NewCleanupServiceWithConfig(manager *RoomManager, config CleanupConfig) *CleanupService {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &CleanupService{
		manager:   manager,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		cleanupCh: make(chan string, 100),
	}
}

// Start begins the cleanup service background goroutine
func (cs *CleanupService) Start() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if cs.running {
		return nil // Already running
	}
	
	cs.running = true
	cs.wg.Add(1)
	
	go cs.cleanupWorker()
	
	if cs.config.EnableLogging {
		log.Printf("Room cleanup service started with interval: %v", cs.config.CleanupInterval)
	}
	
	return nil
}

// Stop gracefully stops the cleanup service
func (cs *CleanupService) Stop() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if !cs.running {
		return nil // Already stopped
	}
	
	cs.running = false
	cs.cancel()
	cs.wg.Wait()
	
	if cs.config.EnableLogging {
		log.Println("Room cleanup service stopped")
	}
	
	return nil
}

// RequestCleanup manually requests cleanup of a specific room
func (cs *CleanupService) RequestCleanup(roomCode string) {
	select {
	case cs.cleanupCh <- roomCode:
		// Successfully queued cleanup request
	default:
		// Channel is full, log warning
		if cs.config.EnableLogging {
			log.Printf("Cleanup request queue full, dropping cleanup request for room: %s", roomCode)
		}
	}
}

// cleanupWorker is the main background goroutine that performs periodic cleanup
func (cs *CleanupService) cleanupWorker() {
	defer cs.wg.Done()
	
	ticker := time.NewTicker(cs.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-cs.ctx.Done():
			return
		case <-ticker.C:
			cs.performCleanup()
		case roomCode := <-cs.cleanupCh:
			cs.cleanupSpecificRoom(roomCode)
		}
	}
}

// performCleanup performs a full cleanup sweep of all rooms
func (cs *CleanupService) performCleanup() {
	start := time.Now()
	cleanedRooms := 0
	totalRooms := cs.manager.GetRoomCount()
	
	if cs.config.EnableLogging && totalRooms > 0 {
		log.Printf("Starting cleanup sweep of %d rooms", totalRooms)
	}
	
	rooms := cs.manager.GetAllRooms()
	
	for roomCode, room := range rooms {
		if cs.shouldCleanupRoom(room) {
			err := cs.cleanupRoom(roomCode, room)
			if err != nil {
				if cs.config.EnableLogging {
					log.Printf("Error cleaning up room %s: %v", roomCode, err)
				}
			} else {
				cleanedRooms++
			}
		}
	}
	
	duration := time.Since(start)
	if cs.config.EnableLogging && cleanedRooms > 0 {
		log.Printf("Cleanup completed: removed %d rooms in %v", cleanedRooms, duration)
	}
}

// cleanupSpecificRoom performs cleanup on a specific room
func (cs *CleanupService) cleanupSpecificRoom(roomCode string) {
	room, err := cs.manager.GetRoom(roomCode)
	if err != nil {
		// Room doesn't exist, nothing to clean up
		return
	}
	
	if cs.shouldCleanupRoom(room) {
		err := cs.cleanupRoom(roomCode, room)
		if err != nil && cs.config.EnableLogging {
			log.Printf("Error cleaning up room %s: %v", roomCode, err)
		}
	}
}

// shouldCleanupRoom determines if a room should be cleaned up based on various criteria
func (cs *CleanupService) shouldCleanupRoom(room *game.Room) bool {
	room.RLock()
	defer room.RUnlock()
	
	now := time.Now()
	
	// Check if room is empty and has exceeded empty timeout
	if len(room.Players) == 0 {
		return now.Sub(room.UpdatedAt) > cs.config.EmptyRoomTimeout
	}
	
	// Check if game is finished and has exceeded finished timeout
	room.GameState.RLock()
	gameFinished := room.GameState.Status == game.GameStatusFinished
	var finishedAt time.Time
	if room.GameState.FinishedAt != nil {
		finishedAt = *room.GameState.FinishedAt
	}
	room.GameState.RUnlock()
	
	if gameFinished && !finishedAt.IsZero() {
		return now.Sub(finishedAt) > cs.config.FinishedGameTimeout
	}
	
	// Check if room has been inactive (no player activity) for too long
	lastActivity := cs.getLastPlayerActivity(room)
	if now.Sub(lastActivity) > cs.config.InactiveRoomTimeout {
		return true
	}
	
	// Check for disconnected players that haven't been active
	disconnectedCount := 0
	for _, player := range room.Players {
		if player.Status == game.PlayerStatusDisconnected {
			if now.Sub(player.LastActivity) > cs.config.InactiveRoomTimeout {
				disconnectedCount++
			}
		}
	}
	
	// If all players are disconnected and inactive, cleanup the room
	return disconnectedCount == len(room.Players)
}

// getLastPlayerActivity returns the most recent activity time among all players
func (cs *CleanupService) getLastPlayerActivity(room *game.Room) time.Time {
	var lastActivity time.Time
	
	for _, player := range room.Players {
		if player.LastActivity.After(lastActivity) {
			lastActivity = player.LastActivity
		}
	}
	
	// If no player activity found, use room's UpdatedAt
	if lastActivity.IsZero() {
		lastActivity = room.UpdatedAt
	}
	
	return lastActivity
}

// cleanupRoom performs the actual cleanup of a room
func (cs *CleanupService) cleanupRoom(roomCode string, room *game.Room) error {
	room.Lock()
	
	// Log room cleanup details
	if cs.config.EnableLogging {
		playerCount := len(room.Players)
		gameStatus := room.GameState.Status
		timeSinceUpdate := time.Since(room.UpdatedAt)
		
		log.Printf("Cleaning up room %s: %d players, status: %s, inactive for: %v",
			roomCode, playerCount, gameStatus, timeSinceUpdate)
	}
	
	// Perform resource cleanup for each player
	for playerID, player := range room.Players {
		cs.cleanupPlayer(roomCode, playerID, player)
	}
	
	// Clear room data
	room.Players = make(map[string]*game.Player)
	room.Unlock()
	
	// Remove room from manager
	return cs.manager.RemoveRoom(roomCode)
}

// cleanupPlayer performs cleanup for a specific player
func (cs *CleanupService) cleanupPlayer(roomCode, playerID string, player *game.Player) {
	if cs.config.EnableLogging {
		log.Printf("Cleaning up player %s (%s) from room %s", player.Name, playerID, roomCode)
	}
	
	// Clear player's guesses to free memory
	player.Guesses = nil
	
	// Mark player as disconnected
	player.Status = game.PlayerStatusDisconnected
}

// GetCleanupStats returns statistics about the cleanup service
func (cs *CleanupService) GetCleanupStats() CleanupStats {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	return CleanupStats{
		IsRunning:       cs.running,
		CleanupInterval: cs.config.CleanupInterval,
		PendingRequests: len(cs.cleanupCh),
	}
}

// CleanupStats holds statistics about the cleanup service
type CleanupStats struct {
	IsRunning       bool
	CleanupInterval time.Duration
	PendingRequests int
}

// ForceCleanupExpiredRooms immediately cleans up all expired rooms
func (cs *CleanupService) ForceCleanupExpiredRooms() (int, error) {
	rooms := cs.manager.GetAllRooms()
	cleanedCount := 0
	
	for roomCode, room := range rooms {
		if cs.shouldCleanupRoom(room) {
			err := cs.cleanupRoom(roomCode, room)
			if err != nil {
				return cleanedCount, err
			}
			cleanedCount++
		}
	}
	
	if cs.config.EnableLogging {
		log.Printf("Force cleanup completed: removed %d rooms", cleanedCount)
	}
	
	return cleanedCount, nil
}

// UpdateConfig updates the cleanup service configuration
func (cs *CleanupService) UpdateConfig(config CleanupConfig) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.config = config
	
	if cs.config.EnableLogging {
		log.Printf("Cleanup service configuration updated: interval=%v, inactive_timeout=%v, empty_timeout=%v",
			config.CleanupInterval, config.InactiveRoomTimeout, config.EmptyRoomTimeout)
	}
}