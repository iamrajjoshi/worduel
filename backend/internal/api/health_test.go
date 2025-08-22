package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	dictionary := game.NewDictionary()
	handler := NewHealthHandler(roomManager, dictionary)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test health check endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Validate response structure
	if response.Status == "" {
		t.Error("Expected health status to be set")
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}

	if response.Uptime == "" {
		t.Error("Expected uptime to be set")
	}

	// Check system metrics
	if response.System.Goroutines <= 0 {
		t.Error("Expected goroutines count to be positive")
	}

	if response.System.CPUCount <= 0 {
		t.Error("Expected CPU count to be positive")
	}

	if response.System.Memory.Allocated == 0 {
		t.Error("Expected allocated memory to be positive")
	}

	// Check application metrics
	if response.Application.Rooms.Total < 0 {
		t.Error("Expected room total to be non-negative")
	}

	// Check dependencies
	if _, exists := response.Dependencies["dictionary"]; !exists {
		t.Error("Expected dictionary dependency to be checked")
	}

	dictHealth := response.Dependencies["dictionary"]
	if dictHealth.Status == "" {
		t.Error("Expected dictionary health status to be set")
	}
}

func TestHealthCheckWithRooms(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	dictionary := game.NewDictionary()
	handler := NewHealthHandler(roomManager, dictionary)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create some rooms for metrics testing
	room1, err := roomManager.CreateRoom("Test Room 1", 2)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	_, err = roomManager.CreateRoom("Test Room 2", 4)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Add a player to one room
	_, err = roomManager.JoinRoom(room1.ID, "player1", "Player One")
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}

	// Test health check with rooms
	req := httptest.NewRequest("GET", "/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Check room metrics
	if response.Application.Rooms.Total != 2 {
		t.Errorf("Expected 2 total rooms, got %d", response.Application.Rooms.Total)
	}

	if response.Application.Rooms.Active != 1 {
		t.Errorf("Expected 1 active room, got %d", response.Application.Rooms.Active)
	}

	if response.Application.Rooms.TotalPlayers != 1 {
		t.Errorf("Expected 1 total player, got %d", response.Application.Rooms.TotalPlayers)
	}

	// Check room distribution
	if response.Application.Rooms.Distribution.Empty != 1 {
		t.Errorf("Expected 1 empty room, got %d", response.Application.Rooms.Distribution.Empty)
	}

	if response.Application.Rooms.Distribution.Single != 1 {
		t.Errorf("Expected 1 single-player room, got %d", response.Application.Rooms.Distribution.Single)
	}
}

func TestLivenessProbe(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	dictionary := game.NewDictionary()
	handler := NewHealthHandler(roomManager, dictionary)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test liveness probe
	req := httptest.NewRequest("GET", "/health/liveness", nil)
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

	if response["status"] != "alive" {
		t.Errorf("Expected status 'alive', got '%v'", response["status"])
	}

	if _, exists := response["timestamp"]; !exists {
		t.Error("Expected timestamp field in liveness response")
	}
}

func TestReadinessProbe(t *testing.T) {
	// Setup
	roomManager := room.NewRoomManager()
	dictionary := game.NewDictionary()
	handler := NewHealthHandler(roomManager, dictionary)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test readiness probe
	req := httptest.NewRequest("GET", "/health/readiness", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response (should be ready with healthy dictionary)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ready" {
		t.Errorf("Expected status 'ready', got '%v'", response["status"])
	}

	if _, exists := response["dependencies"]; !exists {
		t.Error("Expected dependencies field in readiness response")
	}
}

func TestReadinessProbeUnhealthy(t *testing.T) {
	// Setup with nil dictionary to simulate unhealthy dependency
	roomManager := room.NewRoomManager()
	handler := NewHealthHandler(roomManager, nil) // Nil dictionary
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test readiness probe with unhealthy dependency
	req := httptest.NewRequest("GET", "/health/readiness", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response (should be not ready with unhealthy dictionary)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "not_ready" {
		t.Errorf("Expected status 'not_ready', got '%v'", response["status"])
	}
}

func TestDictionaryHealthCheck(t *testing.T) {
	// Test with healthy dictionary
	roomManager := room.NewRoomManager()
	dictionary := game.NewDictionary()
	handler := NewHealthHandler(roomManager, dictionary)

	health := handler.checkDictionaryHealth()

	if health.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", health.Status)
	}

	if health.ResponseTime == "" {
		t.Error("Expected response time to be recorded")
	}

	// Test with nil dictionary
	handler2 := NewHealthHandler(roomManager, nil)
	health2 := handler2.checkDictionaryHealth()

	if health2.Status != HealthStatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", health2.Status)
	}

	if health2.Message == "" {
		t.Error("Expected error message for unhealthy dictionary")
	}
}

func TestSystemMetricsCollection(t *testing.T) {
	roomManager := room.NewRoomManager()
	dictionary := game.NewDictionary()
	handler := NewHealthHandler(roomManager, dictionary)

	metrics := handler.collectSystemMetrics()

	if metrics.Goroutines <= 0 {
		t.Error("Expected positive goroutine count")
	}

	if metrics.CPUCount <= 0 {
		t.Error("Expected positive CPU count")
	}

	if metrics.Memory.Allocated == 0 {
		t.Error("Expected non-zero allocated memory")
	}

	if metrics.Memory.HeapAlloc == 0 {
		t.Error("Expected non-zero heap allocation")
	}
}

func TestOverallHealthDetermination(t *testing.T) {
	roomManager := room.NewRoomManager()
	dictionary := game.NewDictionary()
	handler := NewHealthHandler(roomManager, dictionary)

	// Test healthy scenario
	systemMetrics := SystemMetrics{
		Goroutines: 10,
		Memory: MemoryMetrics{
			HeapAlloc: 1024 * 1024, // 1MB
		},
	}
	appMetrics := ApplicationMetrics{}
	dependencies := map[string]DependencyHealth{
		"test": {Status: HealthStatusHealthy},
	}

	status := handler.determineOverallHealth(systemMetrics, appMetrics, dependencies)
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", status)
	}

	// Test degraded scenario (high memory)
	systemMetrics.Memory.HeapAlloc = 200 * 1024 * 1024 // 200MB
	status = handler.determineOverallHealth(systemMetrics, appMetrics, dependencies)
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status, got %s", status)
	}

	// Test unhealthy scenario (unhealthy dependency)
	dependencies["test"] = DependencyHealth{Status: HealthStatusUnhealthy}
	status = handler.determineOverallHealth(systemMetrics, appMetrics, dependencies)
	if status != HealthStatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", status)
	}
}