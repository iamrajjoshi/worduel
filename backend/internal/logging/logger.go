package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	sentryhandler "github.com/getsentry/sentry-go/slog"
)

type Logger struct {
	*slog.Logger
}

type LogConfig struct {
	Level       string
	Environment string
	Service     string
	SentryDSN   string
	AddSource   bool
}

func NewLogger(config LogConfig) (*Logger, error) {
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	var handler slog.Handler

	if config.Environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	if config.SentryDSN != "" {
		sentryOpts := sentryhandler.Option{
			Level: level,
		}
		handler = sentryOpts.NewSentryHandler(context.Background())
	}

	logger := slog.New(handler)
	logger = logger.With(
		"service", config.Service,
		"environment", config.Environment,
	)

	return &Logger{Logger: logger}, nil
}

func (l *Logger) WithContext(ctx context.Context) *slog.Logger {
	return l.Logger.With("correlation_id", getCorrelationID(ctx))
}

func (l *Logger) WithFields(fields map[string]interface{}) *slog.Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return l.Logger.With(args...)
}

func (l *Logger) LogError(ctx context.Context, err error, msg string, fields ...interface{}) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]interface{}, 0, len(fields)+4)
	args = append(args, "error", err)
	args = append(args, "correlation_id", getCorrelationID(ctx))
	args = append(args, fields...)
	l.Logger.Error(msg, args...)
}

func (l *Logger) LogInfo(ctx context.Context, msg string, fields ...interface{}) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]interface{}, 0, len(fields)+2)
	args = append(args, "correlation_id", getCorrelationID(ctx))
	args = append(args, fields...)
	l.Logger.Info(msg, args...)
}

func (l *Logger) LogDebug(ctx context.Context, msg string, fields ...interface{}) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]interface{}, 0, len(fields)+2)
	args = append(args, "correlation_id", getCorrelationID(ctx))
	args = append(args, fields...)
	l.Logger.Debug(msg, args...)
}

func (l *Logger) LogWarn(ctx context.Context, msg string, fields ...interface{}) {
	if l == nil || l.Logger == nil {
		return
	}
	args := make([]interface{}, 0, len(fields)+2)
	args = append(args, "correlation_id", getCorrelationID(ctx))
	args = append(args, fields...)
	l.Logger.Warn(msg, args...)
}

type contextKey string

const correlationIDKey contextKey = "correlation_id"

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

func getCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return generateCorrelationID()
}

func generateCorrelationID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

type RequestFields struct {
	Method    string
	URL       string
	UserAgent string
	IP        string
	Duration  time.Duration
	Status    int
}

func (l *Logger) LogRequest(ctx context.Context, fields RequestFields) {
	if l == nil || l.Logger == nil {
		return
	}
	l.Logger.Info("HTTP request completed",
		"correlation_id", getCorrelationID(ctx),
		"method", fields.Method,
		"url", fields.URL,
		"user_agent", fields.UserAgent,
		"ip", fields.IP,
		"duration_ms", fields.Duration.Milliseconds(),
		"status", fields.Status,
	)
}

type GameEventFields struct {
	EventType string
	RoomID    string
	PlayerID  string
	GameState string
}

func (l *Logger) LogGameEvent(ctx context.Context, fields GameEventFields) {
	if l == nil || l.Logger == nil {
		return
	}
	l.Logger.Info("Game event",
		"correlation_id", getCorrelationID(ctx),
		"event_type", fields.EventType,
		"room_id", fields.RoomID,
		"player_id", fields.PlayerID,
		"game_state", fields.GameState,
	)
}

type WSEventFields struct {
	EventType    string
	ClientID     string
	RoomID       string
	MessageType  string
	ConnectionIP string
}

func (l *Logger) LogWebSocketEvent(ctx context.Context, fields WSEventFields) {
	if l == nil || l.Logger == nil {
		return
	}
	l.Logger.Info("WebSocket event",
		"correlation_id", getCorrelationID(ctx),
		"event_type", fields.EventType,
		"client_id", fields.ClientID,
		"room_id", fields.RoomID,
		"message_type", fields.MessageType,
		"connection_ip", fields.ConnectionIP,
	)
}