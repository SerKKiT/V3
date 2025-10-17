FROM golang:1.25-alpine AS builder

# Install FFmpeg
RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o stream-service ./cmd

# Final stage
FROM alpine:latest

# Install FFmpeg in final image
RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /root/

COPY --from=builder /app/stream-service .

# Create HLS output directory
RUN mkdir -p /var/www/hls

EXPOSE 8082 6000/udp

CMD ["./stream-service"]
