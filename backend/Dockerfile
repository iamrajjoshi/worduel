# Multi-stage Docker build for Worduel Backend
# Stage 1: Build stage with full Go toolchain
FROM golang:1.21.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
# CGO_ENABLED=0 for static binary, GOOS=linux for Linux containers
# -ldflags='-w -s' strips debug information and symbol table to reduce size
# -a forces rebuilding of all packages
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o worduel-backend \
    ./main.go

# Verify the binary was built
RUN ls -la worduel-backend && file worduel-backend

# Stage 2: Runtime stage with minimal base image
FROM scratch AS runtime

# Copy timezone data and certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Create a minimal passwd file for the non-root user
COPY --from=builder /etc/passwd /etc/passwd

# Copy the static binary from builder stage
COPY --from=builder /app/worduel-backend /worduel-backend

# Create non-root user for security
# Note: In scratch image, we need to create user differently
USER nobody

# Expose the port (configurable via environment)
EXPOSE 8080

# Add health check for container orchestration
# Note: scratch image doesn't have curl, so we need a different approach
# We'll rely on orchestrator health checks via HTTP endpoints
# For Kubernetes/Docker Swarm, use: http://localhost:8080/health

# Set the entrypoint
ENTRYPOINT ["/worduel-backend"]

# Alternative runtime stage using minimal Alpine for debugging purposes
FROM alpine:3.19 AS runtime-debug

# Install minimal runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 worduel && \
    adduser -u 1001 -G worduel -s /bin/sh -D worduel

# Set working directory
WORKDIR /app

# Copy the static binary from builder stage
COPY --from=builder /app/worduel-backend /app/worduel-backend

# Change ownership and permissions
RUN chown -R worduel:worduel /app && \
    chmod +x /app/worduel-backend

# Switch to non-root user
USER worduel

# Expose the port
EXPOSE 8080

# Health check using curl
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:${PORT:-8080}/health || exit 1

# Set the entrypoint
ENTRYPOINT ["/app/worduel-backend"]

# Production stage (default) - uses scratch for minimal size
FROM runtime AS production