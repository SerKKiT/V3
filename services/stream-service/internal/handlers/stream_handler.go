package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/SerKKiT/streaming-platform/stream-service/internal/models"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/storage"
	"github.com/SerKKiT/streaming-platform/stream-service/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StreamHandler struct {
	streamRepo    *repository.StreamRepository
	srtServerAddr string
	minioStorage  *storage.MinIOStorage
	publicBaseURL string
}

// NewStreamHandler - ОБНОВЛЕННАЯ СИГНАТУРА
func NewStreamHandler(
	streamRepo *repository.StreamRepository,
	srtServerAddr string,
	minioStorage *storage.MinIOStorage,
	publicBaseURL string,
) *StreamHandler {
	return &StreamHandler{
		streamRepo:    streamRepo,
		srtServerAddr: srtServerAddr,
		minioStorage:  minioStorage,
		publicBaseURL: publicBaseURL,
	}
}

// CreateStream creates a new stream for the user
func (h *StreamHandler) CreateStream(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req models.CreateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	streamKey, err := utils.GenerateStreamKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate stream key"})
		return
	}

	stream, err := h.streamRepo.CreateStream(userID, streamKey, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	response := models.CreateStreamResponse{
		Stream:    stream,
		StreamURL: h.buildSRTURL(streamKey),
		HLSURL:    h.buildMinIOHLSURL(streamKey),
	}

	c.JSON(http.StatusCreated, response)
}

// GetStreamPlaybackInfo returns HLS URL for public viewing
func (h *StreamHandler) GetStreamPlaybackInfo(c *gin.Context) {
	streamID := c.Param("id")

	id, err := uuid.Parse(streamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid stream ID"})
		return
	}

	stream, err := h.streamRepo.GetStreamByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Stream not found"})
		return
	}

	if stream.Status != "live" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Stream is not currently live"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream_id":           stream.ID,
		"title":               stream.Title,
		"description":         stream.Description,
		"username":            stream.Username, // ✅ ДОБАВЛЕНО
		"status":              stream.Status,
		"hls_url":             h.buildMinIOHLSURL(stream.StreamKey),
		"viewer_count":        stream.ViewerCount,
		"started_at":          stream.StartedAt,
		"thumbnail_url":       stream.ThumbnailURL,
		"available_qualities": stream.AvailableQualities,
		"is_live":             true,
	})
}

// GetStream retrieves stream by ID
func (h *StreamHandler) GetStream(c *gin.Context) {
	streamIDStr := c.Param("id")
	streamID, err := uuid.Parse(streamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid stream ID"})
		return
	}

	stream, err := h.streamRepo.GetStreamByID(streamID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Stream not found"})
		return
	}

	c.JSON(http.StatusOK, stream)
}

// GetUserStreams retrieves all streams for authenticated user
func (h *StreamHandler) GetUserStreams(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	log.Printf("📋 Getting streams for user: %s", userID)

	streams, err := h.streamRepo.GetUserStreams(userID)
	if err != nil {
		log.Printf("❌ Failed to get user streams: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	log.Printf("✅ Found %d streams for user %s", len(streams), userID)
	c.JSON(http.StatusOK, gin.H{
		"streams": streams,
		"total":   len(streams),
	})
}

// DeleteStream deletes a stream and its HLS files
func (h *StreamHandler) DeleteStream(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	streamIDStr := c.Param("id")

	streamID, err := uuid.Parse(streamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid stream ID"})
		return
	}

	// ✅ Получаем stream перед удалением чтобы узнать stream_key
	stream, err := h.streamRepo.GetStreamByID(streamID)
	if err != nil {
		log.Printf("❌ Failed to get stream: %v", err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Stream not found"})
		return
	}

	// ✅ Проверяем ownership
	if stream.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Not authorized to delete this stream"})
		return
	}

	log.Printf("🗑️  Deleting stream %s (key: %s) for user %s", streamID, stream.StreamKey, userID)

	// ✅ Удаляем из БД
	err = h.streamRepo.DeleteStream(streamID, userID)
	if err != nil {
		log.Printf("❌ Failed to delete stream from DB: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// ✅ Удаляем HLS файлы из MinIO (в фоне, не блокируем ответ)
	go func() {
		ctx := context.Background()
		hlsFolder := fmt.Sprintf("live-segments/%s/", stream.StreamKey)

		if err := h.minioStorage.DeleteFolder(ctx, hlsFolder); err != nil {
			log.Printf("❌ Failed to delete HLS files for stream %s: %v", stream.StreamKey, err)
		} else {
			log.Printf("✅ Deleted HLS files for stream %s", stream.StreamKey)
		}
	}()

	log.Printf("✅ Stream %s deleted successfully", streamID)
	c.JSON(http.StatusOK, gin.H{"message": "Stream deleted successfully"})
}

// GetLiveStreams returns all live streams
func (h *StreamHandler) GetLiveStreams(c *gin.Context) {
	streams, err := h.streamRepo.GetLiveStreams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"streams": streams})
}

// UpdateStream updates stream title and description
func (h *StreamHandler) UpdateStream(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	streamIDStr := c.Param("id")

	streamID, err := uuid.Parse(streamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid stream ID"})
		return
	}

	stream, err := h.streamRepo.GetStreamByID(streamID)
	if err != nil {
		log.Printf("❌ Failed to get stream: %v", err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Stream not found"})
		return
	}

	if stream.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Not authorized to update this stream"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Failed to parse request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	log.Printf("📝 Updating stream %s: title=%s, description=%s", streamID, req.Title, req.Description)

	stream.Title = req.Title
	stream.Description = req.Description

	if err := h.streamRepo.UpdateStream(stream); err != nil {
		log.Printf("❌ Failed to update stream in DB: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update stream"})
		return
	}

	log.Printf("✅ Stream %s updated successfully", streamID)
	c.JSON(http.StatusOK, gin.H{"stream": stream})
}

// GetStreamByKey retrieves stream by stream_key
func (h *StreamHandler) GetStreamByKey(c *gin.Context) {
	streamKey := c.Param("key")
	if streamKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stream key is required"})
		return
	}

	stream, err := h.streamRepo.GetStreamByKey(streamKey)
	if err != nil {
		log.Printf("❌ Stream not found by key: %s", streamKey)
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}

	c.JSON(http.StatusOK, stream)
}

// Health check
func (h *StreamHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// GetStreamThumbnail streams thumbnail from MinIO
func (h *StreamHandler) GetStreamThumbnail(c *gin.Context) {
	streamID := c.Param("id")

	log.Printf("📸 Thumbnail request for stream: %s", streamID)

	id, err := uuid.Parse(streamID)
	if err != nil {
		log.Printf("❌ Invalid stream ID: %s", streamID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stream ID"})
		return
	}

	stream, err := h.streamRepo.GetStreamByID(id)
	if err != nil {
		log.Printf("❌ Stream not found: %s", streamID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}

	objectName := filepath.Join("live-segments", stream.StreamKey, "thumbnail.jpg")

	log.Printf("✅ Streaming thumbnail from MinIO: %s", objectName)

	ctx := c.Request.Context()
	object, err := h.minioStorage.GetObject(ctx, objectName)
	if err != nil {
		log.Printf("❌ Failed to get thumbnail: %v", err)
		c.Status(http.StatusNotFound)
		return
	}
	defer object.Close()

	// ✅ Короткий cache для live thumbnails (30 секунд)
	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=30")                                    // Обновляется каждые 30 секунд
	c.Header("ETag", fmt.Sprintf("\"%s-%d\"", stream.StreamKey, time.Now().Unix()/30)) // ETag меняется каждые 30 секунд
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, object)
	if err != nil {
		log.Printf("❌ Failed to stream thumbnail: %v", err)
	}
}

// Helper functions
func (h *StreamHandler) buildSRTURL(streamKey string) string {
	return "srt://" + h.srtServerAddr + "?streamid=" + streamKey
}

func (h *StreamHandler) buildMinIOHLSURL(streamKey string) string {
	// ✅ Возвращаем master.m3u8 для ABR
	return fmt.Sprintf("%s/live-streams/live-segments/%s/master.m3u8",
		h.publicBaseURL, streamKey)
}

// В stream_handler.go добавить:
func (h *StreamHandler) GetStreamQualities(c *gin.Context) {
	streamID := c.Param("id")

	id, err := uuid.Parse(streamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stream ID"})
		return
	}

	stream, err := h.streamRepo.GetStreamByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream_id":           stream.ID,
		"available_qualities": stream.AvailableQualities,
		"status":              stream.Status,
	})
}
