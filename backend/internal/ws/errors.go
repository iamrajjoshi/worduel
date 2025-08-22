package ws

import "errors"

// WebSocket specific errors
var (
	ErrClientNotFound      = errors.New("client not found")
	ErrRoomFull           = errors.New("room is full")
	ErrInvalidMessage     = errors.New("invalid message format")
	ErrClientNotInRoom    = errors.New("client not in room")
	ErrUnauthorized       = errors.New("unauthorized operation")
	ErrConnectionClosed   = errors.New("connection is closed")
	ErrMessageQueueFull   = errors.New("message queue is full")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	
	// Security-related errors
	ErrInvalidOrigin       = errors.New("invalid origin")
	ErrMessageTooLarge     = errors.New("message too large")
	ErrTooManyConnections  = errors.New("too many connections from IP")
	ErrServerOverloaded    = errors.New("server overloaded")
	ErrConnectionNotFound  = errors.New("connection not found in rate limiter")
)