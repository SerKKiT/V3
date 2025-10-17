package recorder

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
)

type FFmpegRecorder struct {
	recordingsPath string
	minioClient    *minio.Client
	minioBucket    string
}

func NewFFmpegRecorder(recordingsPath string, minioClient *minio.Client, segmentsBucket string) *FFmpegRecorder {
	os.MkdirAll(recordingsPath, 0755)
	return &FFmpegRecorder{
		recordingsPath: recordingsPath,
		minioClient:    minioClient,
		minioBucket:    segmentsBucket, // live-streams
	}
}

// ProcessRecordingFromMinIO скачивает сегменты из MinIO и создает MP4
func (r *FFmpegRecorder) ProcessRecordingFromMinIO(ctx context.Context, streamKey, recordingID string) (string, error) {
	log.Printf("📹 Processing recording for stream %s", streamKey)

	// Создать временную директорию для сегментов
	tempDir := filepath.Join(r.recordingsPath, "temp", recordingID)
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir) // Очистить после

	// ✅ Скачать сегменты ОДНОГО качества (приоритет: 1080p -> 720p -> 480p -> 360p)
	segmentFiles, selectedQuality, err := r.downloadSingleQualitySegments(ctx, streamKey, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to download segments: %w", err)
	}

	if len(segmentFiles) == 0 {
		return "", fmt.Errorf("no segments found for stream %s", streamKey)
	}

	log.Printf("✅ Downloaded %d segments from quality '%s' for stream %s", len(segmentFiles), selectedQuality, streamKey)

	// Создать concat file для FFmpeg
	concatFile := filepath.Join(tempDir, "concat.txt")
	err = r.createConcatFile(segmentFiles, concatFile)
	if err != nil {
		return "", fmt.Errorf("failed to create concat file: %w", err)
	}

	// Конкатенировать сегменты в MP4
	outputPath := filepath.Join(r.recordingsPath, fmt.Sprintf("%s.mp4", recordingID))
	err = r.concatenateSegments(concatFile, outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to concatenate segments: %w", err)
	}

	log.Printf("✅ Recording completed: %s", outputPath)
	return outputPath, nil
}

// downloadSingleQualitySegments скачивает сегменты ОДНОГО качества
func (r *FFmpegRecorder) downloadSingleQualitySegments(ctx context.Context, streamKey, tempDir string) ([]string, string, error) {
	// Приоритет качеств: 1080p (лучший) -> 1080p -> 480p -> 360p
	qualityPriority := []string{"1080p", "720p", "480p", "360p"}

	for _, quality := range qualityPriority {
		prefix := fmt.Sprintf("live-segments/%s/%s/", streamKey, quality)
		log.Printf("🔍 Checking for segments in: %s", prefix)

		objectCh := r.minioClient.ListObjects(ctx, r.minioBucket, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: false,
		})

		var segmentFiles []string
		for object := range objectCh {
			if object.Err != nil {
				log.Printf("⚠️ Error listing objects: %v", object.Err)
				continue
			}

			// Скачать только .ts файлы
			if !strings.HasSuffix(object.Key, ".ts") {
				continue
			}

			fileName := filepath.Base(object.Key)
			localPath := filepath.Join(tempDir, fileName)

			err := r.minioClient.FGetObject(ctx, r.minioBucket, object.Key, localPath, minio.GetObjectOptions{})
			if err != nil {
				log.Printf("❌ Failed to download segment %s: %v", object.Key, err)
				continue
			}

			segmentFiles = append(segmentFiles, localPath)
			log.Printf("📥 Downloaded: %s", fileName)
		}

		// Если нашли сегменты в этом качестве - используем его
		if len(segmentFiles) > 0 {
			// Сортировать по номеру сегмента (segment_000.ts, segment_001.ts, ...)
			sort.Slice(segmentFiles, func(i, j int) bool {
				return extractSegmentNumber(segmentFiles[i]) < extractSegmentNumber(segmentFiles[j])
			})

			log.Printf("✅ Found %d segments in quality '%s'", len(segmentFiles), quality)
			return segmentFiles, quality, nil
		}

		log.Printf("⚠️ No segments found in quality '%s', trying next...", quality)
	}

	return nil, "", fmt.Errorf("no segments found in any quality for stream %s", streamKey)
}

// extractSegmentNumber извлекает номер из имени сегмента (segment_123.ts -> 123)
func extractSegmentNumber(filename string) int {
	base := filepath.Base(filename)
	// segment_000.ts -> 000
	parts := strings.Split(base, "_")
	if len(parts) < 2 {
		return 0
	}
	numStr := strings.TrimSuffix(parts[1], ".ts")
	num, _ := strconv.Atoi(numStr)
	return num
}

// createConcatFile создает файл списка для FFmpeg concat
func (r *FFmpegRecorder) createConcatFile(segmentFiles []string, concatFile string) error {
	file, err := os.Create(concatFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, segmentFile := range segmentFiles {
		_, err := fmt.Fprintf(file, "file '%s'\n", segmentFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// concatenateSegments объединяет сегменты в один MP4 файл
func (r *FFmpegRecorder) concatenateSegments(concatFile, outputPath string) error {
	args := []string{
		"-hide_banner",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	}

	log.Printf("🎬 Concatenating segments: ffmpeg %v", args)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	return nil
}

// GetFileDuration получает длительность видео через ffprobe
func (r *FFmpegRecorder) GetFileDuration(filePath string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var duration float64
	fmt.Sscanf(string(output), "%f", &duration)
	return int(duration), nil
}

// GetFileInfo получает информацию о файле
func (r *FFmpegRecorder) GetFileInfo(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}

// GenerateThumbnail создаёт thumbnail из видео файла
func (r *FFmpegRecorder) GenerateThumbnail(videoPath, thumbnailPath string) error {
	log.Printf("🖼️ Generating thumbnail for: %s", videoPath)

	// Используем thumbnail filter для автоматического выбора лучшего кадра
	args := []string{
		"-hide_banner",
		"-i", videoPath,
		"-vf", "thumbnail=300,scale=1920:1080:force_original_aspect_ratio=decrease",
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		thumbnailPath,
	}

	log.Printf("Running: ffmpeg %v", args)
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg thumbnail generation failed: %w, output: %s", err, string(output))
	}

	log.Printf("✅ Thumbnail generated: %s", thumbnailPath)
	return nil
}

// GenerateThumbnailSimple создаёт thumbnail из первого кадра (fallback метод)
func (r *FFmpegRecorder) GenerateThumbnailSimple(videoPath, thumbnailPath string) error {
	log.Printf("🖼️ Generating simple thumbnail for: %s", videoPath)

	args := []string{
		"-hide_banner",
		"-ss", "00:00:01",
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		"-vf", "scale=1920:1080:force_original_aspect_ratio=decrease",
		"-y",
		thumbnailPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg simple thumbnail failed: %w, output: %s", err, string(output))
	}

	log.Printf("✅ Simple thumbnail generated: %s", thumbnailPath)
	return nil
}
