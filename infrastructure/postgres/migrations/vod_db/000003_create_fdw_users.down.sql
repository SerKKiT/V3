-- infrastructure/postgres/migrations/vod_db/000003_create_fdw_users.down.sql

-- Rollback Migration: Remove Foreign Data Wrapper setup

BEGIN;

-- Drop foreign table
DROP FOREIGN TABLE IF EXISTS users CASCADE;

-- Drop user mapping
DROP USER MAPPING IF EXISTS FOR streaming_user SERVER auth_server;

-- Drop foreign server
DROP SERVER IF EXISTS auth_server CASCADE;

DO $$
BEGIN
  RAISE NOTICE '✅ Migration 000003 rolled back: Foreign Data Wrapper removed';
END $$;

COMMIT;
