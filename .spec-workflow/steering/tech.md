# Technical Steering

## Architecture Overview

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Frontend│    │   Go Backend    │    │  Docker Host    │
│                 │    │                 │    │                 │
│  • Game UI      │◄──►│  • WebSocket    │    │  • Single       │
│  • Room Mgmt    │    │  • REST API     │    │    Container    │
│  • Real-time    │    │  • Game Logic   │    │  • Embedded     │
│    Updates      │    │  • In-Memory    │    │    Frontend     │
│                 │    │    Storage      │    │  • Port 8080    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Technology Stack Rationale

**Backend: Go**
- **Single Binary Deployment**: Compiles to self-contained executable
- **Excellent Concurrency**: Goroutines perfect for real-time multiplayer
- **Fast Startup**: Critical for container scaling and quick deployment
- **Minimal Dependencies**: Reduces Docker image size and complexity
- **Strong WebSocket Support**: Built-in `gorilla/websocket` library
- **Cross-Platform**: Builds for any target architecture

**Frontend: React + TypeScript**
- **Static File Serving**: No server-side rendering complexity
- **Component Reusability**: Clean separation of game components
- **Type Safety**: TypeScript prevents runtime errors in game logic
- **Responsive Design**: React's ecosystem for mobile-first development
- **Bundle Optimization**: Vite for fast builds and small bundle sizes

**Communication Layer**
- **WebSocket Primary**: Real-time game updates and opponent progress
- **REST API Secondary**: Room creation, health checks, static content
- **JSON Protocol**: Simple, debuggable message format

### Data Flow Design

```
Frontend                    Backend                     
─────────                   ───────
┌─────────┐    WebSocket    ┌─────────┐    Goroutine     ┌─────────┐
│ Game UI │◄──────────────►│   Hub   │◄───────────────►│  Room   │
└─────────┘                └─────────┘                 └─────────┘
     │                           │                           │
     │ HTTP/REST                 │                           │
     │                           ▼                           ▼
┌─────────┐                 ┌─────────┐                 ┌─────────┐
│Room Join│                 │API Layer│                 │Game     │
│Page     │◄────────────────┤         │◄────────────────│State    │
└─────────┘                 └─────────┘                 └─────────┘
```

### Integration Patterns

**Event-Driven Architecture**
- Game state changes trigger events
- Events broadcast to connected clients
- Immutable state updates for consistency

**Pub/Sub Pattern for Real-Time Updates**
- Room-scoped channels for game events
- Client subscription management via WebSocket
- Automatic cleanup on disconnection

**Stateless Request Handling**
- REST endpoints are stateless
- Game state lives in memory structures
- Session affinity not required (single instance)

## Development Standards

### Go Backend Standards

**Code Structure**
```go
// Package organization
package main // main.go - application entry point
package game // game logic, state management
package ws   // WebSocket handling
package api  // REST endpoint handlers
package room // room management
```

**Error Handling Pattern**
```go
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    result, err := h.processRequest(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        log.Printf("Error processing request: %v", err)
        return
    }
    json.NewEncoder(w).Encode(result)
}
```

**Concurrency Guidelines**
- Use channels for goroutine communication
- Protect shared state with sync.RWMutex
- Context cancellation for request timeouts
- Worker pool pattern for CPU-intensive tasks

**Testing Requirements**
- Unit tests for all game logic functions
- Integration tests for WebSocket communication
- Table-driven tests for game state transitions
- Mock interfaces for external dependencies
- Minimum 80% code coverage

### React Frontend Standards

**Component Architecture**
```typescript
// Hook-based functional components
interface GameBoardProps {
  guesses: Guess[];
  currentGuess: string;
  onGuessSubmit: (guess: string) => void;
}

const GameBoard: React.FC<GameBoardProps> = ({ guesses, currentGuess, onGuessSubmit }) => {
  // Component logic
};
```

**State Management**
- React Context for game state
- useReducer for complex state transitions  
- Custom hooks for WebSocket communication
- Local state for UI-only concerns

**TypeScript Standards**
- Strict mode enabled
- Interface-first design for data structures
- No `any` types in production code
- Generic types for reusable components

**Testing Requirements**
- Jest + React Testing Library
- Component unit tests
- Custom hook testing
- WebSocket mock testing
- E2E tests with Playwright

### Security Guidelines

**Input Validation**
- Sanitize all user inputs on both client and server
- Validate word guesses against approved dictionary
- Rate limiting on guess submissions
- Room code format validation

**WebSocket Security**
- Connection rate limiting per IP
- Message size limits
- Origin validation
- Automatic disconnection for malformed messages

**Data Protection**
- No persistent storage of game data
- Automatic memory cleanup after games
- No personal information collection
- CORS properly configured for frontend origins

**Container Security**
- Non-root user in Docker container
- Minimal base image (alpine or scratch)
- No sensitive information in environment variables
- Security scanning in build pipeline

### Performance Standards

**Backend Performance**
- **Memory Usage**: <50MB per 100 concurrent games
- **CPU Usage**: <5% during normal operation
- **Response Time**: <10ms for game state updates
- **Throughput**: Support 1000+ concurrent WebSocket connections

**Frontend Performance**
- **Initial Load**: <3 seconds to interactive
- **Bundle Size**: <500KB compressed
- **Runtime Performance**: 60fps animations
- **Memory Leaks**: Zero detected leaks during testing

**Network Optimization**
- WebSocket message batching for rapid updates
- Gzip compression for HTTP responses  
- CDN-ready static asset serving
- Efficient JSON message protocols

## Technology Choices

### Programming Languages & Versions

**Go Backend: Go 1.21+**
- Module support for dependency management
- Generics for type-safe collections
- Improved performance and memory management
- Security patches and stability improvements

**Frontend: Node 18+ LTS**
- Native ES modules support
- Improved npm workspaces
- Better TypeScript integration
- Long-term stability guarantees

### Frameworks & Libraries

**Backend Dependencies (go.mod)**
```go
module worduel

go 1.21

require (
    github.com/gorilla/websocket v1.5.0  // WebSocket support
    github.com/gorilla/mux v1.8.0        // HTTP routing
    github.com/rs/cors v1.10.0           // CORS handling
)
```

**Frontend Dependencies**
```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0"
  },
  "devDependencies": {
    "vite": "^4.4.0",
    "typescript": "^5.0.0",
    "@vitejs/plugin-react": "^4.0.0"
  }
}
```

### Development Tools

**Build Tools**
- **Go**: Native `go build` with cross-compilation
- **Frontend**: Vite for fast development and optimized builds
- **Docker**: Multi-stage builds for minimal image size

**Code Quality**
- **Go**: `golangci-lint`, `gofmt`, `go vet`
- **Frontend**: ESLint, Prettier, TypeScript strict mode
- **Both**: Pre-commit hooks for code formatting

**Monitoring & Logging**
- **Structured Logging**: JSON format for easy parsing
- **Metrics**: Basic HTTP and WebSocket metrics
- **Health Checks**: Simple endpoint for container orchestration
- **Debug Mode**: Verbose logging for development

### Deployment Infrastructure

**Docker Strategy**
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS backend-builder
# Build Go binary

FROM node:18-alpine AS frontend-builder  
# Build React app

FROM alpine:latest
# Copy binary and static files
# Expose port 8080
# Run as non-root user
```

**Single Container Design**
- Embedded static file serving from Go binary
- Configuration via environment variables
- Graceful shutdown handling
- Health check endpoint at `/health`

**Resource Requirements**
- **Minimum**: 512MB RAM, 1 CPU core
- **Recommended**: 1GB RAM, 2 CPU cores  
- **Storage**: <100MB disk space
- **Network**: HTTP/HTTPS (80/443) and custom port (8080)

## Patterns & Best Practices

### Recommended Code Patterns

**Game State Management**
```go
type GameState struct {
    mu       sync.RWMutex
    players  map[string]*Player
    word     string
    status   GameStatus
    created  time.Time
}

func (gs *GameState) UpdatePlayer(playerID string, guess Guess) error {
    gs.mu.Lock()
    defer gs.mu.Unlock()
    // Update logic with proper locking
}
```

**WebSocket Message Handling**
```go
type MessageHandler interface {
    Handle(conn *websocket.Conn, msg Message) error
}

type GameMessageHandler struct {
    roomManager *RoomManager
}

func (h *GameMessageHandler) Handle(conn *websocket.Conn, msg Message) error {
    switch msg.Type {
    case "guess":
        return h.handleGuess(conn, msg)
    case "join":
        return h.handleJoin(conn, msg)
    default:
        return fmt.Errorf("unknown message type: %s", msg.Type)
    }
}
```

**React State Patterns**
```typescript
// Game state context
interface GameState {
  room: Room | null;
  player: Player | null;
  opponent: Opponent | null;
  gameStatus: 'waiting' | 'playing' | 'finished';
  guesses: Guess[];
}

const GameContext = createContext<GameState | null>(null);

// Custom hook for game operations
function useGameOperations() {
  const dispatch = useContext(GameDispatchContext);
  
  const submitGuess = useCallback((word: string) => {
    dispatch({ type: 'SUBMIT_GUESS', payload: { word } });
  }, [dispatch]);
  
  return { submitGuess };
}
```

### Error Handling Approaches

**Backend Error Strategy**
```go
// Custom error types
type GameError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e GameError) Error() string {
    return e.Message
}

// Error wrapping
func validateGuess(word string) error {
    if len(word) != 5 {
        return GameError{
            Code:    "INVALID_WORD_LENGTH",
            Message: "Word must be exactly 5 letters",
            Details: fmt.Sprintf("provided: %d letters", len(word)),
        }
    }
    return nil
}
```

**Frontend Error Handling**
```typescript
// Error boundary for React components
class GameErrorBoundary extends React.Component {
  state = { hasError: false, error: null };
  
  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }
  
  render() {
    if (this.state.hasError) {
      return <ErrorDisplay error={this.state.error} />;
    }
    return this.props.children;
  }
}

// WebSocket error handling
function useWebSocket() {
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    ws.onerror = (event) => {
      setError('Connection lost. Attempting to reconnect...');
      // Implement exponential backoff reconnection
    };
  }, []);
}
```

### Logging and Monitoring

**Structured Logging Format**
```go
log.Printf(`{"level":"info","msg":"player joined","roomID":"%s","playerID":"%s","timestamp":"%s"}`, 
    roomID, playerID, time.Now().UTC().Format(time.RFC3339))
```

**Key Metrics to Track**
- Active connections count
- Rooms created per hour
- Average game duration
- Guess validation errors
- WebSocket disconnection reasons

### Documentation Standards

**API Documentation**
```go
// CreateRoom creates a new game room and returns the room ID
// POST /api/rooms
// Response: {"roomID": "ABC123", "created": "2023-10-01T12:00:00Z"}
func (h *APIHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

**Component Documentation**
```typescript
/**
 * GameBoard component displays the current game state and handles user input
 * 
 * @param guesses - Array of previous guess attempts
 * @param currentGuess - The word currently being typed
 * @param onGuessSubmit - Callback fired when user submits a guess
 * @param disabled - Whether the board should accept input
 */
interface GameBoardProps {
  guesses: Guess[];
  currentGuess: string;
  onGuessSubmit: (guess: string) => void;
  disabled?: boolean;
}
```

This technical steering document provides the architectural foundation and development standards for building a high-performance, easily deployable multiplayer Wordle game that meets all specified requirements while maintaining code quality and operational simplicity.