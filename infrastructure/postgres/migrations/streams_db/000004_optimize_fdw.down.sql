-- infrastructure/postgres/migrations/streams_db/000004_optimize_fdw.down.sql

BEGIN;

-- Drop function
DROP FUNCTION IF EXISTS refresh_streams_cache();

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS streams_with_users_cache CASCADE;

-- Revert FDW settings to defaults
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
