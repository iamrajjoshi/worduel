# Structure Steering

## Project Organization

### Directory Structure

```
worduel/
├── README.md                    # Project overview and quick start
├── Dockerfile                   # Single-stage container build  
├── docker-compose.yml          # Development environment setup
├── .gitignore                  # Go and Node.js ignore patterns
├── .github/                    # GitHub Actions workflows
│   └── workflows/
│       ├── ci.yml              # Build, test, and security scanning
│       └── release.yml         # Docker image publishing
│
├── backend/                    # Go application root
│   ├── main.go                 # Application entry point
│   ├── go.mod                  # Go module definition
│   ├── go.sum                  # Dependency checksums
│   ├── embed.go                # Static file embedding
│   │
│   ├── internal/               # Private application packages
│   │   ├── game/               # Core game logic
│   │   │   ├── state.go        # Game state management
│   │   │   ├── logic.go        # Word validation and scoring
│   │   │   ├── dictionary.go   # Word list management
│   │   │   └── types.go        # Game data structures
│   │   │
│   │   ├── room/               # Room management
│   │   │   ├── manager.go      # Room lifecycle management
│   │   │   ├── storage.go      # In-memory room storage
│   │   │   └── cleanup.go      # Automatic room cleanup
│   │   │
│   │   ├── ws/                 # WebSocket handling
│   │   │   ├── hub.go          # Connection hub and broadcasting
│   │   │   ├── client.go       # Individual client management
│   │   │   ├── messages.go     # Message type definitions
│   │   │   └── handlers.go     # Message routing and handling
│   │   │
│   │   ├── api/                # REST API handlers
│   │   │   ├── rooms.go        # Room creation and info endpoints
│   │   │   ├── health.go       # Health check endpoint
│   │   │   └── middleware.go   # CORS, logging, rate limiting
│   │   │
│   │   └── config/             # Configuration management
│   │       ├── config.go       # Environment variable handling
│   │       └── defaults.go     # Default configuration values
│   │
│   ├── assets/                 # Embedded static assets
│   │   └── words/              # Dictionary files
│   │       ├── common.txt      # Most common 5-letter words
│   │       ├── valid.txt       # All valid guess words
│   │       └── excluded.txt    # Inappropriate words to exclude
│   │
│   ├── scripts/                # Development and build scripts
│   │   ├── build.sh            # Cross-platform build script
│   │   ├── test.sh             # Comprehensive test runner
│   │   └── lint.sh             # Code quality checks
│   │
│   └── tests/                  # Go test files
│       ├── integration/        # Integration tests
│       ├── unit/              # Unit tests
│       └── mocks/             # Test mocks and fixtures
│
├── frontend/                   # React application
│   ├── package.json           # NPM dependencies and scripts
│   ├── package-lock.json      # Dependency lock file
│   ├── tsconfig.json          # TypeScript configuration
│   ├── vite.config.ts         # Vite build configuration
│   ├── index.html             # Entry HTML file
│   │
│   ├── src/                   # Source code
│   │   ├── main.tsx           # React application entry point
│   │   ├── App.tsx            # Root application component
│   │   ├── index.css          # Global styles and CSS variables
│   │   │
│   │   ├── components/        # Reusable UI components
│   │   │   ├── common/        # Generic components
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Input.tsx
│   │   │   │   ├── Modal.tsx
│   │   │   │   └── LoadingSpinner.tsx
│   │   │   │
│   │   │   ├── game/          # Game-specific components
│   │   │   │   ├── GameBoard.tsx      # Main game grid
│   │   │   │   ├── GuessRow.tsx       # Individual guess row
│   │   │   │   ├── LetterTile.tsx     # Single letter display
│   │   │   │   ├── Keyboard.tsx       # Virtual keyboard
│   │   │   │   ├── GameStatus.tsx     # Win/lose display
│   │   │   │   └── OpponentProgress.tsx # Opponent state
│   │   │   │
│   │   │   └── room/          # Room management components
│   │   │       ├── RoomCreate.tsx     # Room creation form
│   │   │       ├── RoomJoin.tsx       # Room joining form
│   │   │       ├── RoomLobby.tsx      # Pre-game waiting area
│   │   │       └── RoomCode.tsx       # Room code display/sharing
│   │   │
│   │   ├── hooks/             # Custom React hooks
│   │   │   ├── useWebSocket.ts        # WebSocket connection management
│   │   │   ├── useGameState.ts        # Game state management
│   │   │   ├── useKeyboard.ts         # Keyboard input handling
│   │   │   ├── useLocalStorage.ts     # Browser storage utilities
│   │   │   └── useResponsive.ts       # Responsive design helpers
│   │   │
│   │   ├── contexts/          # React context providers
│   │   │   ├── GameContext.tsx        # Game state and actions
│   │   │   ├── SocketContext.tsx      # WebSocket connection
│   │   │   └── ThemeContext.tsx       # UI theme management
│   │   │
│   │   ├── types/             # TypeScript type definitions
│   │   │   ├── game.ts               # Game-related types
│   │   │   ├── socket.ts             # WebSocket message types
│   │   │   ├── room.ts               # Room-related types
│   │   │   └── api.ts                # REST API types
│   │   │
│   │   ├── utils/             # Utility functions
│   │   │   ├── gameLogic.ts          # Client-side game logic
│   │   │   ├── validation.ts         # Input validation
│   │   │   ├── formatting.ts         # Display formatting
│   │   │   └── constants.ts          # Application constants
│   │   │
│   │   ├── styles/            # Styling files
│   │   │   ├── globals.css           # Global styles
│   │   │   ├── components.css        # Component-specific styles
│   │   │   ├── animations.css        # CSS animations
│   │   │   └── responsive.css        # Media query definitions
│   │   │
│   │   └── assets/            # Static frontend assets
│   │       ├── icons/                # SVG icons
│   │       ├── sounds/               # Audio files (optional)
│   │       └── images/               # Images and graphics
│   │
│   ├── public/               # Static public assets
│   │   ├── favicon.ico       # Browser favicon
│   │   ├── manifest.json     # PWA manifest
│   │   └── robots.txt        # Search engine directives
│   │
│   └── tests/               # Frontend tests
│       ├── components/       # Component tests
│       ├── hooks/           # Custom hook tests
│       ├── utils/           # Utility function tests
│       └── e2e/             # End-to-end tests
│
├── docs/                    # Documentation
│   ├── api/                 # API documentation
│   │   ├── rest-api.md      # REST endpoint documentation
│   │   └── websocket-api.md # WebSocket message documentation
│   │
│   ├── deployment/          # Deployment guides
│   │   ├── docker.md        # Docker deployment guide
│   │   ├── hosting.md       # Hosting platform guides
│   │   └── troubleshooting.md # Common deployment issues
│   │
│   ├── development/         # Development documentation
│   │   ├── setup.md         # Local development setup
│   │   ├── architecture.md  # System architecture overview
│   │   ├── contributing.md  # Contribution guidelines
│   │   └── testing.md       # Testing strategy and guide
│   │
│   └── user/               # User-facing documentation
│       ├── gameplay.md      # How to play guide
│       ├── hosting.md       # How to host your own instance
│       └── faq.md          # Frequently asked questions
│
└── .spec-workflow/         # Spec-driven development artifacts
    ├── steering/           # Steering documents (this file)
    ├── specs/              # Feature specifications
    └── bugs/              # Bug fix specifications
```

### File Naming Conventions

**Go Files**
- `snake_case.go` for file names
- `PascalCase` for exported functions and types
- `camelCase` for unexported functions and variables
- `SCREAMING_SNAKE_CASE` for constants

**TypeScript/React Files**
- `PascalCase.tsx` for React components
- `camelCase.ts` for utilities and hooks
- `kebab-case.css` for stylesheets
- `camelCase` for variables and functions
- `PascalCase` for types and interfaces

**Documentation Files**
- `kebab-case.md` for all documentation
- Clear, descriptive names (e.g., `rest-api.md`, `docker-deployment.md`)

**Configuration Files**
- Use conventional names where possible (`package.json`, `Dockerfile`)
- `kebab-case` for custom configuration files

### Module Organization

**Go Package Structure**
- `internal/` for private packages not meant for external use
- One responsibility per package
- Circular dependencies forbidden
- Interface definitions in using package, implementations in separate packages

**React Component Organization**
- One component per file
- Barrel exports in `index.ts` files for clean imports
- Co-locate tests next to source files (e.g., `Button.tsx` and `Button.test.tsx`)
- Separate concerns: presentation components vs. container components

### Configuration Management

**Environment Variables (Backend)**
```go
// config/config.go
type Config struct {
    Port          string `env:"PORT" default:"8080"`
    Debug         bool   `env:"DEBUG" default:"false"`
    MaxRooms      int    `env:"MAX_ROOMS" default:"1000"`
    RoomTimeout   int    `env:"ROOM_TIMEOUT_MINUTES" default:"30"`
    RateLimit     int    `env:"RATE_LIMIT_PER_MINUTE" default:"60"`
    LogLevel      string `env:"LOG_LEVEL" default:"info"`
}
```

**Build-time Configuration (Frontend)**
```typescript
// vite.config.ts
const config = {
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  }
};
```

## Development Workflow

### Git Branching Strategy

**Branch Types**
- `main`: Production-ready code, protected branch
- `develop`: Integration branch for features (if using GitFlow)
- `feature/feature-name`: Individual feature development
- `bugfix/bug-description`: Bug fixes
- `hotfix/critical-fix`: Emergency production fixes

**Naming Conventions**
```
feature/multiplayer-rooms
feature/opponent-progress-display
bugfix/websocket-reconnection
hotfix/memory-leak-game-cleanup
```

**Commit Message Format**
```
emoji(scope): brief description

Longer explanation if needed

Fixes #issue-number
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
Emoji: `✨`, `🐛`, `📝`, `♻️`, `🧪`, `🧹`

### Code Review Process

**Pull Request Requirements**
- All tests pass
- Code coverage maintained above 80%
- Documentation updated for API changes
- Security scan passes
- Performance impact assessed

**Review Checklist**
- [ ] Code follows established conventions
- [ ] Tests adequately cover new functionality
- [ ] No hardcoded secrets or configuration
- [ ] Error handling is appropriate
- [ ] Performance implications considered
- [ ] Documentation updated

**Approval Requirements**
- At least one approval from code owner
- All automated checks pass
- No unresolved conversations

### Testing Workflow

**Backend Testing Strategy**
```bash
# Run all tests with coverage
./scripts/test.sh

# Unit tests only
go test ./internal/...

# Integration tests
go test ./tests/integration/...

# Benchmarks
go test -bench=. ./internal/game/
```

**Frontend Testing Strategy**
```bash
# Unit and integration tests
npm test

# Component tests with visual regression
npm run test:components

# End-to-end tests
npm run test:e2e

# Type checking
npm run type-check
```

**Test Coverage Requirements**
- Backend: Minimum 80% line coverage
- Frontend: Minimum 75% line coverage
- Critical paths (game logic): 95% coverage
- Integration tests for all API endpoints

### Deployment Process

**Development Deployment**
1. Local development with hot reload
2. Docker Compose for integration testing
3. Feature branch deployment to staging (optional)

**Production Deployment**
1. Code review and merge to main
2. Automated CI/CD pipeline triggers
3. Docker image build and security scan
4. Automated testing in staging environment
5. Manual approval for production deployment
6. Zero-downtime deployment with health checks

**Release Process**
```bash
# Create release branch
git checkout -b release/v1.0.0

# Update version numbers
./scripts/update-version.sh 1.0.0

# Create release notes
./scripts/generate-release-notes.sh

# Tag and push
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## Documentation Structure

### Where to Find What

**For Developers**
- **Quick Start**: `README.md` in project root
- **Architecture**: `docs/development/architecture.md`
- **API Reference**: `docs/api/` directory
- **Contributing**: `docs/development/contributing.md`
- **Local Setup**: `docs/development/setup.md`

**For Deployers**
- **Docker Guide**: `docs/deployment/docker.md`
- **Hosting Options**: `docs/deployment/hosting.md`
- **Troubleshooting**: `docs/deployment/troubleshooting.md`
- **Environment Variables**: `README.md` configuration section

**For Users**
- **How to Play**: `docs/user/gameplay.md`
- **Hosting Guide**: `docs/user/hosting.md`
- **FAQ**: `docs/user/faq.md`

### How to Update Documentation

**Documentation Principles**
- Keep documentation close to code
- Update docs in same PR as code changes
- Use examples and diagrams where helpful
- Maintain consistency in formatting and style

**Documentation Workflow**
1. Identify what documentation needs updating
2. Update relevant files in same branch as code changes
3. Include documentation changes in PR description
4. Review documentation changes as part of code review
5. Validate documentation by following instructions

**API Documentation Standards**
```go
// CreateRoom creates a new multiplayer game room
//
// POST /api/rooms
// Request: {} (empty body)
// Response: {
//   "roomID": "ABC123",
//   "created": "2023-10-01T12:00:00Z",
//   "status": "waiting"
// }
//
// Returns 201 on success, 500 on server error
func (h *APIHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### Spec Organization

**Specification Structure (via spec-workflow)**
- **Requirements**: User needs and acceptance criteria
- **Design**: Technical approach and architecture decisions
- **Tasks**: Implementation breakdown and progress tracking

**Bug Tracking Process**
- **Investigation**: Reproduction steps and impact assessment
- **Analysis**: Root cause analysis and fix approach
- **Implementation**: Code changes and testing approach
- **Verification**: Fix validation and regression prevention

### Knowledge Sharing

**Internal Documentation**
- **Architecture Decision Records (ADRs)**: Major technical decisions
- **Runbooks**: Operational procedures and troubleshooting
- **Code Comments**: Complex logic explanation and context
- **Inline Documentation**: API and function documentation

**External Documentation**
- **README**: Project overview and quick start
- **User Guides**: End-user documentation
- **API Documentation**: Developer integration guides
- **Deployment Guides**: Hosting and operational documentation

## Team Conventions

### Communication Guidelines

**Development Communication**
- **Code Review**: Constructive, specific feedback in PRs
- **Issue Tracking**: Clear, actionable issues with reproduction steps
- **Documentation**: Update docs with code changes
- **Questions**: Ask in appropriate channels with context

**Project Coordination**
- **Standup Updates**: Progress, blockers, and next steps
- **Sprint Planning**: Feature prioritization and capacity planning
- **Retrospectives**: Process improvements and lessons learned
- **Architecture Reviews**: Technical decision documentation

### Meeting Structure

**Weekly Planning**
- Review current sprint progress
- Plan upcoming work and priorities
- Address blockers and dependencies
- Align on technical decisions

**Code Review Sessions**
- Review complex or critical PRs together
- Discuss architecture and design decisions
- Share knowledge and best practices
- Ensure security and performance standards

**Retrospectives**
- What went well
- What could be improved
- Action items for process improvements
- Technical debt prioritization

### Decision-Making Process

**Technical Decisions**
1. **Proposal**: Document the decision with context and options
2. **Discussion**: Team review with pros/cons analysis
3. **Decision**: Consensus or lead engineer decision
4. **Documentation**: Record decision in ADR or steering docs
5. **Implementation**: Execute decision with monitoring

**Product Decisions**
1. **User Research**: Understand user needs and pain points
2. **Options Analysis**: Evaluate different approaches
3. **Impact Assessment**: Consider development effort and user value
4. **Decision**: Product owner decision with team input
5. **Implementation**: Execute with user feedback collection

This structure steering document provides the organizational foundation for developing a maintainable, scalable multiplayer Wordle game while ensuring consistency across team members and clear documentation for future contributors and operators.