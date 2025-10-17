-- infrastructure/postgres/migrations/streams_db/000003_create_fdw_users.down.sql

-- Rollback Migration: Remove Foreign Data Wrapper setup

BEGIN;

-- Drop foreign table
DROP FOREIGN TABLE IF EXISTS users CASCADE;
RAISE NOTICE '✅ Dropped foreign table users';

-- Drop user mapping
DROP USER MAPPING IF EXISTS FOR streaming_user SERVER auth_server;
RAISE NOTICE '✅ Dropped user mapping for streaming_user';

-- Drop foreign server
DROP SERVER IF EXISTS auth_server CASCADE;
RAISE NOTICE '✅ Dropped foreign server auth_server';

-- Log rollback
RAISE NOTICE '✅ Migration 000003 rolled back: Foreign Data Wrapper removed';

COMMIT;
