package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/SerKKiT/streaming-platform/vod-service/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type VideoRepository struct {
	db *sql.DB
}

func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// Create creates a new video
func (r *VideoRepository) Create(video *models.Video) error {
	query := `
		INSERT INTO videos (
			id, user_id, recording_id, stream_id,
			title, description, category, tags,
			source, status, visibility,
			file_path, thumbnail_path, duration, file_size,
			view_count, like_count,
			created_at, updated_at, published_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`

	_, err := r.db.Exec(query,
		video.ID, video.UserID, video.RecordingID, video.StreamID,
		video.Title, video.Description, video.Category, pq.Array(video.Tags),
		video.Source, video.Status, video.Visibility,
		video.FilePath, video.ThumbnailPath, video.Duration, video.FileSize,
		video.ViewCount, video.LikeCount,
		video.CreatedAt, video.UpdatedAt, video.PublishedAt,
	)

	return err
}

// ✅ ОПТИМИЗИРОВАНО: GetByID с CTE
func (r *VideoRepository) GetByID(id uuid.UUID) (*models.Video, error) {
	query := `
		WITH target_video AS (
			SELECT 
				id, user_id, recording_id, stream_id, title, description, category, tags,
				source, status, visibility, file_path, thumbnail_path, duration, file_size,
				view_count, like_count, created_at, updated_at, published_at
			FROM videos
			WHERE id = $1
		)
		SELECT
			tv.id, tv.user_id, tv.recording_id, tv.stream_id,
			tv.title, tv.description, tv.category, tv.tags,
			tv.source, tv.status, tv.visibility,
			tv.file_path, tv.thumbnail_path, tv.duration, tv.file_size,
			tv.view_count, tv.like_count,
			tv.created_at, tv.updated_at, tv.published_at,
			COALESCE(u.username, 'Unknown') as username
		FROM target_video tv
		LEFT JOIN users u ON tv.user_id = u.id
	`

	video := &models.Video{}
	var tags pq.StringArray

	err := r.db.QueryRow(query, id).Scan(
		&video.ID, &video.UserID, &video.RecordingID, &video.StreamID,
		&video.Title, &video.Description, &video.Category, &tags,
		&video.Source, &video.Status, &video.Visibility,
		&video.FilePath, &video.ThumbnailPath, &video.Duration, &video.FileSize,
		&video.ViewCount, &video.LikeCount,
		&video.CreatedAt, &video.UpdatedAt, &video.PublishedAt,
		&video.Username,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("video not found")
	}

	if err != nil {
		return nil, err
	}

	video.Tags = []string(tags)
	return video, nil
}

// ✅ ОПТИМИЗИРОВАНО: ListUserVideos с CTE
func (r *VideoRepository) ListUserVideos(userID uuid.UUID, limit, offset int) ([]*models.Video, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM videos WHERE user_id = $1`
	if err := r.db.QueryRow(countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get videos with CTE
	query := `
		WITH filtered_videos AS (
			SELECT 
				id, user_id, recording_id, stream_id, title, description, category, tags,
				source, status, visibility, file_path, thumbnail_path, duration, file_size,
				view_count, like_count, created_at, updated_at, published_at
			FROM videos
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		)
		SELECT
			fv.id, fv.user_id, fv.recording_id, fv.stream_id,
			fv.title, fv.description, fv.category, fv.tags,
			fv.source, fv.status, fv.visibility,
			fv.file_path, fv.thumbnail_path, fv.duration, fv.file_size,
			fv.view_count, fv.like_count,
			fv.created_at, fv.updated_at, fv.published_at,
			COALESCE(u.username, 'Unknown') as username
		FROM filtered_videos fv
		LEFT JOIN users u ON fv.user_id = u.id
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var videos []*models.Video

	for rows.Next() {
		video := &models.Video{}
		var tags pq.StringArray

		err := rows.Scan(
			&video.ID, &video.UserID, &video.RecordingID, &video.StreamID,
			&video.Title, &video.Description, &video.Category, &tags,
			&video.Source, &video.Status, &video.Visibility,
			&video.FilePath, &video.ThumbnailPath, &video.Duration, &video.FileSize,
			&video.ViewCount, &video.LikeCount,
			&video.CreatedAt, &video.UpdatedAt, &video.PublishedAt,
			&video.Username,
		)

		if err != nil {
			return nil, 0, err
		}

		video.Tags = []string(tags)
		videos = append(videos, video)
	}

	return videos, total, nil
}

// Update updates video metadata
func (r *VideoRepository) Update(video *models.Video) error {
	query := `
		UPDATE videos 
		SET title = $1, description = $2, category = $3, tags = $4,
			visibility = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(query,
		video.Title, video.Description, video.Category, pq.Array(video.Tags),
		video.Visibility, time.Now(), video.ID,
	)

	return err
}

// Delete deletes a video
func (r *VideoRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM videos WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("video not found")
	}

	return nil
}

// IncrementViewCount increments view count
func (r *VideoRepository) IncrementViewCount(id uuid.UUID) error {
	query := `UPDATE videos SET view_count = view_count + 1 WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// IncrementLikeCount increments like count
func (r *VideoRepository) IncrementLikeCount(id uuid.UUID) error {
	query := `UPDATE videos SET like_count = like_count + 1 WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// ✅ ИСПРАВЛЕНО: GetByRecordingID с JOIN для получения username
func (r *VideoRepository) GetByRecordingID(recordingID uuid.UUID) (*models.Video, error) {
	query := `
		SELECT 
			v.id, v.user_id, v.recording_id, v.stream_id,
			v.title, v.description, v.category, v.tags,
			v.source, v.status, v.visibility,
			v.file_path, v.thumbnail_path, v.duration, v.file_size,
			v.view_count, v.like_count,
			v.created_at, v.updated_at, v.published_at,
			COALESCE(u.username, 'Unknown') as username
		FROM videos v
		LEFT JOIN users u ON v.user_id = u.id
		WHERE v.recording_id = $1
	`

	video := &models.Video{}
	var tags pq.StringArray
	err := r.db.QueryRow(query, recordingID).Scan(
		&video.ID, &video.UserID, &video.RecordingID, &video.StreamID,
		&video.Title, &video.Description, &video.Category, &tags,
		&video.Source, &video.Status, &video.Visibility,
		&video.FilePath, &video.ThumbnailPath, &video.Duration, &video.FileSize,
		&video.ViewCount, &video.LikeCount,
		&video.CreatedAt, &video.UpdatedAt, &video.PublishedAt,
		&video.Username, // ✅ ДОБАВЛЕНО
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("video not found")
	}

	if err != nil {
		return nil, err
	}

	video.Tags = []string(tags)
	return video, nil
}

// ✅ ОПТИМИЗИРОВАНО: ListAllVideos с CTE
func (r *VideoRepository) ListAllVideos(userID *uuid.UUID, limit, offset int) ([]*models.Video, int, error) {
	var videos []*models.Video
	var total int

	// Count total
	var countQuery string
	var countArgs []interface{}

	if userID != nil {
		countQuery = `SELECT COUNT(*) FROM videos WHERE visibility = 'public' OR user_id = $1`
		countArgs = []interface{}{userID}
	} else {
		countQuery = `SELECT COUNT(*) FROM videos WHERE visibility = 'public'`
	}

	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		log.Printf("❌ Failed to count videos: %v", err)
		return nil, 0, err
	}

	// Get videos with CTE
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			WITH filtered_videos AS (
				SELECT 
					id, user_id, recording_id, stream_id, title, description, category, tags,
					source, status, visibility, file_path, thumbnail_path, duration, file_size,
					view_count, like_count, created_at, updated_at, published_at
				FROM videos
				WHERE visibility = 'public' OR user_id = $1
				ORDER BY created_at DESC
				LIMIT $2 OFFSET $3
			)
			SELECT
				fv.id, fv.user_id, fv.recording_id, fv.stream_id,
				fv.title, fv.description, fv.category, fv.tags,
				fv.source, fv.status, fv.visibility,
				fv.file_path, fv.thumbnail_path, fv.duration, fv.file_size,
				fv.view_count, fv.like_count,
				fv.created_at, fv.updated_at, fv.published_at,
				COALESCE(u.username, 'Unknown') as username
			FROM filtered_videos fv
			LEFT JOIN users u ON fv.user_id = u.id
		`
		args = []interface{}{userID, limit, offset}
	} else {
		query = `
			WITH filtered_videos AS (
				SELECT 
					id, user_id, recording_id, stream_id, title, description, category, tags,
					source, status, visibility, file_path, thumbnail_path, duration, file_size,
					view_count, like_count, created_at, updated_at, published_at
				FROM videos
				WHERE visibility = 'public'
				ORDER BY created_at DESC
				LIMIT $1 OFFSET $2
			)
			SELECT
				fv.id, fv.user_id, fv.recording_id, fv.stream_id,
				fv.title, fv.description, fv.category, fv.tags,
				fv.source, fv.status, fv.visibility,
				fv.file_path, fv.thumbnail_path, fv.duration, fv.file_size,
				fv.view_count, fv.like_count,
				fv.created_at, fv.updated_at, fv.published_at,
				COALESCE(u.username, 'Unknown') as username
			FROM filtered_videos fv
			LEFT JOIN users u ON fv.user_id = u.id
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		log.Printf("❌ Failed to query videos: %v", err)
		return nil, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		video := &models.Video{}
		var tags pq.StringArray

		err := rows.Scan(
			&video.ID, &video.UserID, &video.RecordingID, &video.StreamID,
			&video.Title, &video.Description, &video.Category, &tags,
			&video.Source, &video.Status, &video.Visibility,
			&video.FilePath, &video.ThumbnailPath, &video.Duration, &video.FileSize,
			&video.ViewCount, &video.LikeCount,
			&video.CreatedAt, &video.UpdatedAt, &video.PublishedAt,
			&video.Username,
		)

		if err != nil {
			log.Printf("❌ Failed to scan video: %v", err)
			return nil, 0, err
		}

		video.Tags = []string(tags)
		videos = append(videos, video)
	}

	log.Printf("✅ Found %d videos (total: %d)", len(videos), total)
	return videos, total, nil
}
