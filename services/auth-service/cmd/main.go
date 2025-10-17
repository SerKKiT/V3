package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/SerKKiT/streaming-platform/auth-service/internal/config"
	"github.com/SerKKiT/streaming-platform/auth-service/internal/handlers"
	"github.com/SerKKiT/streaming-platform/auth-service/internal/middleware"
	"github.com/SerKKiT/streaming-platform/auth-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/auth-service/internal/service"
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

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Initialize repository, service, and handler
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Gin router
	router := gin.Default()
	//router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", authHandler.Health)

	// Auth routes - ИСПРАВЛЕНО: добавили /auth prefix
	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)

	// Protected routes
	protected := router.Group("/auth")
	protected.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/verify", authHandler.Verify)
		protected.GET("/profile", authHandler.GetProfile)              // ✅ Get profile
		protected.PUT("/profile", authHandler.UpdateProfile)           // ✅ Update profile
		protected.POST("/change-password", authHandler.ChangePassword) // ✅ Change password
	}

	// Start server
	log.Printf("🚀 Auth service starting on port %s", cfg.Port)
	log.Println("📋 Routes:")
	log.Println("  POST   /auth/register")
	log.Println("  POST   /auth/login")
	log.Println("  GET    /auth/verify (protected)")
	log.Println("  GET    /health")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
