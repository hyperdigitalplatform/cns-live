package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rs/zerolog"
)

// MinIOClient handles MinIO storage operations for playback
type MinIOClient struct {
	client     *minio.Client
	bucketName string
	logger     zerolog.Logger
}

func NewMinIOClient(
	endpoint string,
	accessKey string,
	secretKey string,
	bucketName string,
	useSSL bool,
	logger zerolog.Logger,
) (*MinIOClient, error) {
	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinIOClient{
		client:     client,
		bucketName: bucketName,
		logger:     logger,
	}, nil
}

// ListSegments lists video segments for a camera in a time range
func (m *MinIOClient) ListSegments(
	ctx context.Context,
	cameraID string,
	startTime, endTime time.Time,
) ([]*domain.Segment, error) {
	// Object naming: recordings/{camera_id}/{year}/{month}/{day}/{hour}/{timestamp}.mp4
	// We need to search across multiple prefixes based on the time range

	var segments []*domain.Segment

	// Generate prefixes to search (by day)
	prefixes := m.generatePrefixes(cameraID, startTime, endTime)

	for _, prefix := range prefixes {
		m.logger.Debug().Str("prefix", prefix).Msg("Listing objects")

		// List objects with prefix
		objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				m.logger.Error().Err(object.Err).Msg("Error listing objects")
				continue
			}

			// Parse segment metadata from object
			segment, err := m.parseSegmentFromObject(object, cameraID)
			if err != nil {
				m.logger.Warn().Err(err).Str("key", object.Key).Msg("Failed to parse segment")
				continue
			}

			// Check if segment is within time range
			if segment.StartTime.Before(endTime) && segment.EndTime.After(startTime) {
				segments = append(segments, &segment)
			}
		}
	}

	m.logger.Info().
		Str("camera_id", cameraID).
		Int("segment_count", len(segments)).
		Msg("Listed segments")

	return segments, nil
}

// GetSegment retrieves a video segment from MinIO
func (m *MinIOClient) GetSegment(ctx context.Context, storagePath, outputPath string) error {
	m.logger.Debug().
		Str("storage_path", storagePath).
		Str("output_path", outputPath).
		Msg("Downloading segment")

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get object from MinIO
	object, err := m.client.GetObject(ctx, m.bucketName, storagePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy data
	if _, err := io.Copy(outFile, object); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	m.logger.Debug().
		Str("storage_path", storagePath).
		Msg("Segment downloaded")

	return nil
}

// GetSegmentSize returns the size of a segment
func (m *MinIOClient) GetSegmentSize(ctx context.Context, storagePath string) (int64, error) {
	info, err := m.client.StatObject(ctx, m.bucketName, storagePath, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to stat object: %w", err)
	}
	return info.Size, nil
}

// GetSegmentURL generates a presigned URL for direct segment access
func (m *MinIOClient) GetSegmentURL(ctx context.Context, storagePath string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, storagePath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GetSegmentDownloadURL generates a presigned download URL for a segment by ID
func (m *MinIOClient) GetSegmentDownloadURL(ctx context.Context, segmentID string, expirySeconds int) (string, error) {
	// TODO: Map segmentID to storage path (currently we just use segment ID as path)
	// In a real implementation, you'd look up the segment metadata to get the storage path
	expiry := time.Duration(expirySeconds) * time.Second
	return m.GetSegmentURL(ctx, segmentID, expiry)
}

// generatePrefixes generates MinIO prefixes to search based on time range
func (m *MinIOClient) generatePrefixes(cameraID string, startTime, endTime time.Time) []string {
	var prefixes []string

	// Iterate through each day in the range
	current := startTime.Truncate(24 * time.Hour)
	end := endTime.Truncate(24 * time.Hour).Add(24 * time.Hour)

	for current.Before(end) {
		// recordings/{camera_id}/{year}/{month}/{day}/
		prefix := fmt.Sprintf("recordings/%s/%04d/%02d/%02d/",
			cameraID,
			current.Year(),
			current.Month(),
			current.Day(),
		)
		prefixes = append(prefixes, prefix)
		current = current.Add(24 * time.Hour)
	}

	return prefixes
}

// parseSegmentFromObject parses segment metadata from MinIO object
func (m *MinIOClient) parseSegmentFromObject(object minio.ObjectInfo, cameraID string) (domain.Segment, error) {
	// Object key format: recordings/{camera_id}/{year}/{month}/{day}/{hour}/{timestamp}.mp4
	// Example: recordings/cam-123/2024/01/20/14/1705761600.mp4

	// Extract timestamp from filename
	filename := filepath.Base(object.Key)
	var timestamp int64
	_, err := fmt.Sscanf(filename, "%d.mp4", &timestamp)
	if err != nil {
		return domain.Segment{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	startTime := time.Unix(timestamp, 0)

	// Segments are typically 60 seconds (configured in Recording Service)
	// We can get actual duration from metadata if needed
	duration := 60 // seconds (default)

	// Check for duration in metadata
	if durationStr, ok := object.UserMetadata["X-Amz-Meta-Duration"]; ok {
		fmt.Sscanf(durationStr, "%d", &duration)
	}

	endTime := startTime.Add(time.Duration(duration) * time.Second)

	return domain.Segment{
		ID:              fmt.Sprintf("%s_%d", cameraID, timestamp),
		CameraID:        cameraID,
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: duration,
		SizeBytes:       object.Size,
		StoragePath:     object.Key,
	}, nil
}

// DownloadSegments downloads multiple segments to a local directory
func (m *MinIOClient) DownloadSegments(ctx context.Context, segments []domain.Segment, outputDir string) ([]string, error) {
	var downloadedPaths []string

	for i, segment := range segments {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("segment_%03d.mp4", i))

		if err := m.GetSegment(ctx, segment.StoragePath, outputPath); err != nil {
			return nil, fmt.Errorf("failed to download segment %s: %w", segment.ID, err)
		}

		downloadedPaths = append(downloadedPaths, outputPath)
	}

	return downloadedPaths, nil
}

// CheckBucketExists checks if the bucket exists
func (m *MinIOClient) CheckBucketExists(ctx context.Context) (bool, error) {
	return m.client.BucketExists(ctx, m.bucketName)
}
