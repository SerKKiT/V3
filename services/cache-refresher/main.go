package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	StreamsDBDSN           string
	VodDBDSN               string
	StreamsRefreshInterval time.Duration
	VideosRefreshInterval  time.Duration
}

func loadConfig() *Config {
	// Парсинг интервалов обновления
	streamsInterval, err := time.ParseDuration(getEnv("STREAMS_REFRESH_INTERVAL", "5m"))
	if err != nil {
		log.Fatalf("❌ Invalid STREAMS_REFRESH_INTERVAL: %v", err)
	}

	videosInterval, err := time.ParseDuration(getEnv("VIDEOS_REFRESH_INTERVAL", "10m"))
	if err != nil {
		log.Fatalf("❌ Invalid VIDEOS_REFRESH_INTERVAL: %v", err)
	}

	// Формирование DSN для баз данных
	postgresUser := getEnv("POSTGRES_USER", "streaming_user")
	postgresPassword := getEnv("POSTGRES_PASSWORD", "streaming_pass")
	postgresHost := getEnv("POSTGRES_HOST", "postgres")
	postgresPort := getEnv("POSTGRES_PORT", "5432")

	return &Config{
		StreamsDBDSN: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			postgresUser,
			postgresPassword,
			postgresHost,
			postgresPort,
			getEnv("STREAMS_DB_NAME", "streams_db"),
		),
		VodDBDSN: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			postgresUser,
			postgresPassword,
			postgresHost,
			postgresPort,
			getEnv("VOD_DB_NAME", "vod_db"),
		),
		StreamsRefreshInterval: streamsInterval,
		VideosRefreshInterval:  videosInterval,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Подключение к БД с retry логикой
func connectWithRetry(dsn string, dbName string, maxRetries int) *sql.DB {
	var db *sql.DB
	var err error

	for i := 1; i <= maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			// Попытка ping для проверки соединения
			err = db.Ping()
			if err == nil {
				log.Printf("✅ Connected to %s", dbName)
				// Настройки connection pool
				db.SetMaxOpenConns(5)
				db.SetMaxIdleConns(2)
				db.SetConnMaxLifetime(time.Hour)
				return db
			}
		}

		if i < maxRetries {
			log.Printf("⏳ Waiting for %s... (attempt %d/%d)", dbName, i, maxRetries)
			time.Sleep(5 * time.Second)
		}
	}

	log.Fatalf("❌ Failed to connect to %s after %d attempts: %v", dbName, maxRetries, err)
	return nil
}

// Обновление materialized view
func refreshCache(db *sql.DB, functionName, cacheName string) error {
	start := time.Now()

	_, err := db.Exec(fmt.Sprintf("SELECT %s()", functionName))
	if err != nil {
		return fmt.Errorf("failed to refresh %s: %w", cacheName, err)
	}

	elapsed := time.Since(start)
	log.Printf("✅ %s refreshed successfully in %v", cacheName, elapsed)
	return nil
}

// Worker для периодического обновления кэша
func startRefreshWorker(db *sql.DB, functionName, cacheName string, interval time.Duration, done <-chan struct{}) {
	log.Printf("🔄 Starting refresh worker for %s (interval: %v)", cacheName, interval)

	// Первое обновление сразу при старте (с задержкой 5 сек для инициализации БД)
	time.Sleep(5 * time.Second)
	if err := refreshCache(db, functionName, cacheName); err != nil {
		log.Printf("❌ Initial refresh failed for %s: %v", cacheName, err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := refreshCache(db, functionName, cacheName); err != nil {
				log.Printf("❌ %v", err)
			}
		case <-done:
			log.Printf("🛑 Stopping %s refresh worker", cacheName)
			return
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("🚀 Starting Cache Refresher Service...")
	log.Println("📋 Service purpose: Keep materialized views up-to-date for potential future use")
	log.Println("⚠️  Note: Current application uses real-time CTE queries, not cached views")

	config := loadConfig()

	// Подключение к базам данных (с retry)
	streamsDB := connectWithRetry(config.StreamsDBDSN, "streams_db", 10)
	defer func() {
		if err := streamsDB.Close(); err != nil {
			log.Printf("❌ Error closing streams_db: %v", err)
		}
	}()

	vodDB := connectWithRetry(config.VodDBDSN, "vod_db", 10)
	defer func() {
		if err := vodDB.Close(); err != nil {
			log.Printf("❌ Error closing vod_db: %v", err)
		}
	}()

	// Канал для graceful shutdown
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск worker'ов для обновления кэшей
	go startRefreshWorker(
		streamsDB,
		"refresh_streams_cache",
		"streams_with_users_cache",
		config.StreamsRefreshInterval,
		done,
	)

	go startRefreshWorker(
		vodDB,
		"refresh_videos_cache",
		"videos_with_users_cache",
		config.VideosRefreshInterval,
		done,
	)

	log.Println("📊 Refresh intervals configured:")
	log.Printf("   - Streams cache: every %v", config.StreamsRefreshInterval)
	log.Printf("   - Videos cache: every %v", config.VideosRefreshInterval)
	log.Println("✅ Cache Refresher Service is running")
	log.Println("💡 Tip: Use cached views for analytics/dashboards in the future")

	// Ожидание сигнала остановки
	<-sigChan
	log.Println("🛑 Shutdown signal received, stopping gracefully...")
	close(done)

	// Даём время завершить текущие операции
	time.Sleep(2 * time.Second)
	log.Println("👋 Cache Refresher Service stopped")
}
