package transmux

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// FFmpegTransmuxer handles H.264 to HLS/DASH conversion
type FFmpegTransmuxer struct {
	workDir       string
	segmentDuration int    // seconds per segment
	logger        zerolog.Logger
}

// TransmuxConfig contains configuration for transmuxing
type TransmuxConfig struct {
	InputPath       string
	OutputDir       string
	Format          string  // hls or dash
	SegmentDuration int     // seconds
	Quality         string  // high, medium, low
}

// TransmuxResult contains the result of transmuxing
type TransmuxResult struct {
	ManifestPath string
	SegmentPaths []string
	Duration     float64
	SegmentCount int
}

func NewFFmpegTransmuxer(workDir string, logger zerolog.Logger) *FFmpegTransmuxer {
	return &FFmpegTransmuxer{
		workDir:         workDir,
		segmentDuration: 6, // 6 second segments (standard HLS)
		logger:          logger,
	}
}

// TransmuxToHLS converts H.264 video to HLS format
func (t *FFmpegTransmuxer) TransmuxToHLS(ctx context.Context, config TransmuxConfig) (*TransmuxResult, error) {
	sessionID := uuid.New().String()
	outputDir := filepath.Join(config.OutputDir, sessionID)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	manifestPath := filepath.Join(outputDir, "playlist.m3u8")
	segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")

	t.logger.Info().
		Str("session_id", sessionID).
		Str("input", config.InputPath).
		Str("output", manifestPath).
		Msg("Starting HLS transmux")

	// Build FFmpeg command
	// -c copy: Stream copy (no re-encoding, very fast)
	// -movflags +faststart: Enable fast start for web playback
	// -hls_time: Segment duration
	// -hls_list_size: Number of segments in playlist (0 = all)
	// -hls_flags: append_list+delete_segments for live-like behavior
	args := []string{
		"-i", config.InputPath,
		"-c", "copy",              // Stream copy - NO re-encoding
		"-movflags", "+faststart", // Fast start for web
		"-f", "hls",               // HLS format
		"-hls_time", fmt.Sprintf("%d", config.SegmentDuration),
		"-hls_list_size", "0",     // Keep all segments in playlist
		"-hls_segment_filename", segmentPattern,
		"-hls_flags", "independent_segments", // Each segment is independent
		manifestPath,
	}

	// Add quality-specific settings if needed
	if config.Quality == "low" {
		// For low quality, we might want to limit bitrate
		args = append([]string{
			"-b:v", "500k",
			"-maxrate", "500k",
			"-bufsize", "1000k",
		}, args...)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.logger.Error().
			Err(err).
			Str("output", string(output)).
			Msg("FFmpeg transmux failed")
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	t.logger.Info().
		Str("session_id", sessionID).
		Msg("HLS transmux completed")

	// Get segment files
	segments, err := t.getSegmentFiles(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment files: %w", err)
	}

	// Get video duration
	duration, err := t.getVideoDuration(config.InputPath)
	if err != nil {
		t.logger.Warn().Err(err).Msg("Failed to get video duration")
		duration = 0
	}

	return &TransmuxResult{
		ManifestPath: manifestPath,
		SegmentPaths: segments,
		Duration:     duration,
		SegmentCount: len(segments),
	}, nil
}

// TransmuxToMP4 converts video to MP4 format for direct download
func (t *FFmpegTransmuxer) TransmuxToMP4(ctx context.Context, inputPath, outputPath string) error {
	t.logger.Info().
		Str("input", inputPath).
		Str("output", outputPath).
		Msg("Starting MP4 transmux")

	// Create output directory
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// FFmpeg command for MP4
	// -c copy: Stream copy (no re-encoding)
	// -movflags +faststart: Move moov atom to beginning for web playback
	args := []string{
		"-i", inputPath,
		"-c", "copy",
		"-movflags", "+faststart",
		"-f", "mp4",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.logger.Error().
			Err(err).
			Str("output", string(output)).
			Msg("FFmpeg MP4 transmux failed")
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	t.logger.Info().
		Str("output", outputPath).
		Msg("MP4 transmux completed")

	return nil
}

// ConcatenateAndTransmux concatenates multiple video files and transmuxes to HLS
func (t *FFmpegTransmuxer) ConcatenateAndTransmux(ctx context.Context, inputPaths []string, config TransmuxConfig) (*TransmuxResult, error) {
	if len(inputPaths) == 0 {
		return nil, fmt.Errorf("no input files provided")
	}

	// If only one file, just transmux it
	if len(inputPaths) == 1 {
		config.InputPath = inputPaths[0]
		return t.TransmuxToHLS(ctx, config)
	}

	t.logger.Info().
		Int("file_count", len(inputPaths)).
		Msg("Concatenating videos before transmux")

	// Create concat list file
	concatFile := filepath.Join(t.workDir, fmt.Sprintf("concat_%s.txt", uuid.New().String()))
	defer os.Remove(concatFile)

	var concatList strings.Builder
	for _, path := range inputPaths {
		concatList.WriteString(fmt.Sprintf("file '%s'\n", path))
	}

	if err := os.WriteFile(concatFile, []byte(concatList.String()), 0644); err != nil {
		return nil, fmt.Errorf("failed to create concat file: %w", err)
	}

	// Create temporary concatenated output
	tempOutput := filepath.Join(t.workDir, fmt.Sprintf("concat_%s.mp4", uuid.New().String()))
	defer os.Remove(tempOutput)

	// Concatenate using FFmpeg
	// -f concat: Concatenation demuxer
	// -safe 0: Allow unsafe file names
	// -c copy: Stream copy (no re-encoding)
	concatArgs := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		tempOutput,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", concatArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.logger.Error().
			Err(err).
			Str("output", string(output)).
			Msg("FFmpeg concatenation failed")
		return nil, fmt.Errorf("concatenation failed: %w", err)
	}

	// Now transmux the concatenated file
	config.InputPath = tempOutput
	return t.TransmuxToHLS(ctx, config)
}

// getSegmentFiles retrieves all segment files from output directory
func (t *FFmpegTransmuxer) getSegmentFiles(outputDir string) ([]string, error) {
	var segments []string

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// HLS segments are .ts files
		if strings.HasSuffix(entry.Name(), ".ts") {
			segments = append(segments, filepath.Join(outputDir, entry.Name()))
		}
	}

	return segments, nil
}

// getVideoDuration gets the duration of a video file in seconds
func (t *FFmpegTransmuxer) getVideoDuration(inputPath string) (float64, error) {
	// Use ffprobe to get duration
	// -v error: Only show errors
	// -show_entries format=duration: Show only duration
	// -of default=noprint_wrappers=1:nokey=1: Output only the value
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	}

	cmd := exec.Command("ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var duration float64
	_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &duration)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

// CleanupSession removes all files associated with a session
func (t *FFmpegTransmuxer) CleanupSession(sessionID string) error {
	sessionDir := filepath.Join(t.workDir, sessionID)
	return os.RemoveAll(sessionDir)
}

// GetSegmentDuration returns the configured segment duration
func (t *FFmpegTransmuxer) GetSegmentDuration() int {
	return t.segmentDuration
}
