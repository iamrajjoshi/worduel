package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func validate(config *Config) error {
	if err := validateServerConfig(config.Server); err != nil {
		return err
	}
	if err := validateCORSConfig(config.CORS); err != nil {
		return err
	}
	if err := validateRateLimitConfig(config.Rate); err != nil {
		return err
	}
	if err := validateRoomConfig(config.Room); err != nil {
		return err
	}
	if err := validateGameConfig(config.Game); err != nil {
		return err
	}
	if err := validateSecurityConfig(config.Security); err != nil {
		return err
	}
	if err := validateLoggingConfig(config.Logging); err != nil {
		return err
	}
	if err := validateSentryConfig(config.Sentry); err != nil {
		return err
	}
	return nil
}

func validateServerConfig(config ServerConfig) error {
	if config.Port == "" {
		return errors.New("server port cannot be empty")
	}
	
	if portNum, err := strconv.Atoi(config.Port); err != nil || portNum < 1 || portNum > 65535 {
		return errors.New("server port must be a valid number between 1 and 65535")
	}
	
	if config.Host == "" {
		return errors.New("server host cannot be empty")
	}
	
	if config.ReadTimeout <= 0 {
		return errors.New("read timeout must be positive")
	}
	
	if config.WriteTimeout <= 0 {
		return errors.New("write timeout must be positive")
	}
	
	if config.IdleTimeout <= 0 {
		return errors.New("idle timeout must be positive")
	}
	
	if config.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}
	
	return nil
}

func validateCORSConfig(config CORSConfig) error {
	if len(config.AllowedOrigins) == 0 {
		return errors.New("at least one allowed origin must be specified")
	}
	
	if len(config.AllowedMethods) == 0 {
		return errors.New("at least one allowed method must be specified")
	}
	
	return nil
}

func validateRateLimitConfig(config RateLimitConfig) error {
	if config.WebSocketMessagesPerMinute <= 0 {
		return errors.New("WebSocket messages per minute must be positive")
	}
	
	if config.APIRequestsPerMinute <= 0 {
		return errors.New("API requests per minute must be positive")
	}
	
	if config.MaxConnectionsPerIP <= 0 {
		return errors.New("max connections per IP must be positive")
	}
	
	return nil
}

func validateRoomConfig(config RoomConfig) error {
	if config.MaxConcurrentRooms <= 0 {
		return errors.New("max concurrent rooms must be positive")
	}
	
	if config.RoomInactiveTimeout <= 0 {
		return errors.New("room inactive timeout must be positive")
	}
	
	if config.GameTimeout <= 0 {
		return errors.New("game timeout must be positive")
	}
	
	if config.CleanupInterval <= 0 {
		return errors.New("cleanup interval must be positive")
	}
	
	if config.MaxPlayersPerRoom <= 0 {
		return errors.New("max players per room must be positive")
	}
	
	if config.MaxPlayersPerRoom > 10 {
		return errors.New("max players per room cannot exceed 10")
	}
	
	return nil
}

func validateGameConfig(config GameConfig) error {
	if config.MaxGuesses <= 0 {
		return errors.New("max guesses must be positive")
	}
	
	if config.MaxGuesses > 20 {
		return errors.New("max guesses cannot exceed 20")
	}
	
	if config.WordLength <= 0 {
		return errors.New("word length must be positive")
	}
	
	if config.WordLength != 5 {
		return errors.New("word length must be 5 for Wordle")
	}
	
	if config.GuessTimeoutMS <= 0 {
		return errors.New("guess timeout must be positive")
	}
	
	if config.BroadcastTimeoutMS <= 0 {
		return errors.New("broadcast timeout must be positive")
	}
	
	if config.BroadcastTimeoutMS > 1000 {
		return errors.New("broadcast timeout cannot exceed 1000ms for real-time gameplay")
	}
	
	return nil
}

func validateSecurityConfig(config SecurityConfig) error {
	if config.MaxMessageSize <= 0 {
		return errors.New("max message size must be positive")
	}
	
	if config.MaxMessageSize > 10*1024 {
		return errors.New("max message size cannot exceed 10KB")
	}
	
	if config.ConnectionTimeout <= 0 {
		return errors.New("connection timeout must be positive")
	}
	
	return nil
}

func validateLoggingConfig(config LoggingConfig) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	for _, validLevel := range validLevels {
		if config.Level == validLevel {
			goto levelValid
		}
	}
	return errors.New("log level must be one of: debug, info, warn, error")

levelValid:
	if config.Service == "" {
		return errors.New("service name cannot be empty")
	}
	
	if config.Environment == "" {
		return errors.New("environment cannot be empty")
	}
	
	return nil
}

func validateSentryConfig(config SentryConfig) error {
	if config.TracesSampleRate < 0 || config.TracesSampleRate > 1.0 {
		return errors.New("Sentry traces sample rate must be between 0 and 1.0")
	}
	
	if config.Environment == "" {
		return errors.New("Sentry environment cannot be empty")
	}
	
	if config.Release == "" {
		return errors.New("Sentry release cannot be empty")
	}
	
	return nil
}