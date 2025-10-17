package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL            string
	Port                   string
	StreamServiceURL       string
	MinioEndpoint          string
	MinioAccessKey         string
	MinioSecretKey         string
	MinioUseSSL            bool
	MinioBucketRecording   string
	MinioBucketLiveStreams string // ← Добавить
	RecordingsPath         string
	MonitorInterval        int // секунды
}

func LoadConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	streamServiceURL := os.Getenv("STREAM_SERVICE_URL")
	if streamServiceURL == "" {
		streamServiceURL = "http://stream-service:8082"
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

	return &Config{
		DatabaseURL:            dbURL,
		Port:                   port,
		StreamServiceURL:       streamServiceURL,
		MinioEndpoint:          minioEndpoint,
		MinioAccessKey:         minioAccessKey,
		MinioSecretKey:         minioSecretKey,
		MinioUseSSL:            minioUseSSL,
		MinioBucketRecording:   "recordings",
		MinioBucketLiveStreams: "live-streams", // ← Добавить
		RecordingsPath:         "/tmp/recordings",
		MonitorInterval:        10, // проверять каждые 10 секунд
	}, nil
}
