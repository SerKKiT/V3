package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq" // ✅ Добавляем для PostgreSQL массивов
)

type Stream struct {
	ID                 uuid.UUID      `json:"id" db:"id"`
	UserID             uuid.UUID      `json:"user_id" db:"user_id"`
	StreamKey          string         `json:"stream_key" db:"stream_key"`
	Title              string         `json:"title" db:"title"`
	Description        string         `json:"description" db:"description"`
	Status             string         `json:"status" db:"status"` // live, offline
	ViewerCount        int            `json:"viewer_count" db:"viewer_count"`
	StartedAt          *time.Time     `json:"started_at,omitempty" db:"started_at"`
	EndedAt            *time.Time     `json:"ended_at,omitempty" db:"ended_at"`
	ThumbnailURL       string         `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	HLSURL             string         `json:"hls_url,omitempty" db:"hls_url"`
	AvailableQualities pq.StringArray `json:"available_qualities" db:"available_qualities"` // ✅ NEW
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time     `json:"updated_at,omitempty" db:"updated_at"`
	Username           string         `json:"username,omitempty"`
}

type CreateStreamRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"max=1000"`
}

type CreateStreamResponse struct {
	Stream    *Stream `json:"stream"`
	StreamURL string  `json:"stream_url"`
	HLSURL    string  `json:"hls_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
