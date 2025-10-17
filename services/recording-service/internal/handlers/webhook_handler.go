package handlers

import (
	"log"
	"net/http"

	"github.com/SerKKiT/streaming-platform/recording-service/internal/monitor"
	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	streamMonitor *monitor.StreamMonitor
}

func NewWebhookHandler(streamMonitor *monitor.StreamMonitor) *WebhookHandler {
	return &WebhookHandler{
		streamMonitor: streamMonitor,
	}
}

type StreamEventPayload struct {
	StreamKey string `json:"stream_key"`
	Event     string `json:"event"`
	HLSURL    string `json:"hls_url"`
	Timestamp int64  `json:"timestamp"`
}

func (h *WebhookHandler) HandleStreamEvent(c *gin.Context) {
	var payload StreamEventPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Invalid webhook payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	log.Printf("Received webhook: stream=%s, event=%s, hls=%s", payload.StreamKey, payload.Event, payload.HLSURL)

	switch payload.Event {
	case "started":
		if err := h.streamMonitor.HandleWebhookStart(payload.StreamKey, payload.HLSURL); err != nil {
			log.Printf("❌ Failed to start recording: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start recording"})
			return
		}
		log.Printf("✅ Recording started for stream: %s", payload.StreamKey)

	case "stopped":
		if err := h.streamMonitor.HandleWebhookStop(payload.StreamKey); err != nil {
			log.Printf("❌ Failed to stop recording: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop recording"})
			return
		}
		log.Printf("✅ Recording stopped for stream: %s", payload.StreamKey)

	default:
		log.Printf("⚠️ Unknown webhook event: %s", payload.Event)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
