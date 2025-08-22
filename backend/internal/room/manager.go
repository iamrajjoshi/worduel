package room

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"worduel-backend/internal/game"
)

const (
	DefaultMaxPlayers = 4
	RoomCodeLength    = 6
)

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrRoomFull        = errors.New("room is full")
	ErrPlayerExists    = errors.New("player already exists in room")
	ErrInvalidRoomCode = errors.New("invalid room code")
)

// RoomManager manages all game rooms with thread-safe operations
type RoomManager struct {
	rooms             map[string]*game.Room
	mutex             sync.RWMutex
	maxConcurrentRooms int
}

// NewRoomManager creates a new instance of RoomManager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:             make(map[string]*game.Room),
		maxConcurrentRooms: 1000, // Default max rooms
	}
}

// CreateRoom creates a new room with a unique room code
func (rm *RoomManager) CreateRoom(name string, maxPlayers int) (*game.Room, error) {
	if maxPlayers <= 0 {
		maxPlayers = DefaultMaxPlayers
	}

	roomCode, err := rm.generateUniqueRoomCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate room code: %w", err)
	}

	room := &game.Room{
		ID:         roomCode,
		Name:       name,
		Players:    make(map[string]*game.Player),
		MaxPlayers: maxPlayers,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		GameState: &game.GameState{
			Status:        game.GameStatusWaiting,
			WordLength:    5, // Default Wordle word length
			MaxGuesses:    6, // Default Wordle max guesses
			CurrentRound:  0,
			RoundDuration: 300, // 5 minutes default
		},
	}

	rm.mutex.Lock()
	rm.rooms[roomCode] = room
	rm.mutex.Unlock()

	return room, nil
}

// JoinRoom adds a player to an existing room
func (rm *RoomManager) JoinRoom(roomCode, playerID, playerName string) (*game.Room, error) {
	if !rm.isValidRoomCode(roomCode) {
		return nil, ErrInvalidRoomCode
	}

	rm.mutex.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mutex.RUnlock()

	if !exists {
		return nil, ErrRoomNotFound
	}

	room.Lock()
	defer room.Unlock()

	// Check if room is full
	if len(room.Players) >= room.MaxPlayers {
		return nil, ErrRoomFull
	}

	// Check if player already exists
	if _, exists := room.Players[playerID]; exists {
		return nil, ErrPlayerExists
	}

	// Add player to room
	player := &game.Player{
		ID:           playerID,
		Name:         playerName,
		Status:       game.PlayerStatusActive,
		Guesses:      make([]game.Guess, 0),
		Score:        0,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
	}

	room.Players[playerID] = player
	room.UpdatedAt = time.Now()

	return room, nil
}

// GetRoom retrieves a room by its code
func (rm *RoomManager) GetRoom(roomCode string) (*game.Room, error) {
	if !rm.isValidRoomCode(roomCode) {
		return nil, ErrInvalidRoomCode
	}

	rm.mutex.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mutex.RUnlock()

	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// RemoveRoom removes a room from the manager
func (rm *RoomManager) RemoveRoom(roomCode string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.rooms[roomCode]; !exists {
		return ErrRoomNotFound
	}

	delete(rm.rooms, roomCode)
	return nil
}

// LeaveRoom removes a player from a room
func (rm *RoomManager) LeaveRoom(roomCode, playerID string) error {
	room, err := rm.GetRoom(roomCode)
	if err != nil {
		return err
	}

	room.Lock()
	defer room.Unlock()

	if _, exists := room.Players[playerID]; !exists {
		return errors.New("player not found in room")
	}

	delete(room.Players, playerID)
	room.UpdatedAt = time.Now()

	// Remove room if empty
	if len(room.Players) == 0 {
		go func() {
			time.Sleep(time.Minute * 5) // Wait 5 minutes before removing empty room
			rm.mutex.Lock()
			defer rm.mutex.Unlock()
			if room, exists := rm.rooms[roomCode]; exists && len(room.Players) == 0 {
				delete(rm.rooms, roomCode)
			}
		}()
	}

	return nil
}

// GetRoomCount returns the total number of active rooms
func (rm *RoomManager) GetRoomCount() int {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return len(rm.rooms)
}

// GetAllRooms returns a copy of all rooms (for admin/monitoring purposes)
func (rm *RoomManager) GetAllRooms() map[string]*game.Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	rooms := make(map[string]*game.Room)
	for k, v := range rm.rooms {
		rooms[k] = v
	}
	return rooms
}

// generateUniqueRoomCode generates a unique 6-character alphanumeric room code
func (rm *RoomManager) generateUniqueRoomCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const maxAttempts = 100

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := rm.generateRoomCode(charset)
		if err != nil {
			return "", err
		}

		rm.mutex.RLock()
		_, exists := rm.rooms[code]
		rm.mutex.RUnlock()

		if !exists {
			return code, nil
		}
	}

	return "", errors.New("failed to generate unique room code after maximum attempts")
}

// generateRoomCode generates a random room code using the specified charset
func (rm *RoomManager) generateRoomCode(charset string) (string, error) {
	code := make([]byte, RoomCodeLength)
	if _, err := rand.Read(code); err != nil {
		return "", err
	}

	for i, b := range code {
		code[i] = charset[b%byte(len(charset))]
	}

	return string(code), nil
}

// isValidRoomCode validates that a room code has the correct format
func (rm *RoomManager) isValidRoomCode(code string) bool {
	if len(code) != RoomCodeLength {
		return false
	}

	code = strings.ToUpper(code)
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}

// SetMaxConcurrentRooms sets the maximum number of concurrent rooms
func (rm *RoomManager) SetMaxConcurrentRooms(max int) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.maxConcurrentRooms = max
}

// CleanupExpiredRooms removes rooms that have been inactive for longer than the timeout
func (rm *RoomManager) CleanupExpiredRooms(timeout time.Duration) int {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	now := time.Now()
	var toDelete []string
	
	for roomID, room := range rm.rooms {
		// Check if room has been inactive for too long
		if now.Sub(room.UpdatedAt) > timeout {
			toDelete = append(toDelete, roomID)
		}
	}
	
	// Delete expired rooms
	for _, roomID := range toDelete {
		delete(rm.rooms, roomID)
	}
	
	return len(toDelete)
}

// Shutdown gracefully shuts down the room manager
func (rm *RoomManager) Shutdown() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// Clear all rooms
	rm.rooms = make(map[string]*game.Room)
}