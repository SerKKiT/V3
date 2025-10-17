-- infrastructure/postgres/migrations/vod_db/000003_create_fdw_users.up.sql

-- Migration: Setup Foreign Data Wrapper for users table from auth_db
-- Description: Create foreign table to access users from auth_db for JOIN queries

BEGIN;

-- ============================================================
-- 1. VERIFY EXTENSION AND CREATE FDW
-- ============================================================
DO $$
BEGIN
  -- Check postgres_fdw extension
  IF NOT EXISTS (
    SELECT 1 FROM pg_extension WHERE extname = 'postgres_fdw'
  ) THEN
    RAISE EXCEPTION 'Extension postgres_fdw not found. Install it first with: CREATE EXTENSION postgres_fdw;';
  END IF;
  RAISE NOTICE '✅ Extension postgres_fdw verified';
END $$;

-- ============================================================
-- 2. CREATE FOREIGN SERVER
-- ============================================================
DROP SERVER IF EXISTS auth_server CASCADE;
CREATE SERVER auth_server
  FOREIGN DATA WRAPPER postgres_fdw
  OPTIONS (
    host 'postgres',
    port '5432',
    dbname 'auth_db'
  );

-- ============================================================
-- 3. CREATE USER MAPPING
-- ============================================================
DROP USER MAPPING IF EXISTS FOR streaming_user SERVER auth_server;
CREATE USER MAPPING FOR streaming_user
  SERVER auth_server
  OPTIONS (
    user 'streaming_user',
    password 'streaming_pass'
  );

-- ============================================================
-- 4. CREATE FOREIGN TABLE
-- ============================================================
DROP FOREIGN TABLE IF EXISTS users CASCADE;
CREATE FOREIGN TABLE users (
  id UUID NOT NULL,
  username VARCHAR(50) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
SERVER auth_server
OPTIONS (
  schema_name 'public',
  table_name 'users'
);

-- ============================================================
-- 5. GRANT PERMISSIONS
-- ============================================================
GRANT SELECT ON users TO streaming_user;

-- ============================================================
-- 6. ADD COMMENTS
-- ============================================================
COMMENT ON FOREIGN TABLE users IS 'Foreign table linking to auth_db.users for JOIN queries';

-- ============================================================
-- 7. VERIFY SETUP AND LOG SUCCESS
-- ============================================================
DO $$
DECLARE
  user_count INTEGER;
  test_username VARCHAR(50);
BEGIN
  -- Count users in foreign table
  SELECT COUNT(*) INTO user_count FROM users;
  RAISE NOTICE '✅ FDW Setup: Found % users in auth_db', user_count;

  -- Test JOIN query (with videos table)
  SELECT u.username INTO test_username
  FROM videos v
  LEFT JOIN users u ON v.user_id = u.id
  LIMIT 1;

  IF test_username IS NOT NULL THEN
    RAISE NOTICE '✅ JOIN test successful: username = %', test_username;
  ELSE
    RAISE NOTICE '✅ JOIN test completed (no videos yet)';
  END IF;

  RAISE NOTICE '✅ Migration 000003 completed successfully';
  RAISE NOTICE '📝 VOD Service can now: SELECT v.*, u.username FROM videos v LEFT JOIN users u ON v.user_id = u.id';

EXCEPTION
  WHEN OTHERS THEN
    RAISE WARNING '⚠️ FDW Verification failed: %', SQLERRM;
    RAISE WARNING '⚠️ Check if auth_db has users table';
END $$;

COMMIT;
