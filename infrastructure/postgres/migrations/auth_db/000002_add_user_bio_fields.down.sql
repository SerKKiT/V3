-- infrastructure/postgres/migrations/auth_db/000002_add_user_bio_fields.down.sql
-- Rollback: Remove profile fields from users table

BEGIN;

ALTER TABLE users 
DROP CONSTRAINT IF EXISTS bio_length,
DROP COLUMN IF EXISTS bio,
DROP COLUMN IF EXISTS avatar_url,
DROP COLUMN IF EXISTS display_name;

COMMIT;
