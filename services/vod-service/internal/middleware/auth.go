package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// AuthMiddleware проверяет JWT из Authorization header ИЛИ cookie
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Попытка 1: Прочитать из Authorization header
		authHeader := c.GetHeader("Authorization")
		tokenString := ""

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Попытка 2: Прочитать из cookie
		if tokenString == "" {
			cookieToken, err := c.Cookie("auth_token")
			if err == nil && cookieToken != "" {
				tokenString = cookieToken
				log.Printf("🍪 Using token from cookie")
			}
		}

		// Если токена нет - ошибка
		if tokenString == "" {
			log.Printf("⛔ No token provided (header or cookie)")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Валидация JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			log.Printf("⛔ Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Извлекаем claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
			log.Printf("✅ Authenticated user: %s (via %s)", claims["username"],
				func() string {
					if authHeader != "" {
						return "header"
					}
					return "cookie"
				}())
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
		}
	}
}

// OptionalAuthMiddleware - извлекает user_id если токен есть, но не требует его
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Попытка 1: Authorization header
		authHeader := c.GetHeader("Authorization")
		tokenString := ""

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Попытка 2: Cookie
		if tokenString == "" {
			cookieToken, err := c.Cookie("auth_token")
			if err == nil && cookieToken != "" {
				tokenString = cookieToken
			}
		}

		// Если токена нет - просто продолжаем без auth
		if tokenString == "" {
			c.Next()
			return
		}

		// Валидация JWT (если есть)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user_id", claims["user_id"])
				c.Set("username", claims["username"])
			}
		}

		c.Next()
	}
}
