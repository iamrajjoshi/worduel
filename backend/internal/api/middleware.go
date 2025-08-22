package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// API security constants
const (
	// Rate limiting: 120 requests per minute per IP for REST API
	maxRequestsPerMinute = 120
	apiRateLimitWindow   = time.Minute
	
	// Request size limits
	maxRequestSize = 1024 * 1024 // 1MB max request size
	
	// Rate limit cleanup interval
	apiRateLimitCleanupInterval = time.Minute * 5
)

// APIRateLimiter manages per-IP rate limiting for REST API
type APIRateLimiter struct {
	requests map[string]*IPRateLimit
	mutex    sync.RWMutex
}

// IPRateLimit tracks rate limiting data for a single IP
type IPRateLimit struct {
	requests    []time.Time
	lastRequest time.Time
	violations  int
}

// APIMiddleware provides HTTP middleware for security and logging
type APIMiddleware struct {
	rateLimiter    *APIRateLimiter
	allowedOrigins map[string]bool
}

// NewAPIMiddleware creates a new API middleware instance
func NewAPIMiddleware(allowedOrigins []string) *APIMiddleware {
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
	
	am := &APIMiddleware{
		rateLimiter: &APIRateLimiter{
			requests: make(map[string]*IPRateLimit),
		},
		allowedOrigins: originMap,
	}
	
	// Start cleanup routine
	go am.startCleanupRoutine()
	
	return am
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func (am *APIMiddleware) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		if origin != "" && am.allowedOrigins[strings.ToLower(origin)] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers to responses
func (am *APIMiddleware) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		// Server identification
		w.Header().Set("Server", "Worduel-Backend")
		
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware implements rate limiting per IP address
func (am *APIMiddleware) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := am.getClientIP(r)
		
		// Check rate limit
		if err := am.checkAPIRateLimit(clientIP); err != nil {
			am.logAPIEvent("RATE_LIMIT_EXCEEDED", clientIP, r.URL.Path, err.Error())
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(apiRateLimitWindow).Unix(), 10))
			
			http.Error(w, "Rate limit exceeded. Too many requests.", http.StatusTooManyRequests)
			return
		}
		
		// Add rate limit headers
		remaining := am.getRemainingRequests(clientIP)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(apiRateLimitWindow).Unix(), 10))
		
		next.ServeHTTP(w, r)
	})
}

// RequestValidationMiddleware validates request size and format
func (am *APIMiddleware) RequestValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request size
		if r.ContentLength > maxRequestSize {
			am.logAPIEvent("REQUEST_SIZE_EXCEEDED", am.getClientIP(r), r.URL.Path, 
				"Request size: %d bytes", r.ContentLength)
			http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
			return
		}
		
		// Validate content type for POST requests
		if r.Method == "POST" && r.Header.Get("Content-Type") != "" {
			contentType := r.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				am.logAPIEvent("INVALID_CONTENT_TYPE", am.getClientIP(r), r.URL.Path, 
					"Content-Type: %s", contentType)
				http.Error(w, "Invalid content type. Expected application/json", http.StatusUnsupportedMediaType)
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// RequestLoggingMiddleware logs all API requests
func (am *APIMiddleware) RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Log request
		duration := time.Since(start)
		am.logAPIRequest(r.Method, r.URL.Path, wrapped.statusCode, duration, am.getClientIP(r), r.UserAgent())
	})
}

// ErrorHandlingMiddleware handles panics and provides structured error responses
func (am *APIMiddleware) ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				am.logAPIEvent("PANIC_RECOVERED", am.getClientIP(r), r.URL.Path, 
					"Panic: %v", err)
				
				// Don't expose internal details
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal Server Error","code":"INTERNAL_ERROR","message":"An unexpected error occurred"}`))
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// ApplyMiddlewares applies all API middlewares in the correct order
func (am *APIMiddleware) ApplyMiddlewares(handler http.Handler) http.Handler {
	// Apply middlewares in reverse order (outermost first)
	handler = am.ErrorHandlingMiddleware(handler)
	handler = am.RequestLoggingMiddleware(handler)
	handler = am.SecurityHeadersMiddleware(handler)
	handler = am.RequestValidationMiddleware(handler)
	handler = am.RateLimitMiddleware(handler)
	handler = am.CORSMiddleware(handler)
	
	return handler
}

// getClientIP extracts the real client IP from the request
func (am *APIMiddleware) getClientIP(r *http.Request) string {
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
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

// checkAPIRateLimit validates API rate limits for an IP
func (am *APIMiddleware) checkAPIRateLimit(clientIP string) error {
	am.rateLimiter.mutex.Lock()
	defer am.rateLimiter.mutex.Unlock()
	
	now := time.Now()
	
	// Get or create rate limit entry
	ipLimit, exists := am.rateLimiter.requests[clientIP]
	if !exists {
		ipLimit = &IPRateLimit{
			requests:    make([]time.Time, 0, maxRequestsPerMinute),
			lastRequest: now,
			violations:  0,
		}
		am.rateLimiter.requests[clientIP] = ipLimit
	}
	
	// Remove requests outside the rate limit window
	cutoff := now.Add(-apiRateLimitWindow)
	validRequests := make([]time.Time, 0, len(ipLimit.requests))
	for _, reqTime := range ipLimit.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	ipLimit.requests = validRequests
	
	// Check if rate limit exceeded
	if len(ipLimit.requests) >= maxRequestsPerMinute {
		ipLimit.violations++
		return fmt.Errorf("rate limit exceeded: %d requests in last minute", len(ipLimit.requests))
	}
	
	// Add current request
	ipLimit.requests = append(ipLimit.requests, now)
	ipLimit.lastRequest = now
	
	return nil
}

// getRemainingRequests returns the number of remaining requests for an IP
func (am *APIMiddleware) getRemainingRequests(clientIP string) int {
	am.rateLimiter.mutex.RLock()
	defer am.rateLimiter.mutex.RUnlock()
	
	ipLimit, exists := am.rateLimiter.requests[clientIP]
	if !exists {
		return maxRequestsPerMinute
	}
	
	now := time.Now()
	cutoff := now.Add(-apiRateLimitWindow)
	
	// Count valid requests
	validCount := 0
	for _, reqTime := range ipLimit.requests {
		if reqTime.After(cutoff) {
			validCount++
		}
	}
	
	remaining := maxRequestsPerMinute - validCount
	if remaining < 0 {
		remaining = 0
	}
	
	return remaining
}

// startCleanupRoutine starts background cleanup for rate limiting data
func (am *APIMiddleware) startCleanupRoutine() {
	ticker := time.NewTicker(apiRateLimitCleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		am.cleanupExpiredRateLimits()
	}
}

// cleanupExpiredRateLimits removes stale rate limit data
func (am *APIMiddleware) cleanupExpiredRateLimits() {
	am.rateLimiter.mutex.Lock()
	defer am.rateLimiter.mutex.Unlock()
	
	cutoff := time.Now().Add(-time.Hour) // Remove IPs inactive for 1 hour
	expiredIPs := make([]string, 0)
	
	for ip, ipLimit := range am.rateLimiter.requests {
		if ipLimit.lastRequest.Before(cutoff) {
			expiredIPs = append(expiredIPs, ip)
		}
	}
	
	for _, ip := range expiredIPs {
		delete(am.rateLimiter.requests, ip)
	}
	
	if len(expiredIPs) > 0 {
		log.Printf("API: Cleaned up %d expired rate limit entries", len(expiredIPs))
	}
}

// logAPIRequest logs API requests with structured format
func (am *APIMiddleware) logAPIRequest(method, path string, statusCode int, duration time.Duration, clientIP, userAgent string) {
	log.Printf("API_REQUEST: %s %s - Status: %d - Duration: %dms - IP: %s - UA: %s", 
		method, path, statusCode, duration.Milliseconds(), clientIP, userAgent)
}

// logAPIEvent logs API security and error events
func (am *APIMiddleware) logAPIEvent(eventType, clientIP, path, message string, args ...interface{}) {
	logMessage := fmt.Sprintf("API_EVENT: %s - IP: %s - Path: %s - %s", 
		eventType, clientIP, path, message)
	if len(args) > 0 {
		log.Printf(logMessage, args...)
	} else {
		log.Printf(logMessage)
	}
}

// GetAPIStats returns current API middleware statistics
func (am *APIMiddleware) GetAPIStats() APIStats {
	am.rateLimiter.mutex.RLock()
	defer am.rateLimiter.mutex.RUnlock()
	
	totalViolations := 0
	for _, ipLimit := range am.rateLimiter.requests {
		totalViolations += ipLimit.violations
	}
	
	return APIStats{
		TrackedIPs:      len(am.rateLimiter.requests),
		TotalViolations: totalViolations,
		AllowedOrigins:  len(am.allowedOrigins),
	}
}

// APIStats contains API middleware statistics
type APIStats struct {
	TrackedIPs      int `json:"tracked_ips"`
	TotalViolations int `json:"total_violations"`
	AllowedOrigins  int `json:"allowed_origins"`
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestContext key for storing request metadata
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	ClientIPKey  contextKey = "client_ip"
)

// AddRequestContext middleware adds request context information
func (am *APIMiddleware) AddRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Add client IP to context
		ctx = context.WithValue(ctx, ClientIPKey, am.getClientIP(r))
		
		// Add request ID for tracing (simple implementation)
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}