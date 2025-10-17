# 🎬 Streaming Platform - Текущая Архитектура и Статус

## 📐 Архитектура системы

### **Microservices Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                       FRONTEND (React)                       │
│          http://localhost (Vite + VideoJS + HLS)            │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    NGINX (Reverse Proxy)                     │
│        ├─ /api/* → API Gateway                              │
│        ├─ /live-streams/* → MinIO (HLS segments)           │
│        └─ /storage/* → MinIO (VOD videos)                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
┌───────▼────────┐           ┌────────▼────────┐
│  API Gateway   │           │      MinIO      │
│   (Port 8080)  │           │  Object Storage │
└───────┬────────┘           └─────────────────┘
        │                    • Buckets:
        │                      - live-segments/ (HLS)
   ┌────┴─────┬──────────────  - recordings/
   │          │                - vod-videos/
┌──▼──┐  ┌───▼────┐  ┌─────▼─────┐  ┌────────▼─────┐
│Auth │  │Stream  │  │Recording  │  │VOD           │
│     │  │Service │  │Service    │  │Service       │
└──┬──┘  └───┬────┘  └─────┬─────┘  └────────┬─────┘
   │         │              │                 │
   │         │              │                 │
┌──▼─────────▼──────────────▼─────────────────▼─────┐
│            PostgreSQL (3 Databases)                │
│  ├─ auth_db (users, JWT)                          │
│  ├─ streams_db (streams, FDW → users)             │
│  ├─ recordings_db (recordings)                    │
│  └─ vod_db (videos, FDW → users)                  │
└────────────────────────────────────────────────────┘

External:
  ┌────────────────┐
  │  SRT Server    │  ← Live streaming input
  │  (Port 8890)   │     (OBS Studio)
  └────────────────┘
        │
        ▼
  ┌────────────────┐
  │  Transcoder    │  → FFmpeg → HLS → MinIO
  └────────────────┘
```

***

## 🏗️ Компоненты системы

### **1. Frontend (React + Vite)**
- ✅ **Live Streaming:** VideoJS player с HLS.js, adaptive bitrate (360p-1080p)
- ✅ **VOD:** Просмотр записанных видео с thumbnail preview
- ✅ **Authentication:** JWT-based auth (localStorage + HttpOnly cookies)
- ✅ **UI:** Tailwind CSS, responsive design
- ✅ **Features:**
  - Live stream player с custom controls
  - DVR (перемотка в live стриме)
  - Quality selector (360p, 480p, 720p, 1080p)
  - Username display (через FDW)
  - Keyboard shortcuts (Space, J/L, Arrow keys)

### **2. API Gateway (Golang)**
- ✅ **Routing:** Centralized routing для всех микросервисов
- ✅ **CORS:** Настроенный CORS для frontend
- ✅ **Endpoints:**
  - `/api/auth/*` → Auth Service
  - `/api/streams/*` → Stream Service
  - `/api/videos/*` → VOD Service
  - `/api/recordings/*` → Recording Service

### **3. Auth Service**
- ✅ **JWT Authentication:** Access + Refresh tokens
- ✅ **Database:** `auth_db` (users, sessions)
- ✅ **Features:**
  - Register, Login, Logout
  - Token refresh
  - Password hashing (bcrypt)
  - User profile management

### **4. Stream Service**
- ✅ **Live Streaming:** Управление live стримами
- ✅ **Database:** `streams_db` с **FDW** к `auth_db.users`
- ✅ **Features:**
  - Create/Delete/Update streams
  - Stream key generation
  - Live status tracking
  - Multi-quality HLS (ABR)
  - Viewer count
  - **Username** в ответах через FDW JOIN
- ✅ **SRT Integration:** Приём потока с OBS через SRT

### **5. Recording Service**
- ✅ **Auto-Recording:** Автоматическая запись всех live стримов
- ✅ **Stream Monitoring:** Отслеживание активных стримов (polling)
- ✅ **FFmpeg Integration:** Захват HLS → MP4
- ✅ **Thumbnail Generation:** Автогенерация превью
- ✅ **MinIO Upload:** Загрузка записей в MinIO
- ✅ **Database:** `recordings_db`
- ✅ **VOD Import:** Автоматический импорт в VOD после завершения
- ✅ **Internal Auth:** `X-Internal-API-Key` для service-to-service

### **6. VOD Service**
- ✅ **Video Management:** CRUD для видео
- ✅ **Database:** `vod_db` с **FDW** к `auth_db.users`
- ✅ **Features:**
  - Import recordings from Recording Service
  - Video streaming (direct MP4 playback)
  - Thumbnail serving
  - View counter
  - Like system
  - Tags/categories
  - **Username** в ответах через FDW JOIN
- ✅ **Security:**
  - Optional Auth (public/private videos)
  - Internal API key для recording service
  - Cookie + JWT auth

### **7. Transcoder Service**
- ✅ **Live Transcoding:** SRT → HLS (multi-bitrate)
- ✅ **FFmpeg Pipeline:** 4 качества (360p, 480p, 720p, 1080p)
- ✅ **HLS Generation:**
  - `master.m3u8` (ABR manifest)
  - `{quality}/playlist.m3u8` (per-quality playlists)
  - `.ts` segments (2-4 seconds)
- ✅ **MinIO Integration:** Direct upload к MinIO
- ✅ **Webhook:** Уведомление Stream Service о статусах

### **8. MinIO (Object Storage)**
- ✅ **Buckets:**
  - `live-segments/` - HLS сегменты live стримов
  - `recordings/` - MP4 записи
  - `vod-videos/` - VOD контент
- ✅ **Public Access:** Настроенные policies для публичного доступа
- ✅ **NGINX Proxy:** `/live-streams/*` и `/storage/*`

### **9. PostgreSQL (3 Databases)**

#### **auth_db**
```sql
users (id, username, email, password_hash, created_at)
refresh_tokens (...)
```

#### **streams_db** 
```sql
streams (id, user_id, stream_key, title, status, started_at, available_qualities, ...)
-- FDW: auth_db.users_foreign
```

#### **recordings_db**
```sql
recordings (id, stream_id, file_path, started_at, ended_at, status, video_id, ...)
```

#### **vod_db**
```sql
videos (id, user_id, title, file_path, thumbnail_path, view_count, ...)
video_likes (...)
video_tags (...)
-- FDW: auth_db.users_foreign
```

***

## 🔒 Security Architecture

### **Authentication Layers**

1. **User Authentication (JWT)**
   - Frontend → Backend: `Authorization: Bearer <token>`
   - Cookie-based auth для video streaming
   - Refresh token rotation

2. **Service-to-Service Authentication**
   - `X-Internal-API-Key` header
   - Recording → VOD import
   - Only trusted services

3. **Foreign Data Wrapper (FDW)**
   - Cross-database queries (streams_db → auth_db)
   - Username display без дублирования данных
   - Read-only access

### **CORS Configuration**
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type, X-User-ID, X-Internal-API-Key
```

***

## 🎥 Live Streaming Flow

```
1. OBS Studio 
   ↓ (SRT protocol)
2. SRT Server (:8890)
   ↓
3. Transcoder Service
   ↓ (FFmpeg transcoding)
4. MinIO (live-segments/)
   ├─ master.m3u8
   ├─ 1080p/playlist.m3u8
   ├─ 720p/playlist.m3u8
   ├─ 480p/playlist.m3u8
   └─ 360p/playlist.m3u8
   ↓
5. NGINX (/live-streams/*)
   ↓
6. Frontend (VideoJS HLS player)

Параллельно:
3a. Recording Service
    ↓ (monitors streams)
4a. FFmpeg Recorder (HLS → MP4)
    ↓
5a. MinIO (recordings/)
    ↓
6a. VOD Import (automatic)
```

***

## ✅ Текущий статус функционала

### **Полностью работает:**
- ✅ Live streaming (SRT → HLS → Player)
- ✅ Multi-bitrate adaptive streaming (ABR)
- ✅ Automatic recording всех стримов
- ✅ Thumbnail generation
- ✅ VOD import (recording → video)
- ✅ Username display (через FDW)
- ✅ Authentication (JWT + cookies)
- ✅ Custom video player (keyboard shortcuts, DVR, quality selector)
- ✅ Stream management (create, delete, update)
- ✅ Video management (CRUD, views, likes)

### **Известные особенности:**
- ⚠️ Recording service polling (каждые 10 секунд) - может быть заменён на webhooks
- ⚠️ MinIO public buckets - нужно для демо, в prod добавить signed URLs
- ⚠️ CORS `*` - в prod ограничить домены

***

## 📊 Database Schema Summary

### **FDW Architecture:**
```
auth_db (master)
  └─ users
        ↑ (FDW)
        ├─ streams_db.users_foreign
        └─ vod_db.users_foreign
```

**Преимущества:**
- Единый источник истины (auth_db)
- Автоматические JOIN с username
- Нет дублирования данных
- Read-only доступ (безопасность)

---

## 🚀 Deployment Stack

```yaml
Services (Docker Compose):
  - nginx (reverse proxy)
  - api-gateway
  - auth-service
  - stream-service
  - recording-service
  - vod-service
  - transcoder
  - minio
  - postgres
  - frontend (Vite dev server / production nginx)
```

**Environment:**
- ✅ All services в Docker
- ✅ Isolated networks
- ✅ Shared MinIO buckets
- ✅ Centralized NGINX routing

***

## 🎯 Архитектурные принципы

1. **Microservices:** Каждый сервис - отдельная ответственность
2. **Event-driven:** Recording service → VOD import (асинхронно)
3. **Scalable:** Можно масштабировать transcoder horizontally
4. **Secure:** Multi-layer auth (JWT + Internal keys + FDW)
5. **Real-time:** Live HLS streaming с низкой задержкой
6. **RESTful:** Чистые REST API endpoints

***

## 🔜 Возможные улучшения (future)

- WebSocket для real-time updates (viewer count, chat)
- Redis для caching и session storage
- Webhook-based stream monitoring (вместо polling)
- CDN integration для глобального стриминга
- AI-powered thumbnail selection
- Stream analytics dashboard
- Mobile apps (React Native)

***

**Статус:** ✅ **Fully Operational MVP**

Все core features работают, система готова к тестированию и демонстрации! 🎉