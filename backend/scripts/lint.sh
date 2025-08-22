#!/bin/bash

# Worduel Backend Lint and Static Analysis Script
# Comprehensive code quality checks and static analysis

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FIX_ISSUES="false"
CHECK_SECURITY="true"
CHECK_PERFORMANCE="true"
CHECK_STYLE="true"
VERBOSE="false"
FAIL_ON_WARNING="false"
OUTPUT_FORMAT="colored-line-number"

# Tool versions (for verification)
GOLANGCI_LINT_VERSION="v1.55.2"
GOSEC_VERSION="2.18.2"
STATICCHECK_VERSION="2023.1.6"

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
    echo "Worduel Backend Lint and Static Analysis"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --fix               Fix auto-fixable issues"
    echo "  --no-security       Skip security checks"
    echo "  --no-performance    Skip performance checks"
    echo "  --no-style          Skip style checks"
    echo "  --fail-on-warning   Fail on warnings (not just errors)"
    echo "  --format FORMAT     Output format [default: colored-line-number]"
    echo "  --verbose           Verbose output"
    echo "  --install-tools     Install required linting tools"
    echo "  --help              Show this help message"
    echo ""
    echo "Output formats:"
    echo "  colored-line-number, line-number, json, tab, checkstyle, junit-xml"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run all checks"
    echo "  $0 --fix               # Run checks and fix issues"
    echo "  $0 --no-security       # Skip security checks"
    echo "  $0 --format json       # Output in JSON format"
    echo "  $0 --install-tools     # Install required tools"
    echo ""
    echo "Tools used:"
    echo "  - golangci-lint: Comprehensive Go linter"
    echo "  - gosec: Security scanner"
    echo "  - staticcheck: Advanced static analysis"
    echo "  - go vet: Built-in Go static analysis"
    echo "  - gofmt: Go code formatting"
    echo ""
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install required tools
install_tools() {
    log_info "Installing linting tools..."
    
    # Install golangci-lint
    if ! command_exists golangci-lint; then
        log_info "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" "$GOLANGCI_LINT_VERSION"
    else
        log_info "golangci-lint already installed: $(golangci-lint --version)"
    fi
    
    # Install gosec
    if ! command_exists gosec; then
        log_info "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@v${GOSEC_VERSION}
    else
        log_info "gosec already installed: $(gosec -version 2>&1 | head -1)"
    fi
    
    # Install staticcheck
    if ! command_exists staticcheck; then
        log_info "Installing staticcheck..."
        go install honnef.co/go/tools/cmd/staticcheck@${STATICCHECK_VERSION}
    else
        log_info "staticcheck already installed: $(staticcheck -version)"
    fi
    
    log_success "All tools installed successfully"
}

# Check tool versions
check_tools() {
    local missing_tools=()
    
    if ! command_exists go; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    if ! command_exists golangci-lint; then
        missing_tools+=("golangci-lint")
    fi
    
    if [ "$CHECK_SECURITY" = "true" ] && ! command_exists gosec; then
        missing_tools+=("gosec")
    fi
    
    if ! command_exists staticcheck; then
        missing_tools+=("staticcheck")
    fi
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Install them with: $0 --install-tools"
        exit 1
    fi
}

# Run gofmt check
run_gofmt() {
    log_info "Checking code formatting with gofmt..."
    
    local unformatted_files
    unformatted_files=$(gofmt -l . 2>/dev/null | grep -E '\.go$' | grep -v vendor || true)
    
    if [ -n "$unformatted_files" ]; then
        log_error "The following files are not formatted correctly:"
        echo "$unformatted_files"
        
        if [ "$FIX_ISSUES" = "true" ]; then
            log_info "Auto-formatting files..."
            echo "$unformatted_files" | xargs gofmt -w
            log_success "Files formatted successfully"
        else
            log_error "Run 'gofmt -w .' or use --fix to format these files"
            return 1
        fi
    else
        log_success "All files are properly formatted"
    fi
    
    return 0
}

# Run go vet
run_go_vet() {
    log_info "Running go vet..."
    
    local vet_flags=""
    if [ "$VERBOSE" = "true" ]; then
        vet_flags="-v"
    fi
    
    if go vet $vet_flags ./...; then
        log_success "go vet passed"
        return 0
    else
        log_error "go vet found issues"
        return 1
    fi
}

# Run staticcheck
run_staticcheck() {
    log_info "Running staticcheck..."
    
    local staticcheck_flags=""
    if [ "$FAIL_ON_WARNING" = "true" ]; then
        staticcheck_flags="-fail=all"
    fi
    
    if staticcheck $staticcheck_flags ./...; then
        log_success "staticcheck passed"
        return 0
    else
        log_error "staticcheck found issues"
        return 1
    fi
}

# Run gosec security scanner
run_gosec() {
    if [ "$CHECK_SECURITY" = "false" ]; then
        log_info "Skipping security checks"
        return 0
    fi
    
    log_info "Running gosec security scanner..."
    
    local gosec_flags="-fmt=sonarqube -out=gosec-report.json"
    if [ "$VERBOSE" = "true" ]; then
        gosec_flags="$gosec_flags -verbose"
    fi
    
    # Run gosec and capture exit code
    if gosec $gosec_flags ./...; then
        log_success "gosec security scan passed"
        return 0
    else
        local exit_code=$?
        if [ $exit_code -eq 1 ]; then
            log_error "gosec found security issues"
        else
            log_error "gosec failed to run (exit code: $exit_code)"
        fi
        return 1
    fi
}

# Run golangci-lint
run_golangci_lint() {
    log_info "Running golangci-lint..."
    
    local lint_flags=""
    lint_flags="--out-format=$OUTPUT_FORMAT"
    
    if [ "$VERBOSE" = "true" ]; then
        lint_flags="$lint_flags --verbose"
    fi
    
    if [ "$FIX_ISSUES" = "true" ]; then
        lint_flags="$lint_flags --fix"
    fi
    
    # Configure enabled/disabled linters based on options
    local enable_linters=""
    local disable_linters=""
    
    if [ "$CHECK_STYLE" = "true" ]; then
        enable_linters="$enable_linters,gofmt,goimports,misspell,whitespace"
    else
        disable_linters="$disable_linters,gofmt,goimports,misspell,whitespace"
    fi
    
    if [ "$CHECK_PERFORMANCE" = "true" ]; then
        enable_linters="$enable_linters,prealloc,unconvert"
    else
        disable_linters="$disable_linters,prealloc,unconvert"
    fi
    
    if [ "$CHECK_SECURITY" = "true" ]; then
        enable_linters="$enable_linters,gosec"
    else
        disable_linters="$disable_linters,gosec"
    fi
    
    if [ -n "$enable_linters" ]; then
        lint_flags="$lint_flags --enable=${enable_linters:1}"  # Remove leading comma
    fi
    
    if [ -n "$disable_linters" ]; then
        lint_flags="$lint_flags --disable=${disable_linters:1}"  # Remove leading comma
    fi
    
    # Run golangci-lint
    if golangci-lint run $lint_flags ./...; then
        log_success "golangci-lint passed"
        return 0
    else
        log_error "golangci-lint found issues"
        return 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --fix)
            FIX_ISSUES="true"
            shift
            ;;
        --no-security)
            CHECK_SECURITY="false"
            shift
            ;;
        --no-performance)
            CHECK_PERFORMANCE="false"
            shift
            ;;
        --no-style)
            CHECK_STYLE="false"
            shift
            ;;
        --fail-on-warning)
            FAIL_ON_WARNING="true"
            shift
            ;;
        --format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE="true"
            shift
            ;;
        --install-tools)
            install_tools
            exit 0
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Check if we're in the backend directory
if [ ! -f "main.go" ]; then
    log_error "main.go not found. Please run this script from the backend directory."
    exit 1
fi

# Check required tools
check_tools

# Main execution
log_info "Starting Worduel Backend Lint and Static Analysis..."
log_info "Go Version: $(go version)"

if [ "$FIX_ISSUES" = "true" ]; then
    log_info "Auto-fix mode: Enabled"
fi

if [ "$FAIL_ON_WARNING" = "true" ]; then
    log_info "Fail on warnings: Enabled"
fi

echo ""

# Track overall success
overall_success=true

# Run all checks
if [ "$CHECK_STYLE" = "true" ]; then
    if ! run_gofmt; then
        overall_success=false
    fi
    echo ""
fi

if ! run_go_vet; then
    overall_success=false
fi
echo ""

if ! run_staticcheck; then
    overall_success=false
fi
echo ""

if ! run_gosec; then
    overall_success=false
fi
echo ""

if ! run_golangci_lint; then
    overall_success=false
fi
echo ""

# Final results
if [ "$overall_success" = "true" ]; then
    log_success "All lint and static analysis checks passed!"
    
    if [ "$FIX_ISSUES" = "true" ]; then
        log_info "Auto-fixable issues have been resolved"
    fi
    
    echo ""
    log_info "Code quality summary:"
    echo "  ✅ Format check (gofmt)"
    echo "  ✅ Static analysis (go vet)"
    echo "  ✅ Advanced analysis (staticcheck)"
    if [ "$CHECK_SECURITY" = "true" ]; then
        echo "  ✅ Security scan (gosec)"
    fi
    echo "  ✅ Comprehensive linting (golangci-lint)"
    
    exit 0
else
    log_error "Some checks failed!"
    
    echo ""
    log_info "To fix issues:"
    echo "  • Run with --fix to auto-fix formatting and some issues"
    echo "  • Check the output above for specific issues to resolve"
    echo "  • Use --verbose for more detailed output"
    
    exit 1
fi