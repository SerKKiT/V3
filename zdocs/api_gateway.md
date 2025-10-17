# API Gateway — Документация

## Назначение

API Gateway — это центральная точка входа для всех клиентских запросов к микросервисам платформы. Он выступает в роли обратного прокси (reverse proxy) между клиентскими приложениями и внутренними сервисами.[1][2][3]

### Основные функции

1. **Единая точка входа** — клиенты взаимодействуют только с одним URL (`http://localhost:8080`), вместо того чтобы знать адреса всех микросервисов
2. **Маршрутизация запросов** — автоматическое перенаправление запросов к соответствующему микросервису на основе URL
3. **Централизованная аутентификация** — JWT токены проверяются один раз в Gateway, а не в каждом сервисе отдельно
4. **Защита от DDoS** — rate limiting предотвращает перегрузку системы
5. **CORS управление** — единая настройка для всех frontend приложений
6. **Логирование и мониторинг** — централизованный сбор метрик всех API запросов[3][4]

### Преимущества использования

- **Упрощение клиентской разработки** — frontend не нужно знать внутреннюю структуру микросервисов[2]
- **Безопасность** — внутренние сервисы скрыты от прямого доступа
- **Масштабируемость** — легко добавлять/удалять сервисы без изменения клиентского кода
- **Эволюция архитектуры** — возможность рефакторинга backend без влияния на клиентов[6]

---

## Архитектура

```
┌─────────────────┐
│  Client App     │ (Web, Mobile, Desktop)
│  (Frontend)     │
└────────┬────────┘
         │
         │ HTTP/HTTPS
         ▼
┌─────────────────────────────────────┐
│       API Gateway :8080             │
│  ┌────────────────────────────┐    │
│  │  Middleware Stack:         │    │
│  │  1. CORS                   │    │
│  │  2. Request Logging        │    │
│  │  3. Rate Limiting          │    │
│  │  4. JWT Validation         │    │
│  └────────────────────────────┘    │
│                                     │
│  ┌────────────────────────────┐    │
│  │  Routing & Proxying        │    │
│  └────────────────────────────┘    │
└──────────┬──┬──┬──┬────────────────┘
           │  │  │  │
    ┌──────┘  │  │  └──────┐
    ▼         ▼  ▼         ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│ Auth   │ │ Stream │ │Record  │ │  VOD   │
│ :8081  │ │ :8082  │ │ :8083  │ │ :8084  │
└────────┘ └────────┘ └────────┘ └────────┘
```

***

## API Endpoints

### Базовый URL

```
http://localhost:8080
```

Для production:
```
https://api.yourdomain.com
```

***

## 1. Gateway Endpoints

### Health Check

Проверка работоспособности Gateway.

**Endpoint:** `GET /health`

**Авторизация:** Не требуется

**Пример запроса:**
```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "status": "healthy",
  "service": "api-gateway"
}
```

***

## 2. Authentication Service (`/api/auth/*`)

Все запросы проксируются в Auth Service (порт 8081).

### 2.1. Регистрация пользователя

**Endpoint:** `POST /api/auth/register`

**Авторизация:** Не требуется

**Request Body:**
```json
{
  "username": "string",    // min: 3, max: 50
  "email": "string",       // valid email
  "password": "string"     // min: 6 characters
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Ответ (успех):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2025-10-06T20:00:00Z",
  "user": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

***

### 2.2. Вход (логин)

**Endpoint:** `POST /api/auth/login`

**Авторизация:** Не требуется

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

**Ответ (успех):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2025-10-06T20:00:00Z",
  "user": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

***

## 3. Stream Service (`/api/streams/*`)

Управление live трансляциями.

### 3.1. Получить список live стримов (публичный)

**Endpoint:** `GET /api/streams/live`

**Авторизация:** Опционально (JWT)

**Query Parameters:**
- `limit` (опционально) — количество результатов

**Пример запроса:**
```bash
curl http://localhost:8080/api/streams/live
```

**Ответ:**
```json
{
  "streams": [
    {
      "id": "uuid",
      "stream_key": "abc123...",
      "title": "Gaming Stream",
      "status": "live",
      "viewer_count": 150,
      "hls_url": "http://localhost/hls/abc123/playlist.m3u8"
    }
  ]
}
```

***

### 3.2. Создать новый стрим (защищённый)

**Endpoint:** `POST /api/streams`

**Авторизация:** Требуется JWT

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "string",
  "description": "string"
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/streams \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Gaming Stream",
    "description": "Playing Minecraft"
  }'
```

**Ответ:**
```json
{
  "stream": {
    "id": "uuid",
    "stream_key": "abc123def456...",
    "title": "My Gaming Stream",
    "status": "offline"
  },
  "stream_url": "srt://localhost:6000?streamid=abc123def456",
  "hls_url": "http://localhost/hls/abc123def456/playlist.m3u8"
}
```

***

### 3.3. Получить информацию о стриме (защищённый)

**Endpoint:** `GET /api/streams/:id`

**Авторизация:** Требуется JWT

**Path Parameters:**
- `id` — UUID стрима

**Пример запроса:**
```bash
curl http://localhost:8080/api/streams/uuid-here \
  -H "Authorization: Bearer eyJhbGc..."
```

***

### 3.4. Обновить стрим (защищённый)

**Endpoint:** `PUT /api/streams/:id`

**Авторизация:** Требуется JWT

**Request Body:**
```json
{
  "title": "string",
  "description": "string"
}
```

***

### 3.5. Удалить стрим (защищённый)

**Endpoint:** `DELETE /api/streams/:id`

**Авторизация:** Требуется JWT

***

## 4. Recording Service (`/api/recordings/*`)

Управление записями стримов.

### 4.1. Получить список всех записей (публичный)

**Endpoint:** `GET /api/recordings`

**Авторизация:** Не требуется

**Пример запроса:**
```bash
curl http://localhost:8080/api/recordings
```

**Ответ:**
```json
{
  "recordings": [
    {
      "id": "uuid",
      "stream_id": "uuid",
      "file_path": "recordings/stream_key.mp4",
      "duration": 3600,
      "file_size": 524288000,
      "status": "completed",
      "started_at": "2025-10-05T18:00:00Z",
      "completed_at": "2025-10-05T19:00:00Z"
    }
  ]
}
```

***

### 4.2. Получить запись по ID (публичный)

**Endpoint:** `GET /api/recordings/:id`

**Авторизация:** Не требуется

**Path Parameters:**
- `id` — UUID записи

**Пример запроса:**
```bash
curl http://localhost:8080/api/recordings/uuid-here
```

**Ответ:**
```json
{
  "recording": {
    "id": "uuid",
    "stream_id": "uuid",
    "file_path": "recordings/stream_key.mp4",
    "duration": 3600,
    "file_size": 524288000,
    "status": "completed"
  }
}
```

***

## 5. VOD Service (`/api/videos/*`)

Управление библиотекой видео.

### 5.1. Получить список публичных видео (публичный)

**Endpoint:** `GET /api/videos`

**Авторизация:** Не требуется

**Query Parameters:**
- `limit` (опционально, default: 20, max: 100) — количество результатов
- `offset` (опционально, default: 0) — смещение для пагинации
- `search` (опционально) — поиск по названию и описанию
- `category` (опционально) — фильтр по категории
- `order` (опционально, default: "created_at DESC") — сортировка

**Пример запроса:**
```bash
curl "http://localhost:8080/api/videos?limit=10&search=gaming&category=esports"
```

**Ответ:**
```json
{
  "videos": [
    {
      "id": "uuid",
      "title": "Pro Gaming Highlights",
      "description": "Best moments",
      "category": "esports",
      "tags": ["gaming", "highlights"],
      "status": "ready",
      "visibility": "public",
      "duration": 1200,
      "view_count": 5000,
      "like_count": 250,
      "created_at": "2025-10-05T18:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "limit": 10
}
```

***

### 5.2. Получить видео по ID (публичный)

**Endpoint:** `GET /api/videos/:id`

**Авторизация:** Не требуется (для публичных видео)

**Path Parameters:**
- `id` — UUID видео

**Пример запроса:**
```bash
curl http://localhost:8080/api/videos/uuid-here
```

**Ответ:**
```json
{
  "video": {
    "id": "uuid",
    "title": "My Video",
    "status": "ready",
    "visibility": "public",
    "duration": 1200,
    "view_count": 100
  },
  "hls_url": "http://minio:9000/vod-videos/...",
  "thumbnail_url": "http://minio:9000/vod-videos/..."
}
```

***

### 5.3. Увеличить счётчик просмотров (публичный)

**Endpoint:** `POST /api/videos/:id/view`

**Авторизация:** Не требуется

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/videos/uuid-here/view
```

***

### 5.4. Импортировать запись в VOD (защищённый)

**Endpoint:** `POST /api/videos/import-recording`

**Авторизация:** Требуется JWT

**Request Body:**
```json
{
  "recording_id": "uuid",
  "title": "string",
  "description": "string",
  "category": "string",
  "tags": ["string"],
  "visibility": "public|private|unlisted"
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/videos/import-recording \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "recording_id": "uuid",
    "title": "Stream Highlights",
    "category": "gaming",
    "visibility": "public"
  }'
```

**Ответ:**
```json
{
  "message": "Recording imported successfully",
  "video_id": "uuid"
}
```

***

### 5.5. Получить видео пользователя (защищённый)

**Endpoint:** `GET /api/videos/user`

**Авторизация:** Требуется JWT

**Пример запроса:**
```bash
curl http://localhost:8080/api/videos/user \
  -H "Authorization: Bearer eyJhbGc..."
```

***

### 5.6. Обновить метаданные видео (защищённый)

**Endpoint:** `PUT /api/videos/:id`

**Авторизация:** Требуется JWT (владелец видео)

**Request Body:**
```json
{
  "title": "string",
  "description": "string",
  "category": "string",
  "tags": ["string"],
  "visibility": "public|private|unlisted"
}
```

***

### 5.7. Удалить видео (защищённый)

**Endpoint:** `DELETE /api/videos/:id`

**Авторизация:** Требуется JWT (владелец видео)

***

### 5.8. Лайкнуть видео (защищённый)

**Endpoint:** `POST /api/videos/:id/like`

**Авторизация:** Требуется JWT

***

## Middleware и безопасность

### CORS

**Настройки:**
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

### Rate Limiting

**Лимиты:**
- **Requests per second:** 100
- **Burst:** 200

При превышении лимита возвращается:
```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

**HTTP Status:** 429 Too Many Requests

### JWT Authentication

**Формат токена:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Содержимое токена:**
```json
{
  "user_id": "uuid",
  "username": "string",
  "exp": 1728241200,
  "iat": 1728154800
}
```

**Ошибки аутентификации:**

| HTTP Code | Error | Описание |
|-----------|-------|----------|
| 401 | `Authorization header required` | Отсутствует заголовок Authorization |
| 401 | `Invalid authorization format` | Неверный формат (не "Bearer token") |
| 401 | `Invalid or expired token` | Токен недействителен или истёк |

***

## Коды ответов

| HTTP Code | Описание |
|-----------|----------|
| 200 | OK — запрос выполнен успешно |
| 201 | Created — ресурс создан |
| 204 | No Content — успешно, нет содержимого |
| 400 | Bad Request — некорректный запрос |
| 401 | Unauthorized — требуется аутентификация |
| 403 | Forbidden — доступ запрещён |
| 404 | Not Found — ресурс не найден |
| 429 | Too Many Requests — превышен rate limit |
| 500 | Internal Server Error — внутренняя ошибка сервера |
| 502 | Bad Gateway — сервис недоступен |

***

## Примеры использования

### Полный цикл работы с API

```bash
# 1. Регистрация
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user","email":"user@test.com","password":"password123"}' \
  | jq -r '.token')

# 2. Создание стрима
STREAM=$(curl -s -X POST http://localhost:8080/api/streams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"My Stream","description":"Test"}' \
  | jq -r '.stream.stream_key')

# 3. Получение live стримов
curl http://localhost:8080/api/streams/live

# 4. Получение видео
curl http://localhost:8080/api/videos?limit=10

# 5. Получение записей
curl http://localhost:8080/api/recordings
```

***

## Заключение

API Gateway предоставляет унифицированный интерфейс для взаимодействия со всеми микросервисами платформы, обеспечивая безопасность, масштабируемость и простоту интеграции.[2][3][6]