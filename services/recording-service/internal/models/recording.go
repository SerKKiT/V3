package models

import (
	"time"

	"github.com/google/uuid"
)

type Recording struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	StreamID      uuid.UUID  `json:"stream_id" db:"stream_id"`
	VideoID       *uuid.UUID `json:"video_id" db:"video_id"`
	FilePath      string     `json:"file_path" db:"file_path"`
	ThumbnailPath string     `json:"thumbnail_path" db:"thumbnail_path"`
	Duration      int        `json:"duration" db:"duration"`
	FileSize      int64      `json:"file_size" db:"file_size"`
	Status        string     `json:"status" db:"status"` // recording, processing, completed, failed
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
