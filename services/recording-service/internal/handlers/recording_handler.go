package handlers

import (
	"log"
	"net/http"

	"github.com/SerKKiT/streaming-platform/recording-service/internal/repository"
	"github.com/gin-gonic/gin"
)

type RecordingHandler struct {
	recordingRepo *repository.RecordingRepository
}

func NewRecordingHandler(recordingRepo *repository.RecordingRepository) *RecordingHandler {
	return &RecordingHandler{
		recordingRepo: recordingRepo,
	}
}

// GetAllRecordings возвращает список всех записей
func (h *RecordingHandler) GetAllRecordings(c *gin.Context) {
	recordings, err := h.recordingRepo.GetAllRecordings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recordings": recordings})
}

// Health check
func (h *RecordingHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// GetRecordingByID - получить запись по ID
func (h *RecordingHandler) GetRecordingByID(c *gin.Context) {
	recordingID := c.Param("id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recording ID is required"})
		return
	}

	recording, err := h.recordingRepo.GetByID(recordingID)
	if err != nil {
		log.Printf("❌ Failed to get recording %s: %v", recordingID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Recording not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recording": recording})
}
