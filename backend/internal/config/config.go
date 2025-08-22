package config

import "os"

type Config struct {
	Port string
	AllowedOrigins []string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	allowedOrigins := []string{"http://localhost:3000"}
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		allowedOrigins = []string{origins}
	}

	return &Config{
		Port: port,
		AllowedOrigins: allowedOrigins,
	}
}