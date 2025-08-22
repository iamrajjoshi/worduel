# Worduel Backend - Development Scripts

This directory contains comprehensive development and build scripts for the Worduel Backend project.

## Quick Start

```bash
# Initial setup
./scripts/worduel.sh setup

# Development with hot reload
./scripts/worduel.sh watch

# Run tests with coverage
./scripts/worduel.sh test-coverage

# Run full CI pipeline locally
./scripts/worduel.sh ci
```

## Script Overview

### 🎯 Master Script - `worduel.sh`
The unified interface for all development tasks.

```bash
./scripts/worduel.sh COMMAND [OPTIONS]
```

**Key Commands:**
- `setup` - Initial development environment setup
- `watch` - Development with hot reload
- `test-coverage` - Run tests with coverage reporting
- `ci` - Run full CI pipeline locally
- `pre-commit` - Quick pre-commit validation
- `release` - Prepare for release

### 🔨 Build Script - `build.sh`
Cross-platform build script with optimization flags.

```bash
./scripts/build.sh [OPTIONS]
```

**Features:**
- Cross-compilation for multiple platforms
- Optimized builds with symbol stripping
- Version information embedding
- Race detection support
- Build artifacts organization

**Examples:**
```bash
# Build for current platform
./scripts/build.sh

# Cross-compile for all platforms
./scripts/build.sh --cross

# Development build with race detection
./scripts/build.sh --race --verbose
```

### 🧪 Test Script - `test.sh`
Comprehensive test runner with coverage reporting.

```bash
./scripts/test.sh [OPTIONS] [PATTERN]
```

**Features:**
- Unit and integration test separation
- Coverage reporting with HTML output
- Benchmark testing
- Race detection
- Watch mode for development
- Coverage threshold validation

**Examples:**
```bash
# Run all tests
./scripts/test.sh

# Unit tests with coverage
./scripts/test.sh --unit --coverage

# Integration tests only
./scripts/test.sh --integration

# Watch mode for development
./scripts/test.sh --watch

# Specific test pattern
./scripts/test.sh TestRoomCreation
```

### 🔍 Lint Script - `lint.sh`
Code quality checks and static analysis.

```bash
./scripts/lint.sh [OPTIONS]
```

**Features:**
- Comprehensive Go linting with golangci-lint
- Security scanning with gosec
- Static analysis with staticcheck
- Code formatting validation
- Auto-fix capabilities
- Configurable output formats

**Examples:**
```bash
# Run all checks
./scripts/lint.sh

# Fix auto-fixable issues
./scripts/lint.sh --fix

# Skip security checks
./scripts/lint.sh --no-security

# JSON output format
./scripts/lint.sh --format json
```

### 🚀 Development Script - `dev.sh`
Development utilities and environment management.

```bash
./scripts/dev.sh COMMAND [OPTIONS]
```

**Commands:**
- `setup` - Set up development environment
- `run` - Run in development mode
- `watch` - Hot reload development
- `debug` - Run with debugger
- `test-quick` - Quick test suite
- `profile` - Run with profiling
- `health` - Check application health
- `clean` - Clean build artifacts

**Examples:**
```bash
# Initial setup
./scripts/dev.sh setup

# Hot reload development
./scripts/dev.sh watch

# Debug mode
./scripts/dev.sh debug --port 9090
```

### 🐳 Docker Build Script - `docker-build.sh`
Docker image building and testing.

```bash
./scripts/docker-build.sh [OPTIONS]
```

**Features:**
- Multi-stage Docker builds
- Production and debug image variants
- Multi-platform support
- Image testing and validation
- Registry push capabilities

**Examples:**
```bash
# Build production image
./scripts/docker-build.sh

# Build debug image with tests
./scripts/docker-build.sh --target runtime-debug --test

# Build and push to registry
./scripts/docker-build.sh --push --registry your-registry.com
```

### 🌍 Environment Setup Script - `setup-env.sh`
Interactive environment configuration setup.

```bash
./scripts/setup-env.sh
```

**Features:**
- Interactive environment file creation
- Development and production templates
- Configuration guidance
- Environment-specific recommendations

## Development Workflow

### 1. Initial Setup
```bash
# Clone repository and navigate to backend
cd worduel/backend

# Run initial setup
./scripts/worduel.sh setup

# This will:
# - Install Go dependencies
# - Create development directories
# - Set up environment files
# - Install development tools (air, delve, etc.)
# - Create configuration files
```

### 2. Daily Development
```bash
# Start development with hot reload
./scripts/worduel.sh watch

# In another terminal, run tests
./scripts/worduel.sh test-unit --watch

# Quick health check
./scripts/worduel.sh health
```

### 3. Before Committing
```bash
# Run pre-commit checks
./scripts/worduel.sh pre-commit

# This includes:
# - Code formatting check
# - Linting
# - Quick tests
# - Build verification
```

### 4. Full Validation
```bash
# Run complete CI pipeline
./scripts/worduel.sh ci

# This includes:
# - Dependency check
# - Full linting
# - All tests with coverage
# - Build for current platform
# - Docker image build
```

### 5. Release Preparation
```bash
# Prepare for release
./scripts/worduel.sh release

# This includes:
# - Full CI pipeline
# - Cross-platform builds
# - Multiple Docker image variants
# - Git status verification
```

## Environment Configuration

### Development Environment
The scripts automatically create and use `.env.development` for local development:

```env
PORT=8080
LOG_LEVEL=debug
LOG_FORMAT=text
DEBUG_MODE=true
ALLOWED_ORIGINS=http://localhost:3000
SENTRY_ENABLED=false
```

### Production Environment
Production configuration in `.env.production`:

```env
PORT=8080
LOG_LEVEL=info
LOG_FORMAT=json
DEBUG_MODE=false
ALLOWED_ORIGINS=https://yourdomain.com
SENTRY_ENABLED=true
SENTRY_DSN=your-sentry-dsn
```

## Tool Requirements

### Required Tools
- **Go 1.21+** - Core language and toolchain
- **Git** - Version control
- **Docker** - Container builds (optional)

### Optional Development Tools
The setup script will install these automatically:

- **air** - Hot reload for development
- **delve (dlv)** - Go debugger
- **golangci-lint** - Comprehensive Go linter
- **gosec** - Security scanner
- **staticcheck** - Advanced static analysis
- **fswatch** - File system monitoring (for watch mode)

### Installation Commands
```bash
# Install required tools manually
go install github.com/cosmtrek/air@latest
go install github.com/go-delve/delve/cmd/dlv@latest
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.55.2
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@v2.18.2
go install honnef.co/go/tools/cmd/staticcheck@2023.1.6

# Or run the automated installation
./scripts/lint.sh --install-tools
```

## Output Directories

The scripts create and use the following directories:

```
backend/
├── build/          # Built binaries
├── coverage/       # Test coverage reports
├── logs/           # Application logs
├── tmp/            # Temporary files (hot reload)
└── scripts/        # Development scripts
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run CI Pipeline
  run: ./scripts/worduel.sh ci

- name: Build Docker Image  
  run: ./scripts/docker-build.sh --target production
```

### Local CI Simulation
```bash
# Simulate complete CI pipeline
./scripts/worduel.sh ci

# Quick pre-commit validation
./scripts/worduel.sh pre-commit
```

## Troubleshooting

### Common Issues

**1. Permission Denied**
```bash
chmod +x scripts/*.sh
```

**2. Missing Tools**
```bash
./scripts/lint.sh --install-tools
./scripts/dev.sh setup
```

**3. Port Already in Use**
```bash
./scripts/worduel.sh run --port 8081
```

**4. Tests Failing**
```bash
# Run specific tests
./scripts/test.sh TestRoomCreation --verbose

# Check application health
./scripts/worduel.sh health
```

**5. Build Failures**
```bash
# Clean and rebuild
./scripts/worduel.sh clean
./scripts/worduel.sh build --verbose
```

### Debug Mode
Most scripts support `--verbose` for detailed output:

```bash
./scripts/worduel.sh watch --verbose
./scripts/test.sh --unit --verbose
./scripts/build.sh --verbose
```

## Script Customization

### Environment Variables
```bash
# Customize development settings
export DEV_PORT=9090
export DEV_ENV=staging

# Run with custom settings
./scripts/worduel.sh watch
```

### Configuration Files
- `.air.toml` - Hot reload configuration
- `.env.development` - Development environment
- `.env.production` - Production environment

## Best Practices

1. **Always run setup first** on new environments
2. **Use watch mode** during active development
3. **Run pre-commit checks** before pushing
4. **Use CI pipeline** for comprehensive validation
5. **Keep environment files updated** with new configuration options
6. **Use specific test patterns** for faster feedback during development

## Support

For script-related issues:
1. Check this README
2. Run `./scripts/worduel.sh help`
3. Use `--verbose` flag for debugging
4. Check the individual script help: `./scripts/SCRIPT.sh --help`