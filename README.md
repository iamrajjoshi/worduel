# Worduel 🎯

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

## Architecture

### Backend (Go)
- **WebSocket Server**: Real-time communication using Gorilla WebSocket
- **Game Engine**: Thread-safe game state management and word validation
- **Room Management**: Multi-room support with player matchmaking
- **Dictionary System**: Efficient word validation and selection
- **CORS Support**: Configured for frontend integration

### Key Components
- `internal/game/`: Core game logic, types, and state management
- `internal/handlers/`: HTTP request handlers
- `internal/websocket/`: WebSocket connection management
- `assets/words/`: Word dictionaries (common and valid words)

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

3. **Run the server**
   ```bash
   go run main.go
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

### Environment Variables
- `PORT`: Server port (default: 8080)

### CORS Settings
Currently configured for `http://localhost:3000` - modify in `main.go` for production deployment.

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