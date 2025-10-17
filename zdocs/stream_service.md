# ✅ Stream Service - ПОЛНОЕ ОПИСАНИЕ

## Stream Service управляет FFmpeg процессами ВНУТРИ себя!

**Назначение**: 
- REST API для управления стримами
- **Встроенный FFmpeg менеджер** для транскодирования SRT → HLS
- Запускает и останавливает FFmpeg процессы по команде

***

## Архитектура (ПРАВИЛЬНАЯ)

```
Streamer (OBS)
    ↓ SRT stream
Stream Service (FFmpeg Manager внутри)
    ↓ FFmpeg process (SRT listener → HLS)
    ↓ HLS files → /tmp/hls/
Nginx → serves HLS
    ↓
Viewers
```

**Stream Service = REST API + FFmpeg Manager в одном процессе!**

***

## Конфигурация

**config.go**:
```go
type Config struct {
    Port          string
    DatabaseURL   string
    JWTSecret     string
    
    // FFmpeg конфигурация
    SRTListenerPort string  // "6000" - Stream Service слушает SRT!
    HLSOutputPath   string  // "/tmp/hls" - куда сохранять HLS
    HLSBaseURL      string  // "http://localhost/hls"
    
    // MinIO
    MinioEndpoint  string
    MinioAccessKey string
    MinioSecretKey string
    MinioBucket    string
}
```

**Environment Variables**:
```env
PORT=8002
DATABASE_URL=postgresql://user:pass@postgres:5432/stream_db
JWT_SECRET=secret
SRT_LISTENER_PORT=6000          # ✅ Stream Service слушает SRT
HLS_OUTPUT_PATH=/tmp/hls
HLS_BASE_URL=http://localhost/hls
MINIO_ENDPOINT=minio-storage:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_BUCKET=streams
```

***

## FFmpeg Manager (pkg/ffmpeg/ffmpeg.go)

**Структура**:
```go
type FFmpegManager struct {
    processes map[string]*FFmpegProcess  // streamKey → process
    mu        sync.RWMutex
    config    *FFmpegConfig
}

type FFmpegProcess struct {
    Cmd         *exec.Cmd
    StreamKey   string
    SRTPort     int        // Уникальный SRT порт для этого стрима
    HLSPath     string
    Profile     EncodingProfile
    Status      string
    StartedAt   time.Time
}

type FFmpegConfig struct {
    SRTBasePort   int     // 6000
    HLSOutputPath string  // /tmp/hls
    DefaultProfile EncodingProfile
}
```

**Методы**:

### 1. StartStream - запускает FFmpeg для стрима
```go
func (m *FFmpegManager) StartStream(streamKey string, profile EncodingProfile) error {
    // Выделяет уникальный SRT порт
    srtPort := m.allocatePort()
    
    // HLS output путь
    hlsPath := filepath.Join(m.config.HLSOutputPath, streamKey)
    os.MkdirAll(hlsPath, 0755)
    
    // FFmpeg command:
    // Слушает SRT на выделенном порту и транскодирует в HLS
    args := []string{
        "-re",
        "-i", fmt.Sprintf("srt://0.0.0.0:%d?mode=listener&streamid=%s", srtPort, streamKey),
        "-c:v", profile.VideoCodec,      // libx264
        "-preset", profile.Preset,        // veryfast
        "-b:v", profile.VideoBitrate,     // 5000k
        "-maxrate", profile.VideoBitrate,
        "-bufsize", "10000k",
        "-g", "60",                       // GOP size
        "-sc_threshold", "0",
        "-c:a", profile.AudioCodec,       // aac
        "-b:a", profile.AudioBitrate,     // 192k
        "-ar", "48000",
        "-f", "hls",
        "-hls_time", "2",                 // 2 second segments
        "-hls_list_size", "10",           // 10 segments in playlist
        "-hls_flags", "delete_segments+append_list",
        "-hls_segment_filename", filepath.Join(hlsPath, "segment_%03d.ts"),
        filepath.Join(hlsPath, "playlist.m3u8"),
    }
    
    cmd := exec.Command("ffmpeg", args...)
    cmd.Start()
    
    // Сохраняет процесс
    m.processes[streamKey] = &FFmpegProcess{
        Cmd: cmd,
        StreamKey: streamKey,
        SRTPort: srtPort,
        HLSPath: hlsPath,
        Status: "running",
        StartedAt: time.Now(),
    }
    
    return nil
}
```

### 2. StopStream - останавливает FFmpeg
```go
func (m *FFmpegManager) StopStream(streamKey string) error {
    process := m.processes[streamKey]
    if process == nil {
        return errors.New("stream not found")
    }
    
    // Graceful shutdown
    process.Cmd.Process.Signal(syscall.SIGTERM)
    
    // Удаляет HLS файлы
    os.RemoveAll(process.HLSPath)
    
    // Освобождает порт
    m.releasePort(process.SRTPort)
    
    delete(m.processes, streamKey)
    return nil
}
```

### 3. GetStreamStatus - статус FFmpeg процесса
```go
func (m *FFmpegManager) GetStreamStatus(streamKey string) (*FFmpegProcess, error)
```

### 4. Port allocation - управление SRT портами
```go
// Выделяет уникальный SRT порт для каждого стрима
// Начинает с 6000, потом 1936, 1937, и т.д.
func (m *FFmpegManager) allocatePort() int
func (m *FFmpegManager) releasePort(port int)
```

***

## Stream Handler API

**stream_handler.go** endpoints:

### POST /api/streams/:id/start
```go
func (h *StreamHandler) StartStream(c *gin.Context) {
    streamID := c.Param("id")
    stream := h.repo.GetStreamByID(streamID)
    
    // Запускает FFmpeg процесс
    profile := profiles.ProfileHD  // Можно выбирать
    err := h.ffmpegManager.StartStream(stream.StreamKey, profile)
    
    // Обновляет DB
    stream.Status = "live"
    stream.IsLive = true
    stream.StartedAt = time.Now()
    h.repo.UpdateStream(stream)
    
    // Возвращает SRT URL с выделенным портом
    process := h.ffmpegManager.GetStreamStatus(stream.StreamKey)
    c.JSON(200, gin.H{
        "srt_url": fmt.Sprintf("srt://YOUR_SERVER_IP:%d?streamid=%s", 
                              process.SRTPort, stream.StreamKey),
        "hls_url": fmt.Sprintf("%s/%s/playlist.m3u8", 
                              h.config.HLSBaseURL, stream.StreamKey),
    })
}
```

### POST /api/streams/:id/stop
```go
func (h *StreamHandler) StopStream(c *gin.Context) {
    streamID := c.Param("id")
    stream := h.repo.GetStreamByID(streamID)
    
    // Останавливает FFmpeg
    h.ffmpegManager.StopStream(stream.StreamKey)
    
    // Обновляет DB
    stream.Status = "ended"
    stream.IsLive = false
    stream.EndedAt = time.Now()
    h.repo.UpdateStream(stream)
}
```

### POST /api/streams (Create)
```go
func (h *StreamHandler) CreateStream(c *gin.Context) {
    // Создает запись в DB
    stream := &models.Stream{
        ID:        uuid.New(),
        UserID:    userID,
        Title:     req.Title,
        StreamKey: fmt.Sprintf("user_%s_%s", userID, streamID),
        Status:    "ready",
        IsLive:    false,
    }
    h.repo.CreateStream(stream)
    
    // Возвращает URLs (FFmpeg еще не запущен!)
    c.JSON(200, gin.H{
        "id": stream.ID,
        "stream_key": stream.StreamKey,
        // Порт будет выделен при START
        "message": "Call POST /streams/:id/start to begin streaming",
    })
}
```

***

## Encoding Profiles (pkg/profiles/profiles.go)

```go
type EncodingProfile struct {
    Name            string
    VideoCodec      string
    AudioCodec      string
    VideoResolution string
    VideoBitrate    string
    AudioBitrate    string
    Preset          string
}

var (
    ProfileHD = EncodingProfile{
        Name:            "hd",
        VideoCodec:      "libx264",
        AudioCodec:      "aac",
        VideoResolution: "1920x1080",
        VideoBitrate:    "5000k",
        AudioBitrate:    "192k",
        Preset:          "veryfast",
    }
    
    ProfileSD = EncodingProfile{
        Name:            "sd",
        VideoCodec:      "libx264",
        AudioCodec:      "aac",
        VideoResolution: "1280x720",
        VideoBitrate:    "2500k",
        AudioBitrate:    "128k",
        Preset:          "veryfast",
    }
)
```

***

## Полный Workflow

### 1. Create Stream
```bash
POST /api/streams
{
  "title": "My Stream"
}

Response:
{
  "id": "uuid",
  "stream_key": "user_123_456",
  "status": "ready",
  "message": "Call POST /streams/:id/start to begin"
}
```

### 2. Start Stream (запускает FFmpeg)
```bash
POST /api/streams/:id/start

Response:
{
  "srt_url": "srt://192.168.1.100:6000?streamid=user_123_456",
  "hls_url": "http://192.168.1.100/hls/user_123_456/playlist.m3u8",
  "status": "live"
}
```

### 3. Configure OBS
```
Settings → Stream
Server: srt://192.168.1.100:6000?streamid=user_123_456
```

### 4. Start Streaming from OBS
```
OBS → connects to SRT listener (FFmpeg)
FFmpeg → transcodes to HLS
HLS files → created in /tmp/hls/user_123_456/
```

### 5. Watch Stream
```
Browser → GET http://server/hls/user_123_456/playlist.m3u8
Nginx → serves HLS files
Video.js → plays HLS
```

### 6. Stop Stream
```bash
POST /api/streams/:id/stop

FFmpeg process → terminated
HLS files → deleted
DB → is_live = false, ended_at = NOW()
```

***

## Docker Compose (ПРАВИЛЬНЫЙ)

```yaml
services:
  stream-service:
    build: ./stream-service
    ports:
      - "8002:8002"        # HTTP API
      - "6000-1945:6000-1945/udp"  # SRT ports (10 concurrent streams)
    environment:
      PORT: 8002
      DATABASE_URL: postgresql://user:pass@postgres:5432/stream_db
      JWT_SECRET: secret
      SRT_LISTENER_PORT: 6000      # ✅ Base SRT port
      HLS_OUTPUT_PATH: /tmp/hls
      HLS_BASE_URL: http://YOUR_SERVER_IP/hls
      MINIO_ENDPOINT: minio-storage:9000
      MINIO_ACCESS_KEY: minioadmin
      MINIO_SECRET_KEY: minioadmin123
      MINIO_BUCKET: streams
    volumes:
      - hls-data:/tmp/hls
    depends_on:
      - postgres
      - minio-storage
```

**ВАЖНО**: 
- Stream Service слушает **диапазон SRT портов** (6000-1945)
- Каждый активный стрим получает уникальный порт
- FFmpeg процессы запускаются ВНУТРИ stream-service контейнера

***

## Dockerfile для Stream Service

```dockerfile
FROM golang:1.21-alpine AS builder

# Install FFmpeg
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o stream-service cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app
COPY --from=builder /app/stream-service .

EXPOSE 8002 6000-1945

CMD ["./stream-service"]
```

***

## Ключевые особенности

1. ✅ **Stream Service управляет FFmpeg процессами**
2. ✅ **SRT listener встроен через FFmpeg**
3. ✅ **Динамическое выделение портов** (6000, 1936, 1937...)
4. ✅ **HLS генерируется в реальном времени**
5. ✅ **Graceful shutdown** FFmpeg процессов
6. ✅ **Автоматическая очистка** HLS файлов

***

**Документация обновлена**: 10.10.2025, 02:33 MSK

Теперь описание **полностью корректное** - Stream Service управляет FFmpeg процессами внутри себя!