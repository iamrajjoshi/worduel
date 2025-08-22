# Environment Configuration Guide

This document provides a comprehensive guide to configuring the Worduel backend through environment variables.

## Quick Start

1. **For Development**: `cp .env.example .env.development`
2. **For Production**: `cp .env.example .env.production` 
3. **Use Setup Script**: `./scripts/setup-env.sh` (interactive setup)

## Environment Variable Reference

### 🌐 Server Configuration

| Variable | Default | Description | Production Notes |
|----------|---------|-------------|------------------|
| `PORT` | `8080` | Server port | Use standard HTTP port or reverse proxy |
| `HOST` | `0.0.0.0` | Server host | Keep as `0.0.0.0` for containers |
| `READ_TIMEOUT` | `10s` | HTTP read timeout | Increase for slower clients |
| `WRITE_TIMEOUT` | `10s` | HTTP write timeout | Increase for slower responses |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout | Balance resources vs UX |
| `SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout | Allow time for connections to close |

### 🔒 Security & CORS

| Variable | Default | Description | Production Notes |
|----------|---------|-------------|------------------|
| `ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated CORS origins | **MUST** set to your domain(s) |
| `ALLOWED_METHODS` | `GET,POST,OPTIONS` | Allowed HTTP methods | Minimal set for security |
| `ALLOWED_HEADERS` | `Content-Type,Authorization` | Allowed headers | Add only what's needed |
| `VALIDATE_ORIGIN` | `true` | Enable origin validation | **MUST** be `true` in production |
| `MAX_MESSAGE_SIZE` | `1024` | WebSocket message size limit (bytes) | Prevent memory attacks |
| `CONNECTION_TIMEOUT` | `30s` | WebSocket connection timeout | Balance resources vs UX |

### 🚦 Rate Limiting

| Variable | Default | Description | Production Notes |
|----------|---------|-------------|------------------|
| `WS_RATE_LIMIT` | `60` | WebSocket messages per minute | Adjust based on game pace |
| `API_RATE_LIMIT` | `100` | API requests per minute | Prevent API abuse |
| `MAX_CONNECTIONS_PER_IP` | `10` | Max connections per IP | Prevent connection flooding |

### 🏠 Room & Game Management

| Variable | Default | Description | Production Notes |
|----------|---------|-------------|------------------|
| `MAX_CONCURRENT_ROOMS` | `1000` | Maximum active rooms | Scale based on server capacity |
| `ROOM_INACTIVE_TIMEOUT` | `30m` | Room cleanup timeout | Balance resources vs UX |
| `GAME_TIMEOUT` | `30m` | Game completion timeout | Prevent stuck games |
| `CLEANUP_INTERVAL` | `5m` | Room cleanup frequency | More frequent = more CPU |
| `MAX_PLAYERS_PER_ROOM` | `2` | Players per room | Core game design constraint |
| `MAX_GUESSES` | `6` | Guesses per game | Classic Wordle mechanic |
| `WORD_LENGTH` | `5` | Word length | Fixed for Wordle compatibility |
| `GUESS_TIMEOUT_MS` | `10` | Guess processing timeout | Keep low for responsiveness |
| `BROADCAST_TIMEOUT_MS` | `100` | Message broadcast timeout | Balance speed vs reliability |

### 🔧 Development & Debug

| Variable | Default | Description | Production Notes |
|----------|---------|-------------|------------------|
| `DEBUG_MODE` | `false` | Enable debug features | **MUST** be `false` in production |
| `VERBOSE_LOG` | `false` | Enable verbose logging | Use sparingly in production |
| `PROFILE_MODE` | `false` | Enable performance profiling | For debugging only |

### 📊 Logging Configuration

| Variable | Default | Description | Options |
|----------|---------|-------------|---------|
| `LOG_LEVEL` | `info` | Minimum log level | `debug`, `info`, `warn`, `error` |
| `ENVIRONMENT` | `development` | Environment name | `development`, `staging`, `production` |
| `SERVICE_NAME` | `worduel-backend` | Service identifier | Customize for multi-service deployments |
| `LOG_ADD_SOURCE` | `false` | Add source file/line to logs | Useful for debugging, avoid in production |

### 📈 Sentry Monitoring (Optional)

| Variable | Default | Description | Production Notes |
|----------|---------|-------------|------------------|
| `SENTRY_DSN` | `""` | Sentry project DSN | **Essential** for production monitoring |
| `SENTRY_ENVIRONMENT` | `development` | Environment tag in Sentry | Match your deployment stage |
| `SENTRY_RELEASE` | `1.0.0` | Release version | Use semantic versioning |
| `SENTRY_TRACES_SAMPLE_RATE` | `0.1` | Performance monitoring sample rate (0.0-1.0) | Start low, increase if needed |
| `SENTRY_DEBUG` | `false` | Enable Sentry debug logging | Only for troubleshooting |

## Environment-Specific Recommendations

### 🧪 Development Environment

```bash
# Relaxed settings for easier development
DEBUG_MODE=true
LOG_LEVEL=debug
LOG_ADD_SOURCE=true
VERBOSE_LOG=true
VALIDATE_ORIGIN=false
WS_RATE_LIMIT=120
API_RATE_LIMIT=200
MAX_CONNECTIONS_PER_IP=20
SENTRY_DSN=""  # Disable Sentry locally
```

### 🚀 Production Environment

```bash
# Strict security settings
DEBUG_MODE=false
LOG_LEVEL=info
LOG_ADD_SOURCE=false
VERBOSE_LOG=false
VALIDATE_ORIGIN=true
WS_RATE_LIMIT=60
API_RATE_LIMIT=100
MAX_CONNECTIONS_PER_IP=5
SENTRY_DSN="https://your-dsn@sentry.io/project"  # REQUIRED
```

### 🧪 Staging Environment

```bash
# Balanced settings for testing
DEBUG_MODE=false
LOG_LEVEL=debug
LOG_ADD_SOURCE=true
VERBOSE_LOG=true
VALIDATE_ORIGIN=true
SENTRY_DSN="https://your-staging-dsn@sentry.io/project"
SENTRY_TRACES_SAMPLE_RATE=1.0  # Full sampling for testing
```

## Security Checklist

- [ ] `ALLOWED_ORIGINS` set to actual domain(s)
- [ ] `VALIDATE_ORIGIN=true`
- [ ] `DEBUG_MODE=false`
- [ ] `SENTRY_DSN` configured for error monitoring
- [ ] Rate limits appropriate for expected traffic
- [ ] `MAX_MESSAGE_SIZE` prevents memory attacks
- [ ] Connection timeouts prevent resource exhaustion

## Performance Tuning

### High Traffic
- Increase `MAX_CONCURRENT_ROOMS`
- Reduce `CLEANUP_INTERVAL` for faster cleanup
- Increase rate limits cautiously
- Reduce `SENTRY_TRACES_SAMPLE_RATE`

### Resource Constrained
- Decrease `MAX_CONCURRENT_ROOMS`
- Reduce `ROOM_INACTIVE_TIMEOUT`
- Increase `CLEANUP_INTERVAL`
- Set `LOG_LEVEL=warn` or `LOG_LEVEL=error`

### Debugging Issues
- Set `LOG_LEVEL=debug`
- Enable `LOG_ADD_SOURCE=true`
- Set `SENTRY_TRACES_SAMPLE_RATE=1.0`
- Enable `VERBOSE_LOG=true`

## Monitoring Setup

1. **Create Sentry Project**: Visit [sentry.io](https://sentry.io) and create a new Go project
2. **Copy DSN**: Add your project's DSN to `SENTRY_DSN`
3. **Configure Alerts**: Set up alerts for error rate, performance issues
4. **Release Tracking**: Update `SENTRY_RELEASE` with each deployment
5. **Dashboard**: Monitor real-time errors and performance metrics

## Troubleshooting

### Common Issues

**Server won't start**
- Check `PORT` is not in use
- Verify `HOST` setting (use `localhost` for local dev)

**CORS errors**
- Add your frontend URL to `ALLOWED_ORIGINS`
- Ensure protocol matches (http vs https)

**Rate limiting issues**
- Increase relevant rate limit variables
- Check `MAX_CONNECTIONS_PER_IP`

**Memory issues**
- Reduce `MAX_CONCURRENT_ROOMS`
- Check `MAX_MESSAGE_SIZE`
- Increase cleanup frequency

**Logs not appearing**
- Check `LOG_LEVEL` setting
- Verify `SERVICE_NAME` in log filtering

## Support

For configuration questions or issues:
1. Check the [README.md](../README.md) for general setup
2. Review this document for specific variable details
3. Use the setup script: `./scripts/setup-env.sh`
4. Check Sentry dashboard for runtime issues