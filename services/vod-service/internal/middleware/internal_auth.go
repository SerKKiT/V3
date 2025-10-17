package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// InternalAuth проверяет Internal-API-Key header для service-to-service запросов
func InternalAuth() gin.HandlerFunc {
	internalAPIKey := os.Getenv("INTERNAL_API_KEY")
	if internalAPIKey == "" {
		internalAPIKey = "default-internal-key-change-me" // Fallback для dev
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Internal-API-Key")

		if apiKey != internalAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
