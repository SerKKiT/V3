package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServiceProxy struct {
	targetURL string
}

func NewServiceProxy(targetURL string) *ServiceProxy {
	return &ServiceProxy{targetURL: targetURL}
}

func (p *ServiceProxy) ProxyRequest(c *gin.Context, stripPrefix string) {
	// Убираем prefix из пути
	targetPath := strings.TrimPrefix(c.Request.URL.Path, stripPrefix)

	// Формируем полный URL
	targetURL := p.targetURL + targetPath

	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	log.Printf("🔄 Proxying: %s %s -> %s", c.Request.Method, c.Request.URL.Path, targetURL)

	// Создаём новый запрос
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		log.Printf("❌ Failed to create proxy request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}

	// Копируем headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Добавляем user context из JWT (если есть)
	if userID, exists := c.Get("user_id"); exists {
		userIDStr := convertToString(userID)
		if userIDStr != "" {
			req.Header.Set("X-User-ID", userIDStr)
			log.Printf("📤 Added X-User-ID: %s", userIDStr)
		}
	}

	if username, exists := c.Get("username"); exists {
		usernameStr := convertToString(username)
		if usernameStr != "" {
			req.Header.Set("X-Username", usernameStr)
		}
	}

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to proxy request to %s: %v", targetURL, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	log.Printf("✅ Proxy response: %d from %s", resp.StatusCode, targetURL)

	// Копируем response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// Возвращаем response
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

// convertToString безопасно конвертирует interface{} в string
func convertToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
