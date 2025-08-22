#!/bin/bash

# Worduel Backend Build Script
# Cross-platform build script with proper flags and optimization

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="worduel-backend"
VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
GIT_BRANCH=${GIT_BRANCH:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")}

# Default build settings
BUILD_OS=""
BUILD_ARCH=""
OUTPUT_DIR="build"
CGO_ENABLED="0"
RACE_DETECTION="false"
OPTIMIZATION_LEVEL="2" # -O2 equivalent in Go
STRIP_SYMBOLS="true"
VERBOSE="false"
CROSS_COMPILE="false"

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
    echo "Worduel Backend Build Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --os OS             Target operating system (linux, darwin, windows)"
    echo "  --arch ARCH         Target architecture (amd64, arm64, 386)"
    echo "  --output DIR        Output directory [default: build]"
    echo "  --version VERSION   Version string [default: dev]"
    echo "  --race              Enable race detection (development only)"
    echo "  --cgo               Enable CGO (default: disabled for static builds)"
    echo "  --no-strip          Don't strip debug symbols"
    echo "  --cross             Enable cross-compilation for multiple platforms"
    echo "  --verbose           Verbose build output"
    echo "  --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                              # Build for current platform"
    echo "  $0 --os linux --arch amd64     # Build for Linux x64"
    echo "  $0 --cross                     # Build for all platforms"
    echo "  $0 --race --verbose            # Development build with race detection"
    echo "  $0 --version v1.0.0            # Build with specific version"
    echo ""
    echo "Environment Variables:"
    echo "  VERSION       Version string"
    echo "  GIT_COMMIT    Git commit hash"
    echo "  GIT_BRANCH    Git branch name"
    echo "  CGO_ENABLED   Enable/disable CGO (0 or 1)"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --os)
            BUILD_OS="$2"
            shift 2
            ;;
        --arch)
            BUILD_ARCH="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --race)
            RACE_DETECTION="true"
            CGO_ENABLED="1"  # Race detection requires CGO
            shift
            ;;
        --cgo)
            CGO_ENABLED="1"
            shift
            ;;
        --no-strip)
            STRIP_SYMBOLS="false"
            shift
            ;;
        --cross)
            CROSS_COMPILE="true"
            shift
            ;;
        --verbose)
            VERBOSE="true"
            shift
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

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build ldflags for version information and optimization
LDFLAGS="-X main.Version=$VERSION"
LDFLAGS="$LDFLAGS -X main.GitCommit=$GIT_COMMIT"
LDFLAGS="$LDFLAGS -X main.GitBranch=$GIT_BRANCH"
LDFLAGS="$LDFLAGS -X main.BuildTime=$BUILD_TIME"

# Add optimization flags
if [ "$STRIP_SYMBOLS" = "true" ]; then
    LDFLAGS="$LDFLAGS -w -s"  # Strip debug info and symbol table
fi

# Add static linking for CGO disabled builds
if [ "$CGO_ENABLED" = "0" ]; then
    LDFLAGS="$LDFLAGS -extldflags '-static'"
fi

# Build flags
BUILD_FLAGS="-a -installsuffix cgo"
if [ "$RACE_DETECTION" = "true" ]; then
    BUILD_FLAGS="$BUILD_FLAGS -race"
    log_warning "Race detection enabled - this will make the binary larger and slower"
fi

if [ "$VERBOSE" = "true" ]; then
    BUILD_FLAGS="$BUILD_FLAGS -v"
fi

# Function to build for a specific platform
build_platform() {
    local target_os=$1
    local target_arch=$2
    
    log_info "Building for $target_os/$target_arch..."
    
    local binary_name="$APP_NAME"
    if [ "$target_os" = "windows" ]; then
        binary_name="$binary_name.exe"
    fi
    
    local output_path="$OUTPUT_DIR/$binary_name"
    if [ "$CROSS_COMPILE" = "true" ]; then
        output_path="$OUTPUT_DIR/${APP_NAME}-${target_os}-${target_arch}"
        if [ "$target_os" = "windows" ]; then
            output_path="$output_path.exe"
        fi
    fi
    
    # Set environment for cross-compilation
    export GOOS=$target_os
    export GOARCH=$target_arch
    export CGO_ENABLED=$CGO_ENABLED
    
    # Run the build
    if go build \
        $BUILD_FLAGS \
        -ldflags "$LDFLAGS" \
        -o "$output_path" \
        ./main.go; then
        
        # Get file size
        local file_size=$(du -h "$output_path" | cut -f1)
        log_success "Built $output_path ($file_size)"
        
        # Verify the binary
        if command -v file >/dev/null 2>&1; then
            log_info "Binary info: $(file "$output_path")"
        fi
        
        return 0
    else
        log_error "Build failed for $target_os/$target_arch"
        return 1
    fi
}

# Main build logic
log_info "Starting Worduel Backend build..."
log_info "Version: $VERSION"
log_info "Git Commit: $GIT_COMMIT"
log_info "Git Branch: $GIT_BRANCH"
log_info "Build Time: $BUILD_TIME"
log_info "CGO Enabled: $CGO_ENABLED"
log_info "Output Directory: $OUTPUT_DIR"

# Verify Go installation
if ! command -v go >/dev/null 2>&1; then
    log_error "Go is not installed or not in PATH"
    exit 1
fi

# Show Go version
GO_VERSION=$(go version)
log_info "Go Version: $GO_VERSION"

# Download dependencies
log_info "Downloading dependencies..."
go mod download
go mod verify

if [ "$CROSS_COMPILE" = "true" ]; then
    log_info "Cross-compiling for multiple platforms..."
    
    # Common platforms for Go applications
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    
    failed_builds=0
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        if ! build_platform "$os" "$arch"; then
            ((failed_builds++))
        fi
    done
    
    if [ $failed_builds -eq 0 ]; then
        log_success "All cross-compilation builds completed successfully"
    else
        log_warning "$failed_builds builds failed"
    fi
    
else
    # Single platform build
    if [ -z "$BUILD_OS" ]; then
        BUILD_OS=$(go env GOOS)
    fi
    if [ -z "$BUILD_ARCH" ]; then
        BUILD_ARCH=$(go env GOARCH)
    fi
    
    log_info "Building for $BUILD_OS/$BUILD_ARCH..."
    
    if build_platform "$BUILD_OS" "$BUILD_ARCH"; then
        log_success "Build completed successfully"
    else
        log_error "Build failed"
        exit 1
    fi
fi

# Show build summary
log_info "Build Summary:"
echo "----------------------------------------"
find "$OUTPUT_DIR" -name "${APP_NAME}*" -type f -exec ls -lh {} \;
echo "----------------------------------------"

# Show usage instructions
echo ""
log_info "Usage Instructions:"
if [ "$CROSS_COMPILE" = "false" ]; then
    echo "  Run: ./$OUTPUT_DIR/$APP_NAME"
    echo "  With environment file: ./$OUTPUT_DIR/$APP_NAME --env .env.production"
else
    echo "  Binaries are in the $OUTPUT_DIR/ directory"
    echo "  Choose the appropriate binary for your target platform"
fi

log_success "Build process completed!"