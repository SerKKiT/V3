-- infrastructure/postgres/migrations/vod_db/000002_add_video_tags.up.sql
-- Migration: Add tags support for videos
-- Description: Add tags array column with GIN index for efficient searching

BEGIN;

-- Add tags column
ALTER TABLE videos 
ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

-- Create GIN index for array searching
CREATE INDEX IF NOT EXISTS idx_videos_tags 
ON videos USING GIN(tags);

-- Add constraint to limit number of tags (simple version)
ALTER TABLE videos
ADD CONSTRAINT max_tags_count CHECK (array_length(tags, 1) IS NULL OR array_length(tags, 1) <= 20);

-- Comment
COMMENT ON COLUMN videos.tags IS 'Array of tags for video categorization (max 20 tags, each 2-30 chars)';

-- Create trigger function to validate tag length
CREATE OR REPLACE FUNCTION validate_video_tags()
RETURNS TRIGGER AS $$
DECLARE
    tag TEXT;
BEGIN
    -- ✅ Проверяем что tags не NULL и не пустой массив
    IF NEW.tags IS NULL OR array_length(NEW.tags, 1) IS NULL THEN
        RETURN NEW;
    END IF;
    
    -- Check each tag length
    FOREACH tag IN ARRAY NEW.tags
    LOOP
        IF char_length(tag) < 2 OR char_length(tag) > 30 THEN
            RAISE EXCEPTION 'Tag length must be between 2 and 30 characters: %', tag;
        END IF;
    END LOOP;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS trigger_validate_video_tags ON videos;
CREATE TRIGGER trigger_validate_video_tags
BEFORE INSERT OR UPDATE OF tags ON videos
FOR EACH ROW
EXECUTE FUNCTION validate_video_tags();

-- Log migration
DO $$ 
BEGIN
    RAISE NOTICE 'Migration 000002: Added tags support to videos table with validation trigger';
END $$;

COMMIT;
