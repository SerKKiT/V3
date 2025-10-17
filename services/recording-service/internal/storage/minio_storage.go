package storage

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
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
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("✅ Bucket '%s' created", bucketName)
	}

	return &MinIOStorage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *MinIOStorage) UploadRecording(ctx context.Context, localPath, objectName string) error {
	// Проверяем что файл существует
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", localPath)
	}

	log.Printf("📤 Opening file for upload: %s", localPath)
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", localPath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	log.Printf("📤 Uploading %d bytes to MinIO bucket '%s': %s", stat.Size(), s.bucketName, objectName)
	_, err = s.client.PutObject(ctx, s.bucketName, objectName, file, stat.Size(), minio.PutObjectOptions{
		ContentType: "video/mp4",
	})
	if err != nil {
		return fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	log.Printf("✅ Uploaded to MinIO: %s/%s (%d bytes)", s.bucketName, objectName, stat.Size())
	return nil
}

// UploadThumbnail загружает thumbnail в MinIO
func (s *MinIOStorage) UploadThumbnail(ctx context.Context, localPath, objectName string) error {
	// Проверяем что файл существует
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("thumbnail file does not exist: %s", localPath)
	}

	log.Printf("📤 Uploading thumbnail: %s to %s/%s", localPath, s.bucketName, objectName)
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open thumbnail: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat thumbnail: %w", err)
	}

	_, err = s.client.PutObject(ctx, s.bucketName, objectName, file, stat.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return fmt.Errorf("failed to upload thumbnail to MinIO: %w", err)
	}

	log.Printf("✅ Thumbnail uploaded: %s/%s (%d bytes)", s.bucketName, objectName, stat.Size())
	return nil
}

// GetClient возвращает MinIO client
func (s *MinIOStorage) GetClient() *minio.Client {
	return s.client
}

// ✅ НОВОЕ: GetBucketName возвращает имя bucket
func (s *MinIOStorage) GetBucketName() string {
	return s.bucketName
}
