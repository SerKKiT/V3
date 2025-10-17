-- infrastructure/postgres/migrations/vod_db/000004_optimize_fdw.up.sql

BEGIN;

-- ============================================================
-- 1. OPTIMIZE FDW SERVER SETTINGS
-- ============================================================

DO $$
BEGIN
    ALTER SERVER auth_server OPTIONS (ADD fetch_size '10000');
    RAISE NOTICE '✅ FDW fetch_size increased to 10000';
EXCEPTION
    WHEN duplicate_object THEN
        ALTER SERVER auth_server OPTIONS (SET fetch_size '10000');
        RAISE NOTICE '✅ FDW fetch_size updated to 10000';
END $$;

DO $$
BEGIN
    ALTER SERVER auth_server OPTIONS (ADD use_remote_estimate 'true');
    RAISE NOTICE '✅ FDW use_remote_estimate enabled';
EXCEPTION
    WHEN duplicate_object THEN
        ALTER SERVER auth_server OPTIONS (SET use_remote_estimate 'true');
        RAISE NOTICE '✅ FDW use_remote_estimate already enabled';
END $$;

-- ============================================================
-- 2. CREATE MATERIALIZED VIEW FOR CACHING
-- ============================================================

DROP MATERIALIZED VIEW IF EXISTS videos_with_users_cache CASCADE;

CREATE MATERIALIZED VIEW videos_with_users_cache AS
SELECT 
    v.id,
    v.user_id,
    v.recording_id,
    v.stream_id,
    v.title,
    v.description,
    v.category,
    v.tags,
    v.source,
    v.status,
    v.visibility,
    v.file_path,
    v.thumbnail_path,
    v.duration,
    v.file_size,
    v.view_count,
    v.like_count,
    v.created_at,
    v.updated_at,
    v.published_at,
    COALESCE(u.username, 'Unknown') as username,
    COALESCE(u.email, '') as user_email
FROM videos v
LEFT JOIN users u ON v.user_id = u.id;

-- Создать уникальный индекс для CONCURRENTLY refresh
CREATE UNIQUE INDEX idx_videos_cache_id ON videos_with_users_cache(id);

-- Индексы для materialized view
CREATE INDEX idx_videos_cache_user_id ON videos_with_users_cache(user_id);
CREATE INDEX idx_videos_cache_visibility ON videos_with_users_cache(visibility);
CREATE INDEX idx_videos_cache_created_at ON videos_with_users_cache(created_at DESC);
CREATE INDEX idx_videos_cache_status ON videos_with_users_cache(status);
CREATE INDEX idx_videos_cache_category ON videos_with_users_cache(category);

-- ============================================================
-- 3. CREATE REFRESH FUNCTION
-- ============================================================

CREATE OR REPLACE FUNCTION refresh_videos_cache()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY videos_with_users_cache;
    RAISE NOTICE '✅ Videos cache refreshed at %', NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 4. ADD COMMENTS
-- ============================================================

COMMENT ON MATERIALIZED VIEW videos_with_users_cache IS 
'Cached JOIN of videos with users from auth_db. Refresh periodically for best performance.';

COMMENT ON FUNCTION refresh_videos_cache IS 
'Refreshes the videos_with_users_cache materialized view. Call this periodically (e.g., every 10 minutes).';

-- ============================================================
-- 5. VERIFY OPTIMIZATION
-- ============================================================

DO $$
DECLARE
    fetch_size_value TEXT;
    remote_estimate TEXT;
    cache_count INTEGER;
    server_oid OID;
BEGIN
    -- Получить OID FDW сервера auth_server
    SELECT oid INTO server_oid 
    FROM pg_foreign_server 
    WHERE srvname = 'auth_server';
    
    IF server_oid IS NULL THEN
        RAISE EXCEPTION 'FDW server "auth_server" not found';
    END IF;
    
    -- Проверить настройки FDW через OID
    SELECT option_value INTO fetch_size_value
    FROM pg_options_to_table((SELECT srvoptions FROM pg_foreign_server WHERE oid = server_oid))
    WHERE option_name = 'fetch_size';
    
    SELECT option_value INTO remote_estimate
    FROM pg_options_to_table((SELECT srvoptions FROM pg_foreign_server WHERE oid = server_oid))
    WHERE option_name = 'use_remote_estimate';
    
    -- Подсчитать записи в кэше
    SELECT COUNT(*) INTO cache_count FROM videos_with_users_cache;
    
    RAISE NOTICE '✅ FDW Optimization Summary:';
    RAISE NOTICE '   - Server OID: %', server_oid;
    RAISE NOTICE '   - fetch_size: %', COALESCE(fetch_size_value, 'not set');
    RAISE NOTICE '   - use_remote_estimate: %', COALESCE(remote_estimate, 'not set');
    RAISE NOTICE '   - Cached videos: %', cache_count;
    RAISE NOTICE '📝 Usage:';
    RAISE NOTICE '   - For live data: Use CTE queries (see repository code)';
    RAISE NOTICE '   - For cached data: SELECT * FROM videos_with_users_cache';
    RAISE NOTICE '   - Refresh cache: SELECT refresh_videos_cache()';
END $$;

COMMIT;
