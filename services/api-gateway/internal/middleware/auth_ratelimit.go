package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthRateLimiter для защиты auth endpoints от brute-force
type AuthRateLimiter struct {
	visitors map[string]*AuthVisitor
	mu       sync.RWMutex

	// Настройки
	maxAttempts    int
	windowDuration time.Duration
	banDuration    time.Duration
}

type AuthVisitor struct {
	attempts     int
	firstAttempt time.Time
	bannedUntil  time.Time
}

// NewAuthRateLimiter создаёт rate limiter для auth endpoints
func NewAuthRateLimiter() *AuthRateLimiter {
	limiter := &AuthRateLimiter{
		visitors:       make(map[string]*AuthVisitor),
		maxAttempts:    5,
		windowDuration: 1 * time.Minute,
		banDuration:    15 * time.Minute,
	}

	go limiter.cleanupVisitors()

	return limiter
}

// Limit - middleware для rate limiting auth endpoints
func (a *AuthRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		a.mu.Lock()
		visitor, exists := a.visitors[ip]

		if !exists {
			visitor = &AuthVisitor{
				attempts:     1,
				firstAttempt: time.Now(),
			}
			a.visitors[ip] = visitor
			a.mu.Unlock()

			c.Next()
			return
		}

		// Проверка: забанен ли IP
		if time.Now().Before(visitor.bannedUntil) {
			retryAfter := int(time.Until(visitor.bannedUntil).Seconds())
			a.mu.Unlock()

			log.Printf("🚫 Auth rate limit: IP %s is banned (retry after %ds)", ip, retryAfter)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many failed attempts. Please try again later.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		// Проверка: истёк ли временной интервал
		if time.Since(visitor.firstAttempt) > a.windowDuration {
			visitor.attempts = 1
			visitor.firstAttempt = time.Now()
			a.mu.Unlock()

			c.Next()
			return
		}

		// Инкремент попыток
		visitor.attempts++

		// Проверка: превышен ли лимит
		if visitor.attempts > a.maxAttempts {
			visitor.bannedUntil = time.Now().Add(a.banDuration)
			a.mu.Unlock()

			log.Printf("⚠️ Auth rate limit: IP %s BANNED after %d attempts", ip, visitor.attempts)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many failed attempts. IP temporarily banned.",
				"retry_after": int(a.banDuration.Seconds()),
			})
			c.Abort()
			return
		}

		a.mu.Unlock()
		c.Next()
	}
}

// cleanupVisitors удаляет старые записи
func (a *AuthRateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()

		now := time.Now()
		for ip, visitor := range a.visitors {
			if now.After(visitor.bannedUntil) &&
				now.Sub(visitor.firstAttempt) > 30*time.Minute {
				delete(a.visitors, ip)
			}
		}

		a.mu.Unlock()
	}
}
