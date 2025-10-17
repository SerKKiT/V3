package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx := context.Background()

	// Check if bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		// Create bucket
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created bucket: %s", bucketName)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		useSSL:     useSSL,
	}

	// Set public read policy
	if err := storage.SetPublicReadPolicy(); err != nil {
		log.Printf("Warning: Failed to set public policy: %v", err)
		log.Println("Note: CORS must be configured via MinIO CLI or environment variables")
	}

	log.Printf("MinIO storage initialized: %s (bucket: %s, public: true)", endpoint, bucketName)

	return storage, nil
}

// SetPublicReadPolicy sets the bucket policy to allow public read access
func (s *MinIOStorage) SetPublicReadPolicy() error {
	// Create policy for public read access (GET only)
	bucketPolicy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": "*"},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, s.bucketName)

	ctx := context.Background()
	err := s.client.SetBucketPolicy(ctx, s.bucketName, bucketPolicy)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	log.Printf("Set public read policy for bucket: %s", s.bucketName)
	return nil
}

// GetObject returns object for streaming (ADDED - same as VOD service)
func (s *MinIOStorage) GetObject(ctx context.Context, objectName string) (*minio.Object, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return object, nil
}

// UploadFile uploads a file to MinIO
func (s *MinIOStorage) UploadFile(ctx context.Context, filePath, objectName string) error {
	contentType := "application/octet-stream"

	// Detect content type for HLS files
	ext := filepath.Ext(objectName)
	switch ext {
	case ".m3u8":
		contentType = "application/x-mpegURL"
	case ".ts":
		contentType = "video/mp2t"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".mp4":
		contentType = "video/mp4"
	}

	_, err := s.client.FPutObject(ctx, s.bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"Cache-Control": "no-cache, no-store, must-revalidate",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// GetPublicURL returns the public URL for an object
func (s *MinIOStorage) GetPublicURL(objectName string) string {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucketName, objectName)
}

// DeleteFile deletes a file from MinIO
func (s *MinIOStorage) DeleteFile(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// UploadStream uploads data from an io.Reader to MinIO
func (s *MinIOStorage) UploadStream(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload stream: %w", err)
	}
	return nil
}

// DeleteFolder удаляет все объекты с заданным префиксом (папку)
func (s *MinIOStorage) DeleteFolder(ctx context.Context, prefix string) error {
	log.Printf("🗑️  Deleting folder: %s/%s", s.bucketName, prefix)

	// Создаем канал для объектов
	objectsCh := make(chan minio.ObjectInfo)

	// Запускаем горутину для listing объектов
	go func() {
		defer close(objectsCh)

		opts := minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		}

		for object := range s.client.ListObjects(ctx, s.bucketName, opts) {
			if object.Err != nil {
				log.Printf("❌ Error listing object: %v", object.Err)
				continue
			}
			objectsCh <- object
		}
	}()

	// Удаляем объекты
	errorCh := s.client.RemoveObjects(ctx, s.bucketName, objectsCh, minio.RemoveObjectsOptions{})

	deletedCount := 0
	for err := range errorCh {
		if err.Err != nil {
			log.Printf("❌ Failed to delete %s: %v", err.ObjectName, err.Err)
		} else {
			deletedCount++
		}
	}

	log.Printf("✅ Deleted %d objects from %s", deletedCount, prefix)
	return nil
}

// DeleteStreamSegments удаляет все HLS файлы стрима из MinIO
func (s *MinIOStorage) DeleteStreamSegments(streamKey string) error {
	ctx := context.Background()
	prefix := fmt.Sprintf("live-segments/%s/", streamKey)

	log.Printf("🗑️ Deleting MinIO objects with prefix: %s", prefix)

	// ✅ ИСПРАВЛЕНО: Сначала собираем список объектов
	var objectsToDelete []string
	objectsCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("❌ Error listing object: %v", obj.Err)
			continue
		}
		objectsToDelete = append(objectsToDelete, obj.Key)
	}

	if len(objectsToDelete) == 0 {
		log.Printf("⚠️ No objects found to delete for stream %s", streamKey)
		return nil
	}

	log.Printf("📋 Found %d objects to delete", len(objectsToDelete))

	// ✅ Создаём канал объектов для удаления
	objectsChan := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsChan)
		for _, objName := range objectsToDelete {
			objectsChan <- minio.ObjectInfo{Key: objName}
		}
	}()

	// ✅ Удаляем объекты
	errorCh := s.client.RemoveObjects(ctx, s.bucketName, objectsChan, minio.RemoveObjectsOptions{})

	deletedCount := 0
	errorCount := 0
	for err := range errorCh {
		if err.Err != nil {
			log.Printf("❌ Failed to delete %s: %v", err.ObjectName, err.Err)
			errorCount++
		} else {
			deletedCount++
		}
	}

	log.Printf("✅ Deleted %d objects from MinIO for stream %s (errors: %d)",
		deletedCount, streamKey, errorCount)

	return nil
}
