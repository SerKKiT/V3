-- infrastructure/postgres/migrations/auth_db/000001_initial_auth_schema.down.sql
-- Rollback: Drop users table and related objects

BEGIN;

DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

DROP TABLE IF EXISTS users CASCADE;

COMMIT;
