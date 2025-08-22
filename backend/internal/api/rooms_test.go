package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"worduel-backend/internal/room"
)

func TestCreateRoom(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	handler := NewRoomHandler(roomManager)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test valid request
	reqBody := CreateRoomRequest{
		Name:       "Test Room",
		MaxPlayers: 2,
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/api/rooms", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", recorder.Code)
	}

	var response CreateRoomResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.RoomID == "" {
		t.Error("Expected room ID to be set")
	}

	if response.Name != "Test Room" {
		t.Errorf("Expected room name 'Test Room', got '%s'", response.Name)
	}
}

func TestCreateRoomWithDefaults(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	handler := NewRoomHandler(roomManager)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test request with empty body (should use defaults)
	req := httptest.NewRequest("POST", "/api/rooms", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", recorder.Code)
	}

	var response CreateRoomResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Name != "Game Room" {
		t.Errorf("Expected default room name 'Game Room', got '%s'", response.Name)
	}
}

func TestGetRoom(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	handler := NewRoomHandler(roomManager)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create a room first
	createdRoom, err := roomManager.CreateRoom("Test Room", 2)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Test getting the room
	req := httptest.NewRequest("GET", "/api/rooms/"+createdRoom.ID, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var response GetRoomResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.RoomID != createdRoom.ID {
		t.Errorf("Expected room ID '%s', got '%s'", createdRoom.ID, response.RoomID)
	}

	if response.Name != "Test Room" {
		t.Errorf("Expected room name 'Test Room', got '%s'", response.Name)
	}

	if response.PlayerCount != 0 {
		t.Errorf("Expected player count 0, got %d", response.PlayerCount)
	}

	if response.MaxPlayers != 2 {
		t.Errorf("Expected max players 2, got %d", response.MaxPlayers)
	}
}

func TestGetRoomNotFound(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	handler := NewRoomHandler(roomManager)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test getting non-existent room
	req := httptest.NewRequest("GET", "/api/rooms/NOTFND", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", recorder.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if response.Code != "ROOM_NOT_FOUND" {
		t.Errorf("Expected error code 'ROOM_NOT_FOUND', got '%s'", response.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	handler := NewRoomHandler(roomManager)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test health check
	req := httptest.NewRequest("GET", "/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}

	if _, exists := response["rooms"]; !exists {
		t.Error("Expected 'rooms' field in health response")
	}

	if _, exists := response["timestamp"]; !exists {
		t.Error("Expected 'timestamp' field in health response")
	}

	if response["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%v'", response["version"])
	}
}