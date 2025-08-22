package ws

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gorilla/websocket"
	"worduel-backend/internal/game"
	"worduel-backend/internal/logging"
	"worduel-backend/internal/room"
)

var serverLogger = logging.CreateLogger("ws.server")

// Handler handles WebSocket HTTP requests and upgrades
type Handler struct {
	hub         *Hub
	roomManager *room.RoomManager
	dictionary  *game.Dictionary
	upgrader    websocket.Upgrader
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, roomManager *room.RoomManager, dictionary *game.Dictionary) *Handler {
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
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		serverLogger.Error("WebSocket upgrade failed", "error", err.Error(), "remote_addr", r.RemoteAddr)
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
		serverLogger.Error("Failed to create WebSocket client", "client_id", clientID, "client_ip", clientIP)
		conn.Close()
		return
	}

	// Register client with hub
	h.hub.register <- client

	serverLogger.Info("WebSocket client connected",
		"event_type", "client_connected",
		"client_id", clientID,
		"connection_ip", clientIP)

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