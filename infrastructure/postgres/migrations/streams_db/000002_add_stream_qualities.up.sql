-- infrastructure/postgres/migrations/streams_db/000002_add_stream_qualities.up.sql

-- Migration: Add ABR (Adaptive Bitrate Streaming) support
-- Description: Add available_qualities column for multi-quality streaming

BEGIN;

-- Add available_qualities column with default qualities
ALTER TABLE streams
ADD COLUMN IF NOT EXISTS available_qualities TEXT[] 
DEFAULT ARRAY['360p', '480p', '720p', '1080p'];

-- Create GIN index for efficient array searching
CREATE INDEX IF NOT EXISTS idx_streams_qualities
ON streams USING GIN (available_qualities);

-- Update existing streams with default qualities
UPDATE streams
SET available_qualities = ARRAY['360p', '480p', '720p', '1080p']
WHERE available_qualities IS NULL;

-- Add constraint to ensure valid quality values (with IF NOT EXISTS check)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'valid_qualities' AND conrelid = 'streams'::regclass
    ) THEN
        ALTER TABLE streams
        ADD CONSTRAINT valid_qualities CHECK (
            available_qualities <@ ARRAY['360p', '480p', '720p', '1080p', '1440p', '4K']
        );
        RAISE NOTICE '✅ Constraint valid_qualities created';
    ELSE
        RAISE NOTICE '⚠️ Constraint valid_qualities already exists, skipping';
    END IF;
    
    RAISE NOTICE '✅ Migration 000002 completed: Added ABR support with qualities: 360p, 480p, 720p, 1080p';
END $$;

-- Add comment
COMMENT ON COLUMN streams.available_qualities IS
'Array of available video qualities for ABR streaming (e.g., ["360p", "720p", "1080p"])';

COMMIT;
