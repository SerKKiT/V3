-- infrastructure/postgres/migrations/vod_db/000004_optimize_fdw.down.sql

BEGIN;

DROP FUNCTION IF EXISTS refresh_videos_cache();
DROP MATERIALIZED VIEW IF EXISTS videos_with_users_cache CASCADE;

DO $$
BEGIN
    ALTER SERVER auth_server OPTIONS (DROP fetch_size);
EXCEPTION
    WHEN undefined_object THEN
        RAISE NOTICE 'fetch_size option not found on auth_server';
END $$;

DO $$
BEGIN
    ALTER SERVER auth_server OPTIONS (DROP use_remote_estimate);
EXCEPTION
    WHEN undefined_object THEN
        RAISE NOTICE 'use_remote_estimate option not found on auth_server';
END $$;

COMMIT;
