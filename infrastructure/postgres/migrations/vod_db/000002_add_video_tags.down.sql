-- infrastructure/postgres/migrations/vod_db/000002_add_video_tags.down.sql
-- Rollback: Remove tags support

BEGIN;

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_validate_video_tags ON videos;

-- Drop function
DROP FUNCTION IF EXISTS validate_video_tags();

-- Drop constraint
ALTER TABLE videos DROP CONSTRAINT IF EXISTS max_tags_count;

-- Drop index
DROP INDEX IF EXISTS idx_videos_tags;

-- Drop column
ALTER TABLE videos DROP COLUMN IF EXISTS tags;

-- Log rollback
DO $$ 
BEGIN
    RAISE NOTICE 'Rollback 000002: Removed tags support from videos table';
END $$;

COMMIT;
