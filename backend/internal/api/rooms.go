package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"worduel-backend/internal/room"
)

// RoomHandler handles HTTP requests for room operations
type RoomHandler struct {
	roomManager *room.RoomManager
}

// NewRoomHandler creates a new RoomHandler instance
func NewRoomHandler(roomManager *room.RoomManager) *RoomHandler {
	return &RoomHandler{
		roomManager: roomManager,
	}
}

// CreateRoomRequest represents the request body for creating a room
type CreateRoomRequest struct {
	Name       string `json:"name,omitempty"`
	MaxPlayers int    `json:"maxPlayers,omitempty"`
}

// CreateRoomResponse represents the response body for creating a room
type CreateRoomResponse struct {
	RoomID    string    `json:"roomId"`
	RoomCode  string    `json:"roomCode"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetRoomResponse represents the response body for getting room status
type GetRoomResponse struct {
	RoomID      string                 `json:"roomId"`
	Name        string                 `json:"name"`
	PlayerCount int                    `json:"playerCount"`
	MaxPlayers  int                    `json:"maxPlayers"`
	GameStatus  string                 `json:"gameStatus"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Players     map[string]PlayerInfo  `json:"players,omitempty"`
}

// PlayerInfo represents player information in API responses (without sensitive data)
type PlayerInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Score  int    `json:"score"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// CreateRoom handles POST /api/rooms requests
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	// Set defaults
	if req.Name == "" {
		req.Name = "Game Room"
	}
	if req.MaxPlayers <= 0 {
		req.MaxPlayers = 2 // Default for competitive Wordle
	}

	// Validate max players (reasonable limit)
	if req.MaxPlayers > 4 {
		h.sendError(w, http.StatusBadRequest, "INVALID_MAX_PLAYERS", "Maximum 4 players allowed")
		return
	}

	room, err := h.roomManager.CreateRoom(req.Name, req.MaxPlayers)
	if err != nil {
		log.Printf("Failed to create room: %v", err)
		h.sendError(w, http.StatusInternalServerError, "ROOM_CREATION_FAILED", "Failed to create room")
		return
	}

	response := CreateRoomResponse{
		RoomID:    room.ID,
		RoomCode:  room.ID, // Room code is the same as room ID in this implementation
		Name:      room.Name,
		CreatedAt: room.CreatedAt,
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode create room response: %v", err)
	}
}

// GetRoom handles GET /api/rooms/{id} requests
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	roomID := strings.ToUpper(strings.TrimSpace(vars["id"]))

	if roomID == "" {
		h.sendError(w, http.StatusBadRequest, "MISSING_ROOM_ID", "Room ID is required")
		return
	}

	// Validate room ID format
	if len(roomID) != room.RoomCodeLength {
		h.sendError(w, http.StatusBadRequest, "INVALID_ROOM_ID", "Room ID must be 6 characters")
		return
	}

	gameRoom, err := h.roomManager.GetRoom(roomID)
	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			h.sendError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "Room not found")
		case room.ErrInvalidRoomCode:
			h.sendError(w, http.StatusBadRequest, "INVALID_ROOM_ID", "Invalid room ID format")
		default:
			log.Printf("Failed to get room %s: %v", roomID, err)
			h.sendError(w, http.StatusInternalServerError, "ROOM_FETCH_FAILED", "Failed to retrieve room")
		}
		return
	}

	gameRoom.RLock()
	defer gameRoom.RUnlock()

	// Convert players to API format (exclude sensitive information)
	players := make(map[string]PlayerInfo)
	for _, player := range gameRoom.Players {
		players[player.ID] = PlayerInfo{
			ID:     player.ID,
			Name:   player.Name,
			Status: string(player.Status),
			Score:  player.Score,
		}
	}

	response := GetRoomResponse{
		RoomID:      gameRoom.ID,
		Name:        gameRoom.Name,
		PlayerCount: len(gameRoom.Players),
		MaxPlayers:  gameRoom.MaxPlayers,
		GameStatus:  string(gameRoom.GameState.Status),
		CreatedAt:   gameRoom.CreatedAt,
		UpdatedAt:   gameRoom.UpdatedAt,
		Players:     players,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode get room response: %v", err)
	}
}

// HealthCheck handles GET /health requests
func (h *RoomHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roomCount := h.roomManager.GetRoomCount()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"rooms":     roomCount,
		"version":   "1.0.0",
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Failed to encode health check response: %v", err)
	}
}

// RegisterRoutes registers all room-related routes to the router
func (h *RoomHandler) RegisterRoutes(router *mux.Router) {
	// Room operations
	router.HandleFunc("/api/rooms", h.CreateRoom).Methods("POST")
	router.HandleFunc("/api/rooms/{id}", h.GetRoom).Methods("GET")
	
	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
}

// sendError sends a standardized error response
func (h *RoomHandler) sendError(w http.ResponseWriter, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    code,
		Message: message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}