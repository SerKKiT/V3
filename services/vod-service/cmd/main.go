package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/SerKKiT/streaming-platform/vod-service/internal/config"
	"github.com/SerKKiT/streaming-platform/vod-service/internal/handlers"
	"github.com/SerKKiT/streaming-platform/vod-service/internal/middleware"
	"github.com/SerKKiT/streaming-platform/vod-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/vod-service/internal/storage"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("🚀 Starting VOD Service...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("❌ Failed to ping database:", err)
	}

	log.Println("✅ Connected to vod_db successfully")

	// Initialize MinIO storage
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatal("❌ Failed to initialize MinIO storage:", err)
	}

	// Initialize repository
	videoRepo := repository.NewVideoRepository(db)

	// Initialize handlers
	videoHandler := handlers.NewVideoHandler(
		videoRepo,
		minioStorage,
		cfg.RecordingServiceURL,
		"recordings", // recording bucket для копирования
		"vod-videos", // vod bucket для хранения и стриминга
	)

	// Setup router
	router := gin.Default()

	// CORS middleware
	// router.Use(func(c *gin.Context) {
	// 	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	// 	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	// 	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Internal-API-Key")
	// 	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// 	if c.Request.Method == "OPTIONS" {
	// 		c.AbortWithStatus(204)
	// 		return
	// 	}

	// 	c.Next()
	// })

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// ✅ Optional auth routes (public videos, auth for private via cookie/header)
	optionalAuth := router.Group("/")
	optionalAuth.Use(middleware.OptionalAuthMiddleware())
	{
		optionalAuth.GET("/videos", videoHandler.ListAllVideos)
		optionalAuth.GET("/videos/:id", videoHandler.GetVideo)
		optionalAuth.GET("/videos/:id/stream", videoHandler.GetStreamURL)
		optionalAuth.GET("/videos/:id/play", videoHandler.StreamVideoFile)
		optionalAuth.GET("/videos/:id/thumbnail", videoHandler.StreamThumbnail)
		optionalAuth.POST("/videos/:id/view", videoHandler.IncrementView)
	}

	// ✅ Internal service-to-service routes (require INTERNAL_API_KEY)
	internal := router.Group("/")
	internal.Use(middleware.InternalAuth())
	{
		internal.POST("/videos/import-recording", videoHandler.ImportRecording)
	}

	// ✅ Protected routes (require auth via cookie/header)
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/videos/user", videoHandler.GetUserVideos)
		protected.PUT("/videos/:id", videoHandler.UpdateVideo)
		protected.DELETE("/videos/:id", videoHandler.DeleteVideo)
		protected.POST("/videos/:id/like", videoHandler.LikeVideo)
	}

	log.Printf("✅ VOD Service running on port %s", cfg.Port)
	log.Printf("📦 Using MinIO buckets: recordings (source), vod-videos (storage)")
	log.Println("🔒 Authentication: Cookie (video playback) + JWT Header (API calls) + Internal Key (service-to-service)")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
