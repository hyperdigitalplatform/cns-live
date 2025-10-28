package usecase

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rta/cctv/playback-service/internal/client"
	"github.com/rs/zerolog"
)

// MilestonePlaybackUsecase handles playback operations for Milestone recordings
type MilestonePlaybackUsecase struct {
	milestoneClient *client.MilestoneClient
	cache           PlaybackCache
	logger          zerolog.Logger
}

// PlaybackCache interface for caching query results
type PlaybackCache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
}

// NewMilestonePlaybackUsecase creates a new playback usecase
func NewMilestonePlaybackUsecase(milestoneClient *client.MilestoneClient, cache PlaybackCache, logger zerolog.Logger) *MilestonePlaybackUsecase {
	return &MilestonePlaybackUsecase{
		milestoneClient: milestoneClient,
		cache:           cache,
		logger:          logger,
	}
}

// QueryRequest represents a request to query recordings
type QueryRequest struct {
	CameraID  string    `json:"cameraId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// TimelineData represents recording data for timeline visualization
type TimelineData struct {
	CameraID              string               `json:"cameraId"`
	QueryRange            TimeRange            `json:"queryRange"`
	Sequences             []RecordingSequence  `json:"sequences"`
	Gaps                  []RecordingGap       `json:"gaps"`
	TotalRecordingSeconds int                  `json:"totalRecordingSeconds"`
	TotalGapSeconds       int                  `json:"totalGapSeconds"`
	Coverage              float64              `json:"coverage"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// RecordingSequence represents a continuous recording segment
type RecordingSequence struct {
	SequenceID      string    `json:"sequenceId"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	DurationSeconds int       `json:"durationSeconds"`
	Available       bool      `json:"available"`
	SizeBytes       int64     `json:"sizeBytes"`
}

// RecordingGap represents a gap in recordings
type RecordingGap struct {
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	DurationSeconds int       `json:"durationSeconds"`
}

// QueryRecordings queries available recordings for a time range
func (u *MilestonePlaybackUsecase) QueryRecordings(ctx context.Context, req QueryRequest) (*TimelineData, error) {
	u.logger.Info().
		Str("camera_id", req.CameraID).
		Time("start_time", req.StartTime).
		Time("end_time", req.EndTime).
		Msg("Querying recordings")

	// Generate cache key
	cacheKey := u.generateCacheKey(req)

	// Check cache
	if cachedData, found := u.cache.Get(cacheKey); found {
		u.logger.Debug().Str("camera_id", req.CameraID).Msg("Cache hit for recording query")

		var timelineData TimelineData
		if err := json.Unmarshal(cachedData, &timelineData); err == nil {
			return &timelineData, nil
		}
	}

	// Query Milestone for sequences
	sequenceReq := client.SequenceQueryRequest{
		CameraID:  req.CameraID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	sequenceList, err := u.milestoneClient.QuerySequences(ctx, sequenceReq)
	if err != nil {
		u.logger.Error().
			Err(err).
			Str("camera_id", req.CameraID).
			Msg("Failed to query sequences from Milestone")
		return nil, fmt.Errorf("failed to query sequences: %w", err)
	}

	// Convert to timeline data
	sequences := make([]RecordingSequence, len(sequenceList.Sequences))
	totalRecordingSeconds := 0

	for i, seq := range sequenceList.Sequences {
		sequences[i] = RecordingSequence{
			SequenceID:      seq.SequenceID,
			StartTime:       seq.StartTime,
			EndTime:         seq.EndTime,
			DurationSeconds: seq.DurationSeconds,
			Available:       seq.Available,
			SizeBytes:       seq.SizeBytes,
		}
		totalRecordingSeconds += seq.DurationSeconds
	}

	gaps := make([]RecordingGap, len(sequenceList.Gaps))
	totalGapSeconds := 0

	for i, gap := range sequenceList.Gaps {
		gaps[i] = RecordingGap{
			StartTime:       gap.StartTime,
			EndTime:         gap.EndTime,
			DurationSeconds: gap.DurationSeconds,
		}
		totalGapSeconds += gap.DurationSeconds
	}

	// Calculate coverage
	totalDuration := int(req.EndTime.Sub(req.StartTime).Seconds())
	coverage := 0.0
	if totalDuration > 0 {
		coverage = float64(totalRecordingSeconds) / float64(totalDuration)
	}

	timelineData := &TimelineData{
		CameraID: req.CameraID,
		QueryRange: TimeRange{
			Start: req.StartTime,
			End:   req.EndTime,
		},
		Sequences:             sequences,
		Gaps:                  gaps,
		TotalRecordingSeconds: totalRecordingSeconds,
		TotalGapSeconds:       totalGapSeconds,
		Coverage:              coverage,
	}

	// Cache the result
	if cachedData, err := json.Marshal(timelineData); err == nil {
		u.cache.Set(cacheKey, cachedData, 5*time.Minute) // Cache for 5 minutes
	}

	u.logger.Info().
		Str("camera_id", req.CameraID).
		Int("sequences", len(sequences)).
		Int("gaps", len(gaps)).
		Float64("coverage", coverage).
		Msg("Recording query completed")

	return timelineData, nil
}

// GetTimelineData retrieves timeline data with aggregation for UI
func (u *MilestonePlaybackUsecase) GetTimelineData(ctx context.Context, req QueryRequest, resolution string) (*AggregatedTimelineData, error) {
	// First get raw timeline data
	timelineData, err := u.QueryRecordings(ctx, req)
	if err != nil {
		return nil, err
	}

	// Aggregate based on resolution (minute, hour, day)
	var bucketSize time.Duration
	switch resolution {
	case "minute":
		bucketSize = 1 * time.Minute
	case "hour":
		bucketSize = 1 * time.Hour
	case "day":
		bucketSize = 24 * time.Hour
	default:
		bucketSize = 1 * time.Hour // Default to hour
	}

	// Create time buckets
	buckets := make(map[time.Time]*TimelineBucket)
	current := req.StartTime.Truncate(bucketSize)

	for current.Before(req.EndTime) {
		buckets[current] = &TimelineBucket{
			Timestamp:    current,
			HasRecording: false,
			SegmentCount: 0,
		}
		current = current.Add(bucketSize)
	}

	// Fill buckets with sequence data
	for _, seq := range timelineData.Sequences {
		bucketTime := seq.StartTime.Truncate(bucketSize)
		if bucket, exists := buckets[bucketTime]; exists {
			bucket.HasRecording = true
			bucket.SegmentCount++
		}
	}

	// Convert to slice
	timeline := make([]TimelineBucket, 0, len(buckets))
	for _, bucket := range buckets {
		timeline = append(timeline, *bucket)
	}

	return &AggregatedTimelineData{
		CameraID:   req.CameraID,
		Resolution: resolution,
		Timeline:   timeline,
	}, nil
}

// AggregatedTimelineData represents aggregated timeline for UI display
type AggregatedTimelineData struct {
	CameraID   string           `json:"cameraId"`
	Resolution string           `json:"resolution"`
	Timeline   []TimelineBucket `json:"timeline"`
}

// TimelineBucket represents a time bucket in the timeline
type TimelineBucket struct {
	Timestamp    time.Time `json:"timestamp"`
	HasRecording bool      `json:"hasRecording"`
	SegmentCount int       `json:"segmentCount"`
}

// PlaybackRequest represents a request to start playback
type PlaybackRequest struct {
	CameraID  string    `json:"cameraId"`
	Timestamp time.Time `json:"timestamp"`
	Speed     float64   `json:"speed"`
	Format    string    `json:"format"`
}

// PlaybackSession represents an active playback session
type PlaybackSession struct {
	SessionID     string    `json:"sessionId"`
	CameraID      string    `json:"cameraId"`
	StartTime     time.Time `json:"startTime"`
	CurrentTime   time.Time `json:"currentTime"`
	Speed         float64   `json:"speed"`
	Format        string    `json:"format"`
	StreamURL     string    `json:"streamUrl"`
}

// StartPlayback initiates playback of recorded video
func (u *MilestonePlaybackUsecase) StartPlayback(ctx context.Context, req PlaybackRequest) (*PlaybackSession, error) {
	u.logger.Info().
		Str("camera_id", req.CameraID).
		Time("timestamp", req.Timestamp).
		Float64("speed", req.Speed).
		Str("format", req.Format).
		Msg("Starting playback")

	// Validate speed
	if req.Speed < -8 || req.Speed > 8 || req.Speed == 0 {
		return nil, fmt.Errorf("invalid playback speed: %f (must be between -8 and 8, excluding 0)", req.Speed)
	}

	// Default format
	if req.Format == "" {
		req.Format = "hls"
	}

	// Get video stream from Milestone
	streamReq := client.VideoStreamRequest{
		CameraID:  req.CameraID,
		Timestamp: req.Timestamp,
		Speed:     req.Speed,
		Format:    req.Format,
	}

	stream, err := u.milestoneClient.GetVideoStream(ctx, streamReq)
	if err != nil {
		u.logger.Error().
			Err(err).
			Str("camera_id", req.CameraID).
			Msg("Failed to get video stream from Milestone")
		return nil, fmt.Errorf("failed to get video stream: %w", err)
	}
	defer stream.Close()

	// TODO: Implement FFmpeg transmuxing to HLS
	// For now, return a placeholder stream URL

	sessionID := fmt.Sprintf("playback_%s_%d", req.CameraID, time.Now().Unix())
	streamURL := fmt.Sprintf("/playback/stream/%s.m3u8", sessionID)

	session := &PlaybackSession{
		SessionID:   sessionID,
		CameraID:    req.CameraID,
		StartTime:   req.Timestamp,
		CurrentTime: req.Timestamp,
		Speed:       req.Speed,
		Format:      req.Format,
		StreamURL:   streamURL,
	}

	u.logger.Info().
		Str("session_id", sessionID).
		Str("camera_id", req.CameraID).
		Msg("Playback session started")

	return session, nil
}

// GetSnapshot retrieves a snapshot at a specific timestamp
func (u *MilestonePlaybackUsecase) GetSnapshot(ctx context.Context, cameraID string, timestamp time.Time) ([]byte, error) {
	u.logger.Info().
		Str("camera_id", cameraID).
		Time("timestamp", timestamp).
		Msg("Getting snapshot")

	snapshot, err := u.milestoneClient.GetSnapshot(ctx, cameraID, timestamp)
	if err != nil {
		u.logger.Error().
			Err(err).
			Str("camera_id", cameraID).
			Msg("Failed to get snapshot from Milestone")
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	u.logger.Info().
		Str("camera_id", cameraID).
		Int("size_bytes", len(snapshot)).
		Msg("Snapshot retrieved")

	return snapshot, nil
}

// generateCacheKey generates a cache key for query results
func (u *MilestonePlaybackUsecase) generateCacheKey(req QueryRequest) string {
	data := fmt.Sprintf("%s:%s:%s",
		req.CameraID,
		req.StartTime.Format(time.RFC3339),
		req.EndTime.Format(time.RFC3339))

	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// MilestoneRecordingMetadata represents metadata from Milestone
type MilestoneRecordingMetadata struct {
	Available    bool      `json:"available"`
	SegmentCount int       `json:"segmentCount"`
	TotalSize    int64     `json:"totalSize"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
}
