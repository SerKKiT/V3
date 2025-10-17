package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port      string
	JWTSecret string
	Services  ServiceConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig // ✅ НОВОЕ ПОЛЕ
}

type ServiceConfig struct {
	AuthURL      string
	StreamURL    string
	RecordingURL string
	VODURL       string
}

type RateLimitConfig struct {
	RequestsPerSecond int
	Burst             int
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

func Load() (*Config, error) {
	// Загрузка CORS origins
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		// Дефолтные значения для development
		allowedOriginsStr = "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173"
		log.Println("⚠️ ALLOWED_ORIGINS not set, using default dev origins")
	}

	allowedOrigins := strings.Split(allowedOriginsStr, ",")
	// Trim whitespace
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	corsAllowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"

	// Rate limit
	requestsPerSecond, err := strconv.Atoi(getEnv("RATE_LIMIT_RPS", "100"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
	}

	burst, err := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "200"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BURST: %w", err)
	}

	config := &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Services: ServiceConfig{
			AuthURL:      getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
			StreamURL:    getEnv("STREAM_SERVICE_URL", "http://stream-service:8082"),
			RecordingURL: getEnv("RECORDING_SERVICE_URL", "http://recording-service:8084"),
			VODURL:       getEnv("VOD_SERVICE_URL", "http://vod-service:8083"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: requestsPerSecond,
			Burst:             burst,
		},
		CORS: CORSConfig{
			AllowedOrigins:   allowedOrigins,
			AllowCredentials: corsAllowCredentials,
		},
	}

	log.Printf("✅ Config loaded:")
	log.Printf("   - Port: %s", config.Port)
	log.Printf("   - Allowed CORS Origins: %v", config.CORS.AllowedOrigins)
	log.Printf("   - CORS Allow Credentials: %v", config.CORS.AllowCredentials)

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
