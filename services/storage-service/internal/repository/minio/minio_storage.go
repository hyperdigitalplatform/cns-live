package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog"
)

// MinIOStorage implements StorageRepository using MinIO
type MinIOStorage struct {
	client *minio.Client
	bucket string
	logger zerolog.Logger
}

// NewMinIOStorage creates a new MinIO storage repository
func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool, logger zerolog.Logger) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinIOStorage{
		client: client,
		bucket: bucket,
		logger: logger,
	}, nil
}

// Store uploads a file to MinIO
func (m *MinIOStorage) Store(ctx context.Context, path string, reader io.Reader, size int64) error {
	m.logger.Debug().
		Str("bucket", m.bucket).
		Str("path", path).
		Int64("size", size).
		Msg("Uploading to MinIO")

	_, err := m.client.PutObject(ctx, m.bucket, path, reader, size, minio.PutObjectOptions{
		ContentType: "video/mp2t",
	})
	if err != nil {
		return fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	m.logger.Info().
		Str("bucket", m.bucket).
		Str("path", path).
		Msg("Successfully uploaded to MinIO")

	return nil
}

// Get retrieves a file from MinIO
func (m *MinIOStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	m.logger.Debug().
		Str("bucket", m.bucket).
		Str("path", path).
		Msg("Getting from MinIO")

	object, err := m.client.GetObject(ctx, m.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get from MinIO: %w", err)
	}

	return object, nil
}

// Delete removes a file from MinIO
func (m *MinIOStorage) Delete(ctx context.Context, path string) error {
	m.logger.Debug().
		Str("bucket", m.bucket).
		Str("path", path).
		Msg("Deleting from MinIO")

	err := m.client.RemoveObject(ctx, m.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from MinIO: %w", err)
	}

	m.logger.Info().
		Str("bucket", m.bucket).
		Str("path", path).
		Msg("Successfully deleted from MinIO")

	return nil
}

// GenerateDownloadURL creates a temporary signed URL
func (m *MinIOStorage) GenerateDownloadURL(ctx context.Context, path string, expirySeconds int) (string, error) {
	m.logger.Debug().
		Str("bucket", m.bucket).
		Str("path", path).
		Int("expiry_seconds", expirySeconds).
		Msg("Generating download URL")

	expiry := time.Duration(expirySeconds) * time.Second
	reqParams := make(url.Values)

	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, path, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	m.logger.Debug().
		Str("url", presignedURL.String()).
		Msg("Generated download URL")

	return presignedURL.String(), nil
}

// List lists objects with a prefix
func (m *MinIOStorage) List(ctx context.Context, prefix string) ([]string, error) {
	m.logger.Debug().
		Str("bucket", m.bucket).
		Str("prefix", prefix).
		Msg("Listing objects from MinIO")

	var objects []string

	for object := range m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		objects = append(objects, object.Key)
	}

	m.logger.Debug().
		Int("count", len(objects)).
		Msg("Listed objects from MinIO")

	return objects, nil
}
