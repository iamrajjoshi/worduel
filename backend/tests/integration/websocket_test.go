package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
	"worduel-backend/internal/ws"
)

// TestClient represents a test WebSocket client
type TestClient struct {
	ID           string
	conn         *websocket.Conn
	messages     chan *game.Message
	errors       chan error
	closed       bool
	mu           sync.RWMutex
	t            *testing.T
	messageCount int
}

// TestServer represents a test server setup
type TestServer struct {
	server      *httptest.Server
	hub         *ws.Hub
	roomManager *room.RoomManager
	gameLogic   *game.GameLogic
	stopCh      chan struct{}
}

func setupTestServer(t *testing.T) *TestServer {
	// Create dictionary, room manager and game logic
	dictionary := game.NewDictionary()
	roomManager := room.NewRoomManager()
	gameLogic := game.NewGameLogic(dictionary)
	
	// Create hub
	hub := ws.NewHub(roomManager, gameLogic)
	stopCh := make(chan struct{})
	go func() {
		hub.Run()
		close(stopCh)
	}()
	
	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			clientID = fmt.Sprintf("test-client-%d", time.Now().UnixNano())
		}
		ws.ServeWS(hub, w, r, clientID)
	})
	
	server := httptest.NewServer(mux)
	
	return &TestServer{
		server:      server,
		hub:         hub,
		roomManager: roomManager,
		gameLogic:   gameLogic,
		stopCh:      stopCh,
	}
}

func (ts *TestServer) Close() {
	// Close the HTTP server first to stop new connections
	ts.server.Close()
	
	// Give some time for connections to close naturally
	time.Sleep(50 * time.Millisecond)
	
	// Shutdown the hub gracefully (non-blocking)
	go func() {
		if ts.hub != nil {
			ts.hub.Shutdown()
		}
	}()
	
	// Wait for hub to finish with a short timeout
	select {
	case <-ts.stopCh:
		// Hub finished gracefully
	case <-time.After(100 * time.Millisecond):
		// Timeout waiting for hub to finish - that's okay for tests
	}
}

func (ts *TestServer) ConnectClient(t *testing.T, clientID string) *TestClient {
	if clientID == "" {
		clientID = fmt.Sprintf("test-client-%d", time.Now().UnixNano())
	}
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(ts.server.URL, "http") + "/ws?client_id=" + clientID
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "Failed to connect WebSocket client")
	
	client := &TestClient{
		ID:       clientID,
		conn:     conn,
		messages: make(chan *game.Message, 100),
		errors:   make(chan error, 10),
		t:        t,
	}
	
	// Start reading messages
	go client.readMessages()
	
	return client
}

func (tc *TestClient) readMessages() {
	defer func() {
		tc.mu.Lock()
		tc.closed = true
		tc.mu.Unlock()
		close(tc.messages)
		close(tc.errors)
	}()
	
	for {
		tc.mu.RLock()
		closed := tc.closed
		tc.mu.RUnlock()
		
		if closed {
			return
		}
		
		var msg game.Message
		err := tc.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			select {
			case tc.errors <- err:
			default:
			}
			return
		}
		
		tc.mu.Lock()
		tc.messageCount++
		tc.mu.Unlock()
		
		select {
		case tc.messages <- &msg:
		default:
			tc.t.Logf("Message channel full, dropping message")
		}
	}
}

func (tc *TestClient) SendMessage(msgType game.MessageType, data interface{}) error {
	msg := game.Message{
		Type:      msgType,
		PlayerID:  tc.ID,
		Data:      data,
		Timestamp: time.Now(),
	}
	
	return tc.conn.WriteJSON(msg)
}

func (tc *TestClient) SendJoin(roomID, playerName string) error {
	return tc.SendMessage(game.MessageTypeJoin, map[string]interface{}{
		"room_id":     roomID,
		"player_name": playerName,
	})
}

func (tc *TestClient) SendGuess(word string) error {
	return tc.SendMessage(game.MessageTypeGuess, map[string]interface{}{
		"word": word,
	})
}

func (tc *TestClient) SendLeave() error {
	return tc.SendMessage(game.MessageTypeLeave, nil)
}

func (tc *TestClient) WaitForMessage(timeout time.Duration) (*game.Message, error) {
	select {
	case msg := <-tc.messages:
		if msg == nil {
			return nil, fmt.Errorf("received nil message")
		}
		return msg, nil
	case err := <-tc.errors:
		return nil, err
	case <-time.After(timeout):
		tc.mu.RLock()
		closed := tc.closed
		tc.mu.RUnlock()
		if closed {
			return nil, fmt.Errorf("client connection closed")
		}
		return nil, fmt.Errorf("timeout waiting for message after %v", timeout)
	}
}

func (tc *TestClient) WaitForMessageType(msgType string, timeout time.Duration) (*game.Message, error) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		msg, err := tc.WaitForMessage(time.Until(deadline))
		if err != nil {
			return nil, err
		}
		
		if string(msg.Type) == msgType {
			return msg, nil
		}
	}
	
	return nil, fmt.Errorf("timeout waiting for message type %s", msgType)
}

func (tc *TestClient) GetMessageCount() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.messageCount
}

func (tc *TestClient) Close() {
	tc.mu.Lock()
	if tc.closed {
		tc.mu.Unlock()
		return
	}
	tc.closed = true
	tc.mu.Unlock()
	
	// Close the WebSocket connection gracefully
	tc.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(50 * time.Millisecond) // Give time for close message to be sent
	tc.conn.Close()
}

// Test connection establishment and acknowledgment
func TestWebSocketConnection(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	client := server.ConnectClient(t, "test-client-1")
	defer client.Close()
	
	// Wait for connection acknowledgment
	msg, err := client.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	assert.Equal(t, "connection_ack", string(msg.Type))
	// PlayerID is empty until client joins a room
	assert.Equal(t, "", msg.PlayerID)
	
	data, ok := msg.Data.(map[string]interface{})
	require.True(t, ok, "Connection ack data should be a map")
	assert.Equal(t, "test-client-1", data["client_id"])
	assert.NotEmpty(t, data["connected_at"])
}

// Test room joining flow
func TestRoomJoining(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	// Create a room first
	room, err := server.roomManager.CreateRoom("Test Room", 2)
	require.NoError(t, err)
	
	client := server.ConnectClient(t, "player1")
	defer client.Close()
	
	// Wait for connection ack
	_, err = client.WaitForMessageType("connection_ack", 1*time.Second)
	require.NoError(t, err)
	
	// Join room
	err = client.SendJoin(room.ID, "Player One")
	require.NoError(t, err)
	
	// Wait for join success
	msg, err := client.WaitForMessageType("join_success", 1*time.Second)
	require.NoError(t, err)
	
	assert.Equal(t, room.ID, msg.RoomID)
	assert.Equal(t, "player1", msg.PlayerID)
	
	// Basic validation that we got some data back
	assert.NotNil(t, msg.Data, "Should have data in join success message")
}

// Test multi-client room joining and game start
func TestMultiClientGameStart(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	// Create a room
	room, err := server.roomManager.CreateRoom("Test Room", 2)
	require.NoError(t, err)
	
	// Connect two clients
	client1 := server.ConnectClient(t, "player1")
	defer client1.Close()
	client2 := server.ConnectClient(t, "player2")
	defer client2.Close()
	
	// Wait for connection acks
	_, err = client1.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	// First player joins
	err = client1.SendJoin(room.ID, "Player One")
	require.NoError(t, err)
	
	// Wait for join success
	_, err = client1.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	
	// Second player joins
	err = client2.SendJoin(room.ID, "Player Two")
	require.NoError(t, err)
	
	// Wait for second join success
	_, err = client2.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	
	// Both clients should receive player joined notification
	_, err = client1.WaitForMessageType("player_update", 2*time.Second)
	require.NoError(t, err)
	
	// Both clients should receive game start message
	gameStart1, err := client1.WaitForMessageType("game_started", 3*time.Second)
	require.NoError(t, err)
	gameStart2, err := client2.WaitForMessageType("game_started", 3*time.Second)
	require.NoError(t, err)
	
	assert.Equal(t, room.ID, gameStart1.RoomID)
	assert.Equal(t, room.ID, gameStart2.RoomID)
	
	// Verify game start data
	data1, ok := gameStart1.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 5, int(data1["target_word_length"].(float64)))
	assert.Equal(t, 6, int(data1["max_guesses"].(float64)))
	assert.Equal(t, string(game.GameStatusActive), data1["game_status"])
}

// Test guess processing and broadcasting
func TestGuessProcessingAndBroadcasting(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	// Create room and join two players
	room, err := server.roomManager.CreateRoom("Test Room", 2)
	require.NoError(t, err)
	
	client1 := server.ConnectClient(t, "player1")
	defer client1.Close()
	client2 := server.ConnectClient(t, "player2")
	defer client2.Close()
	
	// Wait for connections and join room
	_, err = client1.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	err = client1.SendJoin(room.ID, "Player One")
	require.NoError(t, err)
	err = client2.SendJoin(room.ID, "Player Two")
	require.NoError(t, err)
	
	// Wait for game to start
	_, err = client1.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	_, err = client1.WaitForMessageType("player_update", 2*time.Second)
	require.NoError(t, err)
	_, err = client1.WaitForMessageType("game_started", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("game_started", 2*time.Second)
	require.NoError(t, err)
	
	// Player 1 makes a guess
	err = client1.SendGuess("apple")
	require.NoError(t, err)
	
	// Player 1 should receive guess result
	guessResult, err := client1.WaitForMessageType("guess_result", 2*time.Second)
	require.NoError(t, err)
	
	assert.Equal(t, "player1", guessResult.PlayerID)
	assert.Equal(t, room.ID, guessResult.RoomID)
	
	resultData, ok := guessResult.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "apple", resultData["word"])
	assert.NotNil(t, resultData["results"])
	assert.NotNil(t, resultData["is_correct"])
	
	// Both clients should receive game update
	gameUpdate1, err := client1.WaitForMessageType("game_update", 2*time.Second)
	require.NoError(t, err)
	gameUpdate2, err := client2.WaitForMessageType("game_update", 2*time.Second)
	require.NoError(t, err)
	
	assert.Equal(t, room.ID, gameUpdate1.RoomID)
	assert.Equal(t, room.ID, gameUpdate2.RoomID)
}

// Test invalid message handling
func TestInvalidMessageHandling(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	client := server.ConnectClient(t, "test-client")
	defer client.Close()
	
	// Wait for connection ack
	_, err := client.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	// Send invalid message format
	err = client.conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	require.NoError(t, err)
	
	// Should receive error message
	errorMsg, err := client.WaitForMessageType("error", 2*time.Second)
	require.NoError(t, err)
	
	errorData, ok := errorMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "INVALID_MESSAGE", errorData["code"])
	assert.Contains(t, errorData["message"], "Invalid message format")
}

// Test connection cleanup and disconnection handling
func TestConnectionCleanupAndDisconnection(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	// Create room and connect two clients
	room, err := server.roomManager.CreateRoom("Test Room", 2)
	require.NoError(t, err)
	
	client1 := server.ConnectClient(t, "player1")
	client2 := server.ConnectClient(t, "player2")
	defer client2.Close()
	
	// Join room with both clients
	_, err = client1.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	err = client1.SendJoin(room.ID, "Player One")
	require.NoError(t, err)
	err = client2.SendJoin(room.ID, "Player Two")
	require.NoError(t, err)
	
	// Wait for successful joins and game start
	_, err = client1.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	_, err = client1.WaitForMessageType("player_update", 2*time.Second)
	require.NoError(t, err)
	_, err = client1.WaitForMessageType("game_started", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("game_started", 2*time.Second)
	require.NoError(t, err)
	
	// Close client1 connection abruptly
	client1.Close()
	
	// Client2 should receive disconnection notification
	disconnectMsg, err := client2.WaitForMessageType("player_update", 3*time.Second)
	require.NoError(t, err)
	
	disconnectData, ok := disconnectMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "player_disconnected", disconnectData["event"])
	assert.Equal(t, "player1", disconnectData["player_id"])
	assert.Equal(t, string(game.PlayerStatusDisconnected), disconnectData["status"])
}

// Test rate limiting and security
func TestRateLimitingAndSecurity(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	// Set up security middleware with allowed origins
	allowedOrigins := []string{"http://localhost:3000"}
	securityMiddleware := ws.NewSecurityMiddleware(allowedOrigins)
	server.hub.SetSecurityMiddleware(securityMiddleware)
	
	client := server.ConnectClient(t, "test-client")
	defer client.Close()
	
	// Wait for connection ack
	_, err := client.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	// Send messages rapidly to trigger rate limit
	for i := 0; i < 10; i++ {
		err = client.SendMessage("unknown", map[string]interface{}{
			"test": fmt.Sprintf("message%d", i),
		})
		require.NoError(t, err)
	}
	
	// Should receive rate limit error
	var rateLimitError *game.Message
	for i := 0; i < 10; i++ {
		msg, err := client.WaitForMessage(1 * time.Second)
		if err != nil {
			break
		}
		if string(msg.Type) == "error" {
			if data, ok := msg.Data.(*game.ErrorData); ok && data.Code == "RATE_LIMIT_EXCEEDED" {
				rateLimitError = msg
				break
			}
		}
	}
	
	require.NotNil(t, rateLimitError, "Should receive rate limit error")
	errorData, ok := rateLimitError.Data.(*game.ErrorData)
	require.True(t, ok)
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorData.Code)
}

// Test message throughput and latency performance
func TestMessageThroughputAndLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	server := setupTestServer(t)
	defer server.Close()
	
	// Create room and join two players
	room, err := server.roomManager.CreateRoom("Performance Test Room", 2)
	require.NoError(t, err)
	
	client1 := server.ConnectClient(t, "perf-player1")
	defer client1.Close()
	client2 := server.ConnectClient(t, "perf-player2")
	defer client2.Close()
	
	// Setup clients and start game
	_, err = client1.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	err = client1.SendJoin(room.ID, "Perf Player 1")
	require.NoError(t, err)
	err = client2.SendJoin(room.ID, "Perf Player 2")
	require.NoError(t, err)
	
	// Wait for game start
	_, err = client1.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err)
	_, err = client1.WaitForMessageType("player_update", 2*time.Second)
	require.NoError(t, err)
	_, err = client1.WaitForMessageType("game_started", 2*time.Second)
	require.NoError(t, err)
	_, err = client2.WaitForMessageType("game_started", 2*time.Second)
	require.NoError(t, err)
	
	// Performance test: measure latency
	numMessages := 10
	var totalLatency time.Duration
	
	for i := 0; i < numMessages; i++ {
		start := time.Now()
		
		err = client1.SendGuess("hello")
		require.NoError(t, err)
		
		// Wait for guess result
		_, err = client1.WaitForMessageType("guess_result", 2*time.Second)
		require.NoError(t, err)
		
		latency := time.Since(start)
		totalLatency += latency
		
		t.Logf("Message %d latency: %v", i+1, latency)
	}
	
	avgLatency := totalLatency / time.Duration(numMessages)
	t.Logf("Average latency: %v", avgLatency)
	
	// Assert performance requirements (< 100ms as per requirements)
	assert.Less(t, avgLatency, 100*time.Millisecond, "Average latency should be under 100ms")
	
	// Check throughput
	start := time.Now()
	initialCount1 := client1.GetMessageCount()
	initialCount2 := client2.GetMessageCount()
	
	// Send rapid messages for 1 second
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.After(1 * time.Second)
		
		for {
			select {
			case <-ticker.C:
				client1.SendGuess("rapid")
			case <-timeout:
				return
			}
		}
	}()
	
	// Wait for test duration
	time.Sleep(1200 * time.Millisecond)
	
	duration := time.Since(start)
	finalCount1 := client1.GetMessageCount()
	finalCount2 := client2.GetMessageCount()
	
	totalMessagesReceived := (finalCount1 - initialCount1) + (finalCount2 - initialCount2)
	messagesPerSecond := float64(totalMessagesReceived) / duration.Seconds()
	
	t.Logf("Messages per second: %.2f", messagesPerSecond)
	t.Logf("Total messages received: %d", totalMessagesReceived)
	
	// Should handle reasonable throughput
	assert.Greater(t, messagesPerSecond, 50.0, "Should handle at least 50 messages per second")
}

// Test concurrent multi-client scenarios
func TestConcurrentMultiClientScenarios(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	// Create multiple rooms
	numRooms := 3
	numClientsPerRoom := 2
	
	rooms := make([]*game.Room, numRooms)
	clients := make([][]*TestClient, numRooms)
	
	// Create rooms and clients
	for i := 0; i < numRooms; i++ {
		room, err := server.roomManager.CreateRoom(fmt.Sprintf("Test Room %d", i+1), numClientsPerRoom)
		require.NoError(t, err)
		rooms[i] = room
		
		clients[i] = make([]*TestClient, numClientsPerRoom)
		for j := 0; j < numClientsPerRoom; j++ {
			clientID := fmt.Sprintf("room%d-player%d", i+1, j+1)
			client := server.ConnectClient(t, clientID)
			clients[i][j] = client
			defer client.Close()
		}
	}
	
	// Connect all clients to their respective rooms concurrently
	var wg sync.WaitGroup
	
	for i := 0; i < numRooms; i++ {
		for j := 0; j < numClientsPerRoom; j++ {
			wg.Add(1)
			go func(roomIdx, clientIdx int) {
				defer wg.Done()
				
				client := clients[roomIdx][clientIdx]
				room := rooms[roomIdx]
				
				// Wait for connection ack
				_, err := client.WaitForMessageType("connection_ack", 5*time.Second)
				require.NoError(t, err)
				
				// Join room
				playerName := fmt.Sprintf("Player %d-%d", roomIdx+1, clientIdx+1)
				err = client.SendJoin(room.ID, playerName)
				require.NoError(t, err)
				
				// Wait for join success
				_, err = client.WaitForMessageType("join_success", 5*time.Second)
				require.NoError(t, err)
				
				t.Logf("Client %s successfully joined room %s", client.ID, room.ID)
			}(i, j)
		}
	}
	
	wg.Wait()
	t.Logf("All clients successfully connected to their rooms")
	
	// Verify all games started (each room should have 2 players)
	for i := 0; i < numRooms; i++ {
		for j := 0; j < numClientsPerRoom; j++ {
			_, err := clients[i][j].WaitForMessageType("game_started", 3*time.Second)
			require.NoError(t, err, "Game should start for room %d", i+1)
		}
	}
	
	// Test concurrent guess processing across all rooms
	for i := 0; i < numRooms; i++ {
		wg.Add(1)
		go func(roomIdx int) {
			defer wg.Done()
			
			client := clients[roomIdx][0] // First client in room
			
			// Make a guess
			err := client.SendGuess("tests")
			require.NoError(t, err)
			
			// Should receive guess result
			_, err = client.WaitForMessageType("guess_result", 3*time.Second)
			require.NoError(t, err)
			
			t.Logf("Guess processed for room %d", roomIdx+1)
		}(i)
	}
	
	wg.Wait()
	t.Logf("All concurrent guesses processed successfully")
}

// Test error recovery scenarios
func TestErrorRecoveryScenarios(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	
	client := server.ConnectClient(t, "recovery-test")
	defer client.Close()
	
	// Wait for connection
	_, err := client.WaitForMessageType("connection_ack", 2*time.Second)
	require.NoError(t, err)
	
	// Test joining non-existent room
	err = client.SendJoin("nonexistent-room", "Test Player")
	require.NoError(t, err)
	
	errorMsg, err := client.WaitForMessageType("error", 2*time.Second)
	require.NoError(t, err)
	
	errorData, ok := errorMsg.Data.(*game.ErrorData)
	require.True(t, ok)
	assert.Equal(t, "ROOM_NOT_FOUND", errorData.Code)
	
	// Test making guess without joining room
	err = client.SendGuess("apple")
	require.NoError(t, err)
	
	errorMsg2, err := client.WaitForMessageType("error", 2*time.Second)
	require.NoError(t, err)
	
	errorData2, ok := errorMsg2.Data.(*game.ErrorData)
	require.True(t, ok)
	assert.Equal(t, "NOT_IN_ROOM", errorData2.Code)
	
	// Connection should still be active after errors
	// Test by creating a valid room and joining
	room, err := server.roomManager.CreateRoom("Recovery Room", 2)
	require.NoError(t, err)
	
	err = client.SendJoin(room.ID, "Recovery Player")
	require.NoError(t, err)
	
	_, err = client.WaitForMessageType("join_success", 2*time.Second)
	require.NoError(t, err, "Client should recover and be able to join room after errors")
}

// Benchmark WebSocket message processing
func BenchmarkWebSocketMessageProcessing(b *testing.B) {
	log.SetOutput(nil) // Disable logging during benchmarks
	defer log.SetOutput(nil)
	
	server := setupTestServer(&testing.T{})
	defer server.Close()
	
	// Create room and connect clients
	room, err := server.roomManager.CreateRoom("Benchmark Room", 2)
	require.NoError(&testing.T{}, err)
	
	client1 := server.ConnectClient(&testing.T{}, "bench-player1")
	defer client1.Close()
	client2 := server.ConnectClient(&testing.T{}, "bench-player2")
	defer client2.Close()
	
	// Setup game
	client1.WaitForMessageType("connection_ack", 2*time.Second)
	client2.WaitForMessageType("connection_ack", 2*time.Second)
	
	client1.SendJoin(room.ID, "Bench Player 1")
	client2.SendJoin(room.ID, "Bench Player 2")
	
	client1.WaitForMessageType("join_success", 2*time.Second)
	client2.WaitForMessageType("join_success", 2*time.Second)
	client1.WaitForMessageType("game_started", 2*time.Second)
	client2.WaitForMessageType("game_started", 2*time.Second)
	
	// Benchmark guess processing
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		word := fmt.Sprintf("test%d", i%10)
		if len(word) < 5 {
			word = "tests"
		}
		
		start := time.Now()
		client1.SendGuess(word)
		client1.WaitForMessageType("guess_result", 1*time.Second)
		latency := time.Since(start)
		
		if latency > 100*time.Millisecond {
			b.Errorf("Message processing too slow: %v", latency)
		}
	}
}