-- infrastructure/postgres/init.sql

-- =================================================================
-- Database Initialization Script
-- Purpose: Create databases, users, and enable extensions
-- Note: Table schemas are managed by migrations in ./migrations/
-- =================================================================

-- Create databases
CREATE DATABASE auth_db;
CREATE DATABASE streams_db;
CREATE DATABASE vod_db;

-- Create unified user for all databases
CREATE USER streaming_user WITH PASSWORD 'streaming_pass';

-- Grant database-level privileges
GRANT ALL PRIVILEGES ON DATABASE auth_db TO streaming_user;
GRANT ALL PRIVILEGES ON DATABASE streams_db TO streaming_user;
GRANT ALL PRIVILEGES ON DATABASE vod_db TO streaming_user;

-- ============================================================
-- AUTH_DB - Extensions and Base Grants
-- ============================================================
\c auth_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Grant default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO streaming_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO streaming_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO streaming_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO streaming_user;
GRANT USAGE, CREATE ON SCHEMA public TO streaming_user;

-- ============================================================
-- STREAMS_DB - Extensions and Base Grants + FDW
-- ============================================================
\c streams_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ✅ ДОБАВЛЕНО: Создать postgres_fdw расширение для Foreign Data Wrapper
CREATE EXTENSION IF NOT EXISTS postgres_fdw;

-- ✅ ДОБАВЛЕНО: Дать права streaming_user на использование FDW
GRANT USAGE ON FOREIGN DATA WRAPPER postgres_fdw TO streaming_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO streaming_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO streaming_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO streaming_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO streaming_user;
GRANT USAGE, CREATE ON SCHEMA public TO streaming_user;

-- ============================================================
-- VOD_DB - Extensions and Base Grants + FDW
-- ============================================================
\c vod_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ✅ ДОБАВЛЕНО: Создать postgres_fdw расширение для Foreign Data Wrapper
CREATE EXTENSION IF NOT EXISTS postgres_fdw;

-- ✅ ДОБАВЛЕНО: Дать права streaming_user на использование FDW
GRANT USAGE ON FOREIGN DATA WRAPPER postgres_fdw TO streaming_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO streaming_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO streaming_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO streaming_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO streaming_user;
GRANT USAGE, CREATE ON SCHEMA public TO streaming_user;

-- Success message
\c postgres;
SELECT 'Database initialization completed successfully!' AS status;
SELECT 'FDW Extension enabled in streams_db and vod_db for cross-database queries' AS fdw_status;
SELECT 'Run migrations to create table schemas.' AS next_step;
