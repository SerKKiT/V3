package models

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Username      string     `json:"username" db:"username"` // ✅ ДОБАВЛЕНО: username владельца
	RecordingID   *uuid.UUID `json:"recording_id,omitempty" db:"recording_id"`
	StreamID      *uuid.UUID `json:"stream_id,omitempty" db:"stream_id"`
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description" db:"description"`
	Category      string     `json:"category" db:"category"`
	Tags          []string   `json:"tags" db:"tags"`
	Source        string     `json:"source" db:"source"`         // "recording", "upload"
	Status        string     `json:"status" db:"status"`         // "ready", "processing", "failed"
	Visibility    string     `json:"visibility" db:"visibility"` // "public", "private", "unlisted"
	FilePath      string     `json:"file_path" db:"file_path"`
	ThumbnailPath string     `json:"thumbnail_path" db:"thumbnail_path"`
	Duration      int        `json:"duration" db:"duration"`   // секунды
	FileSize      int64      `json:"file_size" db:"file_size"` // байты
	ViewCount     int        `json:"view_count" db:"view_count"`
	LikeCount     int        `json:"like_count" db:"like_count"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty" db:"published_at"`
}

// DTOs
type ImportRecordingRequest struct {
	RecordingID string   `json:"recording_id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"` // default: "public"
}

type UpdateVideoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"`
}

type VideoListResponse struct {
	Videos []*Video `json:"videos"`
	Total  int      `json:"total"`
	Page   int      `json:"page"`
	Limit  int      `json:"limit"`
}
