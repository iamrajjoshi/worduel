package game

import (
	"sync"
	"time"
)

// GameStatus represents the current status of a game
type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusActive   GameStatus = "active"
	GameStatusFinished GameStatus = "finished"
)

// PlayerStatus represents the current status of a player
type PlayerStatus string

const (
	PlayerStatusActive      PlayerStatus = "active"
	PlayerStatusDisconnected PlayerStatus = "disconnected"
	PlayerStatusFinished    PlayerStatus = "finished"
)

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	MessageTypeJoin         MessageType = "join"
	MessageTypeLeave        MessageType = "leave"
	MessageTypeGuess        MessageType = "guess"
	MessageTypeGameUpdate   MessageType = "game_update"
	MessageTypePlayerUpdate MessageType = "player_update"
	MessageTypeError        MessageType = "error"
	MessageTypeChat         MessageType = "chat"
)

// LetterResult represents the result of a single letter in a guess
type LetterResult string

const (
	LetterResultCorrect LetterResult = "correct"   // Green - correct letter in correct position
	LetterResultPresent LetterResult = "present"   // Yellow - letter exists but wrong position
	LetterResultAbsent  LetterResult = "absent"    // Gray - letter not in word
)

// Player represents a player in the game
type Player struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Status       PlayerStatus `json:"status"`
	Guesses      []Guess      `json:"guesses"`
	Score        int          `json:"score"`
	ConnectedAt  time.Time    `json:"connected_at"`
	LastActivity time.Time    `json:"last_activity"`
}

// Guess represents a single guess made by a player
type Guess struct {
	Word      string         `json:"word"`
	Results   []LetterResult `json:"results"`
	Timestamp time.Time      `json:"timestamp"`
	IsCorrect bool           `json:"is_correct"`
}

// Room represents a game room where players compete
type Room struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Players   map[string]*Player `json:"players"`
	GameState *GameState        `json:"game_state"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	MaxPlayers int              `json:"max_players"`
	mutex     sync.RWMutex
}

// GameState represents the current state of a game
type GameState struct {
	Status        GameStatus `json:"status"`
	Word          string     `json:"word,omitempty"`          // Hidden from clients during game
	WordLength    int        `json:"word_length"`
	MaxGuesses    int        `json:"max_guesses"`
	CurrentRound  int        `json:"current_round"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	Winner        string     `json:"winner,omitempty"`        // Player ID of winner
	RoundDuration int        `json:"round_duration_seconds"`  // Duration in seconds
	mutex         sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	PlayerID  string      `json:"player_id,omitempty"`
	RoomID    string      `json:"room_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// GameUpdateData represents data sent with game update messages
type GameUpdateData struct {
	GameState *GameState           `json:"game_state"`
	Players   map[string]*Player   `json:"players"`
}

// GuessData represents data sent with guess messages
type GuessData struct {
	Word string `json:"word"`
}

// JoinData represents data sent with join messages
type JoinData struct {
	PlayerName string `json:"player_name"`
	RoomID     string `json:"room_id"`
}

// ErrorData represents error message data
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Methods for Room with proper locking
func (r *Room) Lock() {
	r.mutex.Lock()
}

func (r *Room) Unlock() {
	r.mutex.Unlock()
}

func (r *Room) RLock() {
	r.mutex.RLock()
}

func (r *Room) RUnlock() {
	r.mutex.RUnlock()
}

// Methods for GameState with proper locking
func (g *GameState) Lock() {
	g.mutex.Lock()
}

func (g *GameState) Unlock() {
	g.mutex.Unlock()
}

func (g *GameState) RLock() {
	g.mutex.RLock()
}

func (g *GameState) RUnlock() {
	g.mutex.RUnlock()
}