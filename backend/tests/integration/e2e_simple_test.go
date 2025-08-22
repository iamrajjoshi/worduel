package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSimpleE2EFlow tests the basic end-to-end flow that currently works
func TestSimpleE2EFlow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Close()

	// Create two players
	player1 := NewE2EGameClient(t, suite, "player1")
	player2 := NewE2EGameClient(t, suite, "player2")
	defer player1.Close()
	defer player2.Close()

	t.Log("=== Simple E2E Flow Test ===")

	// Step 1: Player 1 creates a room via API
	t.Log("Step 1: Creating room...")
	roomID, err := player1.CreateRoom("Test Game Room")
	require.NoError(t, err)
	require.NotEmpty(t, roomID)
	t.Logf("✓ Created room: %s", roomID)

	// Step 2: Both players connect to WebSocket
	t.Log("Step 2: Connecting WebSocket...")
	require.NoError(t, player1.ConnectWebSocket())
	require.NoError(t, player2.ConnectWebSocket())
	t.Log("✓ Both players connected")

	// Step 3: Wait for connection acknowledgment
	t.Log("Step 3: Waiting for connection ack...")
	_, err = player1.WaitForMessage("connection_ack", 2*time.Second)
	require.NoError(t, err)
	_, err = player2.WaitForMessage("connection_ack", 2*time.Second)
	require.NoError(t, err)
	t.Log("✓ Connection acknowledgments received")

	// Step 4: Player 1 joins the room
	t.Log("Step 4: Player 1 joining room...")
	require.NoError(t, player1.JoinRoom(roomID, "Alice"))

	joinMsg1, err := player1.WaitForMessage("join_success", 3*time.Second)
	require.NoError(t, err)
	require.Equal(t, "join_success", string(joinMsg1.Type))
	t.Log("✓ Player 1 joined successfully")

	// Step 5: Player 2 joins the room
	t.Log("Step 5: Player 2 joining room...")
	require.NoError(t, player2.JoinRoom(roomID, "Bob"))

	joinMsg2, err := player2.WaitForMessage("join_success", 3*time.Second)
	require.NoError(t, err)
	require.Equal(t, "join_success", string(joinMsg2.Type))
	t.Log("✓ Player 2 joined successfully")

	// Step 6: Verify room state via API
	t.Log("Step 6: Verifying room state...")
	roomInfo, err := player1.GetRoomInfo(roomID)
	require.NoError(t, err)
	require.Equal(t, roomID, roomInfo.RoomID)
	require.Equal(t, 2, roomInfo.PlayerCount)
	require.Equal(t, 2, roomInfo.MaxPlayers)
	t.Log("✓ Room state verified via API")

	t.Log("=== Simple E2E Flow Test PASSED ===")
}

// TestBasicAPIOperations tests the REST API endpoints
func TestBasicAPIOperations(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Close()

	client := NewE2EGameClient(t, suite, "api-test-client")
	defer client.Close()

	t.Log("=== Basic API Operations Test ===")

	// Test room creation
	t.Log("Step 1: Creating room with custom settings...")
	roomID, err := client.CreateRoomWithOptions("Custom Game", 4)
	require.NoError(t, err)
	require.NotEmpty(t, roomID)
	t.Logf("✓ Created custom room: %s", roomID)

	// Test room retrieval
	t.Log("Step 2: Retrieving room information...")
	roomInfo, err := client.GetRoomInfo(roomID)
	require.NoError(t, err)
	require.Equal(t, roomID, roomInfo.RoomID)
	require.Equal(t, "Custom Game", roomInfo.Name)
	require.Equal(t, 4, roomInfo.MaxPlayers)
	require.Equal(t, 0, roomInfo.PlayerCount)
	t.Log("✓ Room information retrieved successfully")

	// Test health endpoints
	t.Log("Step 3: Testing health endpoints...")
	healthOK, err := client.CheckHealth()
	require.NoError(t, err)
	require.True(t, healthOK)
	t.Log("✓ Health check passed")

	t.Log("=== Basic API Operations Test PASSED ===")
}

// TestWebSocketBasicCommunication tests basic WebSocket functionality
func TestWebSocketBasicCommunication(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Close()

	client := NewE2EGameClient(t, suite, "ws-test-client")
	defer client.Close()

	t.Log("=== WebSocket Basic Communication Test ===")

	// Create room
	roomID, err := client.CreateRoom("WebSocket Test Room")
	require.NoError(t, err)
	t.Logf("✓ Created room: %s", roomID)

	// Connect WebSocket
	t.Log("Step 1: Connecting WebSocket...")
	require.NoError(t, client.ConnectWebSocket())
	t.Log("✓ WebSocket connected")

	// Wait for connection ack
	t.Log("Step 2: Waiting for connection acknowledgment...")
	ackMsg, err := client.WaitForMessage("connection_ack", 3*time.Second)
	require.NoError(t, err)
	require.Equal(t, "connection_ack", string(ackMsg.Type))
	t.Log("✓ Connection acknowledgment received")

	// Join room
	t.Log("Step 3: Joining room...")
	require.NoError(t, client.JoinRoom(roomID, "TestPlayer"))

	joinMsg, err := client.WaitForMessage("join_success", 3*time.Second)
	require.NoError(t, err)
	require.Equal(t, "join_success", string(joinMsg.Type))
	t.Log("✓ Join success message received")

	// Verify we can receive the data
	if joinData, ok := joinMsg.Data.(map[string]interface{}); ok {
		if gameState, exists := joinData["game_state"]; exists {
			t.Logf("✓ Game state received: %+v", gameState)
		}
		if players, exists := joinData["players"]; exists {
			t.Logf("✓ Players data received: %+v", players)
		}
	}

	t.Log("=== WebSocket Basic Communication Test PASSED ===")
}