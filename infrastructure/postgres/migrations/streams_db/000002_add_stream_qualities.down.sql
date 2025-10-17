-- infrastructure/postgres/migrations/streams_db/000002_add_stream_qualities.down.sql
-- Rollback: Remove ABR support

BEGIN;

-- Drop constraint
ALTER TABLE streams DROP CONSTRAINT IF EXISTS valid_qualities;

-- Drop index
DROP INDEX IF EXISTS idx_streams_qualities;

-- Drop column
ALTER TABLE streams DROP COLUMN IF EXISTS available_qualities;

-- Log rollback
DO $$ 
BEGIN
    RAISE NOTICE 'Rollback 000002: Removed ABR support (available_qualities column)';
END $$;

COMMIT;
