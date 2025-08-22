# Requirements Document

## Introduction

The backend service is the core engine that powers the multiplayer competitive Wordle experience. Built in Go for optimal performance and Docker deployment, it handles real-time multiplayer game coordination, room management, word validation, and game state synchronization between competing players. The backend serves as both a WebSocket hub for real-time communication and a REST API for room management operations.

This backend enables two players to compete in solving the same Wordle puzzle while seeing each other's progress patterns (green/yellow/gray indicators) without revealing the actual letters guessed, maintaining the competitive tension while preserving game integrity.

## Alignment with Product Vision

This backend directly supports the product steering goals by:
- **Competitive Experience**: Enabling real-time multiplayer racing with opponent progress visibility
- **Privacy-Preserving Competition**: Sharing progress patterns without revealing opponent strategies  
- **Zero-Friction Deployment**: Single Docker container for easy hosting
- **Performance Standards**: Sub-100ms latency and support for 1000+ concurrent connections
- **Accessibility**: Providing clean APIs that support keyboard navigation and screen reader compatibility

The backend architecture follows the technical steering by implementing event-driven real-time updates, stateless REST endpoints, and in-memory storage for optimal performance and simplicity.

## Requirements

### Requirement 1: Real-Time Multiplayer Game Coordination

**User Story:** As a player, I want to compete against another player in real-time Wordle, so that we can race to solve the same word while seeing each other's progress patterns.

#### Acceptance Criteria

1. WHEN two players are connected to the same room THEN the system SHALL start a new game with an identical target word for both players
2. WHEN a player submits a guess THEN the system SHALL validate the word, compute letter results (green/yellow/gray), and broadcast the pattern to the opponent within 100ms
3. WHEN a player guesses the correct word THEN the system SHALL immediately notify both players of the game completion and declare the winner
4. WHEN a player disconnects during gameplay THEN the system SHALL notify the remaining player and offer options to wait or complete solo
5. IF a player submits more than 6 guesses THEN the system SHALL end the game and declare them eliminated
6. WHEN game time exceeds 30 minutes THEN the system SHALL automatically end the game and clean up resources

### Requirement 2: WebSocket Real-Time Communication

**User Story:** As a player, I want instant updates about my opponent's progress, so that I feel the competitive tension and can adjust my strategy accordingly.

#### Acceptance Criteria

1. WHEN a client connects via WebSocket THEN the system SHALL establish a persistent connection and associate it with a room
2. WHEN receiving a guess message THEN the system SHALL validate, process, and broadcast results to all room participants within 10ms
3. WHEN a connection is lost THEN the system SHALL attempt graceful cleanup and notify remaining players
4. IF message rate exceeds 60 messages per minute per connection THEN the system SHALL rate limit and temporarily suspend the connection
5. WHEN invalid message format is received THEN the system SHALL log the error and close the connection
6. WHEN room becomes empty THEN the system SHALL automatically clean up WebSocket resources and game state

### Requirement 3: Room Management System

**User Story:** As a host, I want to create a game room and share a simple code with friends, so that they can easily join and we can start playing immediately.

#### Acceptance Criteria

1. WHEN creating a room via REST API THEN the system SHALL generate a unique 6-character alphanumeric room code and return room details
2. WHEN a room is created THEN the system SHALL initialize empty game state and set room status to "waiting"
3. WHEN second player joins a waiting room THEN the system SHALL change status to "playing" and select a random target word
4. IF room code is invalid or expired THEN the system SHALL return appropriate error response with clear message
5. WHEN room has been inactive for 30 minutes THEN the system SHALL automatically delete the room and free resources
6. WHEN querying room status THEN the system SHALL return current player count, game status, and creation timestamp

### Requirement 4: Word Validation and Dictionary Management

**User Story:** As a player, I want my word guesses to be validated against a comprehensive dictionary, so that only legitimate words are accepted and the game remains fair.

#### Acceptance Criteria

1. WHEN a player submits a 5-letter guess THEN the system SHALL validate it against the accepted words dictionary
2. IF the submitted word is not in the dictionary THEN the system SHALL reject it with an "invalid word" error without consuming a guess
3. WHEN selecting target words THEN the system SHALL randomly choose from a curated list of common 5-letter words
4. WHEN validating guesses THEN the system SHALL compare letters against the target word and return accurate green/yellow/gray results
5. IF target word selection fails THEN the system SHALL fallback to a default word list and log the error
6. WHEN loading dictionaries THEN the system SHALL embed word lists in the binary for offline operation

### Requirement 5: Game State Management

**User Story:** As the system, I need to maintain accurate game state for all active rooms, so that players receive consistent information and game integrity is preserved.

#### Acceptance Criteria

1. WHEN a new game starts THEN the system SHALL initialize game state with target word, empty guess history, and player tracking
2. WHEN a guess is processed THEN the system SHALL update game state atomically and maintain guess history for both players
3. WHEN accessing game state concurrently THEN the system SHALL use proper locking to prevent race conditions
4. WHEN game ends THEN the system SHALL mark final state with winner/loser status and completion timestamp
5. IF memory usage exceeds limits THEN the system SHALL implement cleanup of oldest inactive rooms first
6. WHEN system restarts THEN the system SHALL clear all in-memory game state (no persistence requirement)

### Requirement 6: REST API for Room Operations

**User Story:** As a frontend client, I need REST endpoints for room management operations, so that I can create rooms, check status, and handle non-real-time operations efficiently.

#### Acceptance Criteria

1. WHEN POST /api/rooms is called THEN the system SHALL create a new room and return JSON with roomID and creation timestamp
2. WHEN GET /api/rooms/{id} is called THEN the system SHALL return room status including player count and game state
3. WHEN GET /health is called THEN the system SHALL return system health status and active room count
4. IF invalid room ID format is provided THEN the system SHALL return 400 Bad Request with error details
5. WHEN CORS preflight requests are received THEN the system SHALL respond with appropriate headers for frontend access
6. WHEN API rate limits are exceeded THEN the system SHALL return 429 Too Many Requests with retry-after header

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each package should handle one specific domain (game logic, WebSocket, API, room management)
- **Modular Design**: Internal packages should be isolated with clean interfaces and minimal interdependencies
- **Dependency Management**: Use Go modules with minimal external dependencies (only gorilla/websocket, gorilla/mux, rs/cors)
- **Clear Interfaces**: Define clean contracts between WebSocket hub, game state, and room management layers

### Performance
- **Response Time**: WebSocket message processing must complete within 10ms, REST API responses within 100ms
- **Throughput**: Support 1000+ concurrent WebSocket connections with graceful degradation
- **Memory Usage**: Maintain less than 50MB RAM usage per 100 concurrent games
- **CPU Usage**: Operate under 5% CPU usage during normal operations
- **Startup Time**: Application must start and be ready to serve requests within 2 seconds

### Security
- **Input Validation**: All user inputs must be sanitized and validated before processing
- **Rate Limiting**: Implement rate limiting for both WebSocket messages and REST API calls
- **Origin Validation**: WebSocket connections must validate origin headers to prevent CSRF
- **Resource Protection**: Prevent memory exhaustion attacks through connection limits and automatic cleanup
- **Error Information**: Error messages must not expose internal system details or stack traces

### Reliability
- **Availability**: Target 99.5% uptime with graceful handling of connection failures
- **Error Recovery**: Automatic cleanup of failed connections and corrupted game states
- **Resource Management**: Automatic garbage collection of expired rooms and idle connections
- **Graceful Shutdown**: Handle SIGTERM/SIGINT signals to complete ongoing requests before termination
- **Circuit Breaker**: Implement connection limits to prevent resource exhaustion

### Usability
- **Error Messages**: Provide clear, actionable error messages for invalid operations
- **API Documentation**: REST endpoints must follow consistent patterns with predictable responses
- **Development Mode**: Support debug logging and verbose error reporting for development
- **Health Checks**: Provide comprehensive health check endpoint for monitoring and orchestration
- **Configuration**: Support environment variable configuration for deployment flexibility