# Worduel 🎯

![Tests](https://github.com/iamrajjoshi/worduel/actions/workflows/test.yml/badge.svg?branch=main)

A real-time multiplayer word guessing game where players compete to solve word puzzles faster than their opponents. Built with Go backend and WebSocket support for seamless multiplayer gameplay.

## Features

- **Real-time Multiplayer**: Compete against other players in real-time using WebSocket connections
- **Room-based Gameplay**: Create or join game rooms with multiple players
- **Wordle-style Mechanics**: Classic word guessing with color-coded feedback:
  - 🟢 Green: Correct letter in correct position
  - 🟡 Yellow: Letter exists but in wrong position
  - ⚫ Gray: Letter not in the word
- **Thread-safe Game State**: Concurrent player management with proper synchronization
- **Comprehensive Word Dictionary**: Includes common and valid word lists
- **Player Statistics**: Track scores, guesses, and game history
- **Production-Ready Monitoring**: Structured logging with optional Sentry integration
- **Comprehensive Configuration**: Environment-based configuration for all aspects of the game
- **Security Features**: Rate limiting, CORS protection, and message validation

## Architecture

### Backend (Go)
- **WebSocket Server**: Real-time communication using Gorilla WebSocket
- **Game Engine**: Thread-safe game state management and word validation
- **Room Management**: Multi-room support with player matchmaking
- **Dictionary System**: Efficient word validation and selection
- **REST API**: Room management and health check endpoints
- **Monitoring & Logging**: Structured logging with Sentry integration
- **Security & Rate Limiting**: Protection against abuse and DDoS
- **Configuration Management**: Environment-based configuration system

### Key Components
- `internal/game/`: Core game logic, types, and state management
- `internal/ws/`: WebSocket connection and message handling
- `internal/api/`: REST API endpoints and middleware
- `internal/room/`: Room management and cleanup services  
- `internal/config/`: Configuration loading and validation
- `internal/logging/`: Structured logging and Sentry integration
- `assets/words/`: Word dictionaries (common and valid words)
- `tests/`: Unit and integration test suites

## Quick Start

### Prerequisites
- Go 1.21.5 or later
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd worduel
   ```

2. **Install dependencies**
   ```bash
   cd backend
   go mod download
   ```

3. **Configure environment**
   ```bash
   # Option 1: Use the setup script (recommended)
   ./scripts/setup-env.sh
   
   # Option 2: Manual setup
   cp .env.example .env.development
   
   # Edit .env.development with your preferred settings
   # At minimum, you may want to adjust:
   # - ALLOWED_ORIGINS for your frontend URL
   # - LOG_LEVEL for more/less verbose logging
   # - SENTRY_DSN if you want error monitoring
   ```

4. **Run the server**
   ```bash
   # Development mode (uses .env.development if present)
   go run main.go
   
   # Or build and run
   go build -o worduel-backend
   ./worduel-backend
   ```

The server will start on port 8080 (or the PORT environment variable if set).

### Development

**Run tests**
```bash
go test ./...
```

**Build the project**
```bash
go build ./...
```

## API & WebSocket Messages

### Message Types
- `join`: Join a game room
- `leave`: Leave current room
- `guess`: Submit a word guess
- `game_update`: Receive game state updates
- `player_update`: Receive player status updates
- `error`: Error notifications
- `chat`: In-game messaging

### Game States
- `waiting`: Room waiting for players
- `active`: Game in progress
- `finished`: Game completed

## Configuration

Worduel backend supports comprehensive configuration through environment variables. Copy the appropriate environment file for your setup:

### Quick Setup

**For Development:**
```bash
cd backend
cp .env.example .env.development
# Edit .env.development with your settings
```

**For Production:**
```bash
cd backend
cp .env.example .env.production
# Edit .env.production with your production settings
```

### Environment Files

- **`.env.example`** - Template with all available configuration options
- **`.env.development`** - Development configuration with relaxed settings
- **`.env.production`** - Production configuration with strict security settings

### Configuration Categories

#### Server Configuration
- `PORT` - Server port (default: 8080)
- `HOST` - Server host (default: 0.0.0.0)
- `READ_TIMEOUT` - HTTP read timeout (default: 10s)
- `WRITE_TIMEOUT` - HTTP write timeout (default: 10s)
- `IDLE_TIMEOUT` - HTTP idle timeout (default: 60s)
- `SHUTDOWN_TIMEOUT` - Graceful shutdown timeout (default: 30s)

#### CORS Configuration
- `ALLOWED_ORIGINS` - Comma-separated list of allowed origins
- `ALLOWED_METHODS` - Comma-separated list of allowed HTTP methods
- `ALLOWED_HEADERS` - Comma-separated list of allowed headers

#### Rate Limiting
- `WS_RATE_LIMIT` - WebSocket messages per minute (default: 60)
- `API_RATE_LIMIT` - API requests per minute (default: 100)
- `MAX_CONNECTIONS_PER_IP` - Maximum connections per IP (default: 10)

#### Game & Room Management
- `MAX_CONCURRENT_ROOMS` - Maximum concurrent rooms (default: 1000)
- `ROOM_INACTIVE_TIMEOUT` - Room cleanup timeout (default: 30m)
- `GAME_TIMEOUT` - Game completion timeout (default: 30m)
- `CLEANUP_INTERVAL` - Room cleanup interval (default: 5m)
- `MAX_PLAYERS_PER_ROOM` - Maximum players per room (default: 2)
- `MAX_GUESSES` - Maximum guesses per game (default: 6)
- `WORD_LENGTH` - Word length (fixed: 5)

#### Security
- `VALIDATE_ORIGIN` - Enable origin validation (default: true)
- `MAX_MESSAGE_SIZE` - Maximum WebSocket message size in bytes (default: 1024)
- `CONNECTION_TIMEOUT` - WebSocket connection timeout (default: 30s)

#### Development & Debug
- `DEBUG_MODE` - Enable debug mode (default: false)
- `VERBOSE_LOG` - Enable verbose logging (default: false)
- `PROFILE_MODE` - Enable profiling (default: false)

### Logging & Monitoring

Worduel includes comprehensive logging and monitoring with structured logging and optional Sentry integration.

#### Logging Configuration
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: info)
- `ENVIRONMENT` - Environment name: development, staging, production
- `SERVICE_NAME` - Service identifier for logs (default: worduel-backend)
- `LOG_ADD_SOURCE` - Add source file/line to logs (default: false, useful for debugging)

#### Sentry Monitoring (Optional)
- `SENTRY_DSN` - Sentry Data Source Name (leave empty to disable)
- `SENTRY_ENVIRONMENT` - Environment name for Sentry
- `SENTRY_RELEASE` - Application version/release identifier
- `SENTRY_TRACES_SAMPLE_RATE` - Performance monitoring sample rate (0.0 to 1.0)
- `SENTRY_DEBUG` - Enable Sentry debug mode

**Benefits of Sentry Integration:**
- **Error Tracking**: Automatic error capture with stack traces
- **Performance Monitoring**: Request/response time tracking
- **Real-time Alerts**: Instant notifications for issues
- **Release Tracking**: Monitor issues across deployments
- **User Impact**: Understand which users are affected by issues

#### Log Formats

**Development**: Human-readable text format
```
2024-01-15 10:30:45 INFO Game event event_type=player_joined room_id=ABC123
```

**Production**: Structured JSON format
```json
{
  "time": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "Game event",
  "event_type": "player_joined",
  "room_id": "ABC123",
  "correlation_id": "req_123456"
}
```

### Production Deployment

For production deployment, follow these additional steps:

1. **Configure Production Environment**
   ```bash
   cd backend
   cp .env.production .env
   # Edit .env with your production settings
   ```

2. **Essential Production Settings**
   - Set `ALLOWED_ORIGINS` to your actual frontend domain(s)
   - Configure `SENTRY_DSN` for error monitoring
   - Set `LOG_LEVEL=info` or `LOG_LEVEL=warn` to reduce log volume
   - Enable `VALIDATE_ORIGIN=true` for security
   - Configure appropriate rate limits for your expected traffic

3. **Build for Production**
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worduel-backend .
   ```

4. **Health Checks**
   The server provides health check endpoints for container orchestration:
   - `GET /health` - Overall health status
   - `GET /health/liveness` - Liveness probe
   - `GET /health/readiness` - Readiness probe

## Game Rules

1. Players join a room and wait for the game to start
2. Each player tries to guess the same secret word
3. Players have a limited number of guesses (configurable)
4. First player to guess correctly wins the round
5. Letter feedback helps guide subsequent guesses
6. Scores are tracked across multiple rounds

## Contributing

This project follows conventional commit standards with emoji prefixes. See `CLAUDE.md` for detailed contribution guidelines.

### Development Commands
```bash
# Test all packages
go test ./...

# Build project
go build ./...

# Run with live reload (if using air)
air
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Raj Joshi

## Roadmap

- [ ] Frontend implementation
- [ ] User authentication
- [ ] Persistent game history
- [ ] Tournament modes
- [ ] Custom word lists
- [ ] Mobile app support
