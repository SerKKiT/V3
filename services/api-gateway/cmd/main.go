package main

import (
	"log"

	"github.com/SerKKiT/streaming-platform/api-gateway/internal/config"
	"github.com/SerKKiT/streaming-platform/api-gateway/internal/handlers"
	"github.com/SerKKiT/streaming-platform/api-gateway/internal/middleware"
	"github.com/SerKKiT/streaming-platform/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 Starting API Gateway...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)

	// Initialize proxies
	authProxy := proxy.NewServiceProxy(cfg.Services.AuthURL)
	streamProxy := proxy.NewServiceProxy(cfg.Services.StreamURL)
	recordingProxy := proxy.NewServiceProxy(cfg.Services.RecordingURL)
	vodProxy := proxy.NewServiceProxy(cfg.Services.VODURL)

	// Setup router
	router := gin.Default()

	// CORS middleware (applied globally)
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestLogger())
	router.Use(rateLimiter.Limit())

	// Gateway health
	router.GET("/health", handlers.HealthCheck)

	// ============================================================
	// Auth Service (PUBLIC)
	// ============================================================
	authPublic := router.Group("/api/auth")
	{
		authPublic.POST("/register", func(c *gin.Context) {
			authProxy.ProxyRequest(c, "/api")
		})

		authPublic.POST("/login", func(c *gin.Context) {
			authProxy.ProxyRequest(c, "/api")
		})
	}

	// ============================================================
	// Auth Service (PROTECTED) - User Profile Management
	// ============================================================
	authProtected := router.Group("/api/auth")
	authProtected.Use(authMiddleware.ValidateJWT())
	{
		// Verify JWT token
		authProtected.GET("/verify", func(c *gin.Context) {
			authProxy.ProxyRequest(c, "/api")
		})

		// Get user profile
		authProtected.GET("/profile", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /profile to auth-service")
			authProxy.ProxyRequest(c, "/api")
		})

		// Update user profile (username, email)
		authProtected.PUT("/profile", func(c *gin.Context) {
			log.Printf("🔄 Proxying PUT /profile to auth-service")
			authProxy.ProxyRequest(c, "/api")
		})

		// Change password
		authProtected.POST("/change-password", func(c *gin.Context) {
			log.Printf("🔄 Proxying POST /change-password to auth-service")
			authProxy.ProxyRequest(c, "/api")
		})
	}

	// ============================================================
	// Stream Service (ABR Support)
	// ============================================================
	// Public routes
	streamPublic := router.Group("/api/streams")
	{
		// List all live streams
		streamPublic.GET("/live", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		// Get stream by key (for webhook validation)
		streamPublic.GET("/by-key/:key", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		// Get stream playback info (HLS URL with ABR support - master.m3u8)
		streamPublic.GET("/:id/play", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /:id/play to stream-service (ABR)")
			streamProxy.ProxyRequest(c, "/api")
		})

		// Get stream thumbnail
		streamPublic.GET("/:id/thumbnail", func(c *gin.Context) {
			log.Printf("🔄 Proxying thumbnail request to stream-service")
			streamProxy.ProxyRequest(c, "/api")
		})

		// Get stream info by ID (includes available_qualities)
		streamPublic.GET("/:id", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		// Get available qualities for a stream
		streamPublic.GET("/:id/qualities", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /:id/qualities to stream-service")
			streamProxy.ProxyRequest(c, "/api")
		})
	}

	// Protected routes (require JWT)
	streamProtected := router.Group("/api/streams")
	streamProtected.Use(authMiddleware.ValidateJWT())
	{
		// Create new stream (with default ABR qualities)
		streamProtected.POST("", func(c *gin.Context) {
			log.Printf("🔄 Creating stream with ABR support")
			streamProxy.ProxyRequest(c, "/api")
		})

		// Get user's streams - support both paths
		streamProtected.GET("/user", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		// Alias for GET /api/streams (same as /api/streams/user)
		streamProtected.GET("", func(c *gin.Context) {
			c.Request.URL.Path = "/api/streams/user"
			streamProxy.ProxyRequest(c, "/api")
		})

		// Update stream
		streamProtected.PUT("/:id", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		// Delete stream
		streamProtected.DELETE("/:id", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		// ✅ OPTIONAL: Get stream statistics
		// streamProtected.GET("/:id/stats", func(c *gin.Context) {
		//     log.Printf("🔄 Proxying GET /:id/stats to stream-service")
		//     streamProxy.ProxyRequest(c, "/api")
		// })
	}

	// ============================================================
	// Recording Service
	// ============================================================
	recordingPublic := router.Group("/api/recordings")
	{
		// List recordings
		recordingPublic.GET("", func(c *gin.Context) {
			recordingProxy.ProxyRequest(c, "/api")
		})

		// Get recording by ID
		recordingPublic.GET("/:id", func(c *gin.Context) {
			recordingProxy.ProxyRequest(c, "/api")
		})

		// Webhook from stream-service (no auth needed)
		recordingPublic.POST("/webhook/stream", func(c *gin.Context) {
			log.Printf("🔄 Received recording webhook")
			recordingProxy.ProxyRequest(c, "/api")
		})
	}

	// ============================================================
	// VOD Service
	// ============================================================
	// Public routes
	vodPublic := router.Group("/api/videos")
	{
		// ✅ List all public videos + user's private (optional auth via cookie)
		vodPublic.GET("", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /videos to vod-service")
			vodProxy.ProxyRequest(c, "/api")
		})
		// Get video metadata (public if video is public)
		vodPublic.GET("/:id", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		// Get video stream info (returns HLS/MP4 URL)
		vodPublic.GET("/:id/stream", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		// Stream video file (actual video playback)
		vodPublic.GET("/:id/play", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		// Get video thumbnail
		vodPublic.GET("/:id/thumbnail", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		// Increment view count
		vodPublic.POST("/:id/view", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})
	}

	// Protected routes
	vodProtected := router.Group("/api/videos")
	vodProtected.Use(authMiddleware.ValidateJWT())
	{
		// Get user's videos
		vodProtected.GET("/user", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		// Import recording to VOD
		vodProtected.POST("/import-recording", func(c *gin.Context) {
			log.Printf("🔄 Importing recording to VOD")
			vodProxy.ProxyRequest(c, "/api")
		})

		// Update video metadata
		vodProtected.PUT("/:id", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		// Delete video
		vodProtected.DELETE("/:id", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		// Like video
		vodProtected.POST("/:id/like", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})
	}

	log.Printf("✅ API Gateway running on port %s", cfg.Port)
	log.Printf("🎬 ABR Support: Enabled (4 qualities: 360p-1080p)")
	log.Println("📋 Registered Routes:")
	for _, route := range router.Routes() {
		log.Printf("  %s %s", route.Method, route.Path)
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
