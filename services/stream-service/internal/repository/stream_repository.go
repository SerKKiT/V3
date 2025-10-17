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

// ✅ ОПТИМИЗИРОВАНО: GetStreamByID с CTE для предварительной фильтрации
func (r *StreamRepository) GetStreamByID(streamID uuid.UUID) (*models.Stream, error) {
	// CTE сначала фильтрует streams, затем JOIN с FDW
	query := `
		WITH target_stream AS (
			SELECT 
				id, user_id, stream_key, title, description, status, viewer_count,
				started_at, ended_at, thumbnail_url, hls_url, available_qualities, created_at
			FROM streams
			WHERE id = $1
		)
		SELECT
			ts.id, ts.user_id, ts.stream_key, ts.title, ts.description, 
			ts.status, ts.viewer_count, ts.started_at, ts.ended_at, 
			ts.thumbnail_url, ts.hls_url, ts.available_qualities, ts.created_at,
			COALESCE(u.username, 'Unknown') as username
		FROM target_stream ts
		LEFT JOIN users u ON ts.user_id = u.id
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

// ✅ ОПТИМИЗИРОВАНО: GetUserStreams с CTE
func (r *StreamRepository) GetUserStreams(userID uuid.UUID) ([]*models.Stream, error) {
	query := `
		WITH filtered_streams AS (
			SELECT 
				id, user_id, stream_key, title, description, status, viewer_count,
				started_at, ended_at, thumbnail_url, hls_url, available_qualities, created_at
			FROM streams
			WHERE user_id = $1
			ORDER BY created_at DESC
		)
		SELECT
			fs.id, fs.user_id, fs.stream_key, fs.title, fs.description, 
			fs.status, fs.viewer_count, fs.started_at, fs.ended_at, 
			fs.thumbnail_url, fs.hls_url, fs.available_qualities, fs.created_at,
			COALESCE(u.username, 'Unknown Streamer') as username
		FROM filtered_streams fs
		LEFT JOIN users u ON fs.user_id = u.id
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
		var username string

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
			&username,
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

		stream.Username = username
		stream.AvailableQualities = pq.StringArray(qualities)
		streams = append(streams, stream)
	}

	return streams, nil
}

// ✅ ОПТИМИЗИРОВАНО: GetLiveStreams с CTE и LIMIT
func (r *StreamRepository) GetLiveStreams() ([]*models.Stream, error) {
	query := `
		WITH filtered_streams AS (
			SELECT 
				id, user_id, stream_key, title, description, status, viewer_count,
				started_at, ended_at, thumbnail_url, hls_url, available_qualities, created_at
			FROM streams
			WHERE status = 'live'
			ORDER BY started_at DESC
			LIMIT 100
		)
		SELECT
			fs.id, fs.user_id, fs.stream_key, fs.title, fs.description, 
			fs.status, fs.viewer_count, fs.started_at, fs.ended_at, 
			fs.thumbnail_url, fs.hls_url, fs.available_qualities, fs.created_at,
			COALESCE(u.username, 'Unknown Streamer') as username
		FROM filtered_streams fs
		LEFT JOIN users u ON fs.user_id = u.id
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
		var username string

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
			&username,
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

		stream.Username = username
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
func (r *StreamRepository) UpdateStreamStatus(streamID uuid.UUID, status string) error {
	now := time.Now()

	if status == "live" {
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
