Все Endpoints Auth Service
1. Регистрация пользователя
Endpoint: POST /register

Описание: Создаёт нового пользователя и возвращает JWT токен

Request:

json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
Validation:

username: обязательно, 3-50 символов

email: обязательно, валидный email формат

password: обязательно, минимум 8 символов

Response (201 Created):

json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-06T04:01:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2025-10-05T04:01:00Z",
    "updated_at": "2025-10-05T04:01:00Z"
  }
}
Error Response (400 Bad Request):

json
{
  "error": "username already exists"
}
Curl команда:

text
curl -X POST http://localhost:8081/register -H "Content-Type: application/json" -d "{\"username\":\"newuser\",\"email\":\"new@example.com\",\"password\":\"securepass123\"}"
2. Логин пользователя
Endpoint: POST /login

Описание: Аутентифицирует пользователя и возвращает новый JWT токен

Request:

json
{
  "username": "testuser",
  "password": "password123"
}
Response (200 OK):

json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-06T04:01:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2025-10-05T04:01:00Z",
    "updated_at": "2025-10-05T04:01:00Z"
  }
}
Error Response (401 Unauthorized):

json
{
  "error": "invalid credentials"
}
Curl команда:

text
curl -X POST http://localhost:8081/login -H "Content-Type: application/json" -d "{\"username\":\"testuser\",\"password\":\"password123\"}"
3. Проверка JWT токена
Endpoint: GET /verify

Описание: Проверяет валидность JWT токена и возвращает информацию о пользователе

Headers:

text
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Response (200 OK):

json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "testuser",
  "valid": true
}
Error Response (401 Unauthorized):

json
{
  "error": "Authorization header required"
}
json
{
  "error": "Invalid token"
}
Curl команда:

text
curl -X GET http://localhost:8081/verify -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwidXNlcm5hbWUiOiJ0ZXN0dXNlciIsImV4cCI6MTcyODI4ODAwMCwiaWF0IjoxNzI4MjAxNjAwLCJuYmYiOjE3MjgyMDE2MDB9.signature"
4. Health Check
Endpoint: GET /health

Описание: Проверяет работоспособность сервиса

Response (200 OK):

json
{
  "status": "healthy"
}
Curl команда:

text
curl http://localhost:8081/health
Таблица всех Endpoints
Method	Endpoint	Auth Required	Description
POST	/register	❌	Регистрация нового пользователя
POST	/login	❌	Аутентификация пользователя
GET	/verify	✅	Проверка JWT токена
GET	/health	❌	Health check сервиса
Безопасность и Best Practices
1. Password Hashing (Bcrypt)
Cost factor: 10 раундов (по умолчанию)

Соль: генерируется автоматически для каждого пароля

Время: ~100-200ms на хеширование (защита от brute-force)

2. JWT Token
Алгоритм: HS256 (HMAC-SHA256)

Срок действия: 24 часа

Secret: хранится в переменной окружения JWT_SECRET

Формат: Bearer token в Authorization header

3. Database Security
Connection pooling: 25 активных, 5 idle соединений

Prepared statements: защита от SQL injection

Password: не возвращается в API ответах (json:"-")

4. CORS
Middleware: разрешает все origins (*)

Methods: GET, POST, PUT, DELETE, OPTIONS

Headers: Content-Type, Authorization

Примеры использования
Полный flow регистрации и верификации
text
# 1. Регистрация
curl -X POST http://localhost:8081/register -H "Content-Type: application/json" -d "{\"username\":\"alice\",\"email\":\"alice@example.com\",\"password\":\"alicepass123\"}"

# Ответ:
# {
#   "token": "eyJhbGc...",
#   "expires_at": "2025-10-06T04:01:00Z",
#   "user": {...}
# }

# 2. Сохранить токен
SET TOKEN=eyJhbGc...

# 3. Проверить токен
curl -X GET http://localhost:8081/verify -H "Authorization: Bearer %TOKEN%"

# 4. Логин (получить новый токен)
curl -X POST http://localhost:8081/login -H "Content-Type: application/json" -d "{\"username\":\"alice\",\"password\":\"alicepass123\"}"
Интеграция с другими сервисами
Другие микросервисы могут проверять токены через endpoint /verify:

go
// Пример из Stream Service
func validateToken(token string) error {
    resp, err := http.Get("http://auth-service:8081/verify", 
        headers: {"Authorization": "Bearer " + token})
    
    if resp.StatusCode != 200 {
        return errors.New("invalid token")
    }
    return nil
}