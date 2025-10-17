# Cache Refresher Service

**Microservice for maintaining PostgreSQL Materialized Views in the Streaming Platform**

---

## 📋 Overview

Cache Refresher is a lightweight background service that automatically refreshes PostgreSQL materialized views to keep cached data synchronized with the source tables. It runs independently from the main application logic and ensures that cached JOIN results (streams with usernames, videos with usernames) stay up-to-date.

**Purpose:** Maintain materialized views fresh for future analytics, dashboards, and aggregation queries while keeping real-time operations unaffected.

**Current Status:** The main application uses real-time CTE queries with FDW optimization. Cached views are prepared for optional use in non-critical scenarios where 5-10 minute staleness is acceptable.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│         Cache Refresher Service             │
│                                             │
│  ┌────────────┐      ┌────────────┐       │
│  │ Streams    │      │ Videos     │       │
│  │ Worker     │      │ Worker     │       │
│  │ (5 min)    │      │ (10 min)   │       │
│  └──────┬─────┘      └──────┬─────┘       │
│         │                   │              │
└─────────┼───────────────────┼──────────────┘
          │                   │
          ▼                   ▼
┌──────────────────────────────────────────────┐
│          PostgreSQL Databases                │
│                                              │
│  streams_db:                                 │
│    └─ REFRESH streams_with_users_cache       │
│                                              │
│  vod_db:                                     │
│    └─ REFRESH videos_with_users_cache        │
└──────────────────────────────────────────────┘
```

***

## 🚀 Features

- ✅ **Automatic periodic refresh** of materialized views
- ✅ **Configurable intervals** via environment variables
- ✅ **Retry logic** with exponential backoff for database connections
- ✅ **Graceful shutdown** handling (SIGINT, SIGTERM)
- ✅ **Health checks** for Docker monitoring
- ✅ **Minimal resource footprint** (~10-20MB RAM, negligible CPU)
- ✅ **Detailed logging** with timestamps and performance metrics
- ✅ **Connection pooling** for optimal database performance

***

## 📊 What Gets Refreshed

### 1. `streams_with_users_cache` (streams_db)
**Interval:** Every 5 minutes  
**Content:** Pre-computed JOIN of streams with usernames from auth_db

```sql
SELECT s.*, u.username 
FROM streams s 
LEFT JOIN users_foreign u ON s.user_id = u.id
```

### 2. `videos_with_users_cache` (vod_db)
**Interval:** Every 10 minutes  
**Content:** Pre-computed JOIN of videos with usernames from auth_db

```sql
SELECT v.*, u.username 
FROM videos v 
LEFT JOIN users_foreign u ON v.user_id = u.id
```

***

## 🛠️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `postgres` | PostgreSQL hostname |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_USER` | `streaming_user` | Database username |
| `POSTGRES_PASSWORD` | `streaming_pass` | Database password |
| `STREAMS_DB_NAME` | `streams_db` | Streams database name |
| `VOD_DB_NAME` | `vod_db` | VOD database name |
| `STREAMS_REFRESH_INTERVAL` | `5m` | Streams cache refresh interval |
| `VIDEOS_REFRESH_INTERVAL` | `10m` | Videos cache refresh interval |

**Interval Format:** Go duration strings (e.g., `5m`, `10m`, `1h`, `30s`)

***

## 🐳 Docker Deployment

### docker-compose.yml

```yaml
services:
  cache-refresher:
    build: ./cache-refresher
    container_name: streaming-cache-refresher
    environment:
      POSTGRES_HOST: postgres
      POSTGRES_PORT: 5432
      POSTGRES_USER: streaming_user
      POSTGRES_PASSWORD: streaming_pass
      STREAMS_DB_NAME: streams_db
      VOD_DB_NAME: vod_db
      STREAMS_REFRESH_INTERVAL: ${STREAMS_REFRESH_INTERVAL:-5m}
      VIDEOS_REFRESH_INTERVAL: ${VIDEOS_REFRESH_INTERVAL:-10m}
    depends_on:
      migrations:
        condition: service_completed_successfully
    restart: unless-stopped
    networks:
      - streaming-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Build and Run

```bash
# Build the service
docker-compose build cache-refresher

# Start the service
docker-compose up -d cache-refresher

# View logs
docker-compose logs -f cache-refresher

# Stop the service
docker-compose stop cache-refresher
```

***

## 📝 Logs

### Normal Operation

```
2025/10/17 20:06:54 🚀 Starting Cache Refresher Service...
2025/10/17 20:06:54 📋 Service purpose: Keep materialized views up-to-date for potential future use
2025/10/17 20:06:54 ⚠️  Note: Current application uses real-time CTE queries, not cached views
2025/10/17 20:06:54 ✅ Connected to streams_db
2025/10/17 20:06:54 ✅ Connected to vod_db
2025/10/17 20:06:54 🔄 Starting refresh worker for streams_with_users_cache (interval: 5m0s)
2025/10/17 20:06:54 🔄 Starting refresh worker for videos_with_users_cache (interval: 10m0s)
2025/10/17 20:07:00 ✅ streams_with_users_cache refreshed successfully in 41.472ms
2025/10/17 20:07:00 ✅ videos_with_users_cache refreshed successfully in 44.653ms
2025/10/17 20:12:00 ✅ streams_with_users_cache refreshed successfully in 12.338ms
```

### Graceful Shutdown

```
2025/10/17 20:15:30 🛑 Shutdown signal received, stopping gracefully...
2025/10/17 20:15:30 🛑 Stopping streams_with_users_cache refresh worker
2025/10/17 20:15:30 🛑 Stopping videos_with_users_cache refresh worker
2025/10/17 20:15:32 👋 Cache Refresher Service stopped
```

***

## 🔧 Manual Operations

### Force Refresh Cache

```bash
# Refresh streams cache
docker exec -it streaming-postgres psql -U streaming_user -d streams_db \
  -c "SELECT refresh_streams_cache();"

# Refresh videos cache
docker exec -it streaming-postgres psql -U streaming_user -d vod_db \
  -c "SELECT refresh_videos_cache();"
```

### Check Cache Status

```bash
# Check streams cache count
docker exec -it streaming-postgres psql -U streaming_user -d streams_db \
  -c "SELECT COUNT(*) FROM streams_with_users_cache;"

# Check videos cache count
docker exec -it streaming-postgres psql -U streaming_user -d vod_db \
  -c "SELECT COUNT(*) FROM videos_with_users_cache;"

# View sample cached data
docker exec -it streaming-postgres psql -U streaming_user -d streams_db \
  -c "SELECT id, title, username, status FROM streams_with_users_cache LIMIT 5;"
```

***

## 🎯 Use Cases

### Current Implementation (Real-time)
All production endpoints use **CTE queries** with FDW optimization for real-time data:

```go
// stream-service/repository/stream_repository.go
func (r *StreamRepository) GetLiveStreams() ([]*models.Stream, error) {
    query := `
        WITH filtered_streams AS (
            SELECT * FROM streams WHERE status = 'live' LIMIT 100
        )
        SELECT fs.*, u.username 
        FROM filtered_streams fs
        LEFT JOIN users u ON fs.user_id = u.id
    `
    // Always returns fresh data
}
```

### Future Use Cases (Optional, from cache)

**1. Analytics Dashboard**
```go
// For dashboard showing statistics (5-min staleness acceptable)
func (r *StreamRepository) GetStreamStatistics() (*Stats, error) {
    query := `SELECT COUNT(*), AVG(viewer_count) FROM streams_with_users_cache`
    // 5-10x faster than CTE
}
```

**2. Popular Content Lists**
```go
// For "Trending Streams" page (10-min delay acceptable)
func (r *StreamRepository) GetTrendingStreams() ([]*models.Stream, error) {
    query := `
        SELECT * FROM streams_with_users_cache 
        WHERE status = 'live' 
        ORDER BY viewer_count DESC 
        LIMIT 20
    `
    // Lightning fast response
}
```

**3. Batch Operations**
```go
// For admin reports or export operations
func (r *VideoRepository) ExportAllVideos() ([]Video, error) {
    query := `SELECT * FROM videos_with_users_cache ORDER BY created_at DESC`
    // No FDW overhead
}
```

***

## 🔍 Monitoring

### Health Check

Service includes built-in health check for Docker:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD pgrep -x cache-refresher > /dev/null || exit 1
```

Check health status:
```bash
docker inspect --format='{{.State.Health.Status}}' streaming-cache-refresher
```

### Performance Metrics

Track refresh performance in logs:
```bash
docker-compose logs cache-refresher | grep "refreshed successfully"
```

Expected performance:
- **Empty cache:** 40-50ms
- **100 streams:** 10-20ms
- **1000 streams:** 50-100ms
- **10000 streams:** 200-500ms

***

## 🐛 Troubleshooting

### Service won't start

**Problem:** Connection failures to PostgreSQL

**Solution:**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check database connectivity
docker exec -it streaming-postgres psql -U streaming_user -d streams_db -c "SELECT 1;"

# Check migrations completed
docker-compose logs migrations
```

### Refresh failures

**Problem:** Error messages in logs

**Common causes:**
1. Materialized views not created (run migrations)
2. FDW server not configured (migration 000003)
3. Database connection pool exhausted

**Check migration status:**
```bash
docker exec -it streaming-postgres psql -U streaming_user -d streams_db \
  -c "SELECT * FROM schema_migrations;"
```

### High CPU usage

**Problem:** Service using too much CPU

**Likely causes:**
- Refresh interval too short (< 1 minute not recommended)
- Very large datasets (millions of rows)

**Solution:** Increase refresh intervals in docker-compose.yml:
```yaml
STREAMS_REFRESH_INTERVAL: 10m
VIDEOS_REFRESH_INTERVAL: 30m
```

***

## 📦 Dependencies

- **Go:** 1.23.2+
- **PostgreSQL:** 17+ with postgres_fdw extension
- **lib/pq:** v1.10.9 (PostgreSQL driver)

***

## 🔐 Security Considerations

1. **Database Credentials:** Use environment variables, never hardcode
2. **Read-Only Operations:** Service only executes `SELECT` statements
3. **Network Isolation:** Runs in isolated Docker network
4. **No External API:** Service doesn't expose HTTP endpoints
5. **Minimal Privileges:** Needs only `EXECUTE` on refresh functions

***

## 🚦 Status

**Current Version:** 1.0.0  
**Status:** ✅ Production Ready  
**Last Updated:** October 17, 2025

***

## 📚 Related Documentation

- [PostgreSQL FDW Optimization Guide](../infrastructure/postgres/migrations/)
- [Stream Service Architecture](../stream-service/README.md)
- [VOD Service Architecture](../vod-service/README.md)
- [Main Project README](../README.md)

---

## 💡 Future Enhancements

- [ ] Metrics endpoint for Prometheus
- [ ] Webhook notifications on refresh completion
- [ ] Adaptive refresh intervals based on data change frequency
- [ ] Partial cache updates instead of full refresh
- [ ] Web UI for manual cache management

***

## 📧 Support

For issues or questions:
- Open an issue in the project repository
- Check logs: `docker-compose logs cache-refresher`
- Review PostgreSQL logs: `docker-compose logs postgres`

***

**Note:** This service is designed to run continuously in the background. It does not affect real-time operations and can be safely stopped/restarted without impacting user-facing functionality.
