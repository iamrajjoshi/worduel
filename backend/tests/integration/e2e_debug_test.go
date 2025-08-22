package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"worduel-backend/internal/game"
)

// TestWebSocketTwoPlayerJoin tests two players joining to see game start
func TestWebSocketTwoPlayerJoin(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Close()

	// Create two players
	player1 := NewE2EGameClient(t, suite, "player1")
	player2 := NewE2EGameClient(t, suite, "player2")
	defer player1.Close()
	defer player2.Close()

	// Step 1: Create room
	t.Log("Creating room...")
	roomID, err := player1.CreateRoom("Debug Room")
	require.NoError(t, err)
	require.NotEmpty(t, roomID)
	t.Logf("Created room: %s", roomID)

	// Step 2: Connect WebSocket
	t.Log("Connecting WebSocket...")
	require.NoError(t, player1.ConnectWebSocket())
	require.NoError(t, player2.ConnectWebSocket())

	// Wait for connection ack
	_, err = player1.WaitForMessage("connection_ack", 2*time.Second)
	require.NoError(t, err)
	_, err = player2.WaitForMessage("connection_ack", 2*time.Second)
	require.NoError(t, err)

	// Step 3: Player 1 joins room
	t.Log("Player 1 joining room...")
	require.NoError(t, player1.JoinRoom(roomID, "Alice"))
	
	_, err = player1.WaitForMessage("join_success", 2*time.Second)
	require.NoError(t, err)
	t.Log("Player 1 joined")

	// Step 4: Player 2 joins room
	t.Log("Player 2 joining room...")
	require.NoError(t, player2.JoinRoom(roomID, "Bob"))
	
	_, err = player2.WaitForMessage("join_success", 2*time.Second)
	require.NoError(t, err)
	t.Log("Player 2 joined")

	// Step 5: Collect all messages for 5 seconds to see if game starts
	t.Log("Collecting messages for 5 seconds...")
	
	messages1 := make([]*game.Message, 0)
	messages2 := make([]*game.Message, 0)
	timeout := time.After(5 * time.Second)
	
	for {
		select {
		case msg := <-player1.Messages:
			messages1 = append(messages1, msg)
			t.Logf("Player1 Message: Type=%s, Data=%+v", string(msg.Type), msg.Data)
		case msg := <-player2.Messages:
			messages2 = append(messages2, msg)
			t.Logf("Player2 Message: Type=%s, Data=%+v", string(msg.Type), msg.Data)
		case err := <-player1.Errors:
			t.Fatalf("Player1 WebSocket error: %v", err)
		case err := <-player2.Errors:
			t.Fatalf("Player2 WebSocket error: %v", err)
		case <-timeout:
			t.Logf("Player1 collected %d messages, Player2 collected %d messages", len(messages1), len(messages2))
			return
		}
	}
}