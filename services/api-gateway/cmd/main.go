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
	authRateLimiter := middleware.NewAuthRateLimiter() // ✅ НОВЫЙ
	validator := middleware.NewValidator()             // ✅ НОВЫЙ

	// Initialize proxies
	authProxy := proxy.NewServiceProxy(cfg.Services.AuthURL)
	streamProxy := proxy.NewServiceProxy(cfg.Services.StreamURL)
	recordingProxy := proxy.NewServiceProxy(cfg.Services.RecordingURL)
	vodProxy := proxy.NewServiceProxy(cfg.Services.VODURL)

	router := gin.Default()

	// CORS Configuration
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-User-ID",
			"X-Internal-API-Key",
		},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           3600,
	}

	// Global middleware
	router.Use(middleware.CORSMiddlewareWithConfig(corsConfig))
	router.Use(middleware.RequestLogger())
	router.Use(rateLimiter.Limit())

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// ============================================================
	// Auth Service (PUBLIC) - WITH STRICT RATE LIMITING
	// ============================================================
	authPublic := router.Group("/api/auth")
	authPublic.Use(authRateLimiter.Limit()) // ✅ Строгий rate limit
	{
		authPublic.POST("/register",
			validator.ValidateAuthInput(), // ✅ Validation
			func(c *gin.Context) {
				authProxy.ProxyRequest(c, "/api")
			},
		)

		authPublic.POST("/login",
			validator.ValidateAuthInput(), // ✅ Validation
			func(c *gin.Context) {
				authProxy.ProxyRequest(c, "/api")
			},
		)
	}

	// ============================================================
	// Auth Service (PROTECTED)
	// ============================================================
	authProtected := router.Group("/api/auth")
	authProtected.Use(authMiddleware.ValidateJWT())
	{
		authProtected.GET("/verify", func(c *gin.Context) {
			authProxy.ProxyRequest(c, "/api")
		})

		authProtected.GET("/profile", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /profile to auth-service")
			authProxy.ProxyRequest(c, "/api")
		})

		authProtected.PUT("/profile",
			validator.ValidateAuthInput(), // ✅ Validation
			func(c *gin.Context) {
				log.Printf("🔄 Proxying PUT /profile to auth-service")
				authProxy.ProxyRequest(c, "/api")
			},
		)

		authProtected.POST("/change-password", func(c *gin.Context) {
			log.Printf("🔄 Proxying POST /change-password to auth-service")
			authProxy.ProxyRequest(c, "/api")
		})
	}

	// ============================================================
	// Stream Service
	// ============================================================
	streamPublic := router.Group("/api/streams")
	{
		streamPublic.GET("/live", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		streamPublic.GET("/by-key/:key", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		streamPublic.GET("/:id/play", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /:id/play to stream-service (ABR)")
			streamProxy.ProxyRequest(c, "/api")
		})

		streamPublic.GET("/:id/thumbnail", func(c *gin.Context) {
			log.Printf("🔄 Proxying thumbnail request to stream-service")
			streamProxy.ProxyRequest(c, "/api")
		})

		streamPublic.GET("/:id", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		streamPublic.GET("/:id/qualities", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /:id/qualities to stream-service")
			streamProxy.ProxyRequest(c, "/api")
		})
	}

	streamProtected := router.Group("/api/streams")
	streamProtected.Use(authMiddleware.ValidateJWT())
	{
		streamProtected.POST("",
			validator.ValidateStreamInput(), // ✅ Validation
			func(c *gin.Context) {
				log.Printf("🔄 Creating stream with ABR support")
				streamProxy.ProxyRequest(c, "/api")
			},
		)

		streamProtected.GET("/user", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})

		streamProtected.GET("", func(c *gin.Context) {
			c.Request.URL.Path = "/api/streams/user"
			streamProxy.ProxyRequest(c, "/api")
		})

		streamProtected.PUT("/:id",
			validator.ValidateStreamInput(), // ✅ Validation
			func(c *gin.Context) {
				streamProxy.ProxyRequest(c, "/api")
			},
		)

		streamProtected.DELETE("/:id", func(c *gin.Context) {
			streamProxy.ProxyRequest(c, "/api")
		})
	}

	// ============================================================
	// Recording Service
	// ============================================================
	recordingPublic := router.Group("/api/recordings")
	{
		recordingPublic.GET("", func(c *gin.Context) {
			recordingProxy.ProxyRequest(c, "/api")
		})

		recordingPublic.GET("/:id", func(c *gin.Context) {
			recordingProxy.ProxyRequest(c, "/api")
		})

		recordingPublic.POST("/webhook/stream", func(c *gin.Context) {
			log.Printf("🔄 Received recording webhook")
			recordingProxy.ProxyRequest(c, "/api")
		})
	}

	// ============================================================
	// VOD Service
	// ============================================================
	vodPublic := router.Group("/api/videos")
	{
		vodPublic.GET("", func(c *gin.Context) {
			log.Printf("🔄 Proxying GET /videos to vod-service")
			vodProxy.ProxyRequest(c, "/api")
		})

		vodPublic.GET("/:id", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		vodPublic.GET("/:id/stream", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		vodPublic.GET("/:id/play", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		vodPublic.GET("/:id/thumbnail", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		vodPublic.POST("/:id/view", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})
	}

	vodProtected := router.Group("/api/videos")
	vodProtected.Use(authMiddleware.ValidateJWT())
	{
		vodProtected.GET("/user", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		vodProtected.POST("/import-recording", func(c *gin.Context) {
			log.Printf("🔄 Importing recording to VOD")
			vodProxy.ProxyRequest(c, "/api")
		})

		vodProtected.PUT("/:id",
			validator.ValidateStreamInput(), // ✅ Validation
			func(c *gin.Context) {
				vodProxy.ProxyRequest(c, "/api")
			},
		)

		vodProtected.DELETE("/:id", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})

		vodProtected.POST("/:id/like", func(c *gin.Context) {
			vodProxy.ProxyRequest(c, "/api")
		})
	}

	log.Printf("✅ API Gateway running on port %s", cfg.Port)
	log.Printf("🛡️ Auth Rate Limiting: 5 attempts/minute, 15min ban after exceed")
	log.Printf("✅ Input Validation: Enabled (XSS protection, length limits)")
	log.Printf("🎬 ABR Support: Enabled (4 qualities: 360p-1080p)")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
