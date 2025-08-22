package ws

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityMiddleware_ValidateConnection(t *testing.T) {
	// Create middleware with allowed origins
	allowedOrigins := []string{"http://localhost:3000", "https://example.com"}
	sm := NewSecurityMiddleware(allowedOrigins)

	tests := []struct {
		name           string
		origin         string
		clientID       string
		remoteAddr     string
		expectError    bool
		expectedError  error
	}{
		{
			name:          "allowed origin",
			origin:        "http://localhost:3000",
			clientID:      "client1",
			remoteAddr:    "127.0.0.1:12345",
			expectError:   false,
		},
		{
			name:          "disallowed origin",
			origin:        "http://malicious.com",
			clientID:      "client2",
			remoteAddr:    "192.168.1.100:12345",
			expectError:   true,
			expectedError: ErrInvalidOrigin,
		},
		{
			name:          "no origin header",
			origin:        "",
			clientID:      "client3",
			remoteAddr:    "127.0.0.1:12345",
			expectError:   false,
		},
		{
			name:          "case insensitive origin",
			origin:        "HTTP://LOCALHOST:3000",
			clientID:      "client4",
			remoteAddr:    "127.0.0.1:12345",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			req.RemoteAddr = tt.remoteAddr

			err := sm.ValidateConnection(req, tt.clientID)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedError != nil && err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Clean up the connection if it was successfully validated
			if err == nil {
				sm.OnConnectionClosed(tt.clientID, sm.getClientIP(req))
			}
		})
	}
}

func TestSecurityMiddleware_CheckMessageRate(t *testing.T) {
	sm := NewSecurityMiddleware([]string{})

	// Initialize connection
	req := httptest.NewRequest("GET", "/ws", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	clientID := "test-client"

	err := sm.ValidateConnection(req, clientID)
	if err != nil {
		t.Fatalf("Failed to validate connection: %v", err)
	}

	// Test message size limit
	t.Run("message too large", func(t *testing.T) {
		largeMessage := strings.Repeat("x", maxMessageSize+1)
		err := sm.CheckMessageRate(clientID, len(largeMessage))
		if err != ErrMessageTooLarge {
			t.Errorf("Expected ErrMessageTooLarge, got %v", err)
		}
	})

	t.Run("valid message size", func(t *testing.T) {
		validMessage := "valid message"
		err := sm.CheckMessageRate(clientID, len(validMessage))
		if err != nil {
			t.Errorf("Expected no error for valid message, got %v", err)
		}
	})

	// Test rate limiting
	t.Run("rate limit", func(t *testing.T) {
		// We already sent one message in the "valid message size" test
		// So we need to send maxMessagesPerMinute - 1 more
		for i := 1; i < maxMessagesPerMinute; i++ {
			err := sm.CheckMessageRate(clientID, 10)
			if err != nil {
				t.Errorf("Message %d should not be rate limited, got error: %v", i+1, err)
			}
		}

		// Next message should be rate limited
		err := sm.CheckMessageRate(clientID, 10)
		if err != ErrRateLimitExceeded {
			t.Errorf("Expected ErrRateLimitExceeded, got %v", err)
		}
	})

	// Clean up
	sm.OnConnectionClosed(clientID, sm.getClientIP(req))
}

func TestSecurityMiddleware_IPConnectionLimit(t *testing.T) {
	sm := NewSecurityMiddleware([]string{})
	
	// Create multiple connections from the same IP
	clientIP := "192.168.1.100"
	var successfulConnections []string

	// Connect up to the limit
	for i := 0; i < maxConnectionsPerIP; i++ {
		req := httptest.NewRequest("GET", "/ws", nil)
		req.RemoteAddr = clientIP + ":12345"
		clientID := "client" + string(rune('1'+i))

		err := sm.ValidateConnection(req, clientID)
		if err != nil {
			t.Errorf("Connection %d should be allowed, got error: %v", i+1, err)
		} else {
			successfulConnections = append(successfulConnections, clientID)
		}
	}

	// Next connection should be rejected
	req := httptest.NewRequest("GET", "/ws", nil)
	req.RemoteAddr = clientIP + ":12345"
	err := sm.ValidateConnection(req, "client-overflow")
	if err != ErrTooManyConnections {
		t.Errorf("Expected ErrTooManyConnections, got %v", err)
	}

	// Clean up connections
	for _, clientID := range successfulConnections {
		sm.OnConnectionClosed(clientID, clientIP)
	}
}

func TestSecurityMiddleware_GetClientIP(t *testing.T) {
	sm := NewSecurityMiddleware([]string{})

	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:       "remote addr only",
			remoteAddr: "192.168.1.100:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name:          "x-forwarded-for header",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195, 70.41.3.18, 150.172.238.178",
			expectedIP:    "203.0.113.195",
		},
		{
			name:       "x-real-ip header",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.195",
			expectedIP: "203.0.113.195",
		},
		{
			name:          "x-forwarded-for takes precedence",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195",
			xRealIP:       "70.41.3.18",
			expectedIP:    "203.0.113.195",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			clientIP := sm.getClientIP(req)
			if clientIP != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, clientIP)
			}
		})
	}
}

func TestSecurityMiddleware_GetSecurityStats(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000", "https://example.com"}
	sm := NewSecurityMiddleware(allowedOrigins)

	// Create some connections
	req1 := httptest.NewRequest("GET", "/ws", nil)
	req1.RemoteAddr = "127.0.0.1:12345"
	sm.ValidateConnection(req1, "client1")

	req2 := httptest.NewRequest("GET", "/ws", nil)
	req2.RemoteAddr = "127.0.0.2:12345"
	sm.ValidateConnection(req2, "client2")

	stats := sm.GetSecurityStats()

	if stats.TotalConnections != 2 {
		t.Errorf("Expected 2 total connections, got %d", stats.TotalConnections)
	}

	if stats.RateLimitedConnections != 2 {
		t.Errorf("Expected 2 rate limited connections, got %d", stats.RateLimitedConnections)
	}

	if stats.UniqueIPs != 2 {
		t.Errorf("Expected 2 unique IPs, got %d", stats.UniqueIPs)
	}

	if stats.AllowedOrigins != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", stats.AllowedOrigins)
	}

	// Clean up
	sm.OnConnectionClosed("client1", "127.0.0.1")
	sm.OnConnectionClosed("client2", "127.0.0.2")
}