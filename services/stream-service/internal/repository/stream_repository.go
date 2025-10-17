package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/SerKKiT/streaming-platform/stream-service/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type StreamRepository struct {
	db *sql.DB
}

func NewStreamRepository(db *sql.DB) *StreamRepository {
	return &StreamRepository{db: db}
}

// CreateStream creates a new stream
func (r *StreamRepository) CreateStream(userID uuid.UUID, streamKey, title, description string) (*models.Stream, error) {
	stream := &models.Stream{
		ID:          uuid.New(),
		UserID:      userID,
		StreamKey:   streamKey,
		Title:       title,
		Description: description,
		Status:      "offline",
		ViewerCount: 0,
		CreatedAt:   time.Now(),
	}

	defaultQualities := []string{"360p", "480p", "720p", "1080p"}

	query := `
		INSERT INTO streams (id, user_id, stream_key, title, description, status, viewer_count, available_qualities, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, stream_key, title, description, status, viewer_count, available_qualities, created_at
	`

	var qualities []string

	err := r.db.QueryRow(
		query,
		stream.ID,
		stream.UserID,
		stream.StreamKey,
		stream.Title,
		stream.Description,
		stream.Status,
		stream.ViewerCount,
		pq.Array(defaultQualities),
		stream.CreatedAt,
	).Scan(
		&stream.ID,
		&stream.UserID,
		&stream.StreamKey,
		&stream.Title,
		&stream.Description,
		&stream.Status,
		&stream.ViewerCount,
		pq.Array(&qualities),
		&stream.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	stream.AvailableQualities = pq.StringArray(qualities)
	return stream, nil
}

// GetStreamByID retrieves a stream by its ID with username
func (r *StreamRepository) GetStreamByID(streamID uuid.UUID) (*models.Stream, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.stream_key, s.title, s.description, s.status, s.viewer_count,
			s.started_at, s.ended_at, s.thumbnail_url, s.hls_url, s.available_qualities, s.created_at,
			COALESCE(u.username, 'Unknown') as username
		FROM streams s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.id = $1
	`

	stream := &models.Stream{}
	var startedAt, endedAt sql.NullTime
	var thumbnailURL, hlsURL sql.NullString
	var qualities []string
	var username string

	err := r.db.QueryRow(query, streamID).Scan(
		&stream.ID, &stream.UserID, &stream.StreamKey,
		&stream.Title, &stream.Description, &stream.Status, &stream.ViewerCount,
		&startedAt, &endedAt, &thumbnailURL, &hlsURL,
		pq.Array(&qualities), &stream.CreatedAt,
		&username,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stream not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	if startedAt.Valid {
		stream.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		stream.EndedAt = &endedAt.Time
	}
	if thumbnailURL.Valid {
		stream.ThumbnailURL = thumbnailURL.String
	}
	if hlsURL.Valid {
		stream.HLSURL = hlsURL.String
	}

	stream.Username = username
	stream.AvailableQualities = pq.StringArray(qualities)

	return stream, nil
}

// GetStreamByKey retrieves a stream by stream key
func (r *StreamRepository) GetStreamByKey(streamKey string) (*models.Stream, error) {
	stream := &models.Stream{}
	query := `
		SELECT id, user_id, stream_key, title, description, status, viewer_count,
		       started_at, ended_at, thumbnail_url, hls_url, available_qualities, created_at
		FROM streams
		WHERE stream_key = $1
	`

	var startedAt, endedAt sql.NullTime
	var thumbnailURL, hlsURL sql.NullString
	var qualities []string

	err := r.db.QueryRow(query, streamKey).Scan(
		&stream.ID,
		&stream.UserID,
		&stream.StreamKey,
		&stream.Title,
		&stream.Description,
		&stream.Status,
		&stream.ViewerCount,
		&startedAt,
		&endedAt,
		&thumbnailURL,
		&hlsURL,
		pq.Array(&qualities),
		&stream.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stream by key: %w", err)
	}

	if startedAt.Valid {
		stream.StartedAt = &startedAt.Time
	}

	if endedAt.Valid {
		stream.EndedAt = &endedAt.Time
	}

	if thumbnailURL.Valid {
		stream.ThumbnailURL = thumbnailURL.String
	}

	if hlsURL.Valid {
		stream.HLSURL = hlsURL.String
	}

	stream.AvailableQualities = pq.StringArray(qualities)
	return stream, nil
}

// GetUserStreams retrieves all streams for a user with username
func (r *StreamRepository) GetUserStreams(userID uuid.UUID) ([]*models.Stream, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.stream_key, s.title, s.description, s.status, s.viewer_count,
			s.started_at, s.ended_at, s.thumbnail_url, s.hls_url, s.available_qualities, s.created_at,
			COALESCE(u.username, 'Unknown Streamer') as username
		FROM streams s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user streams: %w", err)
	}
	defer rows.Close()

	var streams []*models.Stream
	for rows.Next() {
		stream := &models.Stream{}
		var startedAt, endedAt sql.NullTime
		var thumbnailURL, hlsURL sql.NullString
		var qualities []string
		var username string // ← ДОБАВЛЕНО

		err := rows.Scan(
			&stream.ID,
			&stream.UserID,
			&stream.StreamKey,
			&stream.Title,
			&stream.Description,
			&stream.Status,
			&stream.ViewerCount,
			&startedAt,
			&endedAt,
			&thumbnailURL,
			&hlsURL,
			pq.Array(&qualities),
			&stream.CreatedAt,
			&username, // ← ДОБАВЛЕНО
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stream: %w", err)
		}

		if startedAt.Valid {
			stream.StartedAt = &startedAt.Time
		}

		if endedAt.Valid {
			stream.EndedAt = &endedAt.Time
		}

		if thumbnailURL.Valid {
			stream.ThumbnailURL = thumbnailURL.String
		}

		if hlsURL.Valid {
			stream.HLSURL = hlsURL.String
		}

		stream.Username = username // ← ДОБАВЛЕНО
		stream.AvailableQualities = pq.StringArray(qualities)
		streams = append(streams, stream)
	}

	return streams, nil
}

// GetLiveStreams retrieves all live streams with username
func (r *StreamRepository) GetLiveStreams() ([]*models.Stream, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.stream_key, s.title, s.description, s.status, s.viewer_count,
			s.started_at, s.ended_at, s.thumbnail_url, s.hls_url, s.available_qualities, s.created_at,
			COALESCE(u.username, 'Unknown Streamer') as username
		FROM streams s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.status = 'live'
		ORDER BY s.started_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get live streams: %w", err)
	}
	defer rows.Close()

	var streams []*models.Stream
	for rows.Next() {
		stream := &models.Stream{}
		var startedAt, endedAt sql.NullTime
		var thumbnailURL, hlsURL sql.NullString
		var qualities []string
		var username string // ← ДОБАВЛЕНО

		err := rows.Scan(
			&stream.ID,
			&stream.UserID,
			&stream.StreamKey,
			&stream.Title,
			&stream.Description,
			&stream.Status,
			&stream.ViewerCount,
			&startedAt,
			&endedAt,
			&thumbnailURL,
			&hlsURL,
			pq.Array(&qualities),
			&stream.CreatedAt,
			&username, // ← ДОБАВЛЕНО
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stream: %w", err)
		}

		if startedAt.Valid {
			stream.StartedAt = &startedAt.Time
		}

		if endedAt.Valid {
			stream.EndedAt = &endedAt.Time
		}

		if thumbnailURL.Valid {
			stream.ThumbnailURL = thumbnailURL.String
		}

		if hlsURL.Valid {
			stream.HLSURL = hlsURL.String
		}

		stream.Username = username // ← ДОБАВЛЕНО
		stream.AvailableQualities = pq.StringArray(qualities)
		streams = append(streams, stream)
	}

	return streams, nil
}

// UpdateStream updates stream title and description
func (r *StreamRepository) UpdateStream(stream *models.Stream) error {
	query := `
		UPDATE streams 
		SET title = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	_, err := r.db.Exec(query, stream.Title, stream.Description, stream.ID)
	if err != nil {
		return fmt.Errorf("failed to update stream: %w", err)
	}
	return nil
}

// UpdateStreamStatus updates stream status (live/offline)
// UpdateStreamStatus updates stream status (live/offline)
func (r *StreamRepository) UpdateStreamStatus(streamID uuid.UUID, status string) error {
	now := time.Now()

	if status == "live" {
		// При переходе в live устанавливаем started_at только если он NULL
		query := `
			UPDATE streams 
			SET status = $1,
			    started_at = COALESCE(started_at, $2),
			    ended_at = NULL,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`
		_, err := r.db.Exec(query, status, now, streamID)
		if err != nil {
			return fmt.Errorf("failed to update stream status to live: %w", err)
		}

		log.Printf("✅ Stream %s status updated to 'live'", streamID)
	} else if status == "offline" {
		// При переходе в offline устанавливаем ended_at
		query := `
			UPDATE streams 
			SET status = $1,
			    ended_at = $2,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`
		_, err := r.db.Exec(query, status, now, streamID)
		if err != nil {
			return fmt.Errorf("failed to update stream status to offline: %w", err)
		}

		log.Printf("✅ Stream %s status updated to 'offline'", streamID)
	} else {
		// Для других статусов просто обновляем status
		query := `
			UPDATE streams 
			SET status = $1,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`
		_, err := r.db.Exec(query, status, streamID)
		if err != nil {
			return fmt.Errorf("failed to update stream status: %w", err)
		}
	}

	return nil
}

// UpdateStreamThumbnail updates stream thumbnail URL
func (r *StreamRepository) UpdateStreamThumbnail(streamID uuid.UUID, thumbnailURL string) error {
	query := `
		UPDATE streams 
		SET thumbnail_url = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := r.db.Exec(query, thumbnailURL, streamID)
	if err != nil {
		return fmt.Errorf("failed to update thumbnail: %w", err)
	}
	return nil
}

// DeleteStream deletes a stream
func (r *StreamRepository) DeleteStream(streamID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM streams WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, streamID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("stream not found or not authorized")
	}

	return nil
}
