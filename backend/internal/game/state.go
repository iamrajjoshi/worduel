package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrPlayerExists     = errors.New("player already exists")
	ErrGameNotStarted   = errors.New("game not started")
	ErrGameAlreadyEnded = errors.New("game already ended")
	ErrRoomFull         = errors.New("room is full")
)

// StateManager manages game states with thread safety
type StateManager struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new room with initial game state
func (sm *StateManager) CreateRoom(roomID, roomCode string, maxPlayers int) *Room {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	room := &Room{
		ID:         roomID,
		Name:       roomCode, // Using code as name for simplicity
		Players:    make(map[string]*Player),
		MaxPlayers: maxPlayers,
		CreatedAt:  now,
		UpdatedAt:  now,
		GameState: &GameState{
			Status:        GameStatusWaiting,
			WordLength:    5,
			MaxGuesses:    6,
			CurrentRound:  0,
			RoundDuration: 1800, // 30 minutes
		},
	}

	sm.rooms[roomID] = room
	return room
}

// GetRoom retrieves a room by ID with thread safety
func (sm *StateManager) GetRoom(roomID string) (*Room, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	room, exists := sm.rooms[roomID]
	return room, exists
}

// RemoveRoom removes a room from the manager
func (sm *StateManager) RemoveRoom(roomID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	delete(sm.rooms, roomID)
}

// GetAllRooms returns a copy of all rooms for cleanup operations
func (sm *StateManager) GetAllRooms() map[string]*Room {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	rooms := make(map[string]*Room, len(sm.rooms))
	for id, room := range sm.rooms {
		rooms[id] = room
	}
	return rooms
}

// GetRoomCount returns the number of active rooms
func (sm *StateManager) GetRoomCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	return len(sm.rooms)
}

// AddPlayer adds a new player to the room with thread safety
func (r *Room) AddPlayer(playerID, playerName string) error {
	r.Lock()
	defer r.Unlock()

	// Check if player already exists first
	if _, exists := r.Players[playerID]; exists {
		return ErrPlayerExists
	}

	// Check if room is full
	if len(r.Players) >= r.MaxPlayers {
		return ErrRoomFull
	}

	now := time.Now()
	player := &Player{
		ID:           playerID,
		Name:         playerName,
		Status:       PlayerStatusActive,
		Guesses:      make([]Guess, 0),
		Score:        0,
		ConnectedAt:  now,
		LastActivity: now,
	}

	r.Players[playerID] = player
	r.UpdatedAt = now

	return nil
}

// UpdatePlayer updates an existing player's information
func (r *Room) UpdatePlayer(playerID string, updater func(*Player) error) error {
	r.Lock()
	defer r.Unlock()

	player, exists := r.Players[playerID]
	if !exists {
		return ErrPlayerNotFound
	}

	if err := updater(player); err != nil {
		return err
	}

	player.LastActivity = time.Now()
	r.UpdatedAt = time.Now()

	return nil
}

// RemovePlayer removes a player from the room
func (r *Room) RemovePlayer(playerID string) error {
	r.Lock()
	defer r.Unlock()

	if _, exists := r.Players[playerID]; !exists {
		return ErrPlayerNotFound
	}

	delete(r.Players, playerID)
	r.UpdatedAt = time.Now()

	return nil
}

// GetPlayer safely retrieves a player
func (r *Room) GetPlayer(playerID string) (*Player, bool) {
	r.RLock()
	defer r.RUnlock()

	player, exists := r.Players[playerID]
	return player, exists
}

// GetAllPlayers returns a copy of all players
func (r *Room) GetAllPlayers() map[string]*Player {
	r.RLock()
	defer r.RUnlock()

	players := make(map[string]*Player, len(r.Players))
	for id, player := range r.Players {
		// Create a deep copy to avoid race conditions
		playerCopy := *player
		playerCopy.Guesses = make([]Guess, len(player.Guesses))
		copy(playerCopy.Guesses, player.Guesses)
		players[id] = &playerCopy
	}
	return players
}

// GetPlayerCount returns the current number of players
func (r *Room) GetPlayerCount() int {
	r.RLock()
	defer r.RUnlock()

	return len(r.Players)
}

// StartGame initializes game state for active play
func (r *Room) StartGame(targetWord string) error {
	r.Lock()
	defer r.Unlock()

	gameState := r.GameState
	gameState.Lock()
	defer gameState.Unlock()

	if gameState.Status != GameStatusWaiting {
		return errors.New("game can only be started from waiting status")
	}

	now := time.Now()
	gameState.Status = GameStatusActive
	gameState.Word = targetWord
	gameState.StartedAt = &now
	gameState.CurrentRound = 1
	gameState.Winner = ""
	gameState.FinishedAt = nil

	// Reset all players to active status
	for _, player := range r.Players {
		player.Status = PlayerStatusActive
		player.Guesses = make([]Guess, 0)
		player.Score = 0
		player.LastActivity = now
	}

	r.UpdatedAt = now
	return nil
}

// EndGame marks the game as finished with optional winner
func (r *Room) EndGame(winner string) error {
	r.Lock()
	defer r.Unlock()

	gameState := r.GameState
	gameState.Lock()
	defer gameState.Unlock()

	if gameState.Status == GameStatusFinished {
		return ErrGameAlreadyEnded
	}

	now := time.Now()
	gameState.Status = GameStatusFinished
	gameState.Winner = winner
	gameState.FinishedAt = &now

	// Update player statuses
	for id, player := range r.Players {
		if id == winner {
			player.Status = PlayerStatusFinished
		} else if player.Status == PlayerStatusActive {
			player.Status = PlayerStatusFinished
		}
		player.LastActivity = now
	}

	r.UpdatedAt = now
	return nil
}

// ResetGame resets game state back to waiting
func (r *Room) ResetGame() error {
	r.Lock()
	defer r.Unlock()

	gameState := r.GameState
	gameState.Lock()
	defer gameState.Unlock()

	now := time.Now()
	gameState.Status = GameStatusWaiting
	gameState.Word = ""
	gameState.CurrentRound = 0
	gameState.Winner = ""
	gameState.StartedAt = nil
	gameState.FinishedAt = nil

	// Reset all players
	for _, player := range r.Players {
		player.Status = PlayerStatusActive
		player.Guesses = make([]Guess, 0)
		player.Score = 0
		player.LastActivity = now
	}

	r.UpdatedAt = now
	return nil
}

// GetGameStatus safely retrieves current game status
func (r *Room) GetGameStatus() GameStatus {
	r.RLock()
	defer r.RUnlock()

	r.GameState.RLock()
	defer r.GameState.RUnlock()

	return r.GameState.Status
}

// GetGameWinner safely retrieves the game winner
func (r *Room) GetGameWinner() string {
	r.RLock()
	defer r.RUnlock()

	r.GameState.RLock()
	defer r.GameState.RUnlock()

	return r.GameState.Winner
}

// GetTargetWord safely retrieves the target word (for game logic only)
func (r *Room) GetTargetWord() string {
	r.RLock()
	defer r.RUnlock()

	r.GameState.RLock()
	defer r.GameState.RUnlock()

	return r.GameState.Word
}

// IsGameActive checks if the game is currently active
func (r *Room) IsGameActive() bool {
	return r.GetGameStatus() == GameStatusActive
}

// IsGameFinished checks if the game is finished
func (r *Room) IsGameFinished() bool {
	return r.GetGameStatus() == GameStatusFinished
}

// GetLastActivity returns the last activity timestamp
func (r *Room) GetLastActivity() time.Time {
	r.RLock()
	defer r.RUnlock()

	return r.UpdatedAt
}

// UpdateActivity updates the room's last activity timestamp
func (r *Room) UpdateActivity() {
	r.Lock()
	defer r.Unlock()

	r.UpdatedAt = time.Now()
}

// SerializeForClient serializes room state for client consumption (excluding sensitive data)
func (r *Room) SerializeForClient(playerID string) ([]byte, error) {
	r.RLock()
	defer r.RUnlock()

	r.GameState.RLock()
	defer r.GameState.RUnlock()

	// Create client-safe version of game state
	clientGameState := struct {
		Status        GameStatus `json:"status"`
		WordLength    int        `json:"word_length"`
		MaxGuesses    int        `json:"max_guesses"`
		CurrentRound  int        `json:"current_round"`
		StartedAt     *time.Time `json:"started_at,omitempty"`
		FinishedAt    *time.Time `json:"finished_at,omitempty"`
		Winner        string     `json:"winner,omitempty"`
		RoundDuration int        `json:"round_duration_seconds"`
	}{
		Status:        r.GameState.Status,
		WordLength:    r.GameState.WordLength,
		MaxGuesses:    r.GameState.MaxGuesses,
		CurrentRound:  r.GameState.CurrentRound,
		StartedAt:     r.GameState.StartedAt,
		FinishedAt:    r.GameState.FinishedAt,
		Winner:        r.GameState.Winner,
		RoundDuration: r.GameState.RoundDuration,
	}

	// Create client-safe version of players
	clientPlayers := make(map[string]interface{})
	for id, player := range r.Players {
		playerData := map[string]interface{}{
			"id":            player.ID,
			"name":          player.Name,
			"status":        player.Status,
			"score":         player.Score,
			"guess_count":   len(player.Guesses),
			"connected_at":  player.ConnectedAt,
			"last_activity": player.LastActivity,
		}

		// Include full guess details only for the requesting player
		if id == playerID {
			playerData["guesses"] = player.Guesses
		} else {
			// For other players, only show guess patterns (results) without words
			patterns := make([][]LetterResult, len(player.Guesses))
			for i, guess := range player.Guesses {
				patterns[i] = guess.Results
			}
			playerData["guess_patterns"] = patterns
		}

		clientPlayers[id] = playerData
	}

	// Create complete room state
	roomState := map[string]interface{}{
		"id":           r.ID,
		"name":         r.Name,
		"max_players":  r.MaxPlayers,
		"created_at":   r.CreatedAt,
		"updated_at":   r.UpdatedAt,
		"game_state":   clientGameState,
		"players":      clientPlayers,
	}

	return json.Marshal(roomState)
}

// SerializeForAdmin serializes complete room state for admin/debug purposes
func (r *Room) SerializeForAdmin() ([]byte, error) {
	r.RLock()
	defer r.RUnlock()

	r.GameState.RLock()
	defer r.GameState.RUnlock()

	// Include everything for admin view (including target word)
	adminGameState := struct {
		Status        GameStatus `json:"status"`
		Word          string     `json:"word"`
		WordLength    int        `json:"word_length"`
		MaxGuesses    int        `json:"max_guesses"`
		CurrentRound  int        `json:"current_round"`
		StartedAt     *time.Time `json:"started_at,omitempty"`
		FinishedAt    *time.Time `json:"finished_at,omitempty"`
		Winner        string     `json:"winner,omitempty"`
		RoundDuration int        `json:"round_duration_seconds"`
	}{
		Status:        r.GameState.Status,
		Word:          r.GameState.Word,
		WordLength:    r.GameState.WordLength,
		MaxGuesses:    r.GameState.MaxGuesses,
		CurrentRound:  r.GameState.CurrentRound,
		StartedAt:     r.GameState.StartedAt,
		FinishedAt:    r.GameState.FinishedAt,
		Winner:        r.GameState.Winner,
		RoundDuration: r.GameState.RoundDuration,
	}

	// Create complete room state with all data
	roomState := map[string]interface{}{
		"id":           r.ID,
		"name":         r.Name,
		"max_players":  r.MaxPlayers,
		"created_at":   r.CreatedAt,
		"updated_at":   r.UpdatedAt,
		"game_state":   adminGameState,
		"players":      r.Players,
	}

	return json.Marshal(roomState)
}

// ValidateRoomState performs comprehensive validation of room state
func (r *Room) ValidateRoomState() error {
	r.RLock()
	defer r.RUnlock()

	if r.ID == "" {
		return errors.New("room ID cannot be empty")
	}

	if r.MaxPlayers <= 0 || r.MaxPlayers > 10 {
		return errors.New("max players must be between 1 and 10")
	}

	if len(r.Players) > r.MaxPlayers {
		return fmt.Errorf("player count (%d) exceeds max players (%d)", len(r.Players), r.MaxPlayers)
	}

	if r.GameState == nil {
		return errors.New("game state cannot be nil")
	}

	r.GameState.RLock()
	defer r.GameState.RUnlock()

	// Validate game state consistency
	if r.GameState.Status == GameStatusActive {
		if r.GameState.Word == "" {
			return errors.New("active game must have a target word")
		}
		if r.GameState.StartedAt == nil {
			return errors.New("active game must have start time")
		}
	}

	if r.GameState.Status == GameStatusFinished {
		if r.GameState.FinishedAt == nil {
			return errors.New("finished game must have end time")
		}
		if r.GameState.Winner != "" {
			if _, exists := r.Players[r.GameState.Winner]; !exists {
				return fmt.Errorf("winner %s not found in players", r.GameState.Winner)
			}
		}
	}

	// Validate players
	for playerID, player := range r.Players {
		if player == nil {
			return fmt.Errorf("player %s cannot be nil", playerID)
		}
		if player.ID != playerID {
			return fmt.Errorf("player ID mismatch: %s != %s", player.ID, playerID)
		}
		if len(player.Guesses) > r.GameState.MaxGuesses {
			return fmt.Errorf("player %s has too many guesses", playerID)
		}
	}

	return nil
}

// GetRoomSummary returns a lightweight summary for listing purposes
func (r *Room) GetRoomSummary() map[string]interface{} {
	r.RLock()
	defer r.RUnlock()

	r.GameState.RLock()
	defer r.GameState.RUnlock()

	return map[string]interface{}{
		"id":           r.ID,
		"name":         r.Name,
		"player_count": len(r.Players),
		"max_players":  r.MaxPlayers,
		"status":       r.GameState.Status,
		"created_at":   r.CreatedAt,
		"updated_at":   r.UpdatedAt,
	}
}