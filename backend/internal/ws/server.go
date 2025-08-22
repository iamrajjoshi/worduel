package ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gorilla/websocket"
	"worduel-backend/internal/game"
	"worduel-backend/internal/logging"
	"worduel-backend/internal/room"
)

// Handler handles WebSocket HTTP requests and upgrades
type Handler struct {
	hub         *Hub
	roomManager *room.RoomManager
	dictionary  *game.Dictionary
	upgrader    websocket.Upgrader
	logger      *logging.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, roomManager *room.RoomManager, dictionary *game.Dictionary, logger *logging.Logger) *Handler {
	return &Handler{
		hub:         hub,
		roomManager: roomManager,
		dictionary:  dictionary,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - this should be configurable
				return true
			},
		},
		logger: logger,
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctx := context.Background()
		h.logger.LogError(ctx, err, "WebSocket upgrade failed", "remote_addr", r.RemoteAddr)
		return
	}

	// Generate client ID and extract IP
	clientID := generateClientID()
	clientIP := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}

	// Create new client
	client := NewClient(conn, h.hub, clientID, clientIP)
	if client == nil {
		ctx := logging.WithCorrelationID(context.Background(), clientID)
		h.logger.LogError(ctx, nil, "Failed to create WebSocket client", "client_ip", clientIP)
		conn.Close()
		return
	}

	// Register client with hub
	h.hub.register <- client

	ctx := logging.WithCorrelationID(context.Background(), clientID)
	h.logger.LogWebSocketEvent(ctx, logging.WSEventFields{
		EventType:    "client_connected",
		ClientID:     clientID,
		RoomID:       "",
		MessageType:  "",
		ConnectionIP: clientIP,
	})

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// generateClientID generates a unique client identifier
func generateClientID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}