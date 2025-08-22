#!/bin/bash

# Worduel Backend Development Utilities
# Development setup and debugging utilities

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEV_PORT=${DEV_PORT:-8080}
DEV_ENV=${DEV_ENV:-"development"}

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

show_help() {
    echo "Worduel Backend Development Utilities"
    echo ""
    echo "Usage: $0 COMMAND [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  setup           Set up development environment"
    echo "  run             Run the application in development mode"
    echo "  watch           Run with auto-reload on file changes"
    echo "  debug           Run with debugging enabled"
    echo "  test-quick      Run quick test suite"
    echo "  deps            Install/update dependencies"
    echo "  clean           Clean build artifacts and caches"
    echo "  profile         Run with profiling enabled"
    echo "  benchmark       Run performance benchmarks"
    echo "  health          Check application health"
    echo "  logs            Show application logs"
    echo "  reset           Reset development environment"
    echo "  help            Show this help message"
    echo ""
    echo "Options:"
    echo "  --port PORT     Development port [default: 8080]"
    echo "  --env ENV       Environment [default: development]"
    echo "  --verbose       Verbose output"
    echo ""
    echo "Examples:"
    echo "  $0 setup               # Initial development setup"
    echo "  $0 run                 # Run in development mode"
    echo "  $0 watch --verbose     # Watch mode with verbose output"
    echo "  $0 debug --port 9090   # Debug on port 9090"
    echo "  $0 test-quick          # Quick test run"
    echo ""
    echo "Environment Variables:"
    echo "  DEV_PORT    Development port (default: 8080)"
    echo "  DEV_ENV     Development environment (default: development)"
    echo ""
}

# Check if required tools are installed
check_requirements() {
    local missing_tools=()
    
    if ! command -v go >/dev/null 2>&1; then
        missing_tools+=("go")
    fi
    
    if ! command -v git >/dev/null 2>&1; then
        missing_tools+=("git")
    fi
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools and try again"
        exit 1
    fi
}

# Set up development environment
setup_dev_environment() {
    log_info "Setting up Worduel Backend development environment..."
    
    # Check requirements
    check_requirements
    
    # Ensure we're in the backend directory
    if [ ! -f "main.go" ]; then
        log_error "main.go not found. Please run this script from the backend directory."
        exit 1
    fi
    
    # Create necessary directories
    log_info "Creating development directories..."
    mkdir -p logs build coverage tmp
    
    # Install Go dependencies
    log_info "Installing Go dependencies..."
    go mod download
    go mod tidy
    
    # Set up environment files
    if [ ! -f ".env.development" ]; then
        log_info "Creating .env.development file..."
        if [ -f ".env.example" ]; then
            cp .env.example .env.development
            log_info "Copied .env.example to .env.development"
            log_warning "Please review and customize .env.development for your setup"
        else
            log_warning ".env.example not found. You may need to create .env.development manually"
        fi
    else
        log_info ".env.development already exists"
    fi
    
    # Run environment setup script if it exists
    if [ -f "scripts/setup-env.sh" ]; then
        log_info "Running environment setup script..."
        bash scripts/setup-env.sh
    fi
    
    # Install development tools
    log_info "Installing development tools..."
    
    # Install air for hot reloading if not present
    if ! command -v air >/dev/null 2>&1; then
        log_info "Installing air (hot reload tool)..."
        go install github.com/cosmtrek/air@latest
    fi
    
    # Install pprof for profiling if not present
    if ! command -v pprof >/dev/null 2>&1; then
        log_info "Installing pprof (profiling tool)..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    fi
    
    # Create .air.toml if it doesn't exist
    if [ ! -f ".air.toml" ]; then
        log_info "Creating .air.toml configuration..."
        cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/worduel-backend"
  cmd = "go build -o ./tmp/worduel-backend ./main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "build", "coverage"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF
        log_success "Created .air.toml configuration"
    fi
    
    log_success "Development environment setup completed!"
    
    echo ""
    log_info "Next steps:"
    echo "  1. Review .env.development file"
    echo "  2. Run '$0 run' to start the application"
    echo "  3. Run '$0 watch' for hot reload during development"
    echo "  4. Run '$0 test-quick' to run tests"
}

# Run application in development mode
run_dev() {
    log_info "Starting Worduel Backend in development mode..."
    log_info "Port: $DEV_PORT"
    log_info "Environment: $DEV_ENV"
    
    # Set environment variables
    export PORT=$DEV_PORT
    export ENVIRONMENT=$DEV_ENV
    export LOG_LEVEL=debug
    export DEBUG_MODE=true
    
    # Load environment file if it exists
    if [ -f ".env.$DEV_ENV" ]; then
        log_info "Loading environment from .env.$DEV_ENV"
        set -a
        source ".env.$DEV_ENV"
        set +a
    fi
    
    log_info "Starting application..."
    go run main.go
}

# Run with hot reload
run_watch() {
    if ! command -v air >/dev/null 2>&1; then
        log_error "air is not installed. Run '$0 setup' to install development tools."
        exit 1
    fi
    
    log_info "Starting Worduel Backend with hot reload..."
    log_info "Port: $DEV_PORT"
    log_info "Environment: $DEV_ENV"
    
    # Set environment variables
    export PORT=$DEV_PORT
    export ENVIRONMENT=$DEV_ENV
    export LOG_LEVEL=debug
    export DEBUG_MODE=true
    
    # Load environment file if it exists
    if [ -f ".env.$DEV_ENV" ]; then
        log_info "Loading environment from .env.$DEV_ENV"
        set -a
        source ".env.$DEV_ENV"
        set +a
    fi
    
    log_info "Starting with hot reload (Ctrl+C to stop)..."
    air
}

# Run with debugging
run_debug() {
    log_info "Starting Worduel Backend with debugging enabled..."
    
    # Check if delve is installed
    if ! command -v dlv >/dev/null 2>&1; then
        log_info "Installing delve debugger..."
        go install github.com/go-delve/delve/cmd/dlv@latest
    fi
    
    export PORT=$DEV_PORT
    export ENVIRONMENT=$DEV_ENV
    export LOG_LEVEL=debug
    export DEBUG_MODE=true
    
    log_info "Starting debugger on port 40000..."
    log_info "Connect your IDE or run: dlv connect :40000"
    
    dlv debug --headless --listen=:40000 --api-version=2 main.go
}

# Run quick tests
run_quick_tests() {
    log_info "Running quick test suite..."
    
    # Run unit tests only, skip integration tests
    if [ -f "scripts/test.sh" ]; then
        ./scripts/test.sh --unit --timeout 10s
    else
        go test -short -timeout 10s ./internal/...
    fi
}

# Install/update dependencies
update_dependencies() {
    log_info "Updating Go dependencies..."
    
    go mod download
    go mod tidy
    go mod verify
    
    log_success "Dependencies updated successfully"
    
    # Show dependency information
    log_info "Current dependencies:"
    go list -m -versions all | head -10
}

# Clean build artifacts
clean_environment() {
    log_info "Cleaning build artifacts and caches..."
    
    # Remove build directories
    rm -rf build tmp coverage logs
    
    # Clean Go cache
    go clean -cache -testcache -modcache
    
    # Remove temporary files
    find . -name "*.log" -type f -delete
    find . -name "*.tmp" -type f -delete
    find . -name "*.prof" -type f -delete
    
    log_success "Environment cleaned successfully"
}

# Run with profiling
run_profile() {
    log_info "Starting application with profiling enabled..."
    
    export PORT=$DEV_PORT
    export ENVIRONMENT=$DEV_ENV
    export LOG_LEVEL=info
    export PPROF_ENABLED=true
    
    log_info "Profiling endpoints available at:"
    echo "  http://localhost:$DEV_PORT/debug/pprof/"
    echo "  http://localhost:$DEV_PORT/debug/pprof/heap"
    echo "  http://localhost:$DEV_PORT/debug/pprof/profile"
    
    go run main.go
}

# Run benchmarks
run_benchmarks() {
    log_info "Running performance benchmarks..."
    
    if [ -f "scripts/test.sh" ]; then
        ./scripts/test.sh --benchmark
    else
        go test -bench=. -benchmem ./...
    fi
}

# Check application health
check_health() {
    local url="http://localhost:$DEV_PORT/health"
    
    log_info "Checking application health at $url..."
    
    if command -v curl >/dev/null 2>&1; then
        if curl -f "$url" >/dev/null 2>&1; then
            log_success "Application is healthy"
            curl -s "$url" | jq . 2>/dev/null || curl -s "$url"
        else
            log_error "Application health check failed"
            log_info "Make sure the application is running with '$0 run'"
        fi
    else
        log_warning "curl not available. Please check $url manually"
    fi
}

# Show application logs
show_logs() {
    log_info "Showing recent application logs..."
    
    if [ -d "logs" ]; then
        find logs -name "*.log" -type f -exec tail -n 50 {} \;
    else
        log_info "No log directory found. Logs may be printed to stdout."
    fi
}

# Reset development environment
reset_environment() {
    log_warning "This will reset your development environment!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Resetting development environment..."
        
        # Clean everything
        clean_environment
        
        # Remove environment files (but keep examples)
        rm -f .env.development .env.production
        
        # Remove air config
        rm -f .air.toml
        
        # Re-setup
        setup_dev_environment
        
        log_success "Development environment reset completed"
    else
        log_info "Reset cancelled"
    fi
}

# Parse command line arguments
COMMAND=""
VERBOSE="false"

while [[ $# -gt 0 ]]; do
    case $1 in
        --port)
            DEV_PORT="$2"
            shift 2
            ;;
        --env)
            DEV_ENV="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE="true"
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        -*)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
        *)
            if [ -z "$COMMAND" ]; then
                COMMAND="$1"
            else
                log_error "Multiple commands specified"
                show_help
                exit 1
            fi
            shift
            ;;
    esac
done

# Default command
if [ -z "$COMMAND" ]; then
    COMMAND="help"
fi

# Execute command
case $COMMAND in
    setup)
        setup_dev_environment
        ;;
    run)
        run_dev
        ;;
    watch)
        run_watch
        ;;
    debug)
        run_debug
        ;;
    test-quick)
        run_quick_tests
        ;;
    deps)
        update_dependencies
        ;;
    clean)
        clean_environment
        ;;
    profile)
        run_profile
        ;;
    benchmark)
        run_benchmarks
        ;;
    health)
        check_health
        ;;
    logs)
        show_logs
        ;;
    reset)
        reset_environment
        ;;
    help)
        show_help
        ;;
    *)
        log_error "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac