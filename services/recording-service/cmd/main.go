package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SerKKiT/streaming-platform/recording-service/internal/config"
	"github.com/SerKKiT/streaming-platform/recording-service/internal/handlers"
	"github.com/SerKKiT/streaming-platform/recording-service/internal/monitor"
	"github.com/SerKKiT/streaming-platform/recording-service/internal/recorder"
	"github.com/SerKKiT/streaming-platform/recording-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/recording-service/internal/storage"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to vod_db successfully")

	// Initialize storage
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucketRecording,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}

	// Initialize recorder
	ffmpegRecorder := recorder.NewFFmpegRecorder(cfg.RecordingsPath, minioStorage.GetClient(), cfg.MinioBucketLiveStreams)

	// Initialize repository
	recordingRepo := repository.NewRecordingRepository(db)

	// VOD Service URL
	vodServiceURL := os.Getenv("VOD_SERVICE_URL")
	if vodServiceURL == "" {
		vodServiceURL = "http://vod-service:8084"
	}

	// Initialize stream monitor
	streamMonitor := monitor.NewStreamMonitor(
		cfg.StreamServiceURL,
		vodServiceURL, // ← Передаём VOD URL
		ffmpegRecorder,
		recordingRepo,
		minioStorage,
		time.Duration(cfg.MonitorInterval)*time.Second,
	)

	// Initialize handlers
	recordingHandler := handlers.NewRecordingHandler(recordingRepo)
	webhookHandler := handlers.NewWebhookHandler(streamMonitor)

	// Setup HTTP server
	router := gin.Default()
	router.GET("/health", recordingHandler.Health)
	router.GET("/recordings", recordingHandler.GetAllRecordings)
	router.GET("/recordings/:id", recordingHandler.GetRecordingByID)
	router.GET("/recording/:id", recordingHandler.GetRecordingByID) // Альтернативный путь (если используется)
	router.POST("/webhook/stream", webhookHandler.HandleStreamEvent)

	// Start stream monitor in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go streamMonitor.Start(ctx)

	// Start HTTP server
	go func() {
		log.Printf("✅ Recording Service running on port %s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal("❌ Failed to start server:", err)
		}
	}()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Shutting down gracefully...")
	cancel()
}
