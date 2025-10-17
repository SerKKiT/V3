package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port      string
	JWTSecret string
	Services  ServicesConfig
	RateLimit RateLimitConfig
}

type ServicesConfig struct {
	AuthURL      string
	StreamURL    string
	RecordingURL string
	VODURL       string
}

type RateLimitConfig struct {
	RequestsPerSecond int
	Burst             int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		Services: ServicesConfig{
			AuthURL:      getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
			StreamURL:    getEnv("STREAM_SERVICE_URL", "http://stream-service:8082"),
			RecordingURL: getEnv("RECORDING_SERVICE_URL", "http://recording-service:8083"),
			VODURL:       getEnv("VOD_SERVICE_URL", "http://vod-service:8084"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: 100,
			Burst:             200,
		},
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
