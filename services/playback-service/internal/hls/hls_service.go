package hls

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rta/cctv/playback-service/internal/client"
	"github.com/rta/cctv/playback-service/internal/domain"
	"github.com/rta/cctv/playback-service/internal/repository"
	"github.com/rs/zerolog"
)

// HLSService handles HLS playlist generation and segment transmuxing
type HLSService struct {
	storageClient *client.StorageClient
	cacheRepo     *repository.CacheRepository
	workDir       string
	logger        zerolog.Logger
}

// NewHLSService creates a new HLS service
func NewHLSService(
	storageClient *client.StorageClient,
	cacheRepo *repository.CacheRepository,
	workDir string,
	logger zerolog.Logger,
) *HLSService {
	return &HLSService{
		storageClient: storageClient,
		cacheRepo:     cacheRepo,
		workDir:       workDir,
		logger:        logger,
	}
}

// CreatePlaybackSession creates a new HLS playback session
func (s *HLSService) CreatePlaybackSession(
	ctx context.Context,
	cameraID string,
	startTime, endTime time.Time,
) (*domain.PlaybackResponse, error) {
	sessionID := uuid.New().String()

	// Fetch segments from storage service
	segments, err := s.storageClient.ListSegments(ctx, cameraID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments found for time range")
	}

	// Sort segments by start time
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].StartTime.Before(segments[j].StartTime)
	})

	// Generate HLS manifest
	manifest, err := s.generateManifest(ctx, sessionID, segments)
	if err != nil {
		return nil, fmt.Errorf("failed to generate manifest: %w", err)
	}

	// Cache manifest (1 hour TTL)
	if err := s.cacheRepo.SetHLSManifest(ctx, sessionID, manifest, time.Hour); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to cache manifest")
	}

	// Extract segment IDs
	segmentIDs := make([]string, len(segments))
	for i, seg := range segments {
		segmentIDs[i] = seg.ID
	}

	response := &domain.PlaybackResponse{
		SessionID:  sessionID,
		CameraID:   cameraID,
		StartTime:  startTime,
		EndTime:    endTime,
		Format:     "hls",
		URL:        fmt.Sprintf("/api/v1/playback/sessions/%s/playlist.m3u8", sessionID),
		ExpiresAt:  time.Now().Add(time.Hour),
		SegmentIDs: segmentIDs,
	}

	// Cache session (1 hour TTL)
	if err := s.cacheRepo.SetSession(ctx, sessionID, response, time.Hour); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to cache session")
	}

	s.logger.Info().
		Str("session_id", sessionID).
		Str("camera_id", cameraID).
		Int("segments", len(segments)).
		Msg("Created HLS playback session")

	return response, nil
}

// GetManifest retrieves the HLS manifest for a session
func (s *HLSService) GetManifest(ctx context.Context, sessionID string) (string, error) {
	// Try to get from cache first
	manifest, err := s.cacheRepo.GetHLSManifest(ctx, sessionID)
	if err == nil {
		return manifest, nil
	}

	// If not in cache, retrieve session and regenerate
	var session domain.PlaybackResponse
	if err := s.cacheRepo.GetSession(ctx, sessionID, &session); err != nil {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	// Fetch segments again
	segments, err := s.storageClient.ListSegments(ctx, session.CameraID, session.StartTime, session.EndTime)
	if err != nil {
		return "", fmt.Errorf("failed to list segments: %w", err)
	}

	return s.generateManifest(ctx, sessionID, segments)
}

// generateManifest creates an HLS m3u8 playlist
func (s *HLSService) generateManifest(ctx context.Context, sessionID string, segments []*domain.Segment) (string, error) {
	var buf bytes.Buffer

	// Write HLS header
	buf.WriteString("#EXTM3U\n")
	buf.WriteString("#EXT-X-VERSION:3\n")
	buf.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")

	// Calculate target duration (max segment duration)
	maxDuration := 0
	for _, seg := range segments {
		if seg.DurationSeconds > maxDuration {
			maxDuration = seg.DurationSeconds
		}
	}
	buf.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", maxDuration))
	buf.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	// Write segment entries
	for i, seg := range segments {
		buf.WriteString(fmt.Sprintf("#EXTINF:%.3f,\n", float64(seg.DurationSeconds)))
		buf.WriteString(fmt.Sprintf("/api/v1/playback/sessions/%s/segment/%d.ts\n", sessionID, i))

		// Cache segment index to ID mapping
		s.cacheRepo.SetHLSSegment(ctx, sessionID, i, seg.ID, time.Hour)
	}

	buf.WriteString("#EXT-X-ENDLIST\n")

	return buf.String(), nil
}

// GetSegment retrieves and transmuxes a video segment
func (s *HLSService) GetSegment(ctx context.Context, sessionID string, segmentIndex int) (io.ReadCloser, error) {
	// Get segment ID from cache
	segmentID, err := s.cacheRepo.GetHLSSegment(ctx, sessionID, segmentIndex)
	if err != nil {
		return nil, fmt.Errorf("segment not found: %w", err)
	}

	// Get download URL from storage service (30 minute expiry)
	downloadURL, err := s.storageClient.GetSegmentDownloadURL(ctx, segmentID, 1800)
	if err != nil {
		return nil, fmt.Errorf("failed to get download URL: %w", err)
	}

	// Download segment to temp file
	tempFile, err := s.downloadSegment(ctx, downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download segment: %w", err)
	}
	defer os.Remove(tempFile)

	// Transmux to HLS-compatible TS (if needed)
	outputFile, err := s.transmuxSegment(ctx, tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to transmux segment: %w", err)
	}

	// Open and return the file
	file, err := os.Open(outputFile)
	if err != nil {
		os.Remove(outputFile)
		return nil, fmt.Errorf("failed to open segment: %w", err)
	}

	// Wrap in a ReadCloser that also deletes the file when closed
	return &fileReadCloser{File: file, path: outputFile}, nil
}

// downloadSegment downloads a segment from a URL to a temp file
func (s *HLSService) downloadSegment(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tempFile := filepath.Join(s.workDir, fmt.Sprintf("download-%s.ts", uuid.New().String()))
	file, err := os.Create(tempFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(tempFile)
		return "", err
	}

	return tempFile, nil
}

// transmuxSegment converts a video segment to HLS-compatible TS format
func (s *HLSService) transmuxSegment(ctx context.Context, inputPath string) (string, error) {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "-hls.ts"

	// Use FFmpeg to transmux (no transcoding, just remux)
	args := []string{
		"-i", inputPath,
		"-c", "copy", // Copy codecs (no transcoding)
		"-bsf:v", "h264_mp4toannexb", // Convert to Annex B format for HLS
		"-f", "mpegts",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.logger.Error().
			Err(err).
			Str("stderr", stderr.String()).
			Msg("FFmpeg transmux failed")
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	s.logger.Debug().Str("output", outputPath).Msg("Segment transmuxed")
	return outputPath, nil
}

// CleanupSession removes session data and temp files
func (s *HLSService) CleanupSession(ctx context.Context, sessionID string) error {
	return s.cacheRepo.DeleteSession(ctx, sessionID)
}

// fileReadCloser wraps os.File and deletes the file when closed
type fileReadCloser struct {
	*os.File
	path string
}

func (f *fileReadCloser) Close() error {
	f.File.Close()
	os.Remove(f.path)
	return nil
}
