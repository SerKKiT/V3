package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/SerKKiT/streaming-platform/recording-service/internal/recorder"
	"github.com/SerKKiT/streaming-platform/recording-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/recording-service/internal/storage"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type StreamInfo struct {
	ID        uuid.UUID `json:"id"`
	StreamKey string    `json:"stream_key"`
	Status    string    `json:"status"`
	HLSURL    string    `json:"hls_url"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
}

type StreamMonitor struct {
	streamServiceURL string
	vodServiceURL    string
	recorder         *recorder.FFmpegRecorder
	recordingRepo    *repository.RecordingRepository
	minioStorage     *storage.MinIOStorage
	activeRecordings map[uuid.UUID]context.CancelFunc
	streamKeyToID    map[string]uuid.UUID
	mu               sync.RWMutex
	interval         time.Duration
}

func NewStreamMonitor(
	streamServiceURL string,
	vodServiceURL string,
	recorder *recorder.FFmpegRecorder,
	recordingRepo *repository.RecordingRepository,
	minioStorage *storage.MinIOStorage,
	interval time.Duration,
) *StreamMonitor {
	if vodServiceURL == "" {
		vodServiceURL = "http://vod-service:8084"
	}

	return &StreamMonitor{
		streamServiceURL: streamServiceURL,
		vodServiceURL:    vodServiceURL,
		recorder:         recorder,
		recordingRepo:    recordingRepo,
		minioStorage:     minioStorage,
		activeRecordings: make(map[uuid.UUID]context.CancelFunc),
		streamKeyToID:    make(map[string]uuid.UUID),
		interval:         interval,
	}
}

func (m *StreamMonitor) Start(ctx context.Context) {
	log.Println("🔍 Stream monitor started")
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Stream monitor stopped")
			return
		case <-ticker.C:
			m.checkStreams()
		}
	}
}

func (m *StreamMonitor) checkStreams() {
	resp, err := http.Get(m.streamServiceURL + "/streams/live")
	if err != nil {
		log.Printf("⚠️ Failed to fetch live streams: %v", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Streams []StreamInfo `json:"streams"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("⚠️ Failed to decode streams: %v", err)
		return
	}

	log.Printf("Found %d live streams", len(result.Streams))

	m.mu.Lock()
	defer m.mu.Unlock()

	currentStreams := make(map[uuid.UUID]StreamInfo)
	for _, stream := range result.Streams {
		currentStreams[stream.ID] = stream
		m.streamKeyToID[stream.StreamKey] = stream.ID
	}

	// Останавливаем записи для стримов которые больше не live
	for streamID := range m.activeRecordings {
		if _, exists := currentStreams[streamID]; !exists {
			log.Printf("🛑 Stream %s is no longer live, stopping recording", streamID)
			m.stopRecordingLocked(streamID)
		}
	}
}

// HandleWebhookStart обрабатывает webhook о начале стрима
func (m *StreamMonitor) HandleWebhookStart(streamKey, hlsURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Получаем stream_id из маппинга или запрашиваем у Stream Service
	streamID, exists := m.streamKeyToID[streamKey]
	if !exists {
		streamInfo, err := m.getStreamInfoByKey(streamKey)
		if err != nil {
			return fmt.Errorf("failed to get stream info: %w", err)
		}
		streamID = streamInfo.ID
		m.streamKeyToID[streamKey] = streamID
	}

	if _, alreadyRecording := m.activeRecordings[streamID]; alreadyRecording {
		log.Printf("⚠️ Stream %s is already being recorded", streamID)
		return nil
	}

	log.Printf("🎬 Starting recording for stream %s (key: %s)", streamID, streamKey)

	// Создаём запись в БД
	recording, err := m.recordingRepo.CreateRecording(streamID, streamKey+".mp4")
	if err != nil {
		return fmt.Errorf("failed to create recording: %w", err)
	}

	// Запускаем запись в фоне
	ctx, cancel := context.WithCancel(context.Background())
	m.activeRecordings[streamID] = cancel

	go m.processRecording(ctx, streamKey, recording.ID, streamID)

	return nil
}

// HandleWebhookStop обрабатывает webhook об остановке стрима
func (m *StreamMonitor) HandleWebhookStop(streamKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamID, exists := m.streamKeyToID[streamKey]
	if !exists {
		log.Printf("⚠️ Unknown stream key: %s", streamKey)
		return nil
	}

	m.stopRecordingLocked(streamID)
	return nil
}

func (m *StreamMonitor) stopRecordingLocked(streamID uuid.UUID) {
	cancelFunc, exists := m.activeRecordings[streamID]
	if !exists {
		return
	}

	log.Printf("🛑 Stopping recording for stream %s", streamID)
	cancelFunc()
	delete(m.activeRecordings, streamID)
}

// waitForSegmentUploadCompletion ждёт когда счётчик файлов стабилизируется
func (m *StreamMonitor) waitForSegmentUploadCompletion(streamKey string, maxWaitTime time.Duration) int {
	log.Printf("⏳ Waiting for segment upload completion for stream %s", streamKey)

	deadline := time.Now().Add(maxWaitTime)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastCount int
	stableChecks := 0
	const requiredStableChecks = 4 // 2 секунды стабильности

	for time.Now().Before(deadline) {
		select {
		case <-ticker.C:
			client := m.minioStorage.GetClient()
			bucketName := m.minioStorage.GetBucketName()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			prefix := fmt.Sprintf("live-segments/%s/", streamKey)

			count := 0
			objectsCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
				Prefix:    prefix,
				Recursive: true,
			})

			for obj := range objectsCh {
				if obj.Err != nil {
					continue
				}
				if strings.HasSuffix(obj.Key, ".ts") || strings.HasSuffix(obj.Key, ".m3u8") {
					count++
				}
			}
			cancel()

			if count == lastCount && count > 0 {
				stableChecks++
				log.Printf("📊 Upload stable check %d/%d (files: %d)", stableChecks, requiredStableChecks, count)

				if stableChecks >= requiredStableChecks {
					log.Printf("✅ Segment upload completed (stable at %d files)", count)
					return count
				}
			} else {
				if count != lastCount {
					log.Printf("📈 File count changed: %d → %d", lastCount, count)
				}
				stableChecks = 0
				lastCount = count
			}
		}
	}

	log.Printf("⚠️ Timed out waiting for upload completion (last count: %d files)", lastCount)
	return lastCount
}

func (m *StreamMonitor) processRecording(ctx context.Context, streamKey string, recordingID uuid.UUID, streamID uuid.UUID) {
	<-ctx.Done()

	fileCount := m.waitForSegmentUploadCompletion(streamKey, 15*time.Second)
	log.Printf("📹 Processing completed recording %s (%d files)", recordingID, fileCount)

	outputPath, err := m.recorder.ProcessRecordingFromMinIO(context.Background(), streamKey, recordingID.String())

	success := err == nil
	defer func() {
		log.Printf("🧹 Sending cleanup webhook (success: %v)...", success)
		m.sendCleanupWebhook(streamKey, streamID, success)
	}()

	if err != nil {
		log.Printf("❌ Failed to process recording: %v", err)
		m.recordingRepo.UpdateRecordingStatus(recordingID, "failed")
		return
	}

	// Генерируем thumbnail
	thumbnailPath := outputPath + ".thumb.jpg"
	thumbnailGenerated := false
	if err := m.recorder.GenerateThumbnail(outputPath, thumbnailPath); err != nil {
		log.Printf("⚠️ Failed to generate thumbnail (trying fallback): %v", err)
		if err := m.recorder.GenerateThumbnailSimple(outputPath, thumbnailPath); err != nil {
			log.Printf("❌ Failed to generate simple thumbnail: %v", err)
		} else {
			thumbnailGenerated = true
		}
	} else {
		thumbnailGenerated = true
	}

	log.Printf("📦 Uploading recording to MinIO: %s", outputPath)

	if err := m.minioStorage.UploadRecording(context.Background(), outputPath, streamKey+".mp4"); err != nil {
		log.Printf("❌ Failed to upload recording: %v", err)
		m.recordingRepo.UpdateRecordingStatus(recordingID, "failed")
		success = false
		return
	}

	log.Printf("✅ Recording uploaded to MinIO: %s.mp4", streamKey)

	if thumbnailGenerated {
		thumbnailObjectName := streamKey + ".jpg"
		if err := m.minioStorage.UploadThumbnail(context.Background(), thumbnailPath, thumbnailObjectName); err != nil {
			log.Printf("⚠️ Failed to upload thumbnail: %v", err)
		} else {
			log.Printf("✅ Thumbnail uploaded to MinIO: %s", thumbnailObjectName)
			if err := m.recordingRepo.UpdateThumbnailPath(recordingID, thumbnailObjectName); err != nil {
				log.Printf("⚠️ Failed to update thumbnail path in DB: %v", err)
			} else {
				log.Printf("✅ Thumbnail path saved to DB: %s", thumbnailObjectName)
			}
		}
		os.Remove(thumbnailPath)
	}

	os.Remove(outputPath)
	m.recordingRepo.UpdateRecordingStatus(recordingID, "completed")
	log.Printf("✅ Recording %s completed successfully", recordingID)

	go m.triggerVODImport(streamID, recordingID)
}

func (m *StreamMonitor) getStreamInfoByKey(streamKey string) (*StreamInfo, error) {
	resp, err := http.Get(fmt.Sprintf("%s/streams/by-key/%s", m.streamServiceURL, streamKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stream service returned status %d", resp.StatusCode)
	}

	var streamInfo StreamInfo
	if err := json.NewDecoder(resp.Body).Decode(&streamInfo); err != nil {
		return nil, err
	}

	return &streamInfo, nil
}

func (m *StreamMonitor) GetStreamIDByKey(streamKey string) (uuid.UUID, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, exists := m.streamKeyToID[streamKey]
	return id, exists
}

func (m *StreamMonitor) sendCleanupWebhook(streamKey string, streamID uuid.UUID, success bool) {
	streamServiceURL := os.Getenv("STREAM_SERVICE_URL")
	if streamServiceURL == "" {
		streamServiceURL = "http://stream-service:8080"
	}

	webhookURL := fmt.Sprintf("%s/webhooks/recording-complete", streamServiceURL)

	payload := map[string]interface{}{
		"stream_key": streamKey,
		"stream_id":  streamID.String(),
		"success":    success,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Failed to marshal webhook payload: %v", err)
		return
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to create webhook request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send cleanup webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Cleanup webhook returned non-OK status: %d", resp.StatusCode)
	} else {
		log.Printf("✅ Cleanup webhook sent successfully for stream %s", streamKey)
	}
}

func (m *StreamMonitor) triggerVODImport(streamID uuid.UUID, recordingID uuid.UUID) {
	log.Printf("📥 Starting VOD import for recording %s", recordingID)

	recording, err := m.recordingRepo.GetByID(recordingID.String())
	if err != nil {
		log.Printf("❌ Failed to get recording %s: %v", recordingID, err)
		return
	}

	if recording.Status != "completed" {
		log.Printf("⚠️ Recording %s is not completed (status: %s), skipping import", recordingID, recording.Status)
		return
	}

	streamInfo, err := m.getStreamInfoByID(streamID)
	if err != nil {
		log.Printf("⚠️ Failed to get stream info: %v, using defaults", err)
		streamInfo = &StreamInfo{
			ID:     streamID,
			UserID: uuid.Nil,
		}
	}

	title := "Stream Recording"
	if streamInfo.Title != "" {
		title = fmt.Sprintf("Recording: %s", streamInfo.Title)
	} else {
		title = fmt.Sprintf("Stream Recording %s", recording.StartedAt.Format("2006-01-02 15:04"))
	}

	log.Printf("📤 Importing recording %s to VOD for user %s", recordingID, streamInfo.UserID)

	payload := map[string]interface{}{
		"recording_id": recordingID.String(),
		"title":        title,
		"description":  "Automatically imported stream recording",
		"visibility":   "public",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Failed to marshal VOD import payload: %v", err)
		return
	}

	importURL := m.vodServiceURL + "/videos/import-recording"
	req, err := http.NewRequest("POST", importURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to create VOD import request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", streamInfo.UserID.String())
	req.Header.Set("X-Internal-API-Key", os.Getenv("INTERNAL_API_KEY")) // ✅ ДОБАВЛЕНО

	client := &http.Client{Timeout: 15 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send VOD import request: %v", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ VOD import failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return
	}

	var result struct {
		VideoID string `json:"video_id"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		log.Printf("✅ Recording %s imported to VOD as video %s", recordingID, result.VideoID)
		videoID, _ := uuid.Parse(result.VideoID)
		recording.VideoID = &videoID
		m.recordingRepo.UpdateRecording(recording)
	} else {
		log.Printf("✅ Recording %s imported to VOD successfully", recordingID)
	}
}

func (m *StreamMonitor) getStreamInfoByID(streamID uuid.UUID) (*StreamInfo, error) {
	resp, err := http.Get(fmt.Sprintf("%s/streams/%s", m.streamServiceURL, streamID.String()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stream service returned status %d", resp.StatusCode)
	}

	var streamInfo StreamInfo
	if err := json.NewDecoder(resp.Body).Decode(&streamInfo); err != nil {
		return nil, err
	}

	return &streamInfo, nil
}
