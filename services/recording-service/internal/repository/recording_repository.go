package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/SerKKiT/streaming-platform/recording-service/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type RecordingRepository struct {
	db *sql.DB
}

func NewRecordingRepository(db *sql.DB) *RecordingRepository {
	return &RecordingRepository{db: db}
}

func (r *RecordingRepository) CreateRecording(streamID uuid.UUID, filePath string) (*models.Recording, error) {
	recording := &models.Recording{
		ID:        uuid.New(),
		StreamID:  streamID,
		FilePath:  filePath,
		Duration:  0,
		FileSize:  0,
		Status:    "recording",
		StartedAt: time.Now(),
	}

	query := `
		INSERT INTO recordings (id, stream_id, file_path, duration, file_size, status, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, stream_id, file_path, duration, file_size, status, started_at
	`

	err := r.db.QueryRow(
		query,
		recording.ID,
		recording.StreamID,
		recording.FilePath,
		recording.Duration,
		recording.FileSize,
		recording.Status,
		recording.StartedAt,
	).Scan(
		&recording.ID,
		&recording.StreamID,
		&recording.FilePath,
		&recording.Duration,
		&recording.FileSize,
		&recording.Status,
		&recording.StartedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create recording: %w", err)
	}

	return recording, nil
}

// UpdateRecording обновляет запись полностью
func (r *RecordingRepository) UpdateRecording(recording *models.Recording) error {
	query := `
		UPDATE recordings
		SET video_id = $1, file_path = $2, duration = $3, file_size = $4, 
		    status = $5, completed_at = $6
		WHERE id = $7
	`
	_, err := r.db.Exec(query,
		recording.VideoID,
		recording.FilePath,
		recording.Duration,
		recording.FileSize,
		recording.Status,
		recording.CompletedAt,
		recording.ID,
	)
	return err
}

// UpdateRecordingStatus обновляет только статус
func (r *RecordingRepository) UpdateRecordingStatus(id uuid.UUID, status string) error {
	now := time.Now()
	query := `
		UPDATE recordings
		SET status = $1, completed_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(query, status, now, id)
	return err
}

func (r *RecordingRepository) GetRecordingByStreamID(streamID uuid.UUID) (*models.Recording, error) {
	recording := &models.Recording{}
	var completedAt sql.NullTime
	var videoID sql.NullString
	var thumbnailPath sql.NullString

	query := `
		SELECT id, stream_id, video_id, file_path, thumbnail_path, duration, file_size, status, started_at, completed_at
		FROM recordings
		WHERE stream_id = $1
		ORDER BY started_at DESC
		LIMIT 1
	`

	err := r.db.QueryRow(query, streamID).Scan(
		&recording.ID,
		&recording.StreamID,
		&videoID,
		&recording.FilePath,
		&thumbnailPath, // ✅ Добавлено
		&recording.Duration,
		&recording.FileSize,
		&recording.Status,
		&recording.StartedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recording not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	if completedAt.Valid {
		recording.CompletedAt = &completedAt.Time
	}

	if videoID.Valid {
		vid, _ := uuid.Parse(videoID.String)
		recording.VideoID = &vid
	}

	if thumbnailPath.Valid {
		recording.ThumbnailPath = thumbnailPath.String
	}

	return recording, nil
}

func (r *RecordingRepository) GetAllRecordings() ([]*models.Recording, error) {
	query := `
		SELECT id, stream_id, video_id, file_path, thumbnail_path, duration, file_size, status, started_at, completed_at
		FROM recordings
		ORDER BY started_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get recordings: %w", err)
	}

	defer rows.Close()
	var recordings []*models.Recording

	for rows.Next() {
		recording := &models.Recording{}
		var completedAt sql.NullTime
		var videoID sql.NullString
		var thumbnailPath sql.NullString

		err := rows.Scan(
			&recording.ID,
			&recording.StreamID,
			&videoID,
			&recording.FilePath,
			&thumbnailPath, // ✅ Добавлено
			&recording.Duration,
			&recording.FileSize,
			&recording.Status,
			&recording.StartedAt,
			&completedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan recording: %w", err)
		}

		if completedAt.Valid {
			recording.CompletedAt = &completedAt.Time
		}

		if videoID.Valid {
			vid, _ := uuid.Parse(videoID.String)
			recording.VideoID = &vid
		}

		if thumbnailPath.Valid {
			recording.ThumbnailPath = thumbnailPath.String
		}

		recordings = append(recordings, recording)
	}

	return recordings, nil
}

// GetByID - получить запись по ID
func (r *RecordingRepository) GetByID(id string) (*models.Recording, error) {
	query := `
		SELECT id, stream_id, video_id, file_path, thumbnail_path, duration, file_size, status, started_at, completed_at
		FROM recordings
		WHERE id = $1
	`

	var rec models.Recording
	var completedAt sql.NullTime
	var videoID sql.NullString
	var thumbnailPath sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&rec.ID,
		&rec.StreamID,
		&videoID,
		&rec.FilePath,
		&thumbnailPath, // ✅ Добавлено
		&rec.Duration,
		&rec.FileSize,
		&rec.Status,
		&rec.StartedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recording not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	if completedAt.Valid {
		rec.CompletedAt = &completedAt.Time
	}

	if videoID.Valid {
		vid, _ := uuid.Parse(videoID.String)
		rec.VideoID = &vid
	}

	if thumbnailPath.Valid {
		rec.ThumbnailPath = thumbnailPath.String
	}

	return &rec, nil
}

// UpdateThumbnailPath обновляет путь к thumbnail для записи
func (r *RecordingRepository) UpdateThumbnailPath(recordingID uuid.UUID, thumbnailPath string) error {
	query := `
		UPDATE recordings
		SET thumbnail_path = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(query, thumbnailPath, recordingID)
	if err != nil {
		return fmt.Errorf("failed to update thumbnail path: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recording not found: %s", recordingID)
	}

	log.Printf("✅ Updated thumbnail_path for recording %s: %s", recordingID, thumbnailPath)
	return nil
}
