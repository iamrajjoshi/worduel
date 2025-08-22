package ws

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Security constants
const (
	// Rate limiting: 5 messages per 5 seconds per connection (for easier testing)
	maxMessagesPerMinute = 5
	rateLimitWindow      = time.Second * 5
	
	// Connection limits per IP
	maxConnectionsPerIP = 10
	
	// DoS protection
	maxConcurrentConnections = 1000
	
	// Rate limit reset interval
	rateLimitCleanupInterval = time.Minute * 5
)

// RateLimiter manages per-connection rate limiting
type RateLimiter struct {
	connections map[string]*ConnectionLimit
	mutex       sync.RWMutex
	
	// Global connection count
	totalConnections int
}

// ConnectionLimit tracks rate limiting data for a single connection
type ConnectionLimit struct {
	messages    []time.Time
	lastMessage time.Time
	violations  int
	createdAt   time.Time
	ipAddress   string
}

// IPConnectionTracker tracks connections per IP address
type IPConnectionTracker struct {
	connections map[string]int
	mutex       sync.RWMutex
}

// SecurityMiddleware provides WebSocket security features
type SecurityMiddleware struct {
	rateLimiter    *RateLimiter
	ipTracker      *IPConnectionTracker
	allowedOrigins map[string]bool
}

// NewSecurityMiddleware creates a new security middleware instance
func NewSecurityMiddleware(allowedOrigins []string) *SecurityMiddleware {
	originMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originMap[strings.ToLower(origin)] = true
	}
	
	// If no origins specified, allow localhost for development
	if len(originMap) == 0 {
		originMap["http://localhost:3000"] = true
		originMap["http://127.0.0.1:3000"] = true
		originMap["https://localhost:3000"] = true
		originMap["https://127.0.0.1:3000"] = true
	}
	
	sm := &SecurityMiddleware{
		rateLimiter: &RateLimiter{
			connections: make(map[string]*ConnectionLimit),
		},
		ipTracker: &IPConnectionTracker{
			connections: make(map[string]int),
		},
		allowedOrigins: originMap,
	}
	
	// Start cleanup routines
	go sm.startCleanupRoutines()
	
	return sm
}

// ValidateConnection performs initial connection validation
func (sm *SecurityMiddleware) ValidateConnection(r *http.Request, clientID string) error {
	// Check origin
	if err := sm.checkOrigin(r); err != nil {
		sm.logSecurityEvent("ORIGIN_REJECTED", clientID, r.RemoteAddr, err.Error())
		return err
	}
	
	// Get client IP
	clientIP := sm.getClientIP(r)
	
	// Check IP-based connection limits
	if err := sm.checkIPConnectionLimit(clientIP); err != nil {
		sm.logSecurityEvent("IP_LIMIT_EXCEEDED", clientID, clientIP, err.Error())
		return err
	}
	
	// Check global connection limits
	if err := sm.checkGlobalConnectionLimit(); err != nil {
		sm.logSecurityEvent("GLOBAL_LIMIT_EXCEEDED", clientID, clientIP, err.Error())
		return err
	}
	
	// Initialize rate limiting for this connection
	sm.initializeRateLimit(clientID, clientIP)
	
	sm.logSecurityEvent("CONNECTION_ACCEPTED", clientID, clientIP, "Connection validated")
	return nil
}

// CheckMessageRate validates message rate limits
func (sm *SecurityMiddleware) CheckMessageRate(clientID string, messageSize int) error {
	// Check message size
	if messageSize > maxMessageSize {
		sm.logSecurityEvent("MESSAGE_SIZE_EXCEEDED", clientID, "", 
			"Message size exceeded: %d bytes", messageSize)
		return ErrMessageTooLarge
	}
	
	// Check rate limit
	if err := sm.checkRateLimit(clientID); err != nil {
		sm.logSecurityEvent("RATE_LIMIT_EXCEEDED", clientID, "", err.Error())
		return err
	}
	
	return nil
}

// OnConnectionClosed handles connection cleanup
func (sm *SecurityMiddleware) OnConnectionClosed(clientID string, clientIP string) {
	sm.rateLimiter.mutex.Lock()
	delete(sm.rateLimiter.connections, clientID)
	sm.rateLimiter.totalConnections--
	sm.rateLimiter.mutex.Unlock()
	
	sm.ipTracker.mutex.Lock()
	if count, exists := sm.ipTracker.connections[clientIP]; exists {
		if count <= 1 {
			delete(sm.ipTracker.connections, clientIP)
		} else {
			sm.ipTracker.connections[clientIP] = count - 1
		}
	}
	sm.ipTracker.mutex.Unlock()
	
	sm.logSecurityEvent("CONNECTION_CLOSED", clientID, clientIP, "Connection cleaned up")
}

// checkOrigin validates the request origin
func (sm *SecurityMiddleware) checkOrigin(r *http.Request) error {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Allow requests without origin header (for direct WebSocket connections)
		return nil
	}
	
	if sm.allowedOrigins[strings.ToLower(origin)] {
		return nil
	}
	
	return ErrInvalidOrigin
}

// getClientIP extracts the real client IP from the request
func (sm *SecurityMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}
	
	// Fall back to remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// checkIPConnectionLimit validates per-IP connection limits
func (sm *SecurityMiddleware) checkIPConnectionLimit(clientIP string) error {
	sm.ipTracker.mutex.Lock()
	defer sm.ipTracker.mutex.Unlock()
	
	currentConnections := sm.ipTracker.connections[clientIP]
	if currentConnections >= maxConnectionsPerIP {
		return ErrTooManyConnections
	}
	
	sm.ipTracker.connections[clientIP] = currentConnections + 1
	return nil
}

// checkGlobalConnectionLimit validates global connection limits
func (sm *SecurityMiddleware) checkGlobalConnectionLimit() error {
	sm.rateLimiter.mutex.Lock()
	defer sm.rateLimiter.mutex.Unlock()
	
	if sm.rateLimiter.totalConnections >= maxConcurrentConnections {
		return ErrServerOverloaded
	}
	
	sm.rateLimiter.totalConnections++
	return nil
}

// initializeRateLimit sets up rate limiting for a new connection
func (sm *SecurityMiddleware) initializeRateLimit(clientID, clientIP string) {
	sm.rateLimiter.mutex.Lock()
	defer sm.rateLimiter.mutex.Unlock()
	
	sm.rateLimiter.connections[clientID] = &ConnectionLimit{
		messages:    make([]time.Time, 0, maxMessagesPerMinute),
		lastMessage: time.Now(),
		violations:  0,
		createdAt:   time.Now(),
		ipAddress:   clientIP,
	}
}

// checkRateLimit validates message rate for a connection
func (sm *SecurityMiddleware) checkRateLimit(clientID string) error {
	sm.rateLimiter.mutex.Lock()
	defer sm.rateLimiter.mutex.Unlock()
	
	connLimit, exists := sm.rateLimiter.connections[clientID]
	if !exists {
		return ErrConnectionNotFound
	}
	
	now := time.Now()
	
	// Remove messages outside the rate limit window
	cutoff := now.Add(-rateLimitWindow)
	validMessages := make([]time.Time, 0, len(connLimit.messages))
	for _, msgTime := range connLimit.messages {
		if msgTime.After(cutoff) {
			validMessages = append(validMessages, msgTime)
		}
	}
	connLimit.messages = validMessages
	
	// Check if rate limit exceeded
	if len(connLimit.messages) >= maxMessagesPerMinute {
		connLimit.violations++
		return ErrRateLimitExceeded
	}
	
	// Add current message
	connLimit.messages = append(connLimit.messages, now)
	connLimit.lastMessage = now
	
	return nil
}

// startCleanupRoutines starts background cleanup processes
func (sm *SecurityMiddleware) startCleanupRoutines() {
	ticker := time.NewTicker(rateLimitCleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.cleanupExpiredConnections()
	}
}

// cleanupExpiredConnections removes stale rate limit data
func (sm *SecurityMiddleware) cleanupExpiredConnections() {
	sm.rateLimiter.mutex.Lock()
	defer sm.rateLimiter.mutex.Unlock()
	
	cutoff := time.Now().Add(-time.Hour) // Remove connections inactive for 1 hour
	expiredConnections := make([]string, 0)
	
	for clientID, connLimit := range sm.rateLimiter.connections {
		if connLimit.lastMessage.Before(cutoff) {
			expiredConnections = append(expiredConnections, clientID)
		}
	}
	
	for _, clientID := range expiredConnections {
		delete(sm.rateLimiter.connections, clientID)
	}
	
	if len(expiredConnections) > 0 {
		log.Printf("Cleaned up %d expired rate limit entries", len(expiredConnections))
	}
}

// logSecurityEvent logs security-related events
func (sm *SecurityMiddleware) logSecurityEvent(eventType, clientID, clientIP, message string, args ...interface{}) {
	logMessage := "SECURITY_EVENT: " + eventType + " - Client: " + clientID + " - IP: " + clientIP + " - " + message
	if len(args) > 0 {
		log.Printf(logMessage, args...)
	} else {
		log.Printf(logMessage)
	}
}

// GetSecurityStats returns current security statistics
func (sm *SecurityMiddleware) GetSecurityStats() SecurityStats {
	sm.rateLimiter.mutex.RLock()
	totalConnections := sm.rateLimiter.totalConnections
	rateLimitedConnections := len(sm.rateLimiter.connections)
	sm.rateLimiter.mutex.RUnlock()
	
	sm.ipTracker.mutex.RLock()
	uniqueIPs := len(sm.ipTracker.connections)
	sm.ipTracker.mutex.RUnlock()
	
	return SecurityStats{
		TotalConnections:       totalConnections,
		RateLimitedConnections: rateLimitedConnections,
		UniqueIPs:             uniqueIPs,
		AllowedOrigins:        len(sm.allowedOrigins),
	}
}

// SecurityStats contains security middleware statistics
type SecurityStats struct {
	TotalConnections       int `json:"total_connections"`
	RateLimitedConnections int `json:"rate_limited_connections"`
	UniqueIPs             int `json:"unique_ips"`
	AllowedOrigins        int `json:"allowed_origins"`
}