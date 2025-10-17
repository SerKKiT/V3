-- infrastructure/postgres/migrations/streams_db/000001_initial_streams_schema.down.sql
-- Rollback: Drop streams table and related objects

BEGIN;

DROP TRIGGER IF EXISTS trigger_update_streams_timestamp ON streams;
DROP FUNCTION IF EXISTS update_streams_updated_at();

DROP INDEX IF EXISTS idx_streams_live;
DROP INDEX IF EXISTS idx_streams_started_at;
DROP INDEX IF EXISTS idx_streams_created_at;
DROP INDEX IF EXISTS idx_streams_stream_key;
DROP INDEX IF EXISTS idx_streams_status;
DROP INDEX IF EXISTS idx_streams_user_id;

DROP TABLE IF EXISTS streams CASCADE;

COMMIT;
