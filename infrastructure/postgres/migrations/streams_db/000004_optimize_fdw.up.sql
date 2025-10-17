-- infrastructure/postgres/migrations/streams_db/000004_optimize_fdw.up.sql

-- Migration: Optimize Foreign Data Wrapper performance
-- Description: Increase fetch_size and use_remote_estimate for better FDW performance

BEGIN;

-- ============================================================
-- 1. OPTIMIZE FDW SERVER SETTINGS
-- ============================================================

DO $$
BEGIN
    -- Увеличить fetch_size с 100 (по умолчанию) до 10000
    -- Это уменьшает количество FETCH операций при больших выборках
    ALTER SERVER auth_server OPTIONS (ADD fetch_size '10000');
    RAISE NOTICE '✅ FDW fetch_size increased to 10000';
EXCEPTION
    WHEN duplicate_object THEN
        ALTER SERVER auth_server OPTIONS (SET fetch_size '10000');
        RAISE NOTICE '✅ FDW fetch_size updated to 10000';
END $$;

DO $$
BEGIN
    -- Включить use_remote_estimate для точных оценок планировщика
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

-- Материализованное представление для кэширования JOIN с users
DROP MATERIALIZED VIEW IF EXISTS streams_with_users_cache CASCADE;

CREATE MATERIALIZED VIEW streams_with_users_cache AS
SELECT 
    s.id,
    s.user_id,
    s.stream_key,
    s.title,
    s.description,
    s.status,
    s.started_at,
    s.ended_at,
    s.available_qualities,
    s.viewer_count,
    s.thumbnail_url,
    s.hls_url,
    s.created_at,
    s.updated_at,
    COALESCE(u.username, 'Unknown Streamer') as username,
    COALESCE(u.email, '') as user_email
FROM streams s
LEFT JOIN users u ON s.user_id = u.id;

-- Создать уникальный индекс для CONCURRENTLY refresh
CREATE UNIQUE INDEX idx_streams_cache_id ON streams_with_users_cache(id);

-- Индексы для materialized view
CREATE INDEX idx_streams_cache_status ON streams_with_users_cache(status);
CREATE INDEX idx_streams_cache_user_id ON streams_with_users_cache(user_id);
CREATE INDEX idx_streams_cache_created_at ON streams_with_users_cache(created_at DESC);
CREATE INDEX idx_streams_cache_live ON streams_with_users_cache(status, started_at DESC) WHERE status = 'live';

-- ============================================================
-- 3. CREATE REFRESH FUNCTION
-- ============================================================

-- Функция для автоматического обновления materialized view
CREATE OR REPLACE FUNCTION refresh_streams_cache()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY streams_with_users_cache;
    RAISE NOTICE '✅ Streams cache refreshed at %', NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 4. ADD COMMENTS
-- ============================================================

COMMENT ON MATERIALIZED VIEW streams_with_users_cache IS 
'Cached JOIN of streams with users from auth_db. Refresh periodically for best performance.';

COMMENT ON FUNCTION refresh_streams_cache IS 
'Refreshes the streams_with_users_cache materialized view. Call this periodically (e.g., every 5 minutes).';

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
    SELECT COUNT(*) INTO cache_count FROM streams_with_users_cache;
    
    RAISE NOTICE '✅ FDW Optimization Summary:';
    RAISE NOTICE '   - Server OID: %', server_oid;
    RAISE NOTICE '   - fetch_size: %', COALESCE(fetch_size_value, 'not set');
    RAISE NOTICE '   - use_remote_estimate: %', COALESCE(remote_estimate, 'not set');
    RAISE NOTICE '   - Cached streams: %', cache_count;
    RAISE NOTICE '📝 Usage:';
    RAISE NOTICE '   - For live data: Use CTE queries (see repository code)';
    RAISE NOTICE '   - For cached data: SELECT * FROM streams_with_users_cache';
    RAISE NOTICE '   - Refresh cache: SELECT refresh_streams_cache()';
END $$;

COMMIT;
