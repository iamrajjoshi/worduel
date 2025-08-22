package logging

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
	Debug            bool
}

func InitSentry(config SentryConfig) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		Release:          config.Release,
		TracesSampleRate: config.TracesSampleRate,
		Debug:            config.Debug,
		EnableLogs:       true, // Enable Sentry logs
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.ServerName = "worduel-backend"
			return event
		},
		AttachStacktrace: true,
		Transport: &sentry.HTTPTransport{
			Timeout: 5 * time.Second,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}
	return nil
}

func SentryHTTPMiddleware() func(http.Handler) http.Handler {
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         false,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	})
	return sentryHandler.Handle
}

func CaptureError(ctx context.Context, err error, tags map[string]string, extra map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			for k, v := range tags {
				scope.SetTag(k, v)
			}

			for k, v := range extra {
				scope.SetExtra(k, v)
			}

			scope.SetLevel(sentry.LevelError)
			hub.CaptureException(err)
		})
	} else {
		sentry.WithScope(func(scope *sentry.Scope) {
			for k, v := range tags {
				scope.SetTag(k, v)
			}

			for k, v := range extra {
				scope.SetExtra(k, v)
			}

			scope.SetLevel(sentry.LevelError)
			sentry.CaptureException(err)
		})
	}
}

func CaptureMessage(ctx context.Context, message string, level sentry.Level, tags map[string]string, extra map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			for k, v := range tags {
				scope.SetTag(k, v)
			}

			for k, v := range extra {
				scope.SetExtra(k, v)
			}

			scope.SetLevel(level)
			hub.CaptureMessage(message)
		})
	} else {
		sentry.WithScope(func(scope *sentry.Scope) {
			for k, v := range tags {
				scope.SetTag(k, v)
			}

			for k, v := range extra {
				scope.SetExtra(k, v)
			}
			scope.SetLevel(level)
			sentry.CaptureMessage(message)
		})
	}
}

func StartTransaction(ctx context.Context, name, operation string) *sentry.Span {
	return sentry.StartTransaction(ctx, name)
}

func StartSpan(ctx context.Context, operation, description string) *sentry.Span {
	return sentry.StartSpan(ctx, operation)
}

func FlushSentry(timeout time.Duration) {
	sentry.Flush(timeout)
}

func AddBreadcrumb(ctx context.Context, category, message, level string, data map[string]interface{}) {
	breadcrumb := &sentry.Breadcrumb{
		Category:  category,
		Message:   message,
		Level:     parseBreadcrumbLevel(level),
		Timestamp: time.Now(),
		Data:      data,
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.AddBreadcrumb(breadcrumb, nil)
	} else {
		sentry.AddBreadcrumb(breadcrumb)
	}
}

func parseBreadcrumbLevel(level string) sentry.Level {
	switch level {
	case "debug":
		return sentry.LevelDebug
	case "info":
		return sentry.LevelInfo
	case "warning":
		return sentry.LevelWarning
	case "error":
		return sentry.LevelError
	case "fatal":
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}

func SetUser(ctx context.Context, user sentry.User) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(user)
		})
	} else {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(user)
		})
	}
}

func SetTag(ctx context.Context, key, value string) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag(key, value)
		})
	} else {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag(key, value)
		})
	}
}

func SetExtra(ctx context.Context, key string, value interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetExtra(key, value)
		})
	} else {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetExtra(key, value)
		})
	}
}

type PerformanceMetrics struct {
	ActiveConnections int64
	ActiveRooms       int64
	MessageThroughput float64
	MemoryUsageMB     float64
}

func RecordPerformanceMetrics(ctx context.Context, metrics PerformanceMetrics) {
	tags := map[string]string{
		"component": "performance_metrics",
	}
	extra := map[string]interface{}{
		"active_connections": metrics.ActiveConnections,
		"active_rooms":       metrics.ActiveRooms,
		"message_throughput": metrics.MessageThroughput,
		"memory_usage_mb":    metrics.MemoryUsageMB,
	}

	CaptureMessage(ctx, "Performance metrics snapshot", sentry.LevelInfo, tags, extra)
}
