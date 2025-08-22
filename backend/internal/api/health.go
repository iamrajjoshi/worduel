package api

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
)

// HealthHandler handles health check and system monitoring endpoints
type HealthHandler struct {
	roomManager *room.RoomManager
	dictionary  *game.Dictionary
	startTime   time.Time
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler(roomManager *room.RoomManager, dictionary *game.Dictionary) *HealthHandler {
	return &HealthHandler{
		roomManager: roomManager,
		dictionary:  dictionary,
		startTime:   time.Now(),
	}
}

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthResponse represents the comprehensive health check response
type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Uptime      string                 `json:"uptime"`
	System      SystemMetrics          `json:"system"`
	Application ApplicationMetrics     `json:"application"`
	Dependencies map[string]DependencyHealth `json:"dependencies"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	Memory    MemoryMetrics `json:"memory"`
	Goroutines int          `json:"goroutines"`
	CPUCount   int          `json:"cpuCount"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	Allocated     uint64 `json:"allocated"`     // bytes allocated and in use
	TotalAlloc    uint64 `json:"totalAlloc"`    // bytes allocated (even if freed)
	Sys           uint64 `json:"sys"`           // bytes obtained from system
	NumGC         uint32 `json:"numGC"`         // number of garbage collections
	HeapAlloc     uint64 `json:"heapAlloc"`     // bytes allocated on heap
	HeapSys       uint64 `json:"heapSys"`       // bytes obtained from system for heap
	HeapObjects   uint64 `json:"heapObjects"`   // number of allocated objects
}

// ApplicationMetrics represents application-specific metrics
type ApplicationMetrics struct {
	Rooms            RoomMetrics       `json:"rooms"`
	RequestCount     int64            `json:"requestCount,omitempty"`     // Future: request counter
	AverageResponseTime float64       `json:"averageResponseTime,omitempty"` // Future: response time tracking
}

// RoomMetrics represents room-related metrics
type RoomMetrics struct {
	Total         int            `json:"total"`
	Active        int            `json:"active"`        // rooms with players
	Waiting       int            `json:"waiting"`       // rooms waiting for players
	Playing       int            `json:"playing"`       // rooms with games in progress
	TotalPlayers  int            `json:"totalPlayers"`
	Distribution  RoomDistribution `json:"distribution"`
}

// RoomDistribution represents distribution of rooms by player count
type RoomDistribution struct {
	Empty     int `json:"empty"`     // 0 players
	Single    int `json:"single"`    // 1 player
	Paired    int `json:"paired"`    // 2 players
	Multiple  int `json:"multiple"`  // 3+ players
}

// DependencyHealth represents the health status of a dependency
type DependencyHealth struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	CheckedAt time.Time    `json:"checkedAt"`
	ResponseTime string    `json:"responseTime,omitempty"`
}

// HealthCheck handles GET /health requests with comprehensive health information
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	startTime := time.Now()
	
	// Collect system metrics
	systemMetrics := h.collectSystemMetrics()
	
	// Collect application metrics
	appMetrics := h.collectApplicationMetrics()
	
	// Check dependencies
	dependencies := h.checkDependencies()
	
	// Determine overall health status
	status := h.determineOverallHealth(systemMetrics, appMetrics, dependencies)
	
	uptime := time.Since(h.startTime)
	
	response := HealthResponse{
		Status:       status,
		Timestamp:    time.Now(),
		Version:      "1.0.0",
		Uptime:       uptime.String(),
		System:       systemMetrics,
		Application:  appMetrics,
		Dependencies: dependencies,
	}

	// Set appropriate HTTP status code based on health
	statusCode := http.StatusOK
	switch status {
	case HealthStatusDegraded:
		statusCode = http.StatusOK // Still return 200 for degraded
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode health check response: %v", err)
	}

	// Log health check duration for monitoring
	duration := time.Since(startTime)
	if duration > 100*time.Millisecond {
		log.Printf("Health check took %v (longer than expected)", duration)
	}
}

// LivenessProbe handles GET /health/liveness for Kubernetes-style liveness probes
func (h *HealthHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Simple liveness check - server is running
	response := map[string]interface{}{
		"status": "alive",
		"timestamp": time.Now(),
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReadinessProbe handles GET /health/readiness for Kubernetes-style readiness probes
func (h *HealthHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check if service is ready to accept traffic
	dependencies := h.checkDependencies()
	
	ready := true
	for _, dep := range dependencies {
		if dep.Status == HealthStatusUnhealthy {
			ready = false
			break
		}
	}
	
	status := "ready"
	statusCode := http.StatusOK
	
	if !ready {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}
	
	response := map[string]interface{}{
		"status": status,
		"timestamp": time.Now(),
		"dependencies": dependencies,
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// collectSystemMetrics gathers system-level metrics
func (h *HealthHandler) collectSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return SystemMetrics{
		Memory: MemoryMetrics{
			Allocated:   m.Alloc,
			TotalAlloc:  m.TotalAlloc,
			Sys:         m.Sys,
			NumGC:       m.NumGC,
			HeapAlloc:   m.HeapAlloc,
			HeapSys:     m.HeapSys,
			HeapObjects: m.HeapObjects,
		},
		Goroutines: runtime.NumGoroutine(),
		CPUCount:   runtime.NumCPU(),
	}
}

// collectApplicationMetrics gathers application-specific metrics
func (h *HealthHandler) collectApplicationMetrics() ApplicationMetrics {
	roomMetrics := h.collectRoomMetrics()
	
	return ApplicationMetrics{
		Rooms: roomMetrics,
		// Future: Add request count and response time tracking
	}
}

// collectRoomMetrics gathers room-related metrics
func (h *HealthHandler) collectRoomMetrics() RoomMetrics {
	rooms := h.roomManager.GetAllRooms()
	
	metrics := RoomMetrics{
		Total: len(rooms),
	}
	
	distribution := RoomDistribution{}
	totalPlayers := 0
	
	for _, room := range rooms {
		room.RLock()
		playerCount := len(room.Players)
		gameStatus := room.GameState.Status
		room.RUnlock()
		
		totalPlayers += playerCount
		
		// Count by game status
		switch gameStatus {
		case game.GameStatusWaiting:
			metrics.Waiting++
		case game.GameStatusActive:
			metrics.Playing++
		}
		
		// Count by player distribution
		switch playerCount {
		case 0:
			distribution.Empty++
		case 1:
			distribution.Single++
			metrics.Active++
		case 2:
			distribution.Paired++
			metrics.Active++
		default:
			distribution.Multiple++
			metrics.Active++
		}
	}
	
	metrics.TotalPlayers = totalPlayers
	metrics.Distribution = distribution
	
	return metrics
}

// checkDependencies performs health checks on system dependencies
func (h *HealthHandler) checkDependencies() map[string]DependencyHealth {
	dependencies := make(map[string]DependencyHealth)
	
	// Check dictionary loading
	dependencies["dictionary"] = h.checkDictionaryHealth()
	
	// Future: Add checks for external services, databases, etc.
	
	return dependencies
}

// checkDictionaryHealth checks if the dictionary service is healthy
func (h *HealthHandler) checkDictionaryHealth() DependencyHealth {
	startTime := time.Now()
	
	if h.dictionary == nil {
		return DependencyHealth{
			Status:    HealthStatusUnhealthy,
			Message:   "Dictionary service not initialized",
			CheckedAt: time.Now(),
		}
	}
	
	// Test dictionary functionality
	testWord := "about" // Use a word known to be in the common words list
	isValid := h.dictionary.IsValidGuess(testWord)
	responseTime := time.Since(startTime)
	
	if !isValid {
		return DependencyHealth{
			Status:       HealthStatusDegraded,
			Message:      "Dictionary validation test failed",
			CheckedAt:    time.Now(),
			ResponseTime: responseTime.String(),
		}
	}
	
	return DependencyHealth{
		Status:       HealthStatusHealthy,
		Message:      "Dictionary service operational",
		CheckedAt:    time.Now(),
		ResponseTime: responseTime.String(),
	}
}

// determineOverallHealth determines the overall system health based on metrics
func (h *HealthHandler) determineOverallHealth(system SystemMetrics, app ApplicationMetrics, deps map[string]DependencyHealth) HealthStatus {
	// Check dependencies first
	unhealthyDeps := 0
	degradedDeps := 0
	
	for _, dep := range deps {
		switch dep.Status {
		case HealthStatusUnhealthy:
			unhealthyDeps++
		case HealthStatusDegraded:
			degradedDeps++
		}
	}
	
	// If any critical dependency is unhealthy, mark as unhealthy
	if unhealthyDeps > 0 {
		return HealthStatusUnhealthy
	}
	
	// Check system resources
	memoryUsageMB := float64(system.Memory.HeapAlloc) / 1024 / 1024
	
	// Consider degraded if:
	// - Memory usage > 100MB
	// - Goroutine count > 1000
	// - Any dependencies degraded
	if memoryUsageMB > 100 || system.Goroutines > 1000 || degradedDeps > 0 {
		return HealthStatusDegraded
	}
	
	return HealthStatusHealthy
}

// RegisterRoutes registers all health-related routes to the router
func (h *HealthHandler) RegisterRoutes(router *mux.Router) {
	// Comprehensive health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
	
	// Kubernetes-style probes
	router.HandleFunc("/health/liveness", h.LivenessProbe).Methods("GET")
	router.HandleFunc("/health/readiness", h.ReadinessProbe).Methods("GET")
}