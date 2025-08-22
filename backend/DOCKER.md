# Worduel Backend - Docker Deployment Guide

This document provides instructions for building and deploying the Worduel Backend using Docker.

## Quick Start

### Development
```bash
# Build and run with docker-compose (debug image with Alpine)
docker-compose up --build

# The service will be available at http://localhost:8080
```

### Production
```bash
# Build production image (minimal scratch-based)
docker-compose -f docker-compose.prod.yml build --target production

# Run in production mode
docker-compose -f docker-compose.prod.yml up -d
```

## Docker Images

The Dockerfile provides two build targets:

### 1. Production Image (`production`)
- **Base:** `scratch` (minimal, no shell, ~5-10MB)
- **Security:** Non-root user, read-only filesystem
- **Use case:** Production deployments
- **Debugging:** Limited (no shell access)

### 2. Debug Image (`runtime-debug`)  
- **Base:** `alpine:3.19` (~15-20MB)
- **Security:** Non-root user with shell access
- **Use case:** Development, debugging, troubleshooting
- **Tools:** Includes `curl` for health checks

## Build Commands

### Using Docker Directly
```bash
# Production build
docker build --target production -t worduel-backend:latest .

# Debug build
docker build --target runtime-debug -t worduel-backend:debug .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 --target production -t worduel-backend:latest .
```

### Using Build Script
```bash
# Make script executable
chmod +x scripts/docker-build.sh

# Build production image
./scripts/docker-build.sh

# Build debug image with tests
./scripts/docker-build.sh --target runtime-debug --test

# Build and push to registry
./scripts/docker-build.sh --push --registry your-registry.com --version v1.0.0
```

## Running the Container

### Basic Run
```bash
# Production image
docker run -p 8080:8080 worduel-backend:latest

# Debug image
docker run -p 8080:8080 worduel-backend:debug
```

### With Environment Variables
```bash
# Using environment file
docker run -p 8080:8080 --env-file .env.production worduel-backend:latest

# Individual environment variables
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e LOG_LEVEL=info \
  -e SENTRY_ENABLED=true \
  -e SENTRY_DSN=your-sentry-dsn \
  worduel-backend:latest
```

## Health Checks

The application provides health check endpoints:

- **Main health check:** `http://localhost:8080/health`
- **Liveness probe:** `http://localhost:8080/health/liveness`
- **Readiness probe:** `http://localhost:8080/health/readiness`

### Container Health Checks
```bash
# Check container health (debug image with curl)
docker run -d --name worduel-test -p 8080:8080 worduel-backend:debug
sleep 5
curl http://localhost:8080/health
docker stop worduel-test && docker rm worduel-test
```

## Environment Configuration

### Development (.env.development)
```env
PORT=8080
LOG_LEVEL=debug
LOG_FORMAT=text
DEBUG_MODE=true
ALLOWED_ORIGINS=http://localhost:3000
SENTRY_ENABLED=false
```

### Production (.env.production)
```env
PORT=8080
LOG_LEVEL=info
LOG_FORMAT=json
DEBUG_MODE=false
ALLOWED_ORIGINS=https://yourdomain.com
SENTRY_ENABLED=true
SENTRY_DSN=your-sentry-dsn
VALIDATE_ORIGIN=true
SECURE_HEADERS=true
```

## Container Orchestration

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: worduel-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: worduel-backend
  template:
    metadata:
      labels:
        app: worduel-backend
    spec:
      containers:
      - name: worduel-backend
        image: worduel-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        livenessProbe:
          httpGet:
            path: /health/liveness
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health/readiness
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          limits:
            memory: "256Mi"
            cpu: "500m"
          requests:
            memory: "128Mi"
            cpu: "250m"
```

### Docker Swarm
```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.prod.yml worduel

# Scale service
docker service scale worduel_worduel-backend=3

# Check service status
docker service ls
docker service ps worduel_worduel-backend
```

## Security Considerations

### Production Security Features
- **Non-root user:** Container runs as `nobody` user
- **Read-only filesystem:** Container filesystem is read-only
- **No shell access:** Scratch-based image has no shell
- **Minimal attack surface:** Only contains the Go binary
- **Resource limits:** Memory and CPU limits configured

### Security Headers
The application includes security middleware that adds:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`

## Troubleshooting

### Common Issues
1. **Port already in use:** Change the host port mapping
   ```bash
   docker run -p 8081:8080 worduel-backend:latest
   ```

2. **Health check failures:** Wait for application startup (30 seconds)
   ```bash
   # Check logs
   docker logs container-name
   ```

3. **Environment variables not loading:** Verify .env file exists and format
   ```bash
   # Test with explicit variables
   docker run -e LOG_LEVEL=debug -p 8080:8080 worduel-backend:debug
   ```

### Debugging
```bash
# Access debug container shell
docker run -it --entrypoint sh worduel-backend:debug

# Check application logs
docker logs -f container-name

# Inspect container
docker inspect container-name

# Test from inside container (debug image)
docker exec -it container-name sh
curl http://localhost:8080/health
```

## Performance Tuning

### Resource Limits
Adjust based on your load requirements:

```yaml
# docker-compose.yml
deploy:
  resources:
    limits:
      memory: 512M      # Increase for high traffic
      cpus: '1.0'       # Increase for CPU-intensive operations
    reservations:
      memory: 256M
      cpus: '0.5'
```

### Scaling
```bash
# Docker Compose scaling
docker-compose up --scale worduel-backend=3

# Monitor resource usage
docker stats
```

## Monitoring and Logging

### Structured Logging
Production containers output structured JSON logs:
```json
{
  "time": "2025-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "Server starting",
  "port": 8080,
  "version": "1.0.0"
}
```

### Sentry Integration
Configure Sentry for error tracking:
```env
SENTRY_ENABLED=true
SENTRY_DSN=https://your-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=production
SENTRY_TRACES_SAMPLE_RATE=0.1
```

For more information, see the main [README.md](README.md) and [ENVIRONMENT.md](ENVIRONMENT.md) files.