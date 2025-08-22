#!/bin/bash

# Worduel Backend Test Runner Script
# Comprehensive test runner with coverage reporting

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_DIR="coverage"
COVERAGE_FILE="$COVERAGE_DIR/coverage.out"
COVERAGE_HTML="$COVERAGE_DIR/coverage.html"
COVERAGE_THRESHOLD=80
TEST_TIMEOUT="30s"
RACE_DETECTION="false"
VERBOSE="false"
BENCHMARK="false"
INTEGRATION_TESTS="false"
UNIT_TESTS_ONLY="false"
WATCH_MODE="false"
FAIL_FAST="false"

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
    echo "Worduel Backend Test Runner"
    echo ""
    echo "Usage: $0 [OPTIONS] [PATTERN]"
    echo ""
    echo "Options:"
    echo "  --unit              Run only unit tests"
    echo "  --integration       Run only integration tests"
    echo "  --benchmark         Run benchmark tests"
    echo "  --coverage          Generate coverage report"
    echo "  --race              Enable race detection"
    echo "  --verbose           Verbose test output"
    echo "  --watch             Watch mode (re-run tests on file changes)"
    echo "  --fail-fast         Stop on first test failure"
    echo "  --timeout DURATION  Test timeout [default: 30s]"
    echo "  --threshold PCT     Coverage threshold percentage [default: 80]"
    echo "  --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Run all tests"
    echo "  $0 --unit --coverage        # Run unit tests with coverage"
    echo "  $0 --integration --verbose  # Run integration tests verbosely"
    echo "  $0 --benchmark              # Run benchmark tests"
    echo "  $0 --race --fail-fast       # Run with race detection, stop on failure"
    echo "  $0 TestRoomCreation         # Run specific test pattern"
    echo "  $0 --watch                  # Watch mode for development"
    echo ""
    echo "Coverage Reports:"
    echo "  HTML report: $COVERAGE_HTML"
    echo "  Text report: $COVERAGE_FILE"
    echo ""
}

# Parse command line arguments
TEST_PATTERN=""
GENERATE_COVERAGE="false"

while [[ $# -gt 0 ]]; do
    case $1 in
        --unit)
            UNIT_TESTS_ONLY="true"
            shift
            ;;
        --integration)
            INTEGRATION_TESTS="true"
            shift
            ;;
        --benchmark)
            BENCHMARK="true"
            shift
            ;;
        --coverage)
            GENERATE_COVERAGE="true"
            shift
            ;;
        --race)
            RACE_DETECTION="true"
            shift
            ;;
        --verbose)
            VERBOSE="true"
            shift
            ;;
        --watch)
            WATCH_MODE="true"
            shift
            ;;
        --fail-fast)
            FAIL_FAST="true"
            shift
            ;;
        --timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        --threshold)
            COVERAGE_THRESHOLD="$2"
            shift 2
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
            TEST_PATTERN="$1"
            shift
            ;;
    esac
done

# Check if we're in the backend directory
if [ ! -f "main.go" ]; then
    log_error "main.go not found. Please run this script from the backend directory."
    exit 1
fi

# Check Go installation
if ! command -v go >/dev/null 2>&1; then
    log_error "Go is not installed or not in PATH"
    exit 1
fi

# Create coverage directory
if [ "$GENERATE_COVERAGE" = "true" ]; then
    mkdir -p "$COVERAGE_DIR"
fi

# Function to run tests
run_tests() {
    local test_type="$1"
    local test_flags=""
    local test_packages=""
    
    log_info "Running $test_type tests..."
    
    # Build test flags
    if [ "$VERBOSE" = "true" ]; then
        test_flags="$test_flags -v"
    fi
    
    if [ "$RACE_DETECTION" = "true" ]; then
        test_flags="$test_flags -race"
        log_info "Race detection enabled"
    fi
    
    if [ "$FAIL_FAST" = "true" ]; then
        test_flags="$test_flags -failfast"
    fi
    
    test_flags="$test_flags -timeout $TEST_TIMEOUT"
    
    # Add test pattern if provided
    if [ -n "$TEST_PATTERN" ]; then
        test_flags="$test_flags -run $TEST_PATTERN"
    fi
    
    # Determine test packages
    case $test_type in
        "unit")
            # Unit tests exclude integration tests
            test_packages="./internal/... ./tests/unit/..."
            if [ "$GENERATE_COVERAGE" = "true" ]; then
                test_flags="$test_flags -coverprofile=$COVERAGE_FILE -covermode=atomic"
            fi
            ;;
        "integration")
            test_packages="./tests/integration/..."
            # Integration tests typically don't include coverage
            ;;
        "all")
            test_packages="./..."
            if [ "$GENERATE_COVERAGE" = "true" ]; then
                test_flags="$test_flags -coverprofile=$COVERAGE_FILE -covermode=atomic"
            fi
            ;;
    esac
    
    # Run the tests
    log_info "Test command: go test $test_flags $test_packages"
    
    if go test $test_flags $test_packages; then
        log_success "$test_type tests passed"
        return 0
    else
        log_error "$test_type tests failed"
        return 1
    fi
}

# Function to run benchmark tests
run_benchmarks() {
    log_info "Running benchmark tests..."
    
    local bench_flags="-bench=. -benchmem"
    if [ "$VERBOSE" = "true" ]; then
        bench_flags="$bench_flags -v"
    fi
    
    # Add test pattern if provided
    if [ -n "$TEST_PATTERN" ]; then
        bench_flags="$bench_flags -run $TEST_PATTERN"
    fi
    
    if go test $bench_flags ./...; then
        log_success "Benchmark tests completed"
        return 0
    else
        log_error "Benchmark tests failed"
        return 1
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    if [ ! -f "$COVERAGE_FILE" ]; then
        log_error "Coverage file not found. Run tests with --coverage first."
        return 1
    fi
    
    log_info "Generating coverage reports..."
    
    # Generate HTML coverage report
    if go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"; then
        log_success "HTML coverage report generated: $COVERAGE_HTML"
    else
        log_error "Failed to generate HTML coverage report"
        return 1
    fi
    
    # Show coverage summary
    local coverage_percent=$(go tool cover -func="$COVERAGE_FILE" | tail -1 | awk '{print $3}' | sed 's/%//')
    
    echo ""
    log_info "Coverage Summary:"
    go tool cover -func="$COVERAGE_FILE" | tail -10
    
    echo ""
    if (( $(echo "$coverage_percent >= $COVERAGE_THRESHOLD" | bc -l) )); then
        log_success "Coverage: $coverage_percent% (threshold: $COVERAGE_THRESHOLD%)"
    else
        log_error "Coverage: $coverage_percent% is below threshold: $COVERAGE_THRESHOLD%"
        return 1
    fi
    
    return 0
}

# Function for watch mode
watch_tests() {
    if ! command -v fswatch >/dev/null 2>&1; then
        log_error "fswatch is required for watch mode. Install with: brew install fswatch"
        exit 1
    fi
    
    log_info "Starting watch mode... (Ctrl+C to stop)"
    log_info "Watching for changes in .go files"
    
    # Initial test run
    run_main_tests
    
    # Watch for changes
    fswatch -o -r --include='.*\.go$' . | while read num_events; do
        echo ""
        log_info "Files changed, re-running tests..."
        run_main_tests || true  # Don't exit on test failure in watch mode
    done
}

# Function to run main test logic
run_main_tests() {
    local test_failed=false
    
    # Download dependencies
    log_info "Ensuring dependencies are up to date..."
    go mod download
    
    if [ "$UNIT_TESTS_ONLY" = "true" ]; then
        if ! run_tests "unit"; then
            test_failed=true
        fi
    elif [ "$INTEGRATION_TESTS" = "true" ]; then
        if ! run_tests "integration"; then
            test_failed=true
        fi
    else
        # Run all tests
        if ! run_tests "all"; then
            test_failed=true
        fi
    fi
    
    # Run benchmarks if requested
    if [ "$BENCHMARK" = "true" ]; then
        if ! run_benchmarks; then
            test_failed=true
        fi
    fi
    
    # Generate coverage report if requested and tests passed
    if [ "$GENERATE_COVERAGE" = "true" ] && [ "$test_failed" = "false" ]; then
        if ! generate_coverage_report; then
            test_failed=true
        fi
    fi
    
    return $([ "$test_failed" = "true" ] && echo 1 || echo 0)
}

# Main execution
log_info "Starting Worduel Backend Test Runner..."
log_info "Go Version: $(go version)"
log_info "Test Timeout: $TEST_TIMEOUT"

if [ "$RACE_DETECTION" = "true" ]; then
    log_info "Race Detection: Enabled"
fi

if [ "$GENERATE_COVERAGE" = "true" ]; then
    log_info "Coverage Reporting: Enabled (threshold: $COVERAGE_THRESHOLD%)"
fi

if [ -n "$TEST_PATTERN" ]; then
    log_info "Test Pattern: $TEST_PATTERN"
fi

echo ""

if [ "$WATCH_MODE" = "true" ]; then
    watch_tests
else
    if run_main_tests; then
        echo ""
        log_success "All tests completed successfully!"
        
        if [ "$GENERATE_COVERAGE" = "true" ]; then
            echo ""
            log_info "Coverage report available at: $COVERAGE_HTML"
        fi
        
        exit 0
    else
        echo ""
        log_error "Some tests failed!"
        exit 1
    fi
fi