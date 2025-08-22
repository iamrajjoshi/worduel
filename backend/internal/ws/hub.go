package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients by client ID
	clients map[string]*Client

	// Room associations - maps room ID to client IDs
	roomClients map[string]map[string]*Client

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Inbound messages from the clients
	broadcast chan *ClientMessage

	// Room manager for game operations
	roomManager *room.RoomManager

	// Message handler for processing WebSocket messages
	messageHandler *MessageHandler

	// Security middleware for rate limiting and validation
	securityMiddleware *SecurityMiddleware

	// Mutex for protecting concurrent access
	mutex sync.RWMutex

	// Hub statistics
	stats HubStats
}

// HubStats contains statistics about the hub
type HubStats struct {
	ConnectedClients int       `json:"connected_clients"`
	ActiveRooms      int       `json:"active_rooms"`
	MessagesPerSec   float64   `json:"messages_per_sec"`
	LastUpdate       time.Time `json:"last_update"`
}

// NewHub creates a new WebSocket hub
func NewHub(roomManager *room.RoomManager, gameLogic *game.GameLogic) *Hub {
	hub := &Hub{
		clients:     make(map[string]*Client),
		roomClients: make(map[string]map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *ClientMessage),
		roomManager: roomManager,
		stats: HubStats{
			LastUpdate: time.Now(),
		},
	}

	// Initialize message handler with hub reference
	hub.messageHandler = NewMessageHandler(hub, roomManager, gameLogic)

	return hub
}

// SetSecurityMiddleware configures security middleware for the hub
func (h *Hub) SetSecurityMiddleware(middleware *SecurityMiddleware) {
	h.securityMiddleware = middleware
}

// Run starts the hub and handles client registration/unregistration and message broadcasting
func (h *Hub) Run() {
	// Start statistics updater
	go h.updateStats()

	for {
		select {
		case client := <-h.register:
			h.handleClientRegister(client)

		case client := <-h.unregister:
			h.handleClientUnregister(client)

		case clientMessage := <-h.broadcast:
			h.messageHandler.HandleMessage(clientMessage)
		}
	}
}

// handleClientRegister handles client registration
func (h *Hub) handleClientRegister(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client.GetID()] = client
	log.Printf("Client %s connected. Total clients: %d", client.GetID(), len(h.clients))

	// Send connection acknowledgment
	response := &game.Message{
		Type:      "connection_ack",
		PlayerID:  client.GetPlayerID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"client_id":    client.GetID(),
			"connected_at": client.connectedAt,
		},
	}
	client.SendJSON(response)
}

// handleClientUnregister handles client unregistration
func (h *Hub) handleClientUnregister(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clientID := client.GetID()
	roomID := client.GetRoomID()

	// Remove from clients map
	if _, exists := h.clients[clientID]; exists {
		delete(h.clients, clientID)
		log.Printf("Client %s disconnected. Total clients: %d", clientID, len(h.clients))
	}

	// Remove from room if associated
	if roomID != "" {
		h.removeClientFromRoom(client, roomID)

		// Handle player disconnection in room
		if client.GetPlayerID() != "" {
			h.handlePlayerDisconnection(client)
		}
	}

	// Notify security middleware about disconnection
	if h.securityMiddleware != nil {
		h.securityMiddleware.OnConnectionClosed(clientID, client.GetClientIP())
	}

	// Close the client connection
	if !client.IsClosed() {
		client.Close()
	}
}

// removeClientFromRoom removes a client from a room's client list
func (h *Hub) removeClientFromRoom(client *Client, roomID string) {
	if roomClients, exists := h.roomClients[roomID]; exists {
		delete(roomClients, client.GetID())
		
		// If room is empty, remove it from roomClients
		if len(roomClients) == 0 {
			delete(h.roomClients, roomID)
		}
	}
}

// handlePlayerDisconnection handles player disconnection in a room
func (h *Hub) handlePlayerDisconnection(client *Client) {
	roomID := client.GetRoomID()
	playerID := client.GetPlayerID()

	// Update player status in room
	if room, err := h.roomManager.GetRoom(roomID); err == nil {
		room.Lock()
		if player, exists := room.Players[playerID]; exists {
			player.Status = game.PlayerStatusDisconnected
			player.LastActivity = time.Now()
		}
		room.Unlock()

		// Notify other players in the room
		disconnectMessage := &game.Message{
			Type:     game.MessageTypePlayerUpdate,
			PlayerID: playerID,
			RoomID:   roomID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"player_id": playerID,
				"status":    string(game.PlayerStatusDisconnected),
				"event":     "player_disconnected",
			},
		}
		h.broadcastToRoom(roomID, disconnectMessage, client.GetID())
	}
}


// broadcastToRoom sends a message to all clients in a specific room
func (h *Hub) broadcastToRoom(roomID string, message *game.Message, excludeClientID string) {
	h.mutex.RLock()
	roomClients := h.roomClients[roomID]
	h.mutex.RUnlock()

	if roomClients == nil {
		return
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for clientID, client := range roomClients {
		if clientID != excludeClientID && !client.IsClosed() {
			if err := client.SendMessage(messageData); err != nil {
				log.Printf("Error sending message to client %s: %v", clientID, err)
				// Client will be cleaned up by the unregister process
			}
		}
	}
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(clientID string, message *game.Message) error {
	h.mutex.RLock()
	client, exists := h.clients[clientID]
	h.mutex.RUnlock()

	if !exists {
		return ErrClientNotFound
	}

	return client.SendJSON(message)
}


// GetStats returns current hub statistics
func (h *Hub) GetStats() HubStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	stats := h.stats
	stats.ConnectedClients = len(h.clients)
	stats.ActiveRooms = len(h.roomClients)
	
	return stats
}

// updateStats periodically updates hub statistics
func (h *Hub) updateStats() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {
		h.mutex.Lock()
		h.stats.ConnectedClients = len(h.clients)
		h.stats.ActiveRooms = len(h.roomClients)
		h.stats.LastUpdate = time.Now()
		h.mutex.Unlock()
	}
}

// CleanupExpiredConnections removes expired and inactive connections
func (h *Hub) CleanupExpiredConnections() {
	h.mutex.RLock()
	clientsToRemove := make([]*Client, 0)
	
	for _, client := range h.clients {
		client.mutex.RLock()
		lastPong := client.lastPong
		client.mutex.RUnlock()
		
		if time.Since(lastPong) > pongWait*2 {
			clientsToRemove = append(clientsToRemove, client)
		}
	}
	h.mutex.RUnlock()

	// Remove expired clients
	for _, client := range clientsToRemove {
		log.Printf("Removing expired client: %s", client.GetID())
		h.unregister <- client
	}
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	h.mutex.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	// Close all client connections
	for _, client := range clients {
		client.Close()
	}

	log.Printf("Hub shutdown complete. Closed %d connections", len(clients))
}