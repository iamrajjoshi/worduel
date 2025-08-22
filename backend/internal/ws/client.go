package ws

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"worduel-backend/internal/game"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512

	// Time to wait for connection to close gracefully
	closeGracePeriod = 10 * time.Second
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Origin checking is now handled by SecurityMiddleware
		// This will be overridden in ServeWS
		return false
	},
}

// Client represents a WebSocket client connection
type Client struct {
	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Client identifier
	id string

	// Room this client is associated with
	roomID string

	// Player information
	playerID string

	// Hub this client is registered with
	hub *Hub

	// Connection metadata
	connectedAt time.Time
	lastPong    time.Time
	clientIP    string

	// Mutex for protecting concurrent access
	mutex sync.RWMutex

	// Connection state
	closed bool
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub, clientID string, clientIP string) *Client {
	return &Client{
		conn:        conn,
		send:        make(chan []byte, 256),
		id:          clientID,
		hub:         hub,
		connectedAt: time.Now(),
		lastPong:    time.Now(),
		clientIP:    clientIP,
		closed:      false,
	}
}

// GetID returns the client identifier
func (c *Client) GetID() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.id
}

// GetRoomID returns the room ID this client is associated with
func (c *Client) GetRoomID() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.roomID
}

// SetRoom associates the client with a room
func (c *Client) SetRoom(roomID string, playerID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.roomID = roomID
	c.playerID = playerID
}

// GetPlayerID returns the player ID associated with this client
func (c *Client) GetPlayerID() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.playerID
}

// GetClientIP returns the client's IP address
func (c *Client) GetClientIP() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.clientIP
}

// IsClosed returns whether the connection is closed
func (c *Client) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.closed
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message []byte) error {
	c.mutex.RLock()
	closed := c.closed
	c.mutex.RUnlock()

	if closed {
		return websocket.ErrCloseSent
	}

	select {
	case c.send <- message:
		return nil
	default:
		// Channel is full, client is slow
		c.Close()
		return websocket.ErrCloseSent
	}
}

// SendJSON sends a JSON message to the client
func (c *Client) SendJSON(msg *game.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.SendMessage(data)
}

// Close closes the client connection
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	close(c.send)

	// Set close deadline
	c.conn.SetWriteDeadline(time.Now().Add(closeGracePeriod))
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})

	// Close the connection
	c.conn.Close()
}

// readPump pumps messages from the WebSocket connection to the hub
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.mutex.Lock()
		c.lastPong = time.Now()
		c.mutex.Unlock()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for client %s: %v", c.id, err)
			}
			break
		}

		// Check rate limiting if hub has security middleware
		if c.hub.securityMiddleware != nil {
			if err := c.hub.securityMiddleware.CheckMessageRate(c.id, len(message)); err != nil {
				if err == ErrRateLimitExceeded {
					c.SendJSON(&game.Message{
						Type:      game.MessageTypeError,
						PlayerID:  c.playerID,
						Timestamp: time.Now(),
						Data: &game.ErrorData{
							Code:    "RATE_LIMIT_EXCEEDED",
							Message: "Rate limit exceeded. Please slow down.",
						},
					})
					continue
				} else if err == ErrMessageTooLarge {
					c.SendJSON(&game.Message{
						Type:      game.MessageTypeError,
						PlayerID:  c.playerID,
						Timestamp: time.Now(),
						Data: &game.ErrorData{
							Code:    "MESSAGE_TOO_LARGE",
							Message: "Message too large. Maximum size is 512 bytes.",
						},
					})
					continue
				}
				// For other errors, log and break
				log.Printf("Security check failed for client %s: %v", c.id, err)
				break
			}
		}

		// Parse and validate message
		var msg game.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format from client %s: %v", c.id, err)
			c.SendJSON(&game.Message{
				Type:      game.MessageTypeError,
				PlayerID:  c.playerID,
				Timestamp: time.Now(),
				Data: &game.ErrorData{
					Code:    "INVALID_MESSAGE",
					Message: "Invalid message format",
				},
			})
			continue
		}

		// Set client information in message
		msg.PlayerID = c.playerID
		msg.Timestamp = time.Now()

		// Send message to hub for processing
		select {
		case c.hub.broadcast <- &ClientMessage{
			client:  c,
			message: &msg,
		}:
		default:
			log.Printf("Hub broadcast channel full, dropping message from client %s", c.id)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Run starts the client's read and write pumps
func (c *Client) Run() {
	// Start read and write pumps in separate goroutines
	go c.readPump()
	go c.writePump()
}

// ClientMessage represents a message from a client with metadata
type ClientMessage struct {
	client  *Client
	message *game.Message
}

// GetClient returns the client that sent the message
func (cm *ClientMessage) GetClient() *Client {
	return cm.client
}

// GetMessage returns the message content
func (cm *ClientMessage) GetMessage() *game.Message {
	return cm.message
}

// ServeWS handles WebSocket requests from clients
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, clientID string) {
	// Perform security validation if middleware is available
	var clientIP string
	if hub.securityMiddleware != nil {
		if err := hub.securityMiddleware.ValidateConnection(r, clientID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		
		// Override the upgrader's CheckOrigin to use our middleware
		upgrader.CheckOrigin = func(req *http.Request) bool {
			return hub.securityMiddleware.checkOrigin(req) == nil
		}
		
		// Extract client IP for later use
		clientIP = hub.securityMiddleware.getClientIP(r)
	} else {
		// Fallback for when security middleware is not configured
		upgrader.CheckOrigin = func(req *http.Request) bool {
			return true // Allow all origins if no security middleware
		}
		
		// Basic IP extraction
		clientIP = r.RemoteAddr
		if ip, _, err := net.SplitHostPort(clientIP); err == nil {
			clientIP = ip
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed for client %s: %v", clientID, err)
		if hub.securityMiddleware != nil {
			hub.securityMiddleware.OnConnectionClosed(clientID, clientIP)
		}
		return
	}

	client := NewClient(conn, hub, clientID, clientIP)
	
	// Register client with hub
	hub.register <- client

	// Start client pumps
	client.Run()
}