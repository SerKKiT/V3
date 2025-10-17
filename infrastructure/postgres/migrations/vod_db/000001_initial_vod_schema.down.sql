-- infrastructure/postgres/migrations/vod_db/000001_initial_vod_schema.down.sql
-- Rollback: Drop VOD schema

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_videos_timestamp ON videos;
DROP TRIGGER IF EXISTS trigger_update_recordings_timestamp ON recordings;

-- Drop functions
DROP FUNCTION IF EXISTS update_videos_updated_at();
DROP FUNCTION IF EXISTS update_recordings_updated_at();

-- Drop indexes for videos
DROP INDEX IF EXISTS idx_videos_public;
DROP INDEX IF EXISTS idx_videos_search;
DROP INDEX IF EXISTS idx_videos_published_at;
DROP INDEX IF EXISTS idx_videos_view_count;
DROP INDEX IF EXISTS idx_videos_created_at;
DROP INDEX IF EXISTS idx_videos_category;
DROP INDEX IF EXISTS idx_videos_visibility;
DROP INDEX IF EXISTS idx_videos_status;
DROP INDEX IF EXISTS idx_videos_stream_id;
DROP INDEX IF EXISTS idx_videos_recording_id;
DROP INDEX IF EXISTS idx_videos_user_id;

-- Drop indexes for recordings
DROP INDEX IF EXISTS idx_recordings_completed_at;
DROP INDEX IF EXISTS idx_recordings_created_at;
DROP INDEX IF EXISTS idx_recordings_video_id;
DROP INDEX IF EXISTS idx_recordings_status;
DROP INDEX IF EXISTS idx_recordings_stream_id;

-- Drop tables
DROP TABLE IF EXISTS videos CASCADE;
DROP TABLE IF EXISTS recordings CASCADE;

-- Drop ENUM types
DROP TYPE IF EXISTS video_visibility;
DROP TYPE IF EXISTS video_source;
DROP TYPE IF EXISTS video_status;
DROP TYPE IF EXISTS recording_status;

COMMIT;
