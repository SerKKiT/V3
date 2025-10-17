package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
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

	log.Printf("✅ Connected to MinIO bucket: %s", bucketName)
	return &MinIOStorage{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
	}, nil
}

// CopyFromRecordings копирует объект из recordings bucket в vod-videos bucket
func (s *MinIOStorage) CopyFromRecordings(ctx context.Context, sourceBucket, sourceObject, destObject string) error {
	log.Printf("📋 Copying from %s/%s to %s/%s", sourceBucket, sourceObject, s.bucketName, destObject)

	src := minio.CopySrcOptions{
		Bucket: sourceBucket,
		Object: sourceObject,
	}

	contentType := "application/octet-stream"
	if strings.HasSuffix(destObject, ".mp4") {
		contentType = "video/mp4"
	} else if strings.HasSuffix(destObject, ".jpg") || strings.HasSuffix(destObject, ".jpeg") {
		contentType = "image/jpeg"
	}

	dst := minio.CopyDestOptions{
		Bucket: s.bucketName,
		Object: destObject,
		UserMetadata: map[string]string{
			"Content-Type": contentType,
		},
	}

	_, err := s.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy from %s to %s: %w", sourceObject, destObject, err)
	}

	log.Printf("✅ Copied to %s: %s", s.bucketName, destObject)
	return nil
}

// UploadFile загружает файл напрямую
func (s *MinIOStorage) UploadFile(ctx context.Context, localPath, objectName string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	_, err = s.client.PutObject(ctx, s.bucketName, objectName, file, stat.Size(), minio.PutObjectOptions{
		ContentType: "video/mp4",
	})
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	log.Printf("✅ Uploaded: %s", objectName)
	return nil
}

// DeleteObject удаляет объект
func (s *MinIOStorage) DeleteObject(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetObject возвращает объект для streaming из указанного bucket
func (s *MinIOStorage) GetObject(ctx context.Context, bucketName, objectName string) (*minio.Object, error) {
	object, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return object, nil
}
