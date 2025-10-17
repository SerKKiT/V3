# VOD Service — Документация

## Назначение

VOD Service (Video on Demand) — это микросервис для управления библиотекой видеоконтента платформы. Он предоставляет функционал для импорта записей стримов, управления метаданными видео, поиска, фильтрации и отслеживания аналитики просмотров.

### Основные функции

1. **Импорт записей** — преобразование записей стримов в VOD библиотеку
2. **Управление метаданными** — редактирование названий, описаний, категорий, тегов
3. **Контроль видимости** — публичные, приватные и unlisted видео
4. **Поиск и фильтрация** — full-text search, фильтры по категориям и тегам
5. **Аналитика** — счётчики просмотров и лайков
6. **Пользовательские коллекции** — управление собственными видео

### Интеграция с другими сервисами

```
┌──────────────────────┐
│  Recording Service   │
│       :8083          │
└──────────┬───────────┘
           │
           │ GET /recording/:id
           │ (получение метаданных)
           ▼
┌──────────────────────┐
│    VOD Service       │ ──────> PostgreSQL (vod_db)
│       :8084          │         (таблица: videos)
└──────────┬───────────┘
           │
           │ MinIO bucket: recordings
           │ (ссылки на HLS файлы)
           ▼
┌──────────────────────┐
│       MinIO          │
│  recordings/         │
│  stream_key.mp4      │
└──────────────────────┘
```

***

## Архитектура

### Компоненты

1. **Video Handler** — HTTP API обработчики
2. **Video Repository** — работа с PostgreSQL
3. **MinIO Storage** — интеграция с объектным хранилищем
4. **Search Engine** — full-text поиск по видео
5. **Analytics Tracker** — счётчики просмотров и лайков

### Модель данных

```
Video {
  - Идентификация: id, user_id, recording_id, stream_id
  - Метаданные: title, description, category, tags
  - Контроль: source, status, visibility
  - Хранилище: hls_path, thumbnail_path, duration, file_size, resolution
  - Аналитика: view_count, like_count
  - Временные метки: created_at, updated_at, published_at
}
```

***

## API Endpoints

### Базовый URL

```
http://localhost:8084
```

Через API Gateway:
```
http://localhost:8080/api/videos
```

***

## 1. Service Endpoints

### 1.1. Health Check

Проверка работоспособности сервиса.

**Endpoint:** `GET /health`

**Авторизация:** Не требуется

**Пример запроса:**
```bash
curl http://localhost:8084/health
```

**Ответ:**
```json
{
  "status": "healthy",
  "service": "vod-service"
}
```

***

## 2. Public Endpoints

### 2.1. Получить список публичных видео

Возвращает список видео с возможностью поиска, фильтрации и пагинации.

**Endpoint:** `GET /videos`

**Авторизация:** Не требуется

**Query Parameters:**

| Параметр | Тип | Обязательный | По умолчанию | Описание |
|----------|-----|--------------|--------------|----------|
| `limit` | integer | Нет | 20 | Количество результатов (max: 100) |
| `offset` | integer | Нет | 0 | Смещение для пагинации |
| `search` | string | Нет | - | Поиск по названию и описанию |
| `category` | string | Нет | - | Фильтр по категории |
| `tags` | array | Нет | - | Фильтр по тегам (можно несколько) |
| `order` | string | Нет | created_at DESC | Сортировка |

**Доступные сортировки:**
- `created_at DESC` — новые первыми
- `created_at ASC` — старые первыми
- `view_count DESC` — популярные первыми
- `published_at DESC` — недавно опубликованные

**Пример запроса:**
```bash
curl "http://localhost:8084/videos?limit=10&search=gaming&category=esports"
```

**Пример с тегами:**
```bash
curl "http://localhost:8084/videos?tags=minecraft&tags=survival&limit=5"
```

**Пример через API Gateway:**
```bash
curl "http://localhost:8080/api/videos?limit=10&offset=0&order=view_count%20DESC"
```

**Ответ:**
```json
{
  "videos": [
    {
      "id": "87d85ca8-6f0b-4037-9ea1-37831bcdbc37",
      "user_id": "a781b655-d4e4-46a0-89da-c4f655bb24f0",
      "recording_id": "e410066c-8d50-4e0f-965f-452af1814ffa",
      "stream_id": "bb52ba50-edc5-4056-a62c-6ef211d72d25",
      "title": "E2E Test Video",
      "description": "Imported from test stream",
      "category": "test",
      "tags": ["e2e", "automated"],
      "source": "recording",
      "status": "ready",
      "visibility": "public",
      "hls_path": "recordings/d1eb2fe800d01c797c84f0b622e474d9.mp4",
      "thumbnail_path": "",
      "duration": 26,
      "file_size": 5672829,
      "resolution": "",
      "view_count": 0,
      "like_count": 0,
      "created_at": "2025-10-05T18:03:19Z",
      "updated_at": "2025-10-05T18:03:19Z",
      "published_at": "2025-10-05T18:03:19Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

**Поля видео:**

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Уникальный идентификатор видео |
| `user_id` | UUID | ID пользователя-владельца |
| `recording_id` | UUID \| null | ID записи (если импортировано из recording) |
| `stream_id` | UUID \| null | ID стрима (если связано) |
| `title` | string | Название видео |
| `description` | string | Описание |
| `category` | string | Категория (gaming, education, music, etc.) |
| `tags` | array | Теги для поиска |
| `source` | enum | Источник: `recording`, `upload`, `import` |
| `status` | enum | Статус: `pending`, `ready`, `failed`, `archived` |
| `visibility` | enum | Видимость: `public`, `private`, `unlisted` |
| `hls_path` | string | Путь к HLS в MinIO |
| `thumbnail_path` | string | Путь к превью |
| `duration` | integer | Длительность в секундах |
| `file_size` | integer | Размер файла в байтах |
| `resolution` | string | Разрешение (1920x1080, 1280x720, etc.) |
| `view_count` | integer | Количество просмотров |
| `like_count` | integer | Количество лайков |
| `created_at` | timestamp | Дата создания |
| `updated_at` | timestamp | Дата последнего обновления |
| `published_at` | timestamp \| null | Дата публикации |

***

### 2.2. Получить видео по ID

Возвращает полную информацию о видео и URL для воспроизведения.

**Endpoint:** `GET /video/:id`

**Авторизация:** Не требуется (для публичных видео)

**Path Parameters:**
- `id` (UUID) — идентификатор видео

**Пример запроса:**
```bash
curl http://localhost:8084/video/87d85ca8-6f0b-4037-9ea1-37831bcdbc37
```

**Пример через API Gateway:**
```bash
curl http://localhost:8080/api/videos/87d85ca8-6f0b-4037-9ea1-37831bcdbc37
```

**Ответ (публичное видео):**
```json
{
  "video": {
    "id": "87d85ca8-6f0b-4037-9ea1-37831bcdbc37",
    "title": "E2E Test Video",
    "description": "Imported from test stream",
    "category": "test",
    "tags": ["e2e", "automated"],
    "status": "ready",
    "visibility": "public",
    "duration": 26,
    "file_size": 5672829,
    "view_count": 0,
    "like_count": 0,
    "created_at": "2025-10-05T18:03:19Z"
  },
  "hls_url": "http://minio:9000/recordings/d1eb2fe800d01c797c84f0b622e474d9.mp4",
  "thumbnail_url": ""
}
```

**Ответ (приватное видео без авторизации):**
```json
{
  "error": "Video is private"
}
```

**HTTP Status:** 403 Forbidden

**Ответ (видео не найдено):**
```json
{
  "error": "Video not found"
}
```

**HTTP Status:** 404 Not Found

***

### 2.3. Увеличить счётчик просмотров

Увеличивает счётчик просмотров видео на 1.

**Endpoint:** `POST /video/:id/view`

**Авторизация:** Не требуется

**Path Parameters:**
- `id` (UUID) — идентификатор видео

**Пример запроса:**
```bash
curl -X POST http://localhost:8084/video/87d85ca8-6f0b-4037-9ea1-37831bcdbc37/view
```

**Пример через API Gateway:**
```bash
curl -X POST http://localhost:8080/api/videos/87d85ca8-6f0b-4037-9ea1-37831bcdbc37/view
```

**Ответ:**
```json
{
  "message": "View count incremented"
}
```

**Использование:**
Этот endpoint должен вызываться при начале воспроизведения видео в плеере.

***

## 3. Protected Endpoints (требуют JWT)

### 3.1. Импортировать запись в VOD

Создаёт видео в VOD библиотеке из существующей записи.

**Endpoint:** `POST /import-recording`

**Авторизация:** Требуется JWT

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "recording_id": "string (UUID)",
  "title": "string (required)",
  "description": "string",
  "category": "string",
  "tags": ["string"],
  "visibility": "public" | "private" | "unlisted"
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8084/import-recording \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "recording_id": "e410066c-8d50-4e0f-965f-452af1814ffa",
    "title": "My Epic Stream Highlights",
    "description": "Best moments from today stream",
    "category": "gaming",
    "tags": ["minecraft", "pvp", "highlights"],
    "visibility": "public"
  }'
```

**Пример через API Gateway:**
```bash
curl -X POST http://localhost:8080/api/videos/import-recording \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "recording_id": "e410066c-8d50-4e0f-965f-452af1814ffa",
    "title": "My Stream Recording",
    "visibility": "public"
  }'
```

**Ответ (успех):**
```json
{
  "message": "Recording imported successfully",
  "video_id": "87d85ca8-6f0b-4037-9ea1-37831bcdbc37"
}
```

**HTTP Status:** 201 Created

**Ответ (запись не найдена):**
```json
{
  "error": "Recording not found"
}
```

**HTTP Status:** 404 Not Found

**Процесс импорта:**
1. VOD Service запрашивает метаданные записи у Recording Service (`GET /recording/:id`)
2. Создаёт запись в таблице `videos` со статусом `ready`
3. Копирует `file_path` из recording в `hls_path` видео
4. Устанавливает `source = "recording"`
5. Возвращает `video_id` для дальнейшего использования

***

### 3.2. Получить видео пользователя

Возвращает все видео текущего пользователя (включая приватные).

**Endpoint:** `GET /user/videos`

**Авторизация:** Требуется JWT

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `limit` (integer, default: 20) — количество результатов
- `offset` (integer, default: 0) — смещение

**Пример запроса:**
```bash
curl http://localhost:8084/user/videos \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Пример через API Gateway:**
```bash
curl http://localhost:8080/api/videos/user \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Ответ:**
```json
{
  "videos": [
    {
      "id": "87d85ca8-6f0b-4037-9ea1-37831bcdbc37",
      "title": "My Private Video",
      "visibility": "private",
      "status": "ready",
      "view_count": 5,
      "created_at": "2025-10-05T18:00:00Z"
    },
    {
      "id": "abc-123-def",
      "title": "Public Stream Recording",
      "visibility": "public",
      "status": "ready",
      "view_count": 1500,
      "created_at": "2025-10-04T15:00:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "limit": 20
}
```

***

### 3.3. Обновить метаданные видео

Обновляет информацию о видео. Доступно только владельцу.

**Endpoint:** `PUT /video/:id`

**Авторизация:** Требуется JWT (владелец видео)

**Path Parameters:**
- `id` (UUID) — идентификатор видео

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "string",
  "description": "string",
  "category": "string",
  "tags": ["string"],
  "visibility": "public" | "private" | "unlisted"
}
```

Все поля опциональны. Обновляются только переданные поля.

**Пример запроса:**
```bash
curl -X PUT http://localhost:8084/video/87d85ca8-6f0b-4037-9ea1-37831bcdbc37 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "description": "New description",
    "tags": ["new", "tags"],
    "visibility": "public"
  }'
```

**Пример через API Gateway:**
```bash
curl -X PUT http://localhost:8080/api/videos/87d85ca8-6f0b-4037-9ea1-37831bcdbc37 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Title"
  }'
```

**Ответ (успех):**
```json
{
  "message": "Video updated successfully"
}
```

**Ответ (не владелец):**
```json
{
  "error": "Not authorized"
}
```

**HTTP Status:** 403 Forbidden

---

### 3.4. Удалить видео

Удаляет видео из библиотеки. Доступно только владельцу.

**Endpoint:** `DELETE /video/:id`

**Авторизация:** Требуется JWT (владелец видео)

**Path Parameters:**
- `id` (UUID) — идентификатор видео

**Headers:**
```
Authorization: Bearer <token>
```

**Пример запроса:**
```bash
curl -X DELETE http://localhost:8084/video/87d85ca8-6f0b-4037-9ea1-37831bcdbc37 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Пример через API Gateway:**
```bash
curl -X DELETE http://localhost:8080/api/videos/87d85ca8-6f0b-4037-9ea1-37831bcdbc37 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Ответ (успех):**
```json
{
  "message": "Video deleted successfully"
}
```

**Примечание:** Удаление только из таблицы `videos`. Файлы в MinIO остаются (можно расширить функционал для полного удаления).

***

### 3.5. Лайкнуть видео

Увеличивает счётчик лайков видео на 1.

**Endpoint:** `POST /video/:id/like`

**Авторизация:** Требуется JWT

**Path Parameters:**
- `id` (UUID) — идентификатор видео

**Headers:**
```
Authorization: Bearer <token>
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8084/video/87d85ca8-6f0b-4037-9ea1-37831bcdbc37/like \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Пример через API Gateway:**
```bash
curl -X POST http://localhost:8080/api/videos/87d85ca8-6f0b-4037-9ea1-37831bcdbc37/like \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Ответ:**
```json
{
  "message": "Video liked"
}
```

**Примечание:** Текущая реализация не проверяет, лайкал ли пользователь уже это видео. Можно расширить с таблицей `user_likes`.

***

## 4. База данных

### 4.1. Таблица `videos`

**Схема:**
```sql
CREATE TABLE videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    recording_id UUID REFERENCES recordings(id) ON DELETE SET NULL,
    stream_id UUID,
    
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    tags TEXT[] DEFAULT '{}',
    
    source video_source DEFAULT 'recording',
    status video_status DEFAULT 'pending',
    visibility video_visibility DEFAULT 'private',
    
    hls_path TEXT NOT NULL,
    thumbnail_path TEXT,
    duration INT DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    resolution VARCHAR(20),
    
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    
    CONSTRAINT fk_recording FOREIGN KEY (recording_id) 
        REFERENCES recordings(id) ON DELETE SET NULL
);
```

***

### 4.2. ENUM типы

```sql
-- Статус видео
CREATE TYPE video_status AS ENUM (
    'pending',   -- В обработке
    'ready',     -- Готово к просмотру
    'failed',    -- Ошибка обработки
    'archived'   -- Архивировано
);

-- Источник видео
CREATE TYPE video_source AS ENUM (
    'recording', -- Импортировано из записи стрима
    'upload',    -- Прямая загрузка (будущая функция)
    'import'     -- Импорт из внешних источников
);

-- Видимость видео
CREATE TYPE video_visibility AS ENUM (
    'public',    -- Доступно всем
    'private',   -- Доступно только владельцу
    'unlisted'   -- Доступно по ссылке
);
```

***

### 4.3. Индексы

```sql
-- Основные индексы
CREATE INDEX idx_videos_user_id ON videos(user_id);
CREATE INDEX idx_videos_recording_id ON videos(recording_id);
CREATE INDEX idx_videos_stream_id ON videos(stream_id);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_videos_visibility ON videos(visibility);
CREATE INDEX idx_videos_category ON videos(category);

-- Индексы для сортировки
CREATE INDEX idx_videos_created_at ON videos(created_at DESC);
CREATE INDEX idx_videos_view_count ON videos(view_count DESC);
CREATE INDEX idx_videos_published_at ON videos(published_at DESC);

-- GIN индекс для массива тегов
CREATE INDEX idx_videos_tags ON videos USING GIN(tags);

-- Full-text search индекс
CREATE INDEX idx_videos_search ON videos 
USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));
```

***

## 5. Поиск и фильтрация

### 5.1. Full-text search

**Запрос:**
```bash
curl "http://localhost:8084/videos?search=minecraft+survival"
```

**SQL:**
```sql
SELECT * FROM videos
WHERE to_tsvector('english', title || ' ' || COALESCE(description, '')) 
      @@ plainto_tsquery('english', 'minecraft survival')
  AND visibility = 'public'
  AND status = 'ready'
ORDER BY created_at DESC
LIMIT 20;
```

***

### 5.2. Фильтр по тегам

**Запрос:**
```bash
curl "http://localhost:8084/videos?tags=gaming&tags=fps"
```

**SQL:**
```sql
SELECT * FROM videos
WHERE tags && ARRAY['gaming', 'fps']
  AND visibility = 'public'
  AND status = 'ready'
ORDER BY created_at DESC;
```

***

### 5.3. Фильтр по категории

**Запрос:**
```bash
curl "http://localhost:8084/videos?category=gaming"
```

**SQL:**
```sql
SELECT * FROM videos
WHERE category = 'gaming'
  AND visibility = 'public'
  AND status = 'ready'
ORDER BY created_at DESC;
```

***

### 5.4. Комбинированные фильтры

**Запрос:**
```bash
curl "http://localhost:8084/videos?search=tutorial&category=education&tags=beginner&order=view_count+DESC&limit=10"
```

**SQL:**
```sql
SELECT * FROM videos
WHERE to_tsvector('english', title || ' ' || COALESCE(description, '')) 
      @@ plainto_tsquery('english', 'tutorial')
  AND category = 'education'
  AND tags && ARRAY['beginner']
  AND visibility = 'public'
  AND status = 'ready'
ORDER BY view_count DESC
LIMIT 10;
```

***

## 6. Конфигурация

### 6.1. Переменные окружения

```bash
# Service
PORT=8084

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
MINIO_BUCKET=vod-videos

# Services
RECORDING_SERVICE_URL=http://recording-service:8083
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

***

### 6.2. Docker Compose

```yaml
vod-service:
  build: ./services/vod-service
  container_name: streaming-vod
  environment:
    PORT: ${VOD_SERVICE_PORT}
    DB_HOST: ${DB_HOST}
    DB_PORT: ${DB_PORT}
    DB_USER: ${DB_USER}
    DB_PASSWORD: ${DB_PASSWORD}
    DB_NAME: ${VOD_DB_NAME}
    MINIO_ENDPOINT: ${MINIO_ENDPOINT}
    MINIO_ACCESS_KEY: ${MINIO_ACCESS_KEY}
    MINIO_SECRET_KEY: ${MINIO_SECRET_KEY}
    MINIO_USE_SSL: ${MINIO_USE_SSL}
    MINIO_BUCKET: ${MINIO_BUCKET_VOD}
    RECORDING_SERVICE_URL: ${RECORDING_SERVICE_URL}
    JWT_SECRET: ${JWT_SECRET}
  ports:
    - "8084:8084"
  depends_on:
    - postgres
    - minio
  restart: unless-stopped
```

***

## 7. Примеры использования

### 7.1. Полный цикл: Запись → VOD

```bash
# 1. Получить список записей
RECORDINGS=$(curl -s http://localhost:8083/recordings)
RECORDING_ID=$(echo $RECORDINGS | jq -r '.recordings[0].id')

echo "Recording ID: $RECORDING_ID"

# 2. Получить JWT токен
TOKEN=$(curl -s -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"password123"}' \
  | jq -r '.token')

echo "Token: ${TOKEN:0:30}..."

# 3. Импортировать запись в VOD
VIDEO_ID=$(curl -s -X POST http://localhost:8084/import-recording \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"recording_id\": \"$RECORDING_ID\",
    \"title\": \"My Stream Highlights\",
    \"description\": \"Best moments\",
    \"category\": \"gaming\",
    \"tags\": [\"minecraft\", \"pvp\"],
    \"visibility\": \"public\"
  }" | jq -r '.video_id')

echo "Video ID: $VIDEO_ID"

# 4. Получить информацию о видео
curl -s http://localhost:8084/video/$VIDEO_ID | jq .

# 5. Просмотр видео (увеличить счётчик)
curl -X POST http://localhost:8084/video/$VIDEO_ID/view

# 6. Лайкнуть видео
curl -X POST http://localhost:8084/video/$VIDEO_ID/like \
  -H "Authorization: Bearer $TOKEN"
```

***

### 7.2. Поиск и фильтрация

```bash
# Поиск по ключевым словам
curl "http://localhost:8084/videos?search=minecraft&limit=5"

# Фильтр по категории
curl "http://localhost:8084/videos?category=gaming&limit=10"

# Популярные видео
curl "http://localhost:8084/videos?order=view_count+DESC&limit=10"

# Новые видео с тегами
curl "http://localhost:8084/videos?tags=tutorial&tags=beginner&order=created_at+DESC"

# Комбинированный поиск
curl "http://localhost:8084/videos?search=guide&category=education&limit=5"
```

***

### 7.3. Управление видео

```bash
# Получить свои видео
curl http://localhost:8084/user/videos \
  -H "Authorization: Bearer $TOKEN"

# Обновить метаданные
curl -X PUT http://localhost:8084/video/$VIDEO_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "description": "New description",
    "visibility": "public"
  }'

# Удалить видео
curl -X DELETE http://localhost:8084/video/$VIDEO_ID \
  -H "Authorization: Bearer $TOKEN"
```

***

## 8. Аналитика

### 8.1. Топ популярных видео

```sql
SELECT id, title, view_count, like_count, created_at
FROM videos
WHERE visibility = 'public'
  AND status = 'ready'
ORDER BY view_count DESC
LIMIT 10;
```

***

### 8.2. Топ видео по лайкам

```sql
SELECT id, title, like_count, view_count,
       ROUND(like_count::float / NULLIF(view_count, 0) * 100, 2) as like_rate
FROM videos
WHERE visibility = 'public'
  AND status = 'ready'
  AND view_count > 10
ORDER BY like_count DESC
LIMIT 10;
```

***

### 8.3. Статистика по категориям

```sql
SELECT category,
       COUNT(*) as video_count,
       SUM(view_count) as total_views,
       AVG(duration) as avg_duration
FROM videos
WHERE visibility = 'public'
  AND status = 'ready'
GROUP BY category
ORDER BY total_views DESC;
```

***

## 9. Мониторинг

### 9.1. Просмотр логов

```bash
docker-compose logs -f vod-service
```

**Ключевые события:**
```
🚀 Starting VOD Service...
✅ Connected to vod_db successfully
✅ Connected to MinIO bucket: vod-videos
✅ VOD Service running on port 8084
📹 Recording imported as video: recording_id -> video_id
👁️ View count incremented for video: video_id
❤️ Video liked: video_id
```

***

### 9.2. Проверка здоровья

```bash
curl http://localhost:8084/health
```

***

### 9.3. Статистика БД

```bash
docker exec streaming-postgres psql -U streaming_user -d vod_db -c "
SELECT 
  COUNT(*) as total_videos,
  COUNT(*) FILTER (WHERE visibility = 'public') as public_videos,
  COUNT(*) FILTER (WHERE status = 'ready') as ready_videos,
  SUM(view_count) as total_views,
  SUM(like_count) as total_likes
FROM videos;"
```

***

## 10. Ограничения и будущие улучшения

### Текущие ограничения

1. **Нет прямой загрузки** — можно импортировать только из recordings
2. **Нет транскодирования** — используется один bitrate из записи
3. **Нет thumbnails** — превью не генерируются автоматически
4. **Лайки без проверки** — можно лайкнуть несколько раз
5. **Нет плейлистов** — невозможно группировать видео

---

### Планируемые улучшения

#### 1. Прямая загрузка видео

```go
POST /video/upload
Content-Type: multipart/form-data

{
  "title": "My Video",
  "video": [binary file],
  "thumbnail": [binary image]
}
```

#### 2. Генерация thumbnails

```go
// Автоматическое создание превью при импорте
func GenerateThumbnail(videoPath string) (string, error) {
    // FFmpeg: извлечь кадр на 10% длительности
    // Загрузить в MinIO
    // Вернуть URL
}
```

#### 3. Adaptive Bitrate Streaming (ABR)

```
vod-videos/
  videos/{video_id}/
    hls/
      master.m3u8
      720p/playlist.m3u8
      480p/playlist.m3u8
      360p/playlist.m3u8
```

#### 4. Плейлисты

```sql
CREATE TABLE playlists (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE playlist_videos (
    playlist_id UUID REFERENCES playlists(id),
    video_id UUID REFERENCES videos(id),
    position INT NOT NULL,
    PRIMARY KEY (playlist_id, video_id)
);
```

#### 5. Комментарии

```sql
CREATE TABLE video_comments (
    id UUID PRIMARY KEY,
    video_id UUID REFERENCES videos(id),
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

***

## Заключение

VOD Service предоставляет полнофункциональную систему управления видеобиблиотекой с поиском, аналитикой и контролем доступа, готовую к интеграции с frontend приложениями.