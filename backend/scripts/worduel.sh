#!/bin/bash

# Worduel Backend - Master Development Script
# Unified interface for all development tasks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

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

show_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════╗"
    echo "║                                              ║"
    echo "║           🎯 Worduel Backend                 ║"
    echo "║         Development Toolkit                  ║"
    echo "║                                              ║"
    echo "╚══════════════════════════════════════════════╝"
    echo -e "${NC}"
}

show_help() {
    show_banner
    echo ""
    echo "Usage: $0 COMMAND [OPTIONS]"
    echo ""
    echo -e "${YELLOW}Development Commands:${NC}"
    echo "  setup           Set up development environment"
    echo "  run             Run application in development mode"
    echo "  watch           Run with hot reload"
    echo "  debug           Run with debugger attached"
    echo ""
    echo -e "${YELLOW}Build Commands:${NC}"
    echo "  build           Build the application"
    echo "  build-cross     Cross-compile for multiple platforms"
    echo "  docker          Build Docker image"
    echo "  docker-dev      Build and run Docker development image"
    echo ""
    echo -e "${YELLOW}Test Commands:${NC}"
    echo "  test            Run all tests"
    echo "  test-unit       Run unit tests only"
    echo "  test-integration Run integration tests only"
    echo "  test-coverage   Run tests with coverage report"
    echo "  benchmark       Run performance benchmarks"
    echo ""
    echo -e "${YELLOW}Quality Commands:${NC}"
    echo "  lint            Run linter and static analysis"
    echo "  lint-fix        Run linter and fix issues"
    echo "  format          Format code"
    echo ""
    echo -e "${YELLOW}Docker Commands:${NC}"
    echo "  docker-build    Build Docker image"
    echo "  docker-run      Run Docker container"
    echo "  docker-dev      Run Docker development container"
    echo ""
    echo -e "${YELLOW}Utility Commands:${NC}"
    echo "  clean           Clean build artifacts"
    echo "  deps            Update dependencies"
    echo "  health          Check application health"
    echo "  logs            Show application logs"
    echo "  env             Set up environment files"
    echo ""
    echo -e "${YELLOW}Workflow Commands:${NC}"
    echo "  ci              Run full CI pipeline locally"
    echo "  pre-commit      Run pre-commit checks"
    echo "  release         Prepare for release"
    echo ""
    echo "Examples:"
    echo "  $0 setup           # Initial setup"
    echo "  $0 watch           # Development with hot reload"
    echo "  $0 test-coverage   # Run tests with coverage"
    echo "  $0 ci              # Run full CI pipeline"
    echo "  $0 docker-dev      # Run in Docker"
    echo ""
    echo "For detailed help on any command, use: $0 COMMAND --help"
    echo ""
}

# Ensure we're in the backend directory
ensure_backend_directory() {
    if [ ! -f "$PROJECT_DIR/main.go" ]; then
        log_error "main.go not found. Please run this script from the backend directory or its parent."
        exit 1
    fi
    cd "$PROJECT_DIR"
}

# Run a script with error handling
run_script() {
    local script_name="$1"
    shift  # Remove script name from arguments
    
    local script_path="$SCRIPT_DIR/$script_name"
    
    if [ ! -f "$script_path" ]; then
        log_error "Script not found: $script_path"
        exit 1
    fi
    
    if [ ! -x "$script_path" ]; then
        chmod +x "$script_path"
    fi
    
    log_info "Running: $script_name $*"
    "$script_path" "$@"
}

# CI Pipeline - run all checks
run_ci_pipeline() {
    log_info "Running full CI pipeline..."
    echo ""
    
    local failed_steps=()
    
    # Step 1: Dependencies
    log_info "Step 1/6: Checking dependencies..."
    if ! run_script "dev.sh" "deps"; then
        failed_steps+=("dependencies")
    fi
    echo ""
    
    # Step 2: Linting
    log_info "Step 2/6: Running linter..."
    if ! run_script "lint.sh" "--fail-on-warning"; then
        failed_steps+=("linting")
    fi
    echo ""
    
    # Step 3: Unit tests
    log_info "Step 3/6: Running unit tests..."
    if ! run_script "test.sh" "--unit" "--coverage"; then
        failed_steps+=("unit-tests")
    fi
    echo ""
    
    # Step 4: Integration tests
    log_info "Step 4/6: Running integration tests..."
    if ! run_script "test.sh" "--integration"; then
        failed_steps+=("integration-tests")
    fi
    echo ""
    
    # Step 5: Build
    log_info "Step 5/6: Building application..."
    if ! run_script "build.sh"; then
        failed_steps+=("build")
    fi
    echo ""
    
    # Step 6: Docker build
    log_info "Step 6/6: Building Docker image..."
    if ! run_script "docker-build.sh" "--target" "production"; then
        failed_steps+=("docker-build")
    fi
    echo ""
    
    # Results
    if [ ${#failed_steps[@]} -eq 0 ]; then
        log_success "CI pipeline completed successfully! 🎉"
    else
        log_error "CI pipeline failed. Failed steps: ${failed_steps[*]}"
        exit 1
    fi
}

# Pre-commit checks
run_pre_commit() {
    log_info "Running pre-commit checks..."
    echo ""
    
    local failed_checks=()
    
    # Format check
    log_info "Checking code formatting..."
    if ! go fmt ./... > /tmp/fmt_output && [ -s /tmp/fmt_output ]; then
        log_error "Code is not properly formatted. Run: go fmt ./..."
        failed_checks+=("format")
    fi
    
    # Quick lint
    log_info "Running quick lint checks..."
    if ! run_script "lint.sh" "--no-security" "--fail-on-warning"; then
        failed_checks+=("lint")
    fi
    
    # Quick tests
    log_info "Running quick tests..."
    if ! run_script "test.sh" "--unit" "--timeout" "10s"; then
        failed_checks+=("tests")
    fi
    
    # Build check
    log_info "Checking build..."
    if ! go build -o /tmp/worduel-test ./main.go; then
        log_error "Build failed"
        failed_checks+=("build")
    else
        rm -f /tmp/worduel-test
    fi
    
    # Results
    if [ ${#failed_checks[@]} -eq 0 ]; then
        log_success "All pre-commit checks passed! ✨"
    else
        log_error "Pre-commit checks failed: ${failed_checks[*]}"
        exit 1
    fi
}

# Release preparation
prepare_release() {
    log_info "Preparing for release..."
    
    # Run full CI first
    run_ci_pipeline
    
    # Check git status
    if [ -n "$(git status --porcelain)" ]; then
        log_error "Working directory is not clean. Please commit your changes first."
        exit 1
    fi
    
    # Build for all platforms
    log_info "Building for all platforms..."
    run_script "build.sh" "--cross"
    
    # Build Docker images
    log_info "Building Docker images..."
    run_script "docker-build.sh" "--target" "production"
    run_script "docker-build.sh" "--target" "runtime-debug"
    
    log_success "Release preparation completed!"
    log_info "Ready to tag and publish the release"
}

# Parse command and execute
COMMAND=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_help
            exit 0
            ;;
        -*)
            # Pass through options to the underlying script
            break
            ;;
        *)
            if [ -z "$COMMAND" ]; then
                COMMAND="$1"
                shift
                break
            fi
            ;;
    esac
done

# Default command
if [ -z "$COMMAND" ]; then
    show_help
    exit 0
fi

# Ensure we're in the right directory
ensure_backend_directory

# Execute commands
case $COMMAND in
    # Development commands
    setup)
        run_script "dev.sh" "setup" "$@"
        ;;
    run)
        run_script "dev.sh" "run" "$@"
        ;;
    watch)
        run_script "dev.sh" "watch" "$@"
        ;;
    debug)
        run_script "dev.sh" "debug" "$@"
        ;;
    
    # Build commands
    build)
        run_script "build.sh" "$@"
        ;;
    build-cross)
        run_script "build.sh" "--cross" "$@"
        ;;
    
    # Test commands
    test)
        run_script "test.sh" "$@"
        ;;
    test-unit)
        run_script "test.sh" "--unit" "$@"
        ;;
    test-integration)
        run_script "test.sh" "--integration" "$@"
        ;;
    test-coverage)
        run_script "test.sh" "--coverage" "$@"
        ;;
    benchmark)
        run_script "test.sh" "--benchmark" "$@"
        ;;
    
    # Quality commands
    lint)
        run_script "lint.sh" "$@"
        ;;
    lint-fix)
        run_script "lint.sh" "--fix" "$@"
        ;;
    format)
        log_info "Formatting code..."
        go fmt ./...
        log_success "Code formatted"
        ;;
    
    # Docker commands
    docker-build)
        run_script "docker-build.sh" "$@"
        ;;
    docker-run)
        log_info "Running Docker container..."
        docker run -p 8080:8080 worduel-backend:latest
        ;;
    docker-dev)
        log_info "Running Docker development container..."
        docker-compose up --build
        ;;
    
    # Utility commands
    clean)
        run_script "dev.sh" "clean" "$@"
        ;;
    deps)
        run_script "dev.sh" "deps" "$@"
        ;;
    health)
        run_script "dev.sh" "health" "$@"
        ;;
    logs)
        run_script "dev.sh" "logs" "$@"
        ;;
    env)
        run_script "setup-env.sh" "$@"
        ;;
    
    # Workflow commands
    ci)
        run_ci_pipeline "$@"
        ;;
    pre-commit)
        run_pre_commit "$@"
        ;;
    release)
        prepare_release "$@"
        ;;
    
    # Help for specific commands
    help)
        if [ $# -gt 0 ]; then
            # Show help for specific command
            case $1 in
                build|docker-build)
                    run_script "build.sh" "--help"
                    ;;
                test|test-*)
                    run_script "test.sh" "--help"
                    ;;
                lint)
                    run_script "lint.sh" "--help"
                    ;;
                setup|run|watch|debug|clean|deps|health|logs)
                    run_script "dev.sh" "help"
                    ;;
                *)
                    show_help
                    ;;
            esac
        else
            show_help
        fi
        ;;
    
    *)
        log_error "Unknown command: $COMMAND"
        echo ""
        show_help
        exit 1
        ;;
esac