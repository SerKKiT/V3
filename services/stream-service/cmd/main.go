package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SerKKiT/streaming-platform/stream-service/internal/config"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/handlers"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/middleware"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/srt"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/storage"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/transcoder"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
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
	log.Println("✅ Successfully connected to database")

	// Initialize MinIO storage with public policy
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucketLive,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}
	log.Println("✅ Successfully connected to MinIO")

	// Set public read policy for live-streams bucket
	if err := minioStorage.SetPublicReadPolicy(); err != nil {
		log.Printf("⚠️  Warning: Failed to set public policy: %v", err)
	}

	// Initialize components
	streamRepo := repository.NewStreamRepository(db)

	// Create FFmpeg transcoder
	ffmpegTranscoder, err := transcoder.NewFFmpegTranscoder(
		"/var/www/hls",
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucketLive,
		cfg.MinioUseSSL,
		streamRepo,        // Передать repository
		cfg.PublicBaseURL, // Передать public base URL
	)

	if err != nil {
		log.Fatalf("Failed to initialize FFmpeg transcoder: %v", err)
	}

	// Recording Service URL for webhooks
	recordingServiceURL := os.Getenv("RECORDING_SERVICE_URL")
	if recordingServiceURL == "" {
		recordingServiceURL = "http://recording-service:8083"
	}

	srtHandler := srt.NewHandler(streamRepo, ffmpegTranscoder, recordingServiceURL)

	// Initialize SRT server
	srtServer, err := srt.NewServer(&srt.Config{
		Address: ":" + cfg.SRTPort,
		Latency: cfg.SRTLatency,
	}, srtHandler)
	if err != nil {
		log.Fatalf("Failed to create SRT server: %v", err)
	}

	// Start SRT server in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srtServer.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("❌ SRT server error: %v", err)
		}
	}()

	// Setup HTTP API
	streamHandler := handlers.NewStreamHandler(
		streamRepo,
		"localhost:"+cfg.SRTPort,
		minioStorage,
		cfg.PublicBaseURL, // ДОБАВЛЕНО: из конфига
	)

	// Cleanup handler
	cleanupHandler := handlers.NewCleanupHandler(
		streamRepo,
		minioStorage,
		"/var/www/hls",
	)

	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.UserContextMiddleware()) // Extract X-User-ID from API Gateway

	// Health check
	router.GET("/health", streamHandler.Health)

	// Public routes (NO AUTH - no user_id required)
	public := router.Group("/streams")
	{
		public.GET("/live", streamHandler.GetLiveStreams)
		public.GET("/by-key/:key", streamHandler.GetStreamByKey)
		public.GET("/:id/play", streamHandler.GetStreamPlaybackInfo)
		public.GET("/:id/thumbnail", streamHandler.GetStreamThumbnail)
		public.GET("/:id", streamHandler.GetStream)
		public.GET("/:id/qualities", streamHandler.GetStreamQualities)

	}

	// Protected routes (require X-User-ID header from API Gateway)
	protected := router.Group("/streams")
	protected.Use(requireUserID()) // Ensure user_id exists in context
	{
		protected.POST("", streamHandler.CreateStream)
		protected.GET("/user", streamHandler.GetUserStreams)
		protected.PUT("/:id", streamHandler.UpdateStream)
		protected.DELETE("/:id", streamHandler.DeleteStream)
	}

	// ✅ НОВОЕ: Webhook endpoint (public - no auth)
	router.POST("/webhooks/recording-complete", cleanupHandler.HandleRecordingComplete)

	// Start HTTP server in goroutine
	go func() {
		log.Printf("🚀 HTTP server starting on port %s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("❌ Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("⏹️  Shutting down gracefully...")
	cancel()
	srtServer.Stop()
	log.Println("✅ Server stopped")
}

// requireUserID middleware ensures user_id is present in context
func requireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			log.Printf("❌ Missing user_id in context for %s %s", c.Request.Method, c.Request.URL.Path)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "User authentication required",
				"details": "X-User-ID header missing or invalid",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
