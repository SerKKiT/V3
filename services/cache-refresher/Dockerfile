# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cache-refresher .

# Final stage - минимальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копируем бинарник из build stage
COPY --from=builder /app/cache-refresher .

# Healthcheck (опционально)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD pgrep -x cache-refresher > /dev/null || exit 1

CMD ["./cache-refresher"]
