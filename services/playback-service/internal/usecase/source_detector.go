package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rs/zerolog"
)

// SourceDetector detects the best source for playback (MinIO vs Milestone)
type SourceDetector struct {
	storageClient   StorageClient
	milestoneClient MilestoneClient
	logger          zerolog.Logger
}

// StorageClient interface for MinIO operations
type StorageClient interface {
	ListSegments(ctx context.Context, cameraID string, startTime, endTime time.Time) ([]*domain.Segment, error)
	GetSegmentDownloadURL(ctx context.Context, segmentID string, expirySeconds int) (string, error)
}

// MilestoneClient interface for Milestone VMS operations
type MilestoneClient interface {
	CheckRecordingAvailability(ctx context.Context, cameraID string, startTime, endTime time.Time) (bool, error)
	GetRecordingMetadata(ctx context.Context, cameraID string, startTime, endTime time.Time) (*MilestoneRecordingMetadata, error)
}

// MilestoneRecordingMetadata contains Milestone recording info
type MilestoneRecordingMetadata struct {
	Available    bool
	SegmentCount int
	TotalSize    int64
	StartTime    time.Time
	EndTime      time.Time
}

func NewSourceDetector(
	storageClient StorageClient,
	milestoneClient MilestoneClient,
	logger zerolog.Logger,
) *SourceDetector {
	return &SourceDetector{
		storageClient:   storageClient,
		milestoneClient: milestoneClient,
		logger:          logger,
	}
}

// DetectSource determines the best playback source for the requested time range
func (sd *SourceDetector) DetectSource(
	ctx context.Context,
	cameraID string,
	startTime, endTime time.Time,
) (*domain.SourceDetectionResult, error) {
	sd.logger.Info().
		Str("camera_id", cameraID).
		Time("start_time", startTime).
		Time("end_time", endTime).
		Msg("Detecting playback source")

	// Strategy: Check local storage first (faster, cheaper)
	// Fall back to Milestone if local is unavailable

	// 1. Check local MinIO storage
	localResult, localErr := sd.checkLocalStorage(ctx, cameraID, startTime, endTime)
	if localErr != nil {
		sd.logger.Warn().Err(localErr).Msg("Error checking local storage")
	}

	// If local storage has the recordings, use it
	if localResult != nil && localResult.Available && localResult.SegmentCount > 0 {
		sd.logger.Info().
			Str("source", string(domain.PlaybackSourceLocal)).
			Int("segments", localResult.SegmentCount).
			Int64("size_bytes", localResult.TotalSize).
			Msg("Using local storage")
		return localResult, nil
	}

	// 2. Check Milestone as fallback
	milestoneResult, milestoneErr := sd.checkMilestoneStorage(ctx, cameraID, startTime, endTime)
	if milestoneErr != nil {
		sd.logger.Error().Err(milestoneErr).Msg("Error checking Milestone storage")
		return nil, fmt.Errorf("no playback source available: local=%v, milestone=%v", localErr, milestoneErr)
	}

	if milestoneResult != nil && milestoneResult.Available {
		sd.logger.Info().
			Str("source", string(domain.PlaybackSourceMilestone)).
			Int("segments", milestoneResult.SegmentCount).
			Msg("Using Milestone storage")
		return milestoneResult, nil
	}

	// No source available
	return &domain.SourceDetectionResult{
		Source:    domain.PlaybackSourceLocal,
		Available: false,
		Reason:    "No recordings found in specified time range",
	}, nil
}

// checkLocalStorage checks MinIO for available segments
func (sd *SourceDetector) checkLocalStorage(
	ctx context.Context,
	cameraID string,
	startTime, endTime time.Time,
) (*domain.SourceDetectionResult, error) {
	segments, err := sd.storageClient.ListSegments(ctx, cameraID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	if len(segments) == 0 {
		return &domain.SourceDetectionResult{
			Source:       domain.PlaybackSourceLocal,
			Available:    false,
			SegmentCount: 0,
			TotalSize:    0,
			Reason:       "No segments found in local storage",
		}, nil
	}

	// Calculate total size
	var totalSize int64
	for _, segment := range segments {
		totalSize += segment.SizeBytes
	}

	// Check coverage - we want at least 80% of requested time range
	requestedDuration := endTime.Sub(startTime)
	var coveredDuration time.Duration
	for _, segment := range segments {
		coveredDuration += segment.EndTime.Sub(segment.StartTime)
	}

	coveragePercent := float64(coveredDuration) / float64(requestedDuration) * 100

	if coveragePercent < 80.0 {
		return &domain.SourceDetectionResult{
			Source:       domain.PlaybackSourceLocal,
			Available:    false,
			SegmentCount: len(segments),
			TotalSize:    totalSize,
			Reason:       fmt.Sprintf("Insufficient coverage: %.1f%% (need 80%%)", coveragePercent),
		}, nil
	}

	return &domain.SourceDetectionResult{
		Source:       domain.PlaybackSourceLocal,
		Available:    true,
		SegmentCount: len(segments),
		TotalSize:    totalSize,
	}, nil
}

// checkMilestoneStorage checks Milestone VMS for available recordings
func (sd *SourceDetector) checkMilestoneStorage(
	ctx context.Context,
	cameraID string,
	startTime, endTime time.Time,
) (*domain.SourceDetectionResult, error) {
	// Check if Milestone has recordings
	available, err := sd.milestoneClient.CheckRecordingAvailability(ctx, cameraID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check Milestone availability: %w", err)
	}

	if !available {
		return &domain.SourceDetectionResult{
			Source:    domain.PlaybackSourceMilestone,
			Available: false,
			Reason:    "No recordings found in Milestone",
		}, nil
	}

	// Get metadata
	metadata, err := sd.milestoneClient.GetRecordingMetadata(ctx, cameraID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get Milestone metadata: %w", err)
	}

	return &domain.SourceDetectionResult{
		Source:       domain.PlaybackSourceMilestone,
		Available:    metadata.Available,
		SegmentCount: metadata.SegmentCount,
		TotalSize:    metadata.TotalSize,
	}, nil
}
