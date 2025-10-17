-- infrastructure/postgres/migrations/auth_db/000002_add_user_bio_fields.up.sql
-- Migration: Add profile fields to users table
-- Description: Add display_name, avatar_url, and bio for user profiles

BEGIN;

-- Add new columns
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS display_name VARCHAR(100),
ADD COLUMN IF NOT EXISTS avatar_url TEXT,
ADD COLUMN IF NOT EXISTS bio TEXT;

-- Add constraints
ALTER TABLE users 
ADD CONSTRAINT bio_length CHECK (char_length(bio) <= 500);

-- Comments
COMMENT ON COLUMN users.display_name IS 'Public display name (optional, max 100 chars)';
COMMENT ON COLUMN users.avatar_url IS 'URL to user avatar image';
COMMENT ON COLUMN users.bio IS 'User biography/description (max 500 chars)';

-- Update test user with display name
UPDATE users 
SET display_name = 'Gateway User'
WHERE username = 'gateway_user' AND display_name IS NULL;

COMMIT;
