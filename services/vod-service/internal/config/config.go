package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                string
	DatabaseURL         string
	MinioEndpoint       string
	MinioAccessKey      string
	MinioSecretKey      string
	MinioUseSSL         bool
	MinioBucket         string
	RecordingServiceURL string
	JWTSecret           string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	// Поддержка как DATABASE_URL, так и отдельных переменных
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Формируем из отдельных переменных
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
			return nil, fmt.Errorf("DATABASE_URL or DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME is required")
		}

		dbURL = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName,
		)
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

	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "recordings" // По умолчанию тот же bucket что у Recording Service
	}

	recordingServiceURL := os.Getenv("RECORDING_SERVICE_URL")
	if recordingServiceURL == "" {
		recordingServiceURL = "http://recording-service:8083"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}

	return &Config{
		Port:                port,
		DatabaseURL:         dbURL,
		MinioEndpoint:       minioEndpoint,
		MinioAccessKey:      minioAccessKey,
		MinioSecretKey:      minioSecretKey,
		MinioUseSSL:         os.Getenv("MINIO_USE_SSL") == "true",
		MinioBucket:         minioBucket,
		RecordingServiceURL: recordingServiceURL,
		JWTSecret:           jwtSecret,
	}, nil
}
