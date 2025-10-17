package srt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/SerKKiT/streaming-platform/stream-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/stream-service/internal/transcoder"
	gosrt "github.com/datarhei/gosrt"
)

type Handler struct {
	streamRepo          *repository.StreamRepository
	transcoder          *transcoder.FFmpegTranscoder
	recordingServiceURL string
}

func NewHandler(streamRepo *repository.StreamRepository, transcoder *transcoder.FFmpegTranscoder, recordingServiceURL string) *Handler {
	return &Handler{
		streamRepo:          streamRepo,
		transcoder:          transcoder,
		recordingServiceURL: recordingServiceURL,
	}
}

// StreamEventPayload represents webhook data
type StreamEventPayload struct {
	StreamKey string `json:"stream_key"`
	Event     string `json:"event"` // "started" or "stopped"
	HLSURL    string `json:"hls_url"`
	Timestamp int64  `json:"timestamp"`
}

// ValidateStreamKey checks if the stream key is valid
func (h *Handler) ValidateStreamKey(streamKey string) bool {
	stream, err := h.streamRepo.GetStreamByKey(streamKey)
	if err != nil {
		log.Printf("❌ Stream key validation failed: %s", streamKey)
		return false
	}

	log.Printf("✅ Stream key validated: %s (stream ID: %s)", streamKey, stream.ID)
	return true
}

// HandlePublish handles incoming SRT stream
func (h *Handler) HandlePublish(req gosrt.ConnRequest) {
	streamKey := req.StreamId()

	log.Printf("📡 Incoming SRT connection: stream_key=%s", streamKey)

	// Validate stream key
	stream, err := h.streamRepo.GetStreamByKey(streamKey)
	if err != nil {
		log.Printf("❌ Invalid stream key: %s", streamKey)
		req.Reject(gosrt.REJ_PEER)
		return
	}

	// Accept connection
	conn, err := req.Accept()
	if err != nil {
		log.Printf("❌ Failed to accept connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("✅ SRT connection accepted for stream %s", streamKey)

	// Update stream status to live
	if err := h.streamRepo.UpdateStreamStatus(stream.ID, "live"); err != nil {
		log.Printf("❌ Failed to update stream status: %v", err)
		return
	}

	// Send webhook: stream started
	hlsURL := fmt.Sprintf("http://localhost/live-streams/live-segments/%s/playlist.m3u8", streamKey)
	h.sendWebhook(streamKey, "started", hlsURL)

	// Start FFmpeg transcoding
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("🎬 Starting transcoding for stream %s", streamKey)
	if err := h.transcoder.TranscodeToHLS(ctx, conn, streamKey); err != nil {
		log.Printf("❌ Transcoding failed for stream %s: %v", streamKey, err)
	}

	// Update stream status to offline
	if err := h.streamRepo.UpdateStreamStatus(stream.ID, "offline"); err != nil {
		log.Printf("❌ Failed to update stream status: %v", err)
	}

	// ADDED: Update thumbnail URL in database
	thumbnailURL := fmt.Sprintf("http://localhost:9000/live-streams/live-segments/%s/thumbnail.jpg", streamKey)
	if err := h.streamRepo.UpdateStreamThumbnail(stream.ID, thumbnailURL); err != nil {
		log.Printf("⚠️  Failed to update thumbnail URL: %v", err)
	} else {
		log.Printf("✅ Updated thumbnail URL for stream %s", streamKey)
	}

	// Send webhook: stream stopped
	h.sendWebhook(streamKey, "stopped", hlsURL)

	log.Printf("⏹️  Stream ended: %s", streamKey)
}

// sendWebhook sends webhook to recording service
func (h *Handler) sendWebhook(streamKey, event, hlsURL string) {
	payload := StreamEventPayload{
		StreamKey: streamKey,
		Event:     event,
		HLSURL:    hlsURL,
		Timestamp: time.Now().Unix(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Failed to marshal webhook payload: %v", err)
		return
	}

	go func() {
		url := fmt.Sprintf("%s/webhook/stream", h.recordingServiceURL)
		log.Printf("📤 Sending webhook to %s: %s", url, event)

		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("❌ Failed to create webhook request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to send webhook: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("⚠️  Webhook returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
			return
		}

		log.Printf("✅ Webhook sent successfully: %s", event)
	}()
}
