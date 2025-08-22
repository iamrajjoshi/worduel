package ws

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
)

// Handler handles WebSocket HTTP requests and upgrades
type Handler struct {
	hub         *Hub
	roomManager *room.RoomManager
	dictionary  *game.Dictionary
	upgrader    websocket.Upgrader
	logger      *log.Logger
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
		logger: log.New(log.Writer(), "[WS] ", log.LstdFlags|log.Lshortfile),
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Printf("WebSocket upgrade failed: %v", err)
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
		h.logger.Println("Failed to create WebSocket client")
		conn.Close()
		return
	}

	// Register client with hub
	h.hub.register <- client

	h.logger.Printf("Client %s connected from %s", client.GetID(), clientIP)

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