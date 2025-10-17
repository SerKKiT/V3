package transcoder

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/SerKKiT/streaming-platform/stream-service/internal/repository"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FFmpegTranscoder struct {
	outputDir     string
	minioClient   *minio.Client
	minioBucket   string
	streamRepo    *repository.StreamRepository
	publicBaseURL string
	abrConfig     ABRConfig
}

func NewFFmpegTranscoder(
	outputDir string,
	minioEndpoint, minioAccessKey, minioSecretKey, minioBucket string,
	useSSL bool,
	streamRepo *repository.StreamRepository,
	publicBaseURL string,
) (*FFmpegTranscoder, error) {
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, minioBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, minioBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &FFmpegTranscoder{
		outputDir:     outputDir,
		minioClient:   minioClient,
		minioBucket:   minioBucket,
		streamRepo:    streamRepo,
		publicBaseURL: publicBaseURL,
		abrConfig:     DefaultABRConfig,
	}, nil
}

// TranscodeToHLS with Adaptive Bitrate (multiple qualities)
func (t *FFmpegTranscoder) TranscodeToHLS(ctx context.Context, input io.Reader, streamKey string) error {
	outputPath := filepath.Join(t.outputDir, streamKey)
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Создаём директории для каждого качества
	for _, profile := range t.abrConfig.Profiles {
		qualityPath := filepath.Join(outputPath, profile.Name)
		if err := os.MkdirAll(qualityPath, 0755); err != nil {
			return fmt.Errorf("failed to create quality directory %s: %w", profile.Name, err)
		}
	}

	// Build FFmpeg command for ABR
	args := t.buildABRCommand(streamKey, outputPath)

	log.Printf("🎬 Starting ABR transcoding for stream %s with qualities: %v",
		streamKey, GetProfileNames())

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdin = input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Запускаем генерацию thumbnail через 10 секунд
	go t.generateThumbnailAfterDelay(ctx, streamKey, outputPath, 10*time.Second)

	// Запускаем мониторинг и загрузку сегментов для всех качеств
	go t.monitorAndUploadABRSegments(ctx, streamKey, outputPath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg ABR failed: %w", err)
	}

	// После завершения загружаем оставшиеся файлы
	t.uploadRemainingABRSegments(streamKey, outputPath)
	log.Printf("✅ ABR transcoding completed for stream %s", streamKey)
	return nil
}

// buildABRCommand создает FFmpeg команду для множественных качеств
func (t *FFmpegTranscoder) buildABRCommand(streamKey, outputPath string) []string {
	profiles := t.abrConfig.Profiles
	numProfiles := len(profiles)

	args := []string{
		"-hide_banner",
		"-i", "pipe:0",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-g", "60",
		"-keyint_min", "60",
		"-sc_threshold", "0",
		"-pix_fmt", "yuv420p",
	}

	// Создаем filter_complex для разных разрешений
	filterComplex := ""
	if numProfiles > 1 {
		filterComplex = fmt.Sprintf("[0:v]split=%d", numProfiles)
		for i := 0; i < numProfiles; i++ {
			filterComplex += fmt.Sprintf("[v%d]", i)
		}
		filterComplex += ";"

		for i, profile := range profiles {
			filterComplex += fmt.Sprintf("[v%d]scale=%d:%d[v%dout]",
				i, profile.Width, profile.Height, i)
			if i < numProfiles-1 {
				filterComplex += ";"
			}
		}
	} else {
		filterComplex = fmt.Sprintf("[0:v]scale=%d:%d[v0out]",
			profiles[0].Width, profiles[0].Height)
	}

	args = append(args, "-filter_complex", filterComplex)

	// Добавляем map и параметры для каждого качества
	var varStreamMap []string
	for i, profile := range profiles {
		// Video mapping
		args = append(args, "-map", fmt.Sprintf("[v%dout]", i))
		args = append(args, fmt.Sprintf("-c:v:%d", i), "libx264")
		args = append(args, fmt.Sprintf("-b:v:%d", i), profile.VideoBitrate)
		args = append(args, fmt.Sprintf("-maxrate:v:%d", i), profile.MaxRate)
		args = append(args, fmt.Sprintf("-bufsize:v:%d", i), profile.BufSize)

		// Audio mapping
		args = append(args, "-map", "a:0")
		args = append(args, fmt.Sprintf("-c:a:%d", i), "aac")
		args = append(args, fmt.Sprintf("-b:a:%d", i), profile.AudioBitrate)
		args = append(args, "-ar", "48000", "-ac", "2")

		varStreamMap = append(varStreamMap,
			fmt.Sprintf("v:%d,a:%d,name:%s", i, i, profile.Name))
	}

	// ✅ ОБНОВЛЕНО: HLS параметры
	args = append(args,
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", t.abrConfig.SegmentTime), // Берется из config (4 сек)
		"-hls_list_size", fmt.Sprintf("%d", t.abrConfig.PlaylistSize), // 0 = все сегменты
		"-hls_flags", "delete_segments+append_list+independent_segments+program_date_time", // ✅ Добавлен program_date_time
		"-hls_playlist_type", t.abrConfig.PlaylistType,
		"-hls_segment_type", "mpegts",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", strings.Join(varStreamMap, " "),
		"-hls_segment_filename", filepath.Join(outputPath, "%v", "segment_%03d.ts"),
		filepath.Join(outputPath, "%v", "playlist.m3u8"),
	)

	return args
}

// monitorAndUploadABRSegments загружает сегменты всех качеств
func (t *FFmpegTranscoder) monitorAndUploadABRSegments(ctx context.Context, streamKey, outputPath string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	uploadedSegments := make(map[string]bool)
	var mu sync.Mutex
	const maxConcurrentUploads = 10
	semaphore := make(chan struct{}, maxConcurrentUploads)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var wg sync.WaitGroup

			// Загружаем сегменты для каждого качества
			for _, profile := range t.abrConfig.Profiles {
				qualityPath := filepath.Join(outputPath, profile.Name)

				// Загружаем .ts сегменты
				files, err := filepath.Glob(filepath.Join(qualityPath, "segment_*.ts"))
				if err != nil {
					continue
				}

				for _, file := range files {
					fileName := filepath.Base(file)
					qualityKey := profile.Name + "/" + fileName

					mu.Lock()
					alreadyUploaded := uploadedSegments[qualityKey]
					mu.Unlock()

					if alreadyUploaded {
						continue
					}

					// Проверяем что файл завершен
					fileInfo, err := os.Stat(file)
					if err != nil {
						continue
					}

					time.Sleep(300 * time.Millisecond)
					fileInfo2, err := os.Stat(file)
					if err != nil || fileInfo.Size() != fileInfo2.Size() {
						continue
					}

					semaphore <- struct{}{}
					wg.Add(1)
					go func(filePath, quality, fileName string) {
						defer wg.Done()
						defer func() { <-semaphore }()

						objectName := fmt.Sprintf("live-segments/%s/%s/%s", streamKey, quality, fileName)
						err := t.uploadFileToMinIO(filePath, objectName, "video/mp2t")
						if err != nil {
							log.Printf("❌ Failed to upload %s/%s: %v", quality, fileName, err)
							return
						}

						mu.Lock()
						uploadedSegments[qualityKey] = true
						mu.Unlock()

						log.Printf("📦 Uploaded %s/%s", quality, fileName)
					}(file, profile.Name, fileName)
				}

				// Загружаем playlist.m3u8 для этого качества
				playlistPath := filepath.Join(qualityPath, "playlist.m3u8")
				if _, err := os.Stat(playlistPath); err == nil {
					objectName := fmt.Sprintf("live-segments/%s/%s/playlist.m3u8", streamKey, profile.Name)
					_ = t.uploadFileToMinIO(playlistPath, objectName, "application/vnd.apple.mpegurl")
				}
			}

			wg.Wait()

			// Загружаем master playlist
			masterPath := filepath.Join(outputPath, "master.m3u8")
			if _, err := os.Stat(masterPath); err == nil {
				objectName := fmt.Sprintf("live-segments/%s/master.m3u8", streamKey)
				_ = t.uploadFileToMinIO(masterPath, objectName, "application/vnd.apple.mpegurl")
			}
		}
	}
}

// uploadFileToMinIO универсальный метод загрузки файла
func (t *FFmpegTranscoder) uploadFileToMinIO(filePath, objectName, contentType string) error {
	ctx := context.Background()
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = t.minioClient.PutObject(ctx, t.minioBucket, objectName, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// uploadRemainingABRSegments загружает все оставшиеся файлы
func (t *FFmpegTranscoder) uploadRemainingABRSegments(streamKey, outputPath string) {
	log.Printf("📤 Uploading remaining ABR segments for stream %s", streamKey)

	var wg sync.WaitGroup

	// Загружаем master playlist
	masterPath := filepath.Join(outputPath, "master.m3u8")
	if _, err := os.Stat(masterPath); err == nil {
		objectName := fmt.Sprintf("live-segments/%s/master.m3u8", streamKey)
		if err := t.uploadFileToMinIO(masterPath, objectName, "application/vnd.apple.mpegurl"); err != nil {
			log.Printf("❌ Failed to upload master.m3u8: %v", err)
		} else {
			log.Printf("✅ Uploaded master.m3u8")
		}
	}

	// Загружаем файлы для каждого качества
	for _, profile := range t.abrConfig.Profiles {
		qualityPath := filepath.Join(outputPath, profile.Name)

		// Playlist для качества
		playlistPath := filepath.Join(qualityPath, "playlist.m3u8")
		if _, err := os.Stat(playlistPath); err == nil {
			objectName := fmt.Sprintf("live-segments/%s/%s/playlist.m3u8", streamKey, profile.Name)
			_ = t.uploadFileToMinIO(playlistPath, objectName, "application/vnd.apple.mpegurl")
		}

		// Все сегменты
		files, _ := filepath.Glob(filepath.Join(qualityPath, "segment_*.ts"))
		for _, file := range files {
			fileName := filepath.Base(file)
			wg.Add(1)
			go func(filePath, quality, fileName string) {
				defer wg.Done()
				objectName := fmt.Sprintf("live-segments/%s/%s/%s", streamKey, quality, fileName)
				if err := t.uploadFileToMinIO(filePath, objectName, "video/mp2t"); err != nil {
					log.Printf("❌ Failed to upload %s: %v", fileName, err)
				}
			}(file, profile.Name, fileName)
		}
	}

	wg.Wait()
	log.Printf("✅ All ABR segments uploaded for stream %s", streamKey)
}

// generateThumbnailAfterDelay генерирует thumbnail через заданную задержку
func (t *FFmpegTranscoder) generateThumbnailAfterDelay(ctx context.Context, streamKey, outputPath string, delay time.Duration) {
	log.Printf("📸 Will generate thumbnail for stream %s in %v", streamKey, delay)
	select {
	case <-time.After(delay):
		// Продолжаем
	case <-ctx.Done():
		log.Printf("⚠️ Stream %s ended before thumbnail generation", streamKey)
		return
	}

	// ✅ Ждем пока появится хотя бы один сегмент в любой из папок качества
	var firstSegment string
	for i := 0; i < 20; i++ {
		// Проверяем все папки качества
		for _, quality := range []string{"1080p", "720p", "480p", "360p"} {
			qualityPath := filepath.Join(outputPath, quality)
			segments, _ := filepath.Glob(filepath.Join(qualityPath, "segment_*.ts"))
			if len(segments) > 0 {
				firstSegment = segments[0]
				log.Printf("✅ Found segment in %s: %s", quality, firstSegment)
				break
			}
		}
		if firstSegment != "" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if firstSegment == "" {
		log.Printf("❌ No segments found in any quality folder for thumbnail generation")
		return
	}

	thumbnailPath := filepath.Join(outputPath, "thumbnail.jpg")

	// FFmpeg команда для извлечения кадра
	args := []string{
		"-hide_banner",
		"-i", firstSegment,
		"-ss", "00:00:01",
		"-vframes", "1",
		"-vf", "scale=640:-1",
		"-q:v", "2",
		"-y",
		thumbnailPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Failed to generate thumbnail: %v\nOutput: %s", err, string(output))
		return
	}

	log.Printf("✅ Thumbnail generated: %s", thumbnailPath)

	// Загружаем thumbnail в MinIO
	if err := t.uploadThumbnailToMinIO(streamKey, thumbnailPath); err != nil {
		log.Printf("❌ Failed to upload thumbnail: %v", err)
	} else {
		log.Printf("✅ Thumbnail uploaded to MinIO for stream %s", streamKey)
	}
}

// uploadThumbnailToMinIO загружает thumbnail и обновляет БД
func (t *FFmpegTranscoder) uploadThumbnailToMinIO(streamKey, thumbnailPath string) error {
	objectName := fmt.Sprintf("live-segments/%s/thumbnail.jpg", streamKey)
	if err := t.uploadFileToMinIO(thumbnailPath, objectName, "image/jpeg"); err != nil {
		return err
	}

	log.Printf("✅ Thumbnail uploaded to MinIO: %s", objectName)
	stream, err := t.streamRepo.GetStreamByKey(streamKey)
	if err != nil {
		log.Printf("⚠️ Failed to get stream by key %s: %v", streamKey, err)
		return err
	}

	thumbnailURL := fmt.Sprintf("%s/api/streams/%s/thumbnail", t.publicBaseURL, stream.ID)
	if err := t.streamRepo.UpdateStreamThumbnail(stream.ID, thumbnailURL); err != nil {
		log.Printf("⚠️ Failed to update thumbnail_url in DB: %v", err)
	} else {
		log.Printf("✅ Updated thumbnail_url in DB for stream %s", stream.ID)
	}

	return nil
}
