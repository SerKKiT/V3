package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SerKKiT/streaming-platform/stream-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/storage"
	"github.com/gin-gonic/gin"
)

type CleanupHandler struct {
	streamRepo   *repository.StreamRepository
	minioStorage *storage.MinIOStorage
	outputDir    string
}

func NewCleanupHandler(
	streamRepo *repository.StreamRepository,
	minioStorage *storage.MinIOStorage,
	outputDir string,
) *CleanupHandler {
	return &CleanupHandler{
		streamRepo:   streamRepo,
		minioStorage: minioStorage,
		outputDir:    outputDir,
	}
}

// RecordingCompleteRequest - запрос от recording-service
type RecordingCompleteRequest struct {
	StreamKey string `json:"stream_key" binding:"required"`
	StreamID  string `json:"stream_id" binding:"required"`
	VideoID   string `json:"video_id"`
	Success   bool   `json:"success"`
}

// HandleRecordingComplete обрабатывает webhook после завершения конвертации
func (h *CleanupHandler) HandleRecordingComplete(c *gin.Context) {
	var req RecordingCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("📩 Received recording complete webhook for stream %s (success: %v)", req.StreamKey, req.Success)

	if !req.Success {
		log.Printf("⚠️ Recording failed for stream %s, skipping cleanup", req.StreamKey)
		c.JSON(http.StatusOK, gin.H{"message": "Recording failed, cleanup skipped"})
		return
	}

	// 1. Удалить файлы из MinIO
	if err := h.minioStorage.DeleteStreamSegments(req.StreamKey); err != nil {
		log.Printf("❌ Failed to delete MinIO segments for %s: %v", req.StreamKey, err)
	}

	// 2. Удалить локальные файлы
	localPath := filepath.Join(h.outputDir, req.StreamKey)
	if err := os.RemoveAll(localPath); err != nil {
		log.Printf("❌ Failed to delete local files for %s: %v", req.StreamKey, err)
	} else {
		log.Printf("✅ Deleted local files: %s", localPath)
	}

	log.Printf("✅ Cleanup completed for stream %s", req.StreamKey)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Cleanup completed",
		"stream_key": req.StreamKey,
	})
}
