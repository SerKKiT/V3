package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	Port            string
	SRTPort         string
	SRTLatency      uint // миллисекунды
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioUseSSL     bool
	MinioBucketLive string
	PublicBaseURL   string // ДОБАВЛЕНО
}

func LoadConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	srtPort := os.Getenv("SRT_PORT")
	if srtPort == "" {
		srtPort = "6000"
	}

	srtLatencyStr := os.Getenv("SRT_LATENCY")
	srtLatency := uint(2000) // default 2000 milliseconds (2 seconds)
	if srtLatencyStr != "" {
		val, err := strconv.ParseUint(srtLatencyStr, 10, 32)
		if err == nil {
			srtLatency = uint(val)
		}
	}

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "minio:9000"
	}

	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY is required")
	}

	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		return nil, fmt.Errorf("MINIO_SECRET_KEY is required")
	}

	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// ДОБАВЛЕНО: Public base URL для HLS endpoints
	publicBaseURL := os.Getenv("PUBLIC_BASE_URL")
	if publicBaseURL == "" {
		publicBaseURL = "http://localhost" // Default для development
	}

	return &Config{
		DatabaseURL:     dbURL,
		JWTSecret:       jwtSecret,
		Port:            port,
		SRTPort:         srtPort,
		SRTLatency:      srtLatency,
		MinioEndpoint:   minioEndpoint,
		MinioAccessKey:  minioAccessKey,
		MinioSecretKey:  minioSecretKey,
		MinioUseSSL:     minioUseSSL,
		MinioBucketLive: "live-streams",
		PublicBaseURL:   publicBaseURL, // ДОБАВЛЕНО
	}, nil
}
