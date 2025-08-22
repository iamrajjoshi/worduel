package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"worduel-backend/internal/api"
	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
)

// APITestServer provides a test server for API integration tests
type APITestServer struct {
	*httptest.Server
	RoomManager   *room.RoomManager
	Dictionary    *game.Dictionary
	APIMiddleware *api.APIMiddleware
}

// NewAPITestServer creates a new test server with all middleware and handlers
func NewAPITestServer() *APITestServer {
	// Initialize core components
	dictionary := game.NewDictionary()
	roomManager := room.NewRoomManager()
	
	// Setup middleware with test origins
	allowedOrigins := []string{"http://localhost:3000", "http://test.example.com"}
	apiMiddleware := api.NewAPIMiddleware(allowedOrigins)
	
	// Create router and register routes
	router := mux.NewRouter()
	
	// Register API handlers
	roomHandler := api.NewRoomHandler(roomManager)
	roomHandler.RegisterRoutes(router)
	
	healthHandler := api.NewHealthHandler(roomManager, dictionary, apiMiddleware)
	healthHandler.RegisterRoutes(router)
	
	// Apply middleware
	handler := apiMiddleware.ApplyMiddlewares(router)
	
	// Create test server
	server := httptest.NewServer(handler)
	
	return &APITestServer{
		Server:        server,
		RoomManager:   roomManager,
		Dictionary:    dictionary,
		APIMiddleware: apiMiddleware,
	}
}

// TestRoomCreation tests POST /api/rooms endpoint
func TestRoomCreation(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	tests := []struct {
		name           string
		body           string
		contentType    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Valid room creation with default values",
			body:           `{}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"roomId", "roomCode", "name", "createdAt"},
		},
		{
			name:           "Valid room creation with custom name",
			body:           `{"name": "Custom Game Room", "maxPlayers": 2}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"roomId", "roomCode", "name", "createdAt"},
		},
		{
			name:           "Invalid JSON body",
			body:           `{"name": "Invalid JSON"`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "code", "message"},
		},
		{
			name:           "Too many players",
			body:           `{"maxPlayers": 10}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "code"},
		},
		{
			name:           "Invalid content type",
			body:           `{"name": "Test"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedFields: []string{}, // Response is plain text, not JSON
		},
		{
			name:           "Empty body",
			body:           ``,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest, // Empty body is invalid JSON
			expectedFields: []string{"error", "code", "message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.body)
			req, err := http.NewRequest("POST", server.URL+"/api/rooms", body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Skip body parsing for non-JSON responses or empty expected fields
			if len(tt.expectedFields) == 0 {
				return
			}

			// Check if response is JSON before attempting to decode
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				// For non-JSON responses, skip JSON parsing
				return
			}

			// Parse response
			var response map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Check expected fields
			for _, field := range tt.expectedFields {
				if _, exists := response[field]; !exists {
					t.Errorf("Expected field '%s' missing in response", field)
				}
			}

			// Verify room creation for successful cases
			if tt.expectedStatus == http.StatusCreated {
				roomCode, ok := response["roomCode"].(string)
				if !ok || len(roomCode) != room.RoomCodeLength {
					t.Errorf("Invalid room code: %v", roomCode)
				}
			}
		})
	}
}

// TestRoomRetrieval tests GET /api/rooms/{id} endpoint
func TestRoomRetrieval(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	// Create a test room first
	testRoom, err := server.RoomManager.CreateRoom("Test Room", 2)
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	tests := []struct {
		name           string
		roomID         string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Valid room retrieval",
			roomID:         testRoom.ID,
			expectedStatus: http.StatusOK,
			expectedFields: []string{"roomId", "name", "playerCount", "maxPlayers", "gameStatus", "createdAt", "updatedAt"},
		},
		{
			name:           "Room not found",
			roomID:         "NONEXT", // Use proper 6-character format
			expectedStatus: http.StatusNotFound,
			expectedFields: []string{"error", "code", "message"},
		},
		{
			name:           "Invalid room ID format - too short",
			roomID:         "ABC",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "code", "message"},
		},
		{
			name:           "Invalid room ID format - too long",
			roomID:         "ABCDEFGHIJ",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "code", "message"},
		},
		{
			name:           "Empty room ID",
			roomID:         "",
			expectedStatus: http.StatusNotFound, // Router returns 404 for empty path
			expectedFields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + "/api/rooms/" + tt.roomID
			if tt.roomID == "" {
				url = server.URL + "/api/rooms/"
			}
			
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Skip body parsing for non-JSON responses or empty expected fields
			if len(tt.expectedFields) == 0 {
				return
			}

			// Check if response is JSON before attempting to decode
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				// For non-JSON responses, skip JSON parsing
				if tt.name == "Invalid content type" {
					// Debug: print the actual content type
					t.Logf("Content-Type header: '%s'", contentType)
				}
				return
			}

			// Parse response
			var response map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Check expected fields
			for _, field := range tt.expectedFields {
				if _, exists := response[field]; !exists {
					t.Errorf("Expected field '%s' missing in response", field)
				}
			}

			// Verify room data for successful case
			if tt.expectedStatus == http.StatusOK {
				if roomId, ok := response["roomId"].(string); !ok || roomId != testRoom.ID {
					t.Errorf("Expected roomId %s, got %v", testRoom.ID, response["roomId"])
				}
				if name, ok := response["name"].(string); !ok || name != "Test Room" {
					t.Errorf("Expected name 'Test Room', got %v", response["name"])
				}
			}
		})
	}
}

// TestHealthEndpoints tests all health check endpoints
func TestHealthEndpoints(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Main health check",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "timestamp", "version", "uptime", "system", "application", "dependencies"},
		},
		{
			name:           "Liveness probe",
			endpoint:       "/health/liveness",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "timestamp"},
		},
		{
			name:           "Readiness probe",
			endpoint:       "/health/readiness",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "timestamp", "dependencies"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.endpoint)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Check expected fields
			for _, field := range tt.expectedFields {
				if _, exists := response[field]; !exists {
					t.Errorf("Expected field '%s' missing in response", field)
				}
			}

			// Verify specific health check content
			if tt.endpoint == "/health" {
				if status, ok := response["status"].(string); !ok || status == "" {
					t.Error("Health status should be non-empty string")
				}
				
				// Check system metrics exist
				if system, ok := response["system"].(map[string]interface{}); ok {
					if _, hasMemory := system["memory"]; !hasMemory {
						t.Error("System metrics should include memory information")
					}
				}

				// Check application metrics exist
				if app, ok := response["application"].(map[string]interface{}); ok {
					if _, hasRooms := app["rooms"]; !hasRooms {
						t.Error("Application metrics should include room information")
					}
					if _, hasAPI := app["api"]; !hasAPI {
						t.Error("Application metrics should include API information")
					}
				}
			}
		})
	}
}

// TestCORSHeaders tests CORS configuration
func TestCORSHeaders(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	tests := []struct {
		name           string
		method         string
		origin         string
		endpoint       string
		expectedStatus int
		expectCORS     bool
	}{
		{
			name:           "Valid origin preflight",
			method:         "OPTIONS",
			origin:         "http://localhost:3000",
			endpoint:       "/api/rooms",
			expectedStatus: http.StatusOK,
			expectCORS:     true,
		},
		{
			name:           "Valid origin POST request",
			method:         "POST",
			origin:         "http://localhost:3000",
			endpoint:       "/api/rooms",
			expectedStatus: http.StatusCreated,
			expectCORS:     true,
		},
		{
			name:           "Invalid origin",
			method:         "POST",
			origin:         "http://malicious.example.com",
			endpoint:       "/api/rooms",
			expectedStatus: http.StatusCreated, // Request still processes, but no CORS headers
			expectCORS:     false,
		},
		{
			name:           "No origin header",
			method:         "GET",
			origin:         "",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
			expectCORS:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.method == "POST" {
				body = strings.NewReader(`{}`)
			}

			req, err := http.NewRequest(tt.method, server.URL+tt.endpoint, body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check CORS headers
			corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			if tt.expectCORS {
				if corsOrigin != tt.origin {
					t.Errorf("Expected CORS origin %s, got %s", tt.origin, corsOrigin)
				}
				
				allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
				if allowMethods == "" {
					t.Error("Expected Access-Control-Allow-Methods header")
				}
			} else {
				if corsOrigin != "" && corsOrigin == tt.origin {
					t.Errorf("Unexpected CORS origin header: %s", corsOrigin)
				}
			}
		})
	}
}

// TestSecurityHeaders tests security middleware headers
func TestSecurityHeaders(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'",
		"Server":                 "Worduel-Backend",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := resp.Header.Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got: %s", header, expectedValue, actualValue)
		}
	}
}

// TestRateLimiting tests API rate limiting functionality
func TestRateLimiting(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	// Make many requests in quick succession to trigger rate limiting
	const numRequests = 125 // Slightly above the 120/minute limit
	var wg sync.WaitGroup
	var mu sync.Mutex
	var rateLimitHit bool

	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			
			resp, err := http.Get(server.URL + "/health")
			if err != nil {
				t.Logf("Request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			// Check for rate limiting
			if resp.StatusCode == http.StatusTooManyRequests {
				mu.Lock()
				rateLimitHit = true
				mu.Unlock()

				// Verify rate limit headers
				retryAfter := resp.Header.Get("Retry-After")
				if retryAfter == "" {
					t.Error("Rate limited response should include Retry-After header")
				}

				rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining")
				if rateLimitRemaining != "0" {
					t.Errorf("Rate limited response should have X-RateLimit-Remaining: 0, got: %s", rateLimitRemaining)
				}
			}

			// All responses should have rate limit headers
			rateLimitLimit := resp.Header.Get("X-RateLimit-Limit")
			if rateLimitLimit == "" {
				t.Error("Response should include X-RateLimit-Limit header")
			}
		}()
	}

	wg.Wait()

	if !rateLimitHit {
		t.Error("Expected rate limiting to be triggered with many concurrent requests")
	}
}

// TestRequestSizeLimits tests request size validation
func TestRequestSizeLimits(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	// Create a large request body (over 1MB)
	largeBody := strings.Repeat("x", 1024*1024+1)
	
	req, err := http.NewRequest("POST", server.URL+"/api/rooms", strings.NewReader(largeBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status %d for large request, got %d", http.StatusRequestEntityTooLarge, resp.StatusCode)
	}
}

// TestErrorHandling tests error handling middleware
func TestErrorHandling(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		body           string
		contentType    string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "Invalid JSON",
			endpoint:       "/api/rooms",
			method:         "POST",
			body:           `{"invalid": json}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  true,
		},
		{
			name:           "Method not allowed",
			endpoint:       "/api/rooms",
			method:         "DELETE",
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "Not found",
			endpoint:       "/api/nonexistent",
			method:         "GET",
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusNotFound,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			req, err := http.NewRequest(tt.method, server.URL+tt.endpoint, body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkResponse {
				var response map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if _, hasError := response["error"]; !hasError {
					t.Error("Error response should contain 'error' field")
				}
			}
		})
	}
}

// TestAPIPerformance tests response time requirements
func TestAPIPerformance(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	endpoints := []struct {
		name   string
		method string
		url    string
		body   string
		maxDuration time.Duration
	}{
		{
			name:   "Health check performance",
			method: "GET",
			url:    "/health",
			body:   "",
			maxDuration: 100 * time.Millisecond,
		},
		{
			name:   "Room creation performance",
			method: "POST",
			url:    "/api/rooms",
			body:   `{}`,
			maxDuration: 100 * time.Millisecond,
		},
	}

	for _, tt := range endpoints {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			// Measure response time
			start := time.Now()
			
			req, err := http.NewRequest(tt.method, server.URL+tt.url, body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			duration := time.Since(start)

			if duration > tt.maxDuration {
				t.Errorf("Response time %v exceeded maximum %v", duration, tt.maxDuration)
			}

			// Ensure successful response
			if resp.StatusCode >= 400 {
				t.Errorf("Expected successful response, got status %d", resp.StatusCode)
			}
		})
	}
}

// TestConcurrentRequests tests handling of concurrent API requests
func TestConcurrentRequests(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	const numConcurrentRequests = 50
	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount, errorCount int

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg.Add(numConcurrentRequests)

	for i := 0; i < numConcurrentRequests; i++ {
		go func(requestID int) {
			defer wg.Done()

			body := fmt.Sprintf(`{"name": "Room %d"}`, requestID)
			req, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/api/rooms", strings.NewReader(body))
			if err != nil {
				t.Logf("Failed to create request %d: %v", requestID, err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Logf("Failed to make request %d: %v", requestID, err)
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			if resp.StatusCode == http.StatusCreated {
				successCount++
			} else {
				errorCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Allow for some rate limiting, but most requests should succeed
	if successCount < numConcurrentRequests/2 {
		t.Errorf("Expected at least %d successful requests, got %d", numConcurrentRequests/2, successCount)
	}

	t.Logf("Concurrent requests: %d successful, %d errors", successCount, errorCount)
}

// TestRequestLogging tests that requests are properly logged
func TestRequestLogging(t *testing.T) {
	server := NewAPITestServer()
	defer server.Close()

	// Make a request that should be logged
	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Note: In a real scenario, we'd capture logs and verify they contain
	// expected information. For this test, we're just ensuring the request
	// completes successfully with logging middleware active.
}

// BenchmarkAPIEndpoints benchmarks key API endpoints
func BenchmarkAPIEndpoints(b *testing.B) {
	server := NewAPITestServer()
	defer server.Close()

	b.Run("HealthCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := http.Get(server.URL + "/health")
			if err != nil {
				b.Fatalf("Failed to make request: %v", err)
			}
			resp.Body.Close()
		}
	})

	b.Run("RoomCreation", func(b *testing.B) {
		body := `{}`
		for i := 0; i < b.N; i++ {
			resp, err := http.Post(server.URL+"/api/rooms", "application/json", strings.NewReader(body))
			if err != nil {
				b.Fatalf("Failed to make request: %v", err)
			}
			resp.Body.Close()
		}
	})
}