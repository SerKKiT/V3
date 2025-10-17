# Recording Service — Документация

## Назначение

Recording Service — это микросервис, отвечающий за автоматическую запись live трансляций. Он мониторит активные стримы, записывает их в реальном времени и сохраняет готовые записи в объектное хранилище MinIO.

### Основные функции

1. **Автоматический мониторинг стримов** — отслеживание начала и окончания трансляций через webhook и polling
2. **Запись HLS сегментов** — скачивание .ts файлов во время трансляции
3. **Конкатенация в MP4** — объединение сегментов в единый видеофайл после окончания стрима
4. **Хранение в MinIO** — загрузка готовых записей в объектное хранилище
5. **Управление метаданными** — сохранение информации о записях в PostgreSQL

### Интеграция с другими сервисами

```
┌──────────────────┐
│  Stream Service  │ ──webhook──> Recording Service
│     :8082        │              (начало/конец стрима)
└──────────────────┘
                                         │
                                         │ скачивает
                                         ▼
                                  ┌─────────────┐
                                  │   MinIO     │
                                  │  (segments) │
                                  └─────────────┘
                                         │
                            FFmpeg       │
                         конкатенация    │
                                         ▼
                                  ┌─────────────┐
                                  │   MinIO     │
                                  │ (recordings)│
                                  └─────────────┘
                                         │
                                         │ импорт
                                         ▼
                                  ┌─────────────┐
                                  │ VOD Service │
                                  │    :8084    │
                                  └─────────────┘
```

***

## Архитектура

### Компоненты

1. **Stream Monitor** — мониторинг активных стримов
2. **FFmpeg Recorder** — запись и конкатенация видео
3. **MinIO Storage** — работа с объектным хранилищем
4. **Recording Repository** — управление данными в PostgreSQL
5. **Recording Handler** — HTTP API обработчики

### Процесс записи

```
1. Stream Service отправляет webhook: "stream started"
   ↓
2. Recording Service создаёт запись в БД (status: recording)
   ↓
3. Stream Monitor начинает скачивание HLS сегментов
   ↓
4. Сегменты сохраняются локально
   ↓
5. Stream Service отправляет webhook: "stream ended"
   ↓
6. FFmpeg Recorder конкатенирует сегменты → MP4
   ↓
7. MP4 загружается в MinIO (bucket: recordings)
   ↓
8. Локальные файлы удаляются
   ↓
9. Запись обновляется в БД (status: completed)
```

***

## API Endpoints

### Базовый URL

```
http://localhost:8083
```

Через API Gateway:
```
http://localhost:8080/api/recordings
```

***

## 1. Service Endpoints

### 1.1. Health Check

Проверка работоспособности сервиса.

**Endpoint:** `GET /health`

**Авторизация:** Не требуется

**Пример запроса:**
```bash
curl http://localhost:8083/health
```

**Ответ:**
```json
{
  "status": "healthy",
  "service": "recording-service"
}
```

***

## 2. Recording Management

### 2.1. Получить список всех записей

Возвращает список всех записей, отсортированных по дате создания (новые первыми).

**Endpoint:** `GET /recordings`

**Авторизация:** Не требуется

**Query Parameters:** Нет

**Пример запроса:**
```bash
curl http://localhost:8083/recordings
```

**Пример через API Gateway:**
```bash
curl http://localhost:8080/api/recordings
```

**Ответ:**
```json
{
  "recordings": [
    {
      "id": "e410066c-8d50-4e0f-965f-452af1814ffa",
      "stream_id": "bb52ba50-edc5-4056-a62c-6ef211d72d25",
      "video_id": null,
      "file_path": "recordings/d1eb2fe800d01c797c84f0b622e474d9.mp4",
      "duration": 1860,
      "file_size": 524288000,
      "status": "completed",
      "started_at": "2025-10-05T18:00:00Z",
      "completed_at": "2025-10-05T18:31:00Z"
    },
    {
      "id": "a1b2c3d4-5678-90ab-cdef-1234567890ab",
      "stream_id": "xyz-stream-id",
      "video_id": "vod-video-uuid",
      "file_path": "recordings/stream_key_123.mp4",
      "duration": 3600,
      "file_size": 1073741824,
      "status": "completed",
      "started_at": "2025-10-04T15:00:00Z",
      "completed_at": "2025-10-04T16:00:00Z"
    }
  ]
}
```

**Поля ответа:**

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Уникальный идентификатор записи |
| `stream_id` | UUID | ID стрима, который был записан |
| `video_id` | UUID \| null | ID видео в VOD (если импортировано) |
| `file_path` | string | Путь к файлу в MinIO |
| `duration` | integer | Длительность записи в секундах |
| `file_size` | integer | Размер файла в байтах |
| `status` | string | Статус: `recording`, `processing`, `completed`, `failed` |
| `started_at` | timestamp | Время начала записи |
| `completed_at` | timestamp \| null | Время завершения обработки |

***

### 2.2. Получить запись по ID

Возвращает информацию о конкретной записи.

**Endpoint:** `GET /recording/:id`

**Авторизация:** Не требуется

**Path Parameters:**
- `id` (UUID) — идентификатор записи

**Пример запроса:**
```bash
curl http://localhost:8083/recording/e410066c-8d50-4e0f-965f-452af1814ffa
```

**Пример через API Gateway:**
```bash
curl http://localhost:8080/api/recordings/e410066c-8d50-4e0f-965f-452af1814ffa
```

**Ответ (успех):**
```json
{
  "recording": {
    "id": "e410066c-8d50-4e0f-965f-452af1814ffa",
    "stream_id": "bb52ba50-edc5-4056-a62c-6ef211d72d25",
    "video_id": null,
    "file_path": "recordings/d1eb2fe800d01c797c84f0b622e474d9.mp4",
    "duration": 1860,
    "file_size": 524288000,
    "status": "completed",
    "started_at": "2025-10-05T18:00:00Z",
    "completed_at": "2025-10-05T18:31:00Z"
  }
}
```

**Ответ (не найдено):**
```json
{
  "error": "Recording not found"
}
```

**HTTP Status:** 404 Not Found

***

## 3. Webhook Endpoints

### 3.1. Stream Webhook

Принимает уведомления от Stream Service о событиях стрима.

**Endpoint:** `POST /webhook/stream`

**Авторизация:** Внутренний endpoint (только для Stream Service)

**Request Body:**
```json
{
  "event": "stream.started" | "stream.ended",
  "stream_id": "uuid",
  "stream_key": "string",
  "user_id": "uuid",
  "timestamp": "2025-10-05T18:00:00Z"
}
```

**События:**

#### `stream.started`

Отправляется при начале трансляции.

**Действия Recording Service:**
1. Создаёт запись в БД со статусом `recording`
2. Запускает мониторинг HLS сегментов
3. Начинает скачивание `.ts` файлов

**Пример webhook:**
```json
{
  "event": "stream.started",
  "stream_id": "bb52ba50-edc5-4056-a62c-6ef211d72d25",
  "stream_key": "d1eb2fe800d01c797c84f0b622e474d9",
  "user_id": "a781b655-d4e4-46a0-89da-c4f655bb24f0",
  "timestamp": "2025-10-05T18:00:00Z"
}
```

#### `stream.ended`

Отправляется при окончании трансляции.

**Действия Recording Service:**
1. Останавливает мониторинг
2. Меняет статус на `processing`
3. Запускает FFmpeg для конкатенации сегментов
4. Загружает MP4 в MinIO
5. Обновляет статус на `completed`
6. Удаляет временные файлы

**Пример webhook:**
```json
{
  "event": "stream.ended",
  "stream_id": "bb52ba50-edc5-4056-a62c-6ef211d72d25",
  "stream_key": "d1eb2fe800d01c797c84f0b622e474d9",
  "user_id": "a781b655-d4e4-46a0-89da-c4f655bb24f0",
  "timestamp": "2025-10-05T18:31:00Z"
}
```

**Ответ (успех):**
```json
{
  "message": "Webhook processed successfully"
}
```

**Ответ (ошибка):**
```json
{
  "error": "Invalid event type"
}
```

***

## 4. Внутренние процессы

### 4.1. Stream Monitor

**Назначение:** Непрерывный мониторинг активных стримов и скачивание сегментов.

**Алгоритм работы:**
1. Каждые 10 секунд проверяет список активных записей (статус `recording`)
2. Для каждой записи:
   - Скачивает новые HLS сегменты из MinIO (bucket: `live-segments`)
   - Сохраняет локально в `/tmp/recordings/{stream_key}/`
   - Обновляет метаданные в БД

**Конфигурация:**
```go
MonitorInterval: 10 * time.Second
SegmentTimeout:  30 * time.Second
MaxRetries:      3
```

**Fallback механизм:**
- Если webhook не сработал, polling обнаружит изменение статуса стрима через 10 секунд
- Если сегмент недоступен, повтор через 5 секунд (максимум 3 попытки)

---

### 4.2. FFmpeg Recorder

**Назначение:** Конкатенация HLS сегментов в единый MP4 файл.

**Процесс:**

1. **Создание списка сегментов** (`segments.txt`):
```
file '/tmp/recordings/stream_key/segment_000.ts'
file '/tmp/recordings/stream_key/segment_001.ts'
file '/tmp/recordings/stream_key/segment_002.ts'
...
```

2. **FFmpeg команда:**
```bash
ffmpeg -f concat -safe 0 -i segments.txt \
  -c copy \
  -movflags +faststart \
  output.mp4
```

**Параметры:**
- `-f concat` — режим конкатенации
- `-safe 0` — разрешить любые пути к файлам
- `-c copy` — копирование без перекодирования (быстро)
- `-movflags +faststart` — оптимизация для веб-проигрывателей (метаданные в начале файла)

**Длительность обработки:**
- Конкатенация: ~10-30 секунд (зависит от количества сегментов)
- Загрузка в MinIO: зависит от размера файла

***

### 4.3. MinIO Storage Integration

**Buckets:**

| Bucket | Содержимое | Используется |
|--------|-----------|-------------|
| `live-segments` | HLS сегменты (.ts, .m3u8) во время стрима | Чтение (источник) |
| `recordings` | Готовые MP4 записи | Запись (результат) |

**Структура в `recordings` bucket:**
```
recordings/
  d1eb2fe800d01c797c84f0b622e474d9.mp4
  abc123def456.mp4
  xyz789.mp4
```

**Именование файлов:**
```
{stream_key}.mp4
```

**MinIO URLs:**
```
Internal: http://minio:9000/recordings/stream_key.mp4
Public:   http://localhost:9000/recordings/stream_key.mp4
```

***

## 5. База данных

### 5.1. Таблица `recordings`

**Схема:**
```sql
CREATE TABLE recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID,
    video_id UUID,
    file_path TEXT NOT NULL,
    duration INT DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    status recording_status DEFAULT 'recording',
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE TYPE recording_status AS ENUM (
    'recording',   -- В процессе записи
    'processing',  -- Обработка (конкатенация)
    'completed',   -- Завершено
    'failed'       -- Ошибка
);
```

**Индексы:**
```sql
CREATE INDEX idx_recordings_stream_id ON recordings(stream_id);
CREATE INDEX idx_recordings_status ON recordings(status);
CREATE INDEX idx_recordings_video_id ON recordings(video_id);
```

***

### 5.2. Жизненный цикл записи

```
1. [recording]   ← Webhook: stream.started
   - file_path: ""
   - duration: 0
   - file_size: 0
   - started_at: NOW
   - completed_at: NULL

2. [recording]   ← Мониторинг скачивает сегменты
   - duration обновляется каждые 10 сек

3. [processing]  ← Webhook: stream.ended
   - FFmpeg конкатенирует сегменты

4. [completed]   ← Успешная загрузка в MinIO
   - file_path: "recordings/stream_key.mp4"
   - duration: финальное значение
   - file_size: размер MP4
   - completed_at: NOW

ИЛИ

4. [failed]      ← Ошибка обработки
   - completed_at: NOW
```

***

## 6. Конфигурация

### 6.1. Переменные окружения

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=streaming_user
DB_PASSWORD=streaming_pass
DB_NAME=vod_db

# MinIO
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MINIO_BUCKET=recordings

# Service
PORT=8083
STREAM_SERVICE_URL=http://stream-service:8082
VOD_SERVICE_URL=http://vod-service:8084

# Monitoring
MONITOR_INTERVAL=10s
SEGMENT_DOWNLOAD_TIMEOUT=30s
MAX_RETRY_ATTEMPTS=3
```

***

### 6.2. Docker Compose

```yaml
recording-service:
  build: ./services/recording-service
  container_name: streaming-recording
  environment:
    DATABASE_URL: ${VOD_DB_URL}
    DB_HOST: ${DB_HOST}
    DB_PORT: ${DB_PORT}
    DB_USER: ${DB_USER}
    DB_PASSWORD: ${DB_PASSWORD}
    DB_NAME: ${VOD_DB_NAME}
    MINIO_ENDPOINT: ${MINIO_ENDPOINT}
    MINIO_ACCESS_KEY: ${MINIO_ACCESS_KEY}
    MINIO_SECRET_KEY: ${MINIO_SECRET_KEY}
    MINIO_USE_SSL: ${MINIO_USE_SSL}
    MINIO_BUCKET: ${MINIO_BUCKET_RECORDINGS}
    PORT: ${RECORDING_SERVICE_PORT}
    STREAM_SERVICE_URL: ${STREAM_SERVICE_URL}
    VOD_SERVICE_URL: ${VOD_SERVICE_URL}
  ports:
    - "8083:8083"
  depends_on:
    - postgres
    - minio
    - stream-service
  volumes:
    - /tmp/recordings:/tmp/recordings  # Для временных файлов
  restart: unless-stopped
```

***

## 7. Примеры использования

### 7.1. Получение всех записей

```bash
curl http://localhost:8083/recordings
```

**Через API Gateway:**
```bash
curl http://localhost:8080/api/recordings
```

***

### 7.2. Получение конкретной записи

```bash
curl http://localhost:8083/recording/e410066c-8d50-4e0f-965f-452af1814ffa
```

***

### 7.3. Проверка статуса записи после стрима

```bash
# Получить список записей
RECORDINGS=$(curl -s http://localhost:8083/recordings)

# Найти последнюю запись
LAST_RECORDING=$(echo $RECORDINGS | jq -r '.recordings[0]')

# Проверить статус
STATUS=$(echo $LAST_RECORDING | jq -r '.status')
echo "Status: $STATUS"

# Если completed, получить путь к файлу
if [ "$STATUS" = "completed" ]; then
  FILE_PATH=$(echo $LAST_RECORDING | jq -r '.file_path')
  echo "File available at: http://localhost:9000/$FILE_PATH"
fi
```

***

### 7.4. Импорт записи в VOD

```bash
# Получить ID записи
RECORDING_ID="e410066c-8d50-4e0f-965f-452af1814ffa"

# Импортировать через VOD Service
curl -X POST http://localhost:8084/import-recording \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"recording_id\": \"$RECORDING_ID\",
    \"title\": \"My Stream Recording\",
    \"description\": \"Great stream!\",
    \"category\": \"gaming\",
    \"visibility\": \"public\"
  }"
```

***

## 8. Мониторинг и логи

### 8.1. Просмотр логов

```bash
docker-compose logs -f recording-service
```

**Ключевые события в логах:**

```
🚀 Starting Recording Service...
✅ Connected to vod_db successfully
✅ Connected to MinIO bucket: recordings
✅ Stream Monitor started
📹 Webhook received: stream.started for stream abc123
📥 Downloading segment: segment_000.ts
📥 Downloading segment: segment_001.ts
🎬 Stream ended, starting processing...
⚙️ Concatenating 120 segments...
✅ Recording completed: recordings/abc123.mp4
```

***

### 8.2. Проверка здоровья сервиса

```bash
curl http://localhost:8083/health
```

**Ожидаемый ответ:**
```json
{"service":"recording-service","status":"healthy"}
```

***

### 8.3. Проверка MinIO

```bash
# Список файлов в bucket recordings
docker exec streaming-minio mc ls minio/recordings/
```

***

## 9. Troubleshooting

### Проблема: Запись не создаётся

**Возможные причины:**
1. Webhook от Stream Service не доходит
2. Recording Service недоступен

**Решение:**
```bash
# Проверить логи Stream Service
docker-compose logs stream-service | grep webhook

# Проверить доступность Recording Service
curl http://localhost:8083/health

# Проверить network
docker exec streaming-stream ping recording-service
```

***

### Проблема: Сегменты не скачиваются

**Возможные причины:**
1. MinIO недоступен
2. Неверные credentials
3. Bucket `live-segments` пустой

**Решение:**
```bash
# Проверить MinIO
curl http://localhost:9000/minio/health/live

# Проверить bucket
docker exec streaming-minio mc ls minio/live-segments/

# Проверить логи
docker-compose logs recording-service | grep "Download"
```

***

### Проблема: FFmpeg конкатенация падает

**Возможные причины:**
1. Нехватка места на диске
2. Повреждённые сегменты
3. FFmpeg не установлен в контейнере

**Решение:**
```bash
# Проверить место на диске
docker exec streaming-recording df -h

# Проверить FFmpeg
docker exec streaming-recording ffmpeg -version

# Ручная конкатенация для отладки
docker exec -it streaming-recording /bin/sh
cd /tmp/recordings/stream_key/
ls -la *.ts
ffmpeg -f concat -safe 0 -i segments.txt -c copy output.mp4
```

***

## 10. Ограничения и рекомендации

### Текущие ограничения

1. **Одновременные записи:** Нет жёсткого лимита, но зависит от CPU/RAM
2. **Максимальная длительность:** Ограничена только местом на диске
3. **Формат вывода:** Только MP4 (можно расширить)
4. **Retention policy:** Записи хранятся бессрочно (нужна очистка вручную)

### Рекомендации

1. **Мониторинг места:** Настроить алерты при заполнении диска > 80%
2. **Очистка старых записей:** Периодически удалять записи старше N дней
3. **Backup:** Настроить резервное копирование MinIO
4. **Масштабирование:** При большой нагрузке использовать очередь (RabbitMQ/Kafka)

***

## Заключение

Recording Service обеспечивает полностью автоматическую запись live трансляций с минимальной задержкой и высокой надёжностью благодаря webhook + polling архитектуре.