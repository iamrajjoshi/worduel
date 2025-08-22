package config

import (
	"fmt"
	"time"
)

type Config struct {
	Server   ServerConfig
	CORS     CORSConfig
	Rate     RateLimitConfig
	Room     RoomConfig
	Game     GameConfig
	Security SecurityConfig
	Dev      DevConfig
	Logging  LoggingConfig
	Sentry   SentryConfig
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type RateLimitConfig struct {
	WebSocketMessagesPerMinute int
	APIRequestsPerMinute       int
	MaxConnectionsPerIP        int
}

type RoomConfig struct {
	MaxConcurrentRooms    int
	RoomInactiveTimeout   time.Duration
	GameTimeout           time.Duration
	CleanupInterval       time.Duration
	MaxPlayersPerRoom     int
}

type GameConfig struct {
	MaxGuesses        int
	WordLength        int
	GuessTimeoutMS    int
	BroadcastTimeoutMS int
}

type SecurityConfig struct {
	ValidateOrigin     bool
	MaxMessageSize     int64
	ConnectionTimeout  time.Duration
}

type DevConfig struct {
	DebugMode   bool
	VerboseLog  bool
	ProfileMode bool
}

type LoggingConfig struct {
	Level       string
	Environment string
	Service     string
	AddSource   bool
}

type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
	Debug            bool
}

func Load() (*Config, error) {
	config := &Config{
		Server:   loadServerConfig(),
		CORS:     loadCORSConfig(),
		Rate:     loadRateLimitConfig(),
		Room:     loadRoomConfig(),
		Game:     loadGameConfig(),
		Security: loadSecurityConfig(),
		Dev:      loadDevConfig(),
		Logging:  loadLoggingConfig(),
		Sentry:   loadSentryConfig(),
	}

	if err := validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port:            getEnvString("PORT", "8080"),
		Host:            getEnvString("HOST", "0.0.0.0"),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
	}
}

func loadCORSConfig() CORSConfig {
	defaultOrigins := []string{"http://localhost:3000"}
	origins := getEnvStringSlice("ALLOWED_ORIGINS", defaultOrigins)
	
	return CORSConfig{
		AllowedOrigins: origins,
		AllowedMethods: getEnvStringSlice("ALLOWED_METHODS", []string{"GET", "POST", "OPTIONS"}),
		AllowedHeaders: getEnvStringSlice("ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
	}
}

func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		WebSocketMessagesPerMinute: getEnvInt("WS_RATE_LIMIT", 60),
		APIRequestsPerMinute:       getEnvInt("API_RATE_LIMIT", 100),
		MaxConnectionsPerIP:        getEnvInt("MAX_CONNECTIONS_PER_IP", 10),
	}
}

func loadRoomConfig() RoomConfig {
	return RoomConfig{
		MaxConcurrentRooms:    getEnvInt("MAX_CONCURRENT_ROOMS", 1000),
		RoomInactiveTimeout:   getEnvDuration("ROOM_INACTIVE_TIMEOUT", 30*time.Minute),
		GameTimeout:           getEnvDuration("GAME_TIMEOUT", 30*time.Minute),
		CleanupInterval:       getEnvDuration("CLEANUP_INTERVAL", 5*time.Minute),
		MaxPlayersPerRoom:     getEnvInt("MAX_PLAYERS_PER_ROOM", 2),
	}
}

func loadGameConfig() GameConfig {
	return GameConfig{
		MaxGuesses:        getEnvInt("MAX_GUESSES", 6),
		WordLength:        getEnvInt("WORD_LENGTH", 5),
		GuessTimeoutMS:    getEnvInt("GUESS_TIMEOUT_MS", 10),
		BroadcastTimeoutMS: getEnvInt("BROADCAST_TIMEOUT_MS", 100),
	}
}

func loadSecurityConfig() SecurityConfig {
	return SecurityConfig{
		ValidateOrigin:    getEnvBool("VALIDATE_ORIGIN", true),
		MaxMessageSize:    getEnvInt64("MAX_MESSAGE_SIZE", 1024),
		ConnectionTimeout: getEnvDuration("CONNECTION_TIMEOUT", 30*time.Second),
	}
}

func loadDevConfig() DevConfig {
	return DevConfig{
		DebugMode:   getEnvBool("DEBUG_MODE", false),
		VerboseLog:  getEnvBool("VERBOSE_LOG", false),
		ProfileMode: getEnvBool("PROFILE_MODE", false),
	}
}

func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:       getEnvString("LOG_LEVEL", "info"),
		Environment: getEnvString("ENVIRONMENT", "development"),
		Service:     getEnvString("SERVICE_NAME", "worduel-backend"),
		AddSource:   getEnvBool("LOG_ADD_SOURCE", false),
	}
}

func loadSentryConfig() SentryConfig {
	return SentryConfig{
		DSN:              getEnvString("SENTRY_DSN", ""),
		Environment:      getEnvString("SENTRY_ENVIRONMENT", "development"),
		Release:          getEnvString("SENTRY_RELEASE", "1.0.0"),
		TracesSampleRate: getEnvFloat64("SENTRY_TRACES_SAMPLE_RATE", 0.1),
		Debug:            getEnvBool("SENTRY_DEBUG", false),
	}
}