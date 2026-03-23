# Worduel

![Tests](https://github.com/iamrajjoshi/worduel/actions/workflows/test.yml/badge.svg?branch=main)

A real-time multiplayer Wordle game. Create a room, share the code with a friend, and race to guess the word first.

## Quick Start

### Docker (recommended)

```bash
docker compose up --build
```

Open [http://localhost:8080](http://localhost:8080) and play.

### Local Development

**Prerequisites:** Go 1.21+, Node 20+, pnpm

```bash
# Terminal 1: Backend
cd backend
go run main.go

# Terminal 2: Frontend
cd frontend
pnpm install
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000). The Vite dev server proxies API/WebSocket requests to the backend.

## How to Play

1. Enter your name and create a game
2. Share the 6-character room code with a friend
3. Once both players join, the game starts automatically
4. Guess the 5-letter word in 6 tries
5. First player to guess correctly wins

Feedback colors:
- **Green** -- correct letter, correct position
- **Yellow** -- correct letter, wrong position
- **Gray** -- letter not in the word

You can see your opponent's color patterns (but not their words) in real time.

## Architecture

```
worduel/
  backend/         Go server (WebSocket + REST + static file serving)
  frontend/        React + TypeScript + Vite + Tailwind CSS
  Dockerfile       Multi-stage build (single container)
  docker-compose.yml
```

### Backend (Go)
- WebSocket server for real-time gameplay (Gorilla WebSocket)
- REST API for room creation (`POST /api/rooms`, `GET /api/rooms/{id}`)
- Thread-safe game state with room management
- Curated word dictionary (~2,300 common words, ~10,000 valid guesses)
- Health check endpoints (`/health`, `/health/liveness`, `/health/readiness`)
- Serves the frontend build as static files in production

### Frontend (React + TypeScript)
- Vite build with Tailwind CSS v4
- WebSocket hook with auto-reconnection (exponential backoff)
- Game state via React Context + `useReducer`
- Framer Motion animations (tile flips, row shakes)
- On-screen + physical keyboard with color tracking
- Responsive layout with opponent progress sidebar

## Deploy to Production

The game is designed to run as a single container. The recommended setup is **Railway** (hosting) + **Cloudflare** (DNS).

### Railway + Cloudflare DNS

1. Push this repo to GitHub
2. Go to [railway.app](https://railway.app) and create a new project from the repo
3. Railway auto-detects the `Dockerfile` and deploys
4. In Railway project settings, add your custom domain (e.g. `worduel.xyz`)
5. Railway gives you a CNAME target (e.g. `xyz.up.railway.app`)
6. In Cloudflare DNS for your domain, add:
   - **CNAME** `@` -> the Railway target (set proxy status to **DNS only** / gray cloud)
   - **CNAME** `www` -> `worduel.xyz`
7. Set these env vars in Railway:
   ```
   ALLOWED_ORIGINS=https://worduel.xyz,https://www.worduel.xyz
   VALIDATE_ORIGIN=true
   LOG_LEVEL=info
   ```

> **Note:** Use "DNS only" (gray cloud) in Cloudflare for the CNAME record. Cloudflare's proxy can interfere with WebSocket connections on the free plan. Railway handles SSL automatically.

### Generic Docker Deploy

The image works anywhere that runs containers (Fly.io, DigitalOcean, AWS ECS, etc.):

```bash
docker build -t worduel .
docker run -p 8080:8080 -e ALLOWED_ORIGINS=https://yourdomain.com worduel
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `ALLOWED_ORIGINS` | `localhost` | CORS origins (comma-separated) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `MAX_PLAYERS_PER_ROOM` | `2` | Players per room |
| `MAX_GUESSES` | `6` | Guesses per game |
| `VALIDATE_ORIGIN` | `true` | Enforce origin checking |

See `backend/.env.example` for the full list.

## Development

```bash
# Backend tests
cd backend && go test ./...

# Frontend type check + build
cd frontend && pnpm build
```

## License

MIT -- see [LICENSE](LICENSE).

Copyright (c) 2025 Raj Joshi
