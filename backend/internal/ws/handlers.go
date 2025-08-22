package ws

import (
	"encoding/json"
	"math/rand"
	"time"

	"worduel-backend/internal/game"
	"worduel-backend/internal/logging"
	"worduel-backend/internal/room"
)

var logger = logging.CreateLogger("ws.handlers")

// MessageHandler handles WebSocket messages and coordinates with game logic
type MessageHandler struct {
	hub         *Hub
	roomManager *room.RoomManager
	gameLogic   *game.GameLogic
}

// NewMessageHandler creates a new message handler instance
func NewMessageHandler(hub *Hub, roomManager *room.RoomManager, gameLogic *game.GameLogic) *MessageHandler {
	return &MessageHandler{
		hub:         hub,
		roomManager: roomManager,
		gameLogic:   gameLogic,
	}
}

// HandleMessage routes messages to appropriate handlers based on message type
func (mh *MessageHandler) HandleMessage(clientMessage *ClientMessage) {
	client := clientMessage.GetClient()
	message := clientMessage.GetMessage()

	// Update client activity
	mh.updateClientActivity(client)

	switch message.Type {
	case game.MessageTypeJoin:
		mh.handleJoinMessage(client, message)
	case game.MessageTypeLeave:
		mh.handleLeaveMessage(client, message)
	case game.MessageTypeGuess:
		mh.handleGuessMessage(client, message)
	case game.MessageTypeChat:
		mh.handleChatMessage(client, message)
	default:
		mh.sendError(client, "UNKNOWN_MESSAGE_TYPE", "Unknown message type: "+string(message.Type))
		logger.Warn("Unknown message type received",
			"message_type", string(message.Type),
			"client_id", client.GetID())
	}
}

// handleJoinMessage processes room join requests
func (mh *MessageHandler) handleJoinMessage(client *Client, message *game.Message) {
	// Parse join data
	joinData, ok := message.Data.(map[string]interface{})
	if !ok {
		mh.sendError(client, "INVALID_JOIN_DATA", "Invalid join message data format")
		return
	}

	roomID, ok := joinData["room_id"].(string)
	if !ok || roomID == "" {
		mh.sendError(client, "MISSING_ROOM_ID", "Room ID is required for joining")
		return
	}

	playerName, ok := joinData["player_name"].(string)
	if !ok || playerName == "" {
		mh.sendError(client, "MISSING_PLAYER_NAME", "Player name is required for joining")
		return
	}

	// Check if client is already in a room
	if client.GetRoomID() != "" {
		mh.sendError(client, "ALREADY_IN_ROOM", "Already connected to a room. Leave current room first")
		return
	}

	// Attempt to join room through room manager
	gameRoom, err := mh.roomManager.JoinRoom(roomID, client.GetID(), playerName)
	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			mh.sendError(client, "ROOM_NOT_FOUND", "Room not found")
		case room.ErrRoomFull:
			mh.sendError(client, "ROOM_FULL", "Room is full")
		case room.ErrPlayerExists:
			mh.sendError(client, "PLAYER_EXISTS", "Player already exists in room")
		case room.ErrInvalidRoomCode:
			mh.sendError(client, "INVALID_ROOM_CODE", "Invalid room code format")
		default:
			mh.sendError(client, "JOIN_FAILED", "Failed to join room: "+err.Error())
		}
		return
	}

	// Associate client with room
	mh.hub.mutex.Lock()
	client.SetRoom(roomID, client.GetID())
	
	if mh.hub.roomClients[roomID] == nil {
		mh.hub.roomClients[roomID] = make(map[string]*Client)
	}
	mh.hub.roomClients[roomID][client.GetID()] = client
	mh.hub.mutex.Unlock()

	// Check if we should start the game (2 players joined)
	gameRoom.RLock()
	playerCount := len(gameRoom.Players)
	gameStatus := gameRoom.GameState.Status
	gameRoom.RUnlock()

	if playerCount >= 2 && gameStatus == game.GameStatusWaiting {
		mh.startNewGame(gameRoom)
	}

	// Send join success response with current game state
	response := &game.Message{
		Type:      "join_success",
		PlayerID:  client.GetID(),
		RoomID:    roomID,
		Timestamp: time.Now(),
		Data: &game.GameUpdateData{
			GameState: gameRoom.GameState,
			Players:   gameRoom.Players,
		},
	}
	client.SendJSON(response)

	// Notify other players in room
	playerJoinedMessage := &game.Message{
		Type:      game.MessageTypePlayerUpdate,
		PlayerID:  client.GetID(),
		RoomID:    roomID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"event":       "player_joined",
			"player_id":   client.GetID(),
			"player_name": playerName,
			"player_count": playerCount,
		},
	}
	mh.broadcastToRoom(roomID, playerJoinedMessage, client.GetID())

	logger.Info("Player joined room",
		"event_type", "player_joined",
		"player_name", playerName,
		"room_id", roomID,
		"player_id", client.GetID(),
		"game_state", string(gameStatus),
		"player_count", playerCount)
}

// handleLeaveMessage processes room leave requests
func (mh *MessageHandler) handleLeaveMessage(client *Client, message *game.Message) {
	roomID := client.GetRoomID()
	playerID := client.GetPlayerID()
	
	if roomID == "" {
		mh.sendError(client, "NOT_IN_ROOM", "Not currently in any room")
		return
	}

	// Leave room through room manager
	if err := mh.roomManager.LeaveRoom(roomID, playerID); err != nil {
		logger.Error("Failed to leave room", 
			"error", err.Error(),
			"room_id", roomID,
			"player_id", playerID)
		mh.sendError(client, "LEAVE_FAILED", "Failed to leave room: "+err.Error())
		return
	}

	// Remove client from hub room associations
	mh.hub.mutex.Lock()
	mh.removeClientFromRoom(client, roomID)
	client.SetRoom("", "")
	mh.hub.mutex.Unlock()

	// Send leave confirmation
	response := &game.Message{
		Type:      "leave_success",
		PlayerID:  playerID,
		RoomID:    roomID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message": "Successfully left room",
		},
	}
	client.SendJSON(response)

	// Notify remaining players
	playerLeftMessage := &game.Message{
		Type:      game.MessageTypePlayerUpdate,
		PlayerID:  playerID,
		RoomID:    roomID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"event":     "player_left",
			"player_id": playerID,
		},
	}
	mh.broadcastToRoom(roomID, playerLeftMessage, client.GetID())

	logger.Info("Player left room",
		"event_type", "player_left",
		"room_id", roomID,
		"player_id", playerID,
		"game_state", "unknown")
}

// handleGuessMessage processes guess submissions and integrates with game logic
func (mh *MessageHandler) handleGuessMessage(client *Client, message *game.Message) {
	roomID := client.GetRoomID()
	playerID := client.GetPlayerID()
	
	if roomID == "" {
		mh.sendError(client, "NOT_IN_ROOM", "Must join a room before making guesses")
		return
	}

	// Parse guess data
	guessData, ok := message.Data.(map[string]interface{})
	if !ok {
		mh.sendError(client, "INVALID_GUESS_DATA", "Invalid guess message data format")
		return
	}

	word, ok := guessData["word"].(string)
	if !ok || word == "" {
		mh.sendError(client, "MISSING_WORD", "Word is required for guess")
		return
	}

	// Get room for processing
	gameRoom, err := mh.roomManager.GetRoom(roomID)
	if err != nil {
		mh.sendError(client, "ROOM_NOT_FOUND", "Room not found")
		return
	}

	// Process guess through game logic
	guessResult, err := mh.gameLogic.ProcessGuess(gameRoom, playerID, word)
	if err != nil {
		switch err {
		case game.ErrInvalidWord:
			mh.sendError(client, "INVALID_WORD", "Word not found in dictionary")
		case game.ErrGameNotActive:
			mh.sendError(client, "GAME_NOT_ACTIVE", "Game is not currently active")
		case game.ErrPlayerNotFound:
			mh.sendError(client, "PLAYER_NOT_FOUND", "Player not found in game")
		case game.ErrTooManyGuesses:
			mh.sendError(client, "TOO_MANY_GUESSES", "Maximum number of guesses reached")
		case game.ErrGameAlreadyWon:
			mh.sendError(client, "GAME_ALREADY_WON", "Game has already been won")
		case game.ErrInvalidWordLength:
			mh.sendError(client, "INVALID_WORD_LENGTH", "Word must be exactly 5 letters")
		default:
			mh.sendError(client, "GUESS_PROCESSING_FAILED", "Failed to process guess: "+err.Error())
		}
		return
	}

	// Send guess result back to the player
	guessResponse := &game.Message{
		Type:      "guess_result",
		PlayerID:  playerID,
		RoomID:    roomID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"word":       guessResult.Word,
			"results":    guessResult.Results,
			"is_correct": guessResult.IsCorrect,
			"timestamp":  guessResult.Timestamp,
		},
	}
	client.SendJSON(guessResponse)

	// Broadcast game update to all players in room
	mh.broadcastGameUpdate(gameRoom, playerID)

	// Check if game is complete
	if isComplete, winner := mh.gameLogic.IsComplete(gameRoom); isComplete {
		mh.handleGameCompletion(gameRoom, winner)
	}

	logger.Info("Guess processed",
		"event_type", "guess_processed",
		"room_id", roomID,
		"player_id", playerID,
		"game_state", "active",
		"word", word,
		"is_correct", guessResult.IsCorrect)
}

// handleChatMessage processes chat messages within rooms
func (mh *MessageHandler) handleChatMessage(client *Client, message *game.Message) {
	roomID := client.GetRoomID()
	if roomID == "" {
		mh.sendError(client, "NOT_IN_ROOM", "Must join a room to send chat messages")
		return
	}

	// Parse chat data
	chatData, ok := message.Data.(map[string]interface{})
	if !ok {
		mh.sendError(client, "INVALID_CHAT_DATA", "Invalid chat message data format")
		return
	}

	messageText, ok := chatData["message"].(string)
	if !ok || messageText == "" {
		mh.sendError(client, "MISSING_MESSAGE", "Message text is required")
		return
	}

	// Create chat message for broadcasting
	chatMessage := &game.Message{
		Type:      game.MessageTypeChat,
		PlayerID:  client.GetPlayerID(),
		RoomID:    roomID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message":     messageText,
			"player_id":   client.GetPlayerID(),
			"player_name": mh.getPlayerName(roomID, client.GetPlayerID()),
		},
	}

	// Broadcast to all players in room
	mh.broadcastToRoom(roomID, chatMessage, "")
}

// startNewGame initializes a new game when enough players have joined
func (mh *MessageHandler) startNewGame(room *game.Room) {
	// TODO: Get target word from dictionary service
	// For now, use a simple list of words
	words := []string{"apple", "bread", "chair", "dream", "eagle", "flame", "grape", "house", "image", "juice"}
	targetWord := words[rand.Intn(len(words))]

	if err := mh.gameLogic.StartGame(room, targetWord); err != nil {
		logger.Error("Failed to start game", "error", err.Error(), "room_id", room.ID)
		return
	}

	// Broadcast game start to all players in room
	gameStartMessage := &game.Message{
		Type:      "game_started",
		RoomID:    room.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"target_word_length": 5,
			"max_guesses":        6,
			"game_status":        string(game.GameStatusActive),
		},
	}

	mh.broadcastToRoom(room.ID, gameStartMessage, "")
	logger.Info("Game started", 
		"event_type", "game_started",
		"room_id", room.ID,
		"game_state", string(game.GameStatusActive),
		"target_word", targetWord)
}

// broadcastGameUpdate sends current game state to all players in room
func (mh *MessageHandler) broadcastGameUpdate(room *game.Room, triggeringPlayerID string) {
	room.RLock()
	roomID := room.ID
	room.RUnlock()

	// Send personalized game summaries to each player
	mh.hub.mutex.RLock()
	roomClients := mh.hub.roomClients[roomID]
	mh.hub.mutex.RUnlock()

	if roomClients == nil {
		return
	}

	for clientID, client := range roomClients {
		if client.IsClosed() {
			continue
		}

		// Get personalized game summary for this client
		gameSummary := mh.gameLogic.GetGameSummary(room, clientID)
		
		gameUpdateMessage := &game.Message{
			Type:      game.MessageTypeGameUpdate,
			PlayerID:  clientID,
			RoomID:    roomID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"game_summary":       gameSummary,
				"triggering_player":  triggeringPlayerID,
				"update_reason":      "guess_processed",
			},
		}

		if err := client.SendJSON(gameUpdateMessage); err != nil {
			logger.Error("Failed to send game update", "error", err.Error(), "client_id", clientID)
		}
	}
}

// handleGameCompletion processes game completion and broadcasts results
func (mh *MessageHandler) handleGameCompletion(room *game.Room, winner string) {
	room.RLock()
	roomID := room.ID
	room.RUnlock()

	gameCompletionMessage := &game.Message{
		Type:      "game_completed",
		RoomID:    roomID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"winner":        winner,
			"game_status":   string(game.GameStatusFinished),
			"completed_at":  time.Now(),
		},
	}

	mh.broadcastToRoom(roomID, gameCompletionMessage, "")
	
	// Send final game summary to all players
	mh.broadcastGameUpdate(room, winner)
	
	logger.Info("Game completed",
		"event_type", "game_completed", 
		"room_id", roomID,
		"player_id", winner,
		"game_state", string(game.GameStatusFinished),
		"winner", winner)
}

// Helper methods

// updateClientActivity updates the client's last activity timestamp
func (mh *MessageHandler) updateClientActivity(client *Client) {
	if roomID := client.GetRoomID(); roomID != "" {
		if room, err := mh.roomManager.GetRoom(roomID); err == nil {
			room.RLock()
			if player, exists := room.Players[client.GetPlayerID()]; exists {
				player.LastActivity = time.Now()
			}
			room.RUnlock()
		}
	}
}

// removeClientFromRoom removes a client from a room's client list in the hub
func (mh *MessageHandler) removeClientFromRoom(client *Client, roomID string) {
	if roomClients, exists := mh.hub.roomClients[roomID]; exists {
		delete(roomClients, client.GetID())
		
		if len(roomClients) == 0 {
			delete(mh.hub.roomClients, roomID)
		}
	}
}

// broadcastToRoom sends a message to all clients in a specific room
func (mh *MessageHandler) broadcastToRoom(roomID string, message *game.Message, excludeClientID string) {
	mh.hub.mutex.RLock()
	roomClients := mh.hub.roomClients[roomID]
	mh.hub.mutex.RUnlock()

	if roomClients == nil {
		return
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		logger.Error("Error marshaling broadcast message", "error", err.Error(), "room_id", roomID)
		return
	}

	for clientID, client := range roomClients {
		if clientID != excludeClientID && !client.IsClosed() {
			if err := client.SendMessage(messageData); err != nil {
				logger.Error("Error broadcasting message", "error", err.Error(),
					"client_id", clientID,
					"room_id", roomID)
			}
		}
	}
}

// sendError sends an error message to a specific client
func (mh *MessageHandler) sendError(client *Client, code, message string) {
	errorMsg := &game.Message{
		Type:      game.MessageTypeError,
		PlayerID:  client.GetPlayerID(),
		RoomID:    client.GetRoomID(),
		Timestamp: time.Now(),
		Data: &game.ErrorData{
			Code:    code,
			Message: message,
		},
	}
	client.SendJSON(errorMsg)
}

// getPlayerName retrieves a player's name from room data
func (mh *MessageHandler) getPlayerName(roomID, playerID string) string {
	if room, err := mh.roomManager.GetRoom(roomID); err == nil {
		room.RLock()
		defer room.RUnlock()
		
		if player, exists := room.Players[playerID]; exists {
			return player.Name
		}
	}
	return "Unknown Player"
}