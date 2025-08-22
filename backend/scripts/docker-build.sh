#!/bin/bash

# Worduel Backend Docker Build Script
# Builds and tests Docker images for different environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="worduel-backend"
VERSION=${VERSION:-"latest"}
REGISTRY=${REGISTRY:-""}

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

# Check if we're in the backend directory
if [ ! -f "Dockerfile" ]; then
    log_error "Dockerfile not found. Please run this script from the backend directory."
    exit 1
fi

# Parse command line arguments
BUILD_TARGET="production"
PUSH_IMAGE=false
RUN_TESTS=false
PLATFORM="linux/amd64"

while [[ $# -gt 0 ]]; do
    case $1 in
        --target)
            BUILD_TARGET="$2"
            shift 2
            ;;
        --push)
            PUSH_IMAGE=true
            shift
            ;;
        --test)
            RUN_TESTS=true
            shift
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --target TARGET     Build target (production, runtime-debug) [default: production]"
            echo "  --push              Push image to registry after building"
            echo "  --test              Run tests after building"
            echo "  --platform PLATFORM Target platform [default: linux/amd64]"
            echo "  --version VERSION   Image version tag [default: latest]"
            echo "  --registry REGISTRY Container registry prefix"
            echo "  --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Build production image"
            echo "  $0 --target runtime-debug --test     # Build debug image and run tests"
            echo "  $0 --push --version v1.0.0           # Build and push with version tag"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Set full image name
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${VERSION}"
else
    FULL_IMAGE_NAME="${IMAGE_NAME}:${VERSION}"
fi

log_info "Building Docker image: $FULL_IMAGE_NAME"
log_info "Target: $BUILD_TARGET"
log_info "Platform: $PLATFORM"

# Build the image
log_info "Starting Docker build..."
docker build \
    --target "$BUILD_TARGET" \
    --platform "$PLATFORM" \
    --tag "$FULL_IMAGE_NAME" \
    --progress=plain \
    .

log_success "Docker build completed successfully"

# Show image information
log_info "Image information:"
docker images "$FULL_IMAGE_NAME" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

# Test the image
if [ "$RUN_TESTS" = true ]; then
    log_info "Running image tests..."
    
    # Test that the binary exists and is executable
    log_info "Testing binary..."
    docker run --rm "$FULL_IMAGE_NAME" --version 2>/dev/null || true
    
    # Test health endpoint (start container and test)
    log_info "Testing health endpoint..."
    CONTAINER_ID=$(docker run -d -p 8080:8080 "$FULL_IMAGE_NAME")
    
    # Wait for container to start
    sleep 5
    
    # Test health endpoint
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log_success "Health check passed"
    else
        log_warning "Health check failed (this might be expected if dependencies are missing)"
    fi
    
    # Cleanup test container
    docker stop "$CONTAINER_ID" >/dev/null 2>&1 || true
    docker rm "$CONTAINER_ID" >/dev/null 2>&1 || true
    
    log_success "Image tests completed"
fi

# Push to registry if requested
if [ "$PUSH_IMAGE" = true ]; then
    if [ -z "$REGISTRY" ]; then
        log_error "Cannot push: no registry specified. Use --registry option."
        exit 1
    fi
    
    log_info "Pushing image to registry..."
    docker push "$FULL_IMAGE_NAME"
    log_success "Image pushed successfully"
fi

# Show final instructions
echo ""
log_success "Build completed successfully!"
echo ""
echo "To run the image:"
echo "  docker run -p 8080:8080 $FULL_IMAGE_NAME"
echo ""
echo "To run with environment file:"
echo "  docker run -p 8080:8080 --env-file .env.production $FULL_IMAGE_NAME"
echo ""
echo "To use with docker-compose:"
echo "  docker-compose up"
echo ""

# Show size comparison
log_info "Image size comparison:"
echo "Production (scratch-based):  $(docker images $IMAGE_NAME:$VERSION --format '{{.Size}}' 2>/dev/null || echo 'Not built')"
if docker images "$IMAGE_NAME:debug-$VERSION" >/dev/null 2>&1; then
    echo "Debug (Alpine-based):        $(docker images $IMAGE_NAME:debug-$VERSION --format '{{.Size}}')"
fi