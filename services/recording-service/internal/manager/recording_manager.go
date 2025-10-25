package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rta/cctv/recording-service/internal/domain"
	"github.com/rta/cctv/recording-service/internal/ffmpeg"
	"github.com/rs/zerolog"
)

// RecordingManager manages multiple camera recordings
type RecordingManager struct {
	recordings      map[string]*domain.Recording
	recorders       map[string]*ffmpeg.Recorder
	mu              sync.RWMutex
	outputDir       string
	segmentSeconds  int
	storageURL      string
	vmsURL          string
	streamCounterURL string
	logger          zerolog.Logger
	httpClient      *http.Client
}

// NewRecordingManager creates a new recording manager
func NewRecordingManager(
	outputDir string,
	segmentSeconds int,
	storageURL string,
	vmsURL string,
	streamCounterURL string,
	logger zerolog.Logger,
) *RecordingManager {
	return &RecordingManager{
		recordings:       make(map[string]*domain.Recording),
		recorders:        make(map[string]*ffmpeg.Recorder),
		outputDir:        outputDir,
		segmentSeconds:   segmentSeconds,
		storageURL:       storageURL,
		vmsURL:           vmsURL,
		streamCounterURL: streamCounterURL,
		logger:           logger,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
	}
}

// StartRecording starts recording for a camera
func (m *RecordingManager) StartRecording(ctx context.Context, cameraID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already recording
	if rec, exists := m.recordings[cameraID]; exists {
		if rec.Status == domain.StatusRecording {
			return fmt.Errorf("camera %s is already recording", cameraID)
		}
	}

	// Get camera details from VMS Service
	camera, err := m.getCameraDetails(ctx, cameraID)
	if err != nil {
		return fmt.Errorf("failed to get camera details: %w", err)
	}

	// Get RTSP URL from VMS Service
	rtspURL, err := m.getRTSPURL(ctx, cameraID)
	if err != nil {
		return fmt.Errorf("failed to get RTSP URL: %w", err)
	}

	// Reserve stream quota
	reservationID, err := m.reserveQuota(ctx, cameraID, camera["source"].(string))
	if err != nil {
		return fmt.Errorf("failed to reserve quota: %w", err)
	}

	// Create recording record
	recording := &domain.Recording{
		CameraID:      cameraID,
		CameraName:    camera["name"].(string),
		RTSPURL:       rtspURL,
		Status:        domain.StatusStarting,
		StartedAt:     time.Now(),
		ReservationID: reservationID,
	}

	m.recordings[cameraID] = recording

	// Start FFmpeg recorder
	recorder := ffmpeg.NewRecorder(
		cameraID,
		rtspURL,
		m.outputDir,
		m.segmentSeconds,
		m.logger,
		m.onSegmentComplete,
	)

	if err := recorder.Start(ctx); err != nil {
		recording.Status = domain.StatusError
		recording.Error = err.Error()
		return fmt.Errorf("failed to start recorder: %w", err)
	}

	m.recorders[cameraID] = recorder
	recording.Status = domain.StatusRecording

	m.logger.Info().
		Str("camera_id", cameraID).
		Str("camera_name", recording.CameraName).
		Msg("Recording started")

	// Start heartbeat goroutine
	go m.sendHeartbeats(ctx, cameraID, reservationID)

	return nil
}

// StopRecording stops recording for a camera
func (m *RecordingManager) StopRecording(cameraID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	recording, exists := m.recordings[cameraID]
	if !exists {
		return fmt.Errorf("camera %s is not recording", cameraID)
	}

	recording.Status = domain.StatusStopping

	// Stop FFmpeg recorder
	if recorder, exists := m.recorders[cameraID]; exists {
		if err := recorder.Stop(); err != nil {
			m.logger.Warn().Err(err).Str("camera_id", cameraID).Msg("Error stopping recorder")
		}
		delete(m.recorders, cameraID)
	}

	// Release quota
	if recording.ReservationID != "" {
		if err := m.releaseQuota(context.Background(), recording.ReservationID); err != nil {
			m.logger.Warn().Err(err).Msg("Failed to release quota")
		}
	}

	recording.Status = domain.StatusStopped
	delete(m.recordings, cameraID)

	m.logger.Info().Str("camera_id", cameraID).Msg("Recording stopped")

	return nil
}

// GetRecording gets recording status for a camera
func (m *RecordingManager) GetRecording(cameraID string) (*domain.Recording, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	recording, exists := m.recordings[cameraID]
	if !exists {
		return nil, fmt.Errorf("camera %s is not recording", cameraID)
	}

	return recording, nil
}

// GetAllRecordings gets all active recordings
func (m *RecordingManager) GetAllRecordings() map[string]*domain.Recording {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	recordings := make(map[string]*domain.Recording)
	for k, v := range m.recordings {
		recordings[k] = v
	}

	return recordings
}

// onSegmentComplete is called when a segment is completed
func (m *RecordingManager) onSegmentComplete(segmentPath string, startTime time.Time, endTime time.Time) {
	m.logger.Info().
		Str("segment_path", segmentPath).
		Time("start_time", startTime).
		Time("end_time", endTime).
		Msg("Segment completed, uploading to storage")

	// Get file info
	fileInfo, err := os.Stat(segmentPath)
	if err != nil {
		m.logger.Error().Err(err).Str("segment_path", segmentPath).Msg("Failed to stat segment file")
		return
	}

	// Extract camera ID from path
	cameraID := filepath.Base(filepath.Dir(segmentPath))

	// Generate thumbnail (every 300 seconds = 5 minutes)
	thumbnailPath := ""
	if int(time.Since(startTime).Seconds())%300 == 0 {
		thumbnailPath = segmentPath[:len(segmentPath)-3] + ".jpg"
		if err := ffmpeg.GenerateThumbnail(segmentPath, thumbnailPath, 5); err != nil {
			m.logger.Warn().Err(err).Msg("Failed to generate thumbnail")
		}
	}

	// Upload segment to Storage Service
	segmentInfo := domain.SegmentInfo{
		CameraID:        cameraID,
		FilePath:        segmentPath,
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: int(endTime.Sub(startTime).Seconds()),
		SizeBytes:       fileInfo.Size(),
		ThumbnailPath:   thumbnailPath,
	}

	if err := m.uploadSegment(context.Background(), segmentInfo); err != nil {
		m.logger.Error().Err(err).Str("segment_path", segmentPath).Msg("Failed to upload segment")
		return
	}

	// Update recording stats
	m.mu.Lock()
	if recording, exists := m.recordings[cameraID]; exists {
		recording.LastSegmentAt = endTime
		recording.SegmentCount++
		recording.TotalBytes += fileInfo.Size()
	}
	m.mu.Unlock()

	// Delete local file after successful upload
	if err := os.Remove(segmentPath); err != nil {
		m.logger.Warn().Err(err).Str("segment_path", segmentPath).Msg("Failed to delete local segment")
	}

	if thumbnailPath != "" {
		// Keep thumbnail for a short time (cleanup job will remove it)
	}

	m.logger.Info().
		Str("camera_id", cameraID).
		Str("segment_path", segmentPath).
		Int64("size_bytes", fileInfo.Size()).
		Msg("Segment uploaded successfully")
}

// Helper functions for external service calls

func (m *RecordingManager) getCameraDetails(ctx context.Context, cameraID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/vms/cameras/%s", m.vmsURL, cameraID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VMS service returned status %d", resp.StatusCode)
	}

	var camera map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&camera); err != nil {
		return nil, err
	}

	return camera, nil
}

func (m *RecordingManager) getRTSPURL(ctx context.Context, cameraID string) (string, error) {
	url := fmt.Sprintf("%s/vms/cameras/%s/stream", m.vmsURL, cameraID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("VMS service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	rtspURL, ok := result["rtsp_url"].(string)
	if !ok {
		return "", fmt.Errorf("invalid RTSP URL in response")
	}

	return rtspURL, nil
}

func (m *RecordingManager) reserveQuota(ctx context.Context, cameraID string, source string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/stream/reserve", m.streamCounterURL)

	payload := map[string]interface{}{
		"camera_id": cameraID,
		"user_id":   "recording-service",
		"source":    source,
		"duration":  86400, // 24 hours
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("stream counter returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	reservationID, ok := result["reservation_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid reservation ID in response")
	}

	return reservationID, nil
}

func (m *RecordingManager) releaseQuota(ctx context.Context, reservationID string) error {
	url := fmt.Sprintf("%s/api/v1/stream/release/%s", m.streamCounterURL, reservationID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (m *RecordingManager) uploadSegment(ctx context.Context, segment domain.SegmentInfo) error {
	url := fmt.Sprintf("%s/api/v1/storage/segments", m.storageURL)

	payload := map[string]interface{}{
		"camera_id":        segment.CameraID,
		"start_time":       segment.StartTime.Format(time.RFC3339),
		"end_time":         segment.EndTime.Format(time.RFC3339),
		"file_path":        segment.FilePath,
		"size_bytes":       segment.SizeBytes,
		"duration_seconds": segment.DurationSeconds,
		"thumbnail_path":   segment.ThumbnailPath,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("storage service returned status %d", resp.StatusCode)
	}

	return nil
}

func (m *RecordingManager) sendHeartbeats(ctx context.Context, cameraID string, reservationID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.RLock()
			_, exists := m.recordings[cameraID]
			m.mu.RUnlock()

			if !exists {
				return
			}

			url := fmt.Sprintf("%s/api/v1/stream/heartbeat/%s", m.streamCounterURL, reservationID)
			req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
			if err != nil {
				m.logger.Warn().Err(err).Msg("Failed to create heartbeat request")
				continue
			}

			resp, err := m.httpClient.Do(req)
			if err != nil {
				m.logger.Warn().Err(err).Msg("Failed to send heartbeat")
				continue
			}
			resp.Body.Close()

			m.logger.Debug().Str("camera_id", cameraID).Msg("Heartbeat sent")
		}
	}
}
