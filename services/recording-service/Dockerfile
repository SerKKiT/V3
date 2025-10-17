FROM golang:1.25-alpine AS builder

# Install FFmpeg and ffprobe
RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o recording-service ./cmd

# Final stage
FROM alpine:latest

# Install FFmpeg in final image
RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /root/

COPY --from=builder /app/recording-service .

# Create recordings directory
RUN mkdir -p /tmp/recordings

EXPOSE 8083

CMD ["./recording-service"]
