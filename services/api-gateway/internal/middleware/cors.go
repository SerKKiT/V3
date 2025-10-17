package middleware

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig содержит настройки CORS
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// CORSMiddlewareWithConfig создаёт Gin middleware с whitelist origins
func CORSMiddlewareWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Проверка origin в whitelist
		allowedOrigin := ""
		for _, allowed := range config.AllowedOrigins {
			if origin == allowed {
				allowedOrigin = origin
				break
			}
			// Поддержка wildcard для субдоменов: *.example.com
			if strings.HasPrefix(allowed, "*.") {
				domain := strings.TrimPrefix(allowed, "*")
				if strings.HasSuffix(origin, domain) {
					allowedOrigin = origin
					break
				}
			}
		}

		// Логировать блокированные запросы
		if origin != "" && allowedOrigin == "" {
			log.Printf("⚠️ CORS: Blocked request from unauthorized origin: %s (Path: %s)", origin, c.Request.URL.Path)
		}

		// Устанавливать заголовки только для разрешённых origins
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)

			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			}
		}

		// Обработка preflight request
		if c.Request.Method == "OPTIONS" {
			if allowedOrigin != "" {
				c.AbortWithStatus(200)
			} else {
				c.AbortWithStatus(403)
			}
			return
		}

		c.Next()
	}
}

// CORSMiddleware - простой middleware с wildcard (для обратной совместимости)
// DEPRECATED: Используйте CORSMiddlewareWithConfig для production
func CORSMiddleware() gin.HandlerFunc {
	log.Println("⚠️ Warning: Using legacy CORSMiddleware with wildcard origin (*). Use CORSMiddlewareWithConfig instead.")
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Internal-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
