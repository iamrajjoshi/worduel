package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"PORT":                      "9000",
				"HOST":                      "127.0.0.1",
				"ALLOWED_ORIGINS":           "http://example.com,http://localhost:8080",
				"WS_RATE_LIMIT":             "120",
				"MAX_CONCURRENT_ROOMS":      "500",
				"ROOM_INACTIVE_TIMEOUT":     "45m",
				"DEBUG_MODE":                "true",
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"PORT": "invalid",
			},
			wantErr: true,
		},
		{
			name: "port out of range",
			envVars: map[string]string{
				"PORT": "99999",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config == nil {
					t.Error("Load() returned nil config")
				}
			}
		})
	}
}

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "use default when env not set",
			key:          "TEST_STRING",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
		{
			name:         "use env value when set",
			key:          "TEST_STRING",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvString(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		want         int
	}{
		{
			name:         "use default when env not set",
			key:          "TEST_INT",
			defaultValue: 42,
			envValue:     "",
			want:         42,
		},
		{
			name:         "use env value when set and valid",
			key:          "TEST_INT",
			defaultValue: 42,
			envValue:     "100",
			want:         100,
		},
		{
			name:         "use default when env value invalid",
			key:          "TEST_INT",
			defaultValue: 42,
			envValue:     "invalid",
			want:         42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvInt(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		want         bool
	}{
		{
			name:         "use default when env not set",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "",
			want:         true,
		},
		{
			name:         "parse true",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			want:         true,
		},
		{
			name:         "parse false",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			want:         false,
		},
		{
			name:         "use default when invalid",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "invalid",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvBool(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		want         time.Duration
	}{
		{
			name:         "use default when env not set",
			key:          "TEST_DURATION",
			defaultValue: 5 * time.Minute,
			envValue:     "",
			want:         5 * time.Minute,
		},
		{
			name:         "parse valid duration",
			key:          "TEST_DURATION",
			defaultValue: 5 * time.Minute,
			envValue:     "10m",
			want:         10 * time.Minute,
		},
		{
			name:         "use default when invalid",
			key:          "TEST_DURATION",
			defaultValue: 5 * time.Minute,
			envValue:     "invalid",
			want:         5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvDuration(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvStringSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue []string
		envValue     string
		want         []string
	}{
		{
			name:         "use default when env not set",
			key:          "TEST_SLICE",
			defaultValue: []string{"a", "b"},
			envValue:     "",
			want:         []string{"a", "b"},
		},
		{
			name:         "parse comma-separated values",
			key:          "TEST_SLICE",
			defaultValue: []string{"a", "b"},
			envValue:     "x,y,z",
			want:         []string{"x", "y", "z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvStringSlice(tt.key, tt.defaultValue)
			if len(got) != len(tt.want) {
				t.Errorf("getEnvStringSlice() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getEnvStringSlice() = %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Port:            "8080",
					Host:            "0.0.0.0",
					ReadTimeout:     10 * time.Second,
					WriteTimeout:    10 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 30 * time.Second,
				},
				CORS: CORSConfig{
					AllowedOrigins: []string{"http://localhost:3000"},
					AllowedMethods: []string{"GET", "POST"},
					AllowedHeaders: []string{"Content-Type"},
				},
				Rate: RateLimitConfig{
					WebSocketMessagesPerMinute: 60,
					APIRequestsPerMinute:       100,
					MaxConnectionsPerIP:        10,
				},
				Room: RoomConfig{
					MaxConcurrentRooms:    1000,
					RoomInactiveTimeout:   30 * time.Minute,
					GameTimeout:           30 * time.Minute,
					CleanupInterval:       5 * time.Minute,
					MaxPlayersPerRoom:     2,
				},
				Game: GameConfig{
					MaxGuesses:        6,
					WordLength:        5,
					GuessTimeoutMS:    10,
					BroadcastTimeoutMS: 100,
				},
				Security: SecurityConfig{
					ValidateOrigin:    true,
					MaxMessageSize:    1024,
					ConnectionTimeout: 30 * time.Second,
				},
				Dev: DevConfig{
					DebugMode:   false,
					VerboseLog:  false,
					ProfileMode: false,
				},
				Logging: LoggingConfig{
					Level:       "info",
					Environment: "test",
					Service:     "worduel-backend",
					AddSource:   false,
				},
				Sentry: SentryConfig{
					DSN:              "",
					Environment:      "test",
					Release:          "1.0.0",
					TracesSampleRate: 0.1,
					Debug:            false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: &Config{
				Server: ServerConfig{
					Port: "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid word length",
			config: &Config{
				Server: ServerConfig{
					Port:            "8080",
					Host:            "0.0.0.0",
					ReadTimeout:     10 * time.Second,
					WriteTimeout:    10 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 30 * time.Second,
				},
				CORS: CORSConfig{
					AllowedOrigins: []string{"http://localhost:3000"},
					AllowedMethods: []string{"GET"},
				},
				Rate: RateLimitConfig{
					WebSocketMessagesPerMinute: 60,
					APIRequestsPerMinute:       100,
					MaxConnectionsPerIP:        10,
				},
				Room: RoomConfig{
					MaxConcurrentRooms:    1000,
					RoomInactiveTimeout:   30 * time.Minute,
					GameTimeout:           30 * time.Minute,
					CleanupInterval:       5 * time.Minute,
					MaxPlayersPerRoom:     2,
				},
				Game: GameConfig{
					WordLength: 4,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}