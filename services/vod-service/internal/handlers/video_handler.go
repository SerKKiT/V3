package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/SerKKiT/streaming-platform/vod-service/internal/models"
	"github.com/SerKKiT/streaming-platform/vod-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/vod-service/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VideoHandler struct {
	repo                *repository.VideoRepository
	storage             *storage.MinIOStorage
	recordingServiceURL string
	recordingBucket     string
	vodBucket           string
}

func NewVideoHandler(
	repo *repository.VideoRepository,
	storage *storage.MinIOStorage,
	recordingServiceURL string,
	recordingBucket string,
	vodBucket string,
) *VideoHandler {
	return &VideoHandler{
		repo:                repo,
		storage:             storage,
		recordingServiceURL: recordingServiceURL,
		recordingBucket:     recordingBucket,
		vodBucket:           vodBucket,
	}
}

// getUserID извлекает user_id из контекста, заголовка или JWT токена
// getUserID извлекает user_id из контекста или заголовка
func getUserID(c *gin.Context) string {
	// 1. Из контекста (установлен middleware)
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok && uid != "" {
			return uid
		}
	}

	// 2. Из заголовка X-User-ID (от API Gateway)
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}

	// Нет авторизации
	return ""
}

// ImportRecording импортирует запись из Recording Service
func (h *VideoHandler) ImportRecording(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		log.Println("❌ ImportRecording: missing user_id")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.ImportRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	recordingID, err := uuid.Parse(req.RecordingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recording ID"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("📥 Importing recording %s for user %s", recordingID, userID)

	// Проверяем дубликаты
	existingVideo, _ := h.repo.GetByRecordingID(recordingID)
	if existingVideo != nil {
		log.Printf("⚠️ Recording %s already imported as video %s", recordingID, existingVideo.ID)
		c.JSON(http.StatusOK, gin.H{
			"video_id": existingVideo.ID,
			"message":  "Recording already imported",
		})
		return
	}

	// Получаем информацию о recording
	recording, err := h.getRecordingInfo(recordingID)
	if err != nil {
		log.Printf("❌ Failed to get recording info: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Recording not found"})
		return
	}

	log.Printf("🔍 Recording info: FilePath=%s, ThumbnailPath=%s", recording.FilePath, recording.ThumbnailPath)

	// Генерируем уникальное имя для видео
	videoID := uuid.New()
	videoFileName := fmt.Sprintf("%s.mp4", videoID.String())

	// Копируем видео из recordings в vod-videos bucket
	ctx := context.Background()
	if err := h.storage.CopyFromRecordings(ctx, h.recordingBucket, recording.FilePath, videoFileName); err != nil {
		log.Printf("❌ Failed to copy video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy recording"})
		return
	}

	log.Printf("✅ Video copied: %s -> %s", recording.FilePath, videoFileName)

	// Копируем thumbnail если есть
	var thumbnailFileName string
	if recording.ThumbnailPath != "" && recording.ThumbnailPath != "null" {
		thumbnailFileName = fmt.Sprintf("%s.jpg", videoID.String())
		log.Printf("📋 Copying thumbnail from recordings/%s to vod-videos/%s", recording.ThumbnailPath, thumbnailFileName)

		if err := h.storage.CopyFromRecordings(ctx, h.recordingBucket, recording.ThumbnailPath, thumbnailFileName); err != nil {
			log.Printf("⚠️ Failed to copy thumbnail (non-critical): %v", err)
			thumbnailFileName = ""
		} else {
			log.Printf("✅ Thumbnail copied: %s -> %s", recording.ThumbnailPath, thumbnailFileName)
		}
	} else {
		log.Printf("ℹ️ No thumbnail available for recording %s", recordingID)
	}

	// Устанавливаем visibility
	visibility := req.Visibility
	if visibility == "" {
		visibility = "public" // По умолчанию public
	}

	// Создаём video
	video := &models.Video{
		ID:            videoID,
		UserID:        userUUID,
		RecordingID:   &recordingID,
		StreamID:      recording.StreamID,
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Tags:          req.Tags,
		Source:        "recording",
		Status:        "ready",
		Visibility:    visibility,
		FilePath:      videoFileName,
		ThumbnailPath: thumbnailFileName,
		Duration:      recording.Duration,
		FileSize:      recording.FileSize,
		ViewCount:     0,
		LikeCount:     0,
		CreatedAt:     recording.StartedAt,
		UpdatedAt:     time.Now(),
	}

	if err := h.repo.Create(video); err != nil {
		log.Printf("❌ Failed to create video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import recording"})
		return
	}

	log.Printf("✅ Recording imported as video: %s -> %s", recordingID, video.ID)
	log.Printf("📁 File location: vod-videos/%s", videoFileName)
	if thumbnailFileName != "" {
		log.Printf("🖼️ Thumbnail location: vod-videos/%s", thumbnailFileName)
	}

	c.JSON(http.StatusCreated, gin.H{
		"video_id": video.ID,
		"message":  "Recording imported successfully",
	})
}

// GetUserVideos возвращает все видео пользователя
func (h *VideoHandler) GetUserVideos(c *gin.Context) {
	userID := getUserID(c)

	if userID == "" {
		log.Println("❌ GetUserVideos: missing user_id")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	log.Printf("📹 Received request for user videos: X-User-ID=%s", userID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("❌ Invalid user ID format: %s, error: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	log.Printf("📹 Getting videos for user: %s (limit=%d, offset=%d)", userID, limit, offset)

	videos, total, err := h.repo.ListUserVideos(userUUID, limit, offset)
	if err != nil {
		log.Printf("❌ Failed to get user videos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch videos"})
		return
	}

	page := offset/limit + 1

	log.Printf("✅ Found %d videos for user %s", total, userID)

	c.JSON(http.StatusOK, models.VideoListResponse{
		Videos: videos,
		Total:  total,
		Page:   page,
		Limit:  limit,
	})
}

// GetVideo возвращает одно видео по ID
// ✅ ИСПРАВЛЕНО: Проверка доступа с учетом отсутствия JWT
func (h *VideoHandler) GetVideo(c *gin.Context) {
	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.repo.GetByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// ✅ ИСПРАВЛЕНО: Проверяем права доступа к приватному видео
	if video.Visibility == "private" {
		// Получаем user_id (может быть пустым для неавторизованных)
		userID := getUserID(c)

		// Если нет JWT или это не владелец - запретить
		if userID == "" || userID != video.UserID.String() {
			log.Printf("⛔ Access denied to private video %s: requester=%s, owner=%s",
				videoID, userID, video.UserID.String())
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "This video is private",
				"message": "Only the owner can view this video",
			})
			return
		}

		log.Printf("✅ Owner access granted to private video %s", videoID)
	}

	log.Printf("✅ Returning video %s (visibility=%s)", videoID, video.Visibility)
	c.JSON(http.StatusOK, gin.H{"video": video})
}

// UpdateVideo обновляет метаданные видео
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.repo.GetByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	if video.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	var req models.UpdateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Обновляем поля
	if req.Title != "" {
		video.Title = req.Title
	}
	if req.Description != "" {
		video.Description = req.Description
	}
	if req.Category != "" {
		video.Category = req.Category
	}
	if len(req.Tags) > 0 {
		video.Tags = req.Tags
	}
	if req.Visibility != "" {
		video.Visibility = req.Visibility
	}

	if err := h.repo.Update(video); err != nil {
		log.Printf("❌ Failed to update video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video updated successfully"})
}

// DeleteVideo удаляет видео
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.repo.GetByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	if video.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Удаляем файл из MinIO
	ctx := context.Background()
	if err := h.storage.DeleteObject(ctx, video.FilePath); err != nil {
		log.Printf("⚠️ Failed to delete file from MinIO: %v", err)
	}

	// Удаляем thumbnail
	if video.ThumbnailPath != "" {
		if err := h.storage.DeleteObject(ctx, video.ThumbnailPath); err != nil {
			log.Printf("⚠️ Failed to delete thumbnail from MinIO: %v", err)
		}
	}

	// Удаляем из БД
	if err := h.repo.Delete(videoID); err != nil {
		log.Printf("❌ Failed to delete video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video deleted successfully"})
}

// IncrementView увеличивает счётчик просмотров
func (h *VideoHandler) IncrementView(c *gin.Context) {
	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	if err := h.repo.IncrementViewCount(videoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to increment view count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "View count incremented"})
}

// LikeVideo увеличивает счётчик лайков
func (h *VideoHandler) LikeVideo(c *gin.Context) {
	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	if err := h.repo.IncrementLikeCount(videoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video liked"})
}

// GetStreamURL возвращает URL для воспроизведения через API
func (h *VideoHandler) GetStreamURL(c *gin.Context) {
	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	// Получаем видео из БД
	video, err := h.repo.GetByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// ✅ ИСПРАВЛЕНО: Проверяем права доступа
	if video.Visibility == "private" {
		userID := getUserID(c)
		if userID == "" || userID != video.UserID.String() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Возвращаем URL эндпоинтов
	videoURL := fmt.Sprintf("http://localhost/api/videos/%s/play", video.ID.String())
	thumbnailURL := ""
	if video.ThumbnailPath != "" {
		thumbnailURL = fmt.Sprintf("http://localhost/api/videos/%s/thumbnail", video.ID.String())
	}

	c.JSON(http.StatusOK, gin.H{
		"video_url":     videoURL,
		"thumbnail_url": thumbnailURL,
		"video": gin.H{
			"id":          video.ID,
			"title":       video.Title,
			"description": video.Description,
			"duration":    video.Duration,
			"view_count":  video.ViewCount,
			"like_count":  video.LikeCount,
			"created_at":  video.CreatedAt,
			"visibility":  video.Visibility,
			"tags":        video.Tags,
			"category":    video.Category,
		},
	})
}

// StreamVideoFile streams video file directly with auth check
func (h *VideoHandler) StreamVideoFile(c *gin.Context) {
	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.repo.GetByID(videoID)
	if err != nil {
		log.Printf("❌ Video not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// Check access for private videos
	if video.Visibility == "private" {
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			log.Printf("⛔ No auth provided for private video %s", videoID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		userID := userIDInterface.(string)
		if userID != video.UserID.String() {
			log.Printf("⛔ User %s tried to access private video of %s", userID, video.UserID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		log.Printf("✅ Access granted to owner %s for private video %s", userID, videoID)
	}

	// ✅ Stream file directly from MinIO
	ctx := c.Request.Context()
	object, err := h.storage.GetObject(ctx, h.vodBucket, video.FilePath) // ✅ Передаем bucket
	if err != nil {
		log.Printf("❌ Failed to get object from MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream video"})
		return
	}
	defer object.Close()

	// Get file info
	stat, err := object.Stat()
	if err != nil {
		log.Printf("❌ Failed to stat object: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video info"})
		return
	}

	log.Printf("✅ Streaming video %s directly (size: %d bytes)", videoID, stat.Size)

	// Set headers for video streaming
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=31536000")

	// CORS for video element
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Credentials", "true")

	// Stream to client
	_, err = io.Copy(c.Writer, object)
	if err != nil {
		log.Printf("⚠️ Error streaming video: %v", err)
	}
}

// StreamThumbnail streams thumbnail file directly with auth check
func (h *VideoHandler) StreamThumbnail(c *gin.Context) {
	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.repo.GetByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	if video.ThumbnailPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thumbnail not found"})
		return
	}

	// Check access for private videos
	if video.Visibility == "private" {
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			log.Printf("⛔ No auth provided for private video thumbnail %s", videoID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		userID := userIDInterface.(string)
		if userID != video.UserID.String() {
			log.Printf("⛔ User %s tried to access private video thumbnail of %s", userID, video.UserID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		log.Printf("✅ Access granted to owner %s for private video thumbnail %s", userID, videoID)
	}

	// ✅ Stream thumbnail directly from MinIO
	ctx := c.Request.Context()
	object, err := h.storage.GetObject(ctx, h.vodBucket, video.ThumbnailPath)
	if err != nil {
		log.Printf("❌ Failed to get thumbnail from MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream thumbnail"})
		return
	}
	defer object.Close()

	// Get file info
	stat, err := object.Stat()
	if err != nil {
		log.Printf("❌ Failed to stat thumbnail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get thumbnail info"})
		return
	}

	log.Printf("✅ Streaming thumbnail for video %s (size: %d bytes)", videoID, stat.Size)

	// Set headers for image streaming
	c.Header("Content-Type", "image/jpeg")
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size))
	c.Header("Cache-Control", "public, max-age=31536000")

	// CORS for image element
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Credentials", "true")

	// Stream to client
	_, err = io.Copy(c.Writer, object)
	if err != nil {
		log.Printf("⚠️ Error streaming thumbnail: %v", err)
	}
}

// getRecordingInfo получает информацию о recording
func (h *VideoHandler) getRecordingInfo(recordingID uuid.UUID) (*RecordingInfo, error) {
	url := fmt.Sprintf("%s/recording/%s", h.recordingServiceURL, recordingID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to contact recording service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recording service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Recording *RecordingInfo `json:"recording"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Recording, nil
}

// RecordingInfo - информация о записи из Recording Service
type RecordingInfo struct {
	ID            uuid.UUID  `json:"id"`
	StreamID      *uuid.UUID `json:"stream_id"`
	FilePath      string     `json:"file_path"`
	ThumbnailPath string     `json:"thumbnail_path"`
	Duration      int        `json:"duration"`
	FileSize      int64      `json:"file_size"`
	StartedAt     time.Time  `json:"started_at"`
}

// HealthCheck - проверка здоровья сервиса
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "vod-service",
	})
}

// ListAllVideos возвращает все публичные видео + приватные текущего пользователя
func (h *VideoHandler) ListAllVideos(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	// Получаем user_id (может быть пустым для неавторизованных)
	userID := getUserID(c)

	var userUUID *uuid.UUID
	if userID != "" {
		parsed, err := uuid.Parse(userID)
		if err == nil {
			userUUID = &parsed
		}
	}

	log.Printf("📹 Getting all videos (limit=%d, offset=%d, user_id=%v)", limit, offset, userID)

	videos, total, err := h.repo.ListAllVideos(userUUID, limit, offset)
	if err != nil {
		log.Printf("❌ Failed to get videos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch videos"})
		return
	}

	page := offset/limit + 1

	log.Printf("✅ Found %d videos (public + user's private)", total)

	c.JSON(http.StatusOK, models.VideoListResponse{
		Videos: videos,
		Total:  total,
		Page:   page,
		Limit:  limit,
	})
}
