package logging

import (
	"context"
	"fmt"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"
)

type Logger struct {
	logger    sentry.Logger
	component string
	fields    map[string]interface{}
}

var (
	globalLogger *Logger
	globalMutex  sync.RWMutex
)

type LogConfig struct {
	Level       string
	Environment string
	Service     string
	AddSource   bool
}

func NewLogger(config LogConfig) (*Logger, error) {
	// Create a Sentry logger with context
	ctx := context.Background()
	sentryLogger := sentry.NewLogger(ctx)

	// Set permanent attributes for service and environment
	sentryLogger.SetAttributes(
		attribute.String("service", config.Service),
		attribute.String("environment", config.Environment),
	)

	return &Logger{
		logger: sentryLogger,
		fields: map[string]interface{}{
			"service":     config.Service,
			"environment": config.Environment,
		},
	}, nil
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalLogger = logger
}

// CreateLogger creates a component-scoped logger with additional context fields
func CreateLogger(component string, additionalFields ...interface{}) *Logger {
	globalMutex.RLock()
	defer globalMutex.RUnlock()

	if globalLogger == nil {
		// Fallback: create a basic logger if global logger isn't set
		ctx := context.Background()
		fallbackSentryLogger := sentry.NewLogger(ctx)
		fallbackSentryLogger.SetAttributes(attribute.String("component", component))

		return &Logger{
			logger:    fallbackSentryLogger,
			component: component,
			fields:    map[string]interface{}{"component": component},
		}
	}

	// Create component logger with additional fields
	ctx := context.Background()
	componentLogger := sentry.NewLogger(ctx)

	// Copy global attributes and add component and additional fields
	attributes := make([]attribute.Builder, 0)
	attributes = append(attributes, attribute.String("component", component))

	// Add service and environment from global logger
	if globalLogger.fields != nil {
		if service, ok := globalLogger.fields["service"].(string); ok {
			attributes = append(attributes, attribute.String("service", service))
		}
		if env, ok := globalLogger.fields["environment"].(string); ok {
			attributes = append(attributes, attribute.String("environment", env))
		}
	}

	// Process additional fields
	fields := map[string]interface{}{"component": component}
	for i := 0; i < len(additionalFields); i += 2 {
		if i+1 < len(additionalFields) {
			key := fmt.Sprintf("%v", additionalFields[i])
			value := additionalFields[i+1]
			fields[key] = value

			// Add to Sentry attributes based on type
			switch v := value.(type) {
			case string:
				attributes = append(attributes, attribute.String(key, v))
			case int:
				attributes = append(attributes, attribute.Int(key, v))
			case bool:
				attributes = append(attributes, attribute.Bool(key, v))
			case float64:
				attributes = append(attributes, attribute.Float64(key, v))
			default:
				attributes = append(attributes, attribute.String(key, fmt.Sprintf("%v", v)))
			}
		}
	}

	componentLogger.SetAttributes(attributes...)

	return &Logger{
		logger:    componentLogger,
		component: component,
		fields:    fields,
	}
}

// Info logs an info level message with attributes
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	if l == nil {
		return
	}
	entry := l.logger.Info()
	l.addAttributes(entry, keysAndValues...)
	entry.Emit(msg)
}

// Error logs an error level message with attributes
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	if l == nil {
		return
	}
	entry := l.logger.Error()
	l.addAttributes(entry, keysAndValues...)
	entry.Emit(msg)
}

// Warn logs a warn level message with attributes
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	if l == nil {
		return
	}
	entry := l.logger.Warn()
	l.addAttributes(entry, keysAndValues...)
	entry.Emit(msg)
}

// Debug logs a debug level message with attributes
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	if l == nil {
		return
	}
	entry := l.logger.Debug()
	l.addAttributes(entry, keysAndValues...)
	entry.Emit(msg)
}

// addAttributes adds key-value pairs as attributes to a log entry
func (l *Logger) addAttributes(entry sentry.LogEntry, keysAndValues ...interface{}) {
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			value := keysAndValues[i+1]

			switch v := value.(type) {
			case string:
				entry.String(key, v)
			case int:
				entry.Int(key, v)
			case bool:
				entry.Bool(key, v)
			case float64:
				entry.Float64(key, v)
			default:
				entry.String(key, fmt.Sprintf("%v", v))
			}
		}
	}
}
