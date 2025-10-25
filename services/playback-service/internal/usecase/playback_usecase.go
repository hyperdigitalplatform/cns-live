package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rta/cctv/playback-service/internal/cache"
	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rta/cctv/playback-service/internal/transmux"
	"github.com/rs/zerolog"
)

// PlaybackUseCase orchestrates video playback operations
type PlaybackUseCase struct {
	sourceDetector *SourceDetector
	transmuxer     *transmux.FFmpegTransmuxer
	segmentCache   *cache.SegmentCache
	storageClient  StorageClient
	workDir        string
	hlsBaseURL     string
	logger         zerolog.Logger
}

func NewPlaybackUseCase(
	sourceDetector *SourceDetector,
	transmuxer *transmux.FFmpegTransmuxer,
	segmentCache *cache.SegmentCache,
	storageClient StorageClient,
	workDir string,
	hlsBaseURL string,
	logger zerolog.Logger,
) *PlaybackUseCase {
	return &PlaybackUseCase{
		sourceDetector: sourceDetector,
		transmuxer:     transmuxer,
		segmentCache:   segmentCache,
		storageClient:  storageClient,
		workDir:        workDir,
		hlsBaseURL:     hlsBaseURL,
		logger:         logger,
	}
}

// RequestPlayback handles a playback request
func (p *PlaybackUseCase) RequestPlayback(ctx context.Context, req domain.PlaybackRequest) (*domain.PlaybackResponse, error) {
	p.logger.Info().
		Str("camera_id", req.CameraID).
		Time("start_time", req.StartTime).
		Time("end_time", req.EndTime).
		Str("format", req.Format).
		Msg("Processing playback request")

	// Validate time range
	if req.EndTime.Before(req.StartTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	duration := req.EndTime.Sub(req.StartTime)
	if duration > 24*time.Hour {
		return nil, fmt.Errorf("playback duration cannot exceed 24 hours")
	}

	// 1. Detect best source (MinIO vs Milestone)
	sourceResult, err := p.sourceDetector.DetectSource(ctx, req.CameraID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to detect source: %w", err)
	}

	if !sourceResult.Available {
		return nil, fmt.Errorf("no recordings available: %s", sourceResult.Reason)
	}

	// 2. Handle based on source
	switch sourceResult.Source {
	case domain.PlaybackSourceLocal:
		return p.handleLocalPlayback(ctx, req, sourceResult)
	case domain.PlaybackSourceMilestone:
		return p.handleMilestonePlayback(ctx, req, sourceResult)
	default:
		return nil, fmt.Errorf("unknown source: %s", sourceResult.Source)
	}
}

// handleLocalPlayback handles playback from local MinIO storage
func (p *PlaybackUseCase) handleLocalPlayback(
	ctx context.Context,
	req domain.PlaybackRequest,
	sourceResult *domain.SourceDetectionResult,
) (*domain.PlaybackResponse, error) {
	sessionID := uuid.New().String()

	p.logger.Info().
		Str("session_id", sessionID).
		Str("camera_id", req.CameraID).
		Msg("Starting local playback")

	// 1. Get segments from MinIO
	segments, err := p.storageClient.ListSegments(ctx, req.CameraID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}

	// 2. Download segments to temp directory
	tempDir := filepath.Join(p.workDir, "temp", sessionID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup temp files

	var downloadedPaths []string
	for i, segment := range segments {
		outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", i))

		// Check cache first
		if cachedPath, found := p.segmentCache.Get(segment.ID); found {
			p.logger.Debug().Str("segment_id", segment.ID).Msg("Using cached segment")
			// Copy from cache
			if err := p.copyFile(cachedPath, outputPath); err == nil {
				downloadedPaths = append(downloadedPaths, outputPath)
				continue
			}
		}

		// Download from MinIO
		downloadURL, err := p.storageClient.GetSegmentDownloadURL(ctx, segment.ID, 3600)
		if err != nil {
			return nil, fmt.Errorf("failed to get download URL: %w", err)
		}

		// TODO: Actually download the file from the URL to outputPath
		// For now, we'll skip this and assume segments are already available
		_ = downloadURL

		downloadedPaths = append(downloadedPaths, outputPath)

		// Add to cache
		if _, err := p.segmentCache.Put(segment.ID, outputPath); err != nil {
			p.logger.Warn().Err(err).Msg("Failed to cache segment")
		}
	}

	// 3. Transmux to HLS
	hlsDir := filepath.Join(p.workDir, "hls", sessionID)
	transmuxConfig := transmux.TransmuxConfig{
		OutputDir:       hlsDir,
		Format:          string(req.Format),
		SegmentDuration: 6, // 6 second segments
	}

	var result *transmux.TransmuxResult
	if len(downloadedPaths) == 1 {
		transmuxConfig.InputPath = downloadedPaths[0]
		result, err = p.transmuxer.TransmuxToHLS(ctx, transmuxConfig)
	} else {
		result, err = p.transmuxer.ConcatenateAndTransmux(ctx, downloadedPaths, transmuxConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to transmux: %w", err)
	}

	// 4. Generate manifest URL
	manifestURL := fmt.Sprintf("%s/%s/playlist.m3u8", p.hlsBaseURL, sessionID)

	// 5. Create playback session
	session := &domain.PlaybackSession{
		ID:           sessionID,
		CameraID:     req.CameraID,
		UserID:       "", // TODO: Get from context/auth
		Source:       domain.PlaybackSourceLocal,
		Format:       req.Format,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		ManifestPath: result.ManifestPath,
		SegmentPaths: result.SegmentPaths,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		LastAccessed: time.Now(),
	}

	p.logger.Info().
		Str("session_id", sessionID).
		Str("manifest_url", manifestURL).
		Int("segment_count", result.SegmentCount).
		Msg("Playback session created")

	// 6. Return response
	return &domain.PlaybackResponse{
		SessionID:  sessionID,
		CameraID:   req.CameraID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Format:     req.Format,
		URL:        manifestURL,
		ExpiresAt:  session.ExpiresAt,
		SegmentIDs: p.getSegmentIDs(segments),
	}, nil
}

// handleMilestonePlayback handles playback from external Milestone VMS
func (p *PlaybackUseCase) handleMilestonePlayback(
	ctx context.Context,
	req domain.PlaybackRequest,
	sourceResult *domain.SourceDetectionResult,
) (*domain.PlaybackResponse, error) {
	// TODO: Implement Milestone playback
	// This would involve:
	// 1. Get stream URL from Milestone
	// 2. Transcode stream to HLS
	// 3. Return manifest URL

	return nil, fmt.Errorf("milestone playback not implemented yet")
}

// CreateExport creates a downloadable video export
func (p *PlaybackUseCase) CreateExport(ctx context.Context, req domain.ExportRequest) (*domain.ExportResponse, error) {
	exportID := uuid.New().String()

	p.logger.Info().
		Str("export_id", exportID).
		Str("camera_id", req.CameraID).
		Time("start_time", req.StartTime).
		Time("end_time", req.EndTime).
		Msg("Creating export")

	// 1. Get segments
	segments, err := p.storageClient.ListSegments(ctx, req.CameraID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}

	// 2. Download segments
	tempDir := filepath.Join(p.workDir, "exports", exportID, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	var downloadedPaths []string
	for i, segment := range segments {
		outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", i))

		// Get download URL
		downloadURL, err := p.storageClient.GetSegmentDownloadURL(ctx, segment.ID, 3600)
		if err != nil {
			return nil, fmt.Errorf("failed to get download URL: %w", err)
		}

		// TODO: Actually download the file from the URL to outputPath
		// For now, we'll skip this and assume segments are already available
		_ = downloadURL

		downloadedPaths = append(downloadedPaths, outputPath)
	}

	// 3. Transmux to MP4
	outputPath := filepath.Join(p.workDir, "exports", exportID, fmt.Sprintf("export_%s.mp4", exportID))

	// Concatenate if multiple segments
	var inputPath string
	if len(downloadedPaths) == 1 {
		inputPath = downloadedPaths[0]
	} else {
		// Create concat file
		concatFile := filepath.Join(tempDir, "concat.txt")
		var concatList string
		for _, path := range downloadedPaths {
			concatList += fmt.Sprintf("file '%s'\n", path)
		}
		if err := os.WriteFile(concatFile, []byte(concatList), 0644); err != nil {
			return nil, fmt.Errorf("failed to create concat file: %w", err)
		}
		inputPath = concatFile
	}

	if err := p.transmuxer.TransmuxToMP4(ctx, inputPath, outputPath); err != nil {
		return nil, fmt.Errorf("failed to transmux to MP4: %w", err)
	}

	// 4. Generate download URL
	downloadURL := fmt.Sprintf("%s/exports/%s/export_%s.mp4", p.hlsBaseURL, exportID, exportID)

	// 7. Clean up temp files
	os.RemoveAll(tempDir)

	return &domain.ExportResponse{
		ExportID:    exportID,
		CameraID:    req.CameraID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Format:      req.Format,
		Status:      "ready",
		DownloadURL: downloadURL,
		CreatedAt:   time.Now(),
	}, nil
}

// GetCacheStats returns cache statistics
func (p *PlaybackUseCase) GetCacheStats() cache.CacheStats {
	return p.segmentCache.GetStats()
}

// getSegmentIDs extracts segment IDs from segments
func (p *PlaybackUseCase) getSegmentIDs(segments []*domain.Segment) []string {
	ids := make([]string, len(segments))
	for i, seg := range segments {
		ids[i] = seg.ID
	}
	return ids
}

// copyFile copies a file from src to dst
func (p *PlaybackUseCase) copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, sourceData, 0644)
}
