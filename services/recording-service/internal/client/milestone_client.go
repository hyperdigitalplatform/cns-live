package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// MilestoneClient handles communication with Milestone XProtect Management Server
type MilestoneClient struct {
	baseURL     string
	username    string
	password    string
	authType    string // "basic" or "ntlm"
	token       string
	tokenExpiry time.Time
	httpClient  *http.Client
	mu          sync.RWMutex
	logger      zerolog.Logger
}

// MilestoneConfig holds configuration for Milestone client
type MilestoneConfig struct {
	BaseURL        string
	Username       string
	Password       string
	AuthType       string
	SessionTimeout time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
}

// NewMilestoneClient creates a new Milestone XProtect API client
func NewMilestoneClient(config MilestoneConfig, logger zerolog.Logger) *MilestoneClient {
	if config.SessionTimeout == 0 {
		config.SessionTimeout = 1 * time.Hour
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.AuthType == "" {
		config.AuthType = "basic"
	}

	return &MilestoneClient{
		baseURL:  config.BaseURL,
		username: config.Username,
		password: config.Password,
		authType: config.AuthType,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

// ============================================================================
// Authentication Methods
// ============================================================================

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login response from Milestone
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// Login authenticates with Milestone and obtains session token
func (m *MilestoneClient) Login(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info().
		Str("base_url", m.baseURL).
		Str("username", m.username).
		Str("auth_type", m.authType).
		Msg("Logging into Milestone XProtect")

	loginReq := LoginRequest{
		Username: m.username,
		Password: m.password,
	}

	body, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/rest/v1/login", m.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		m.logger.Error().
			Int("status_code", resp.StatusCode).
			Str("response", string(bodyBytes)).
			Msg("Login failed")
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	m.token = loginResp.Token
	m.tokenExpiry = time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)

	m.logger.Info().
		Time("token_expiry", m.tokenExpiry).
		Msg("Successfully logged into Milestone")

	return nil
}

// Logout terminates the Milestone session
func (m *MilestoneClient) Logout(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.token == "" {
		return nil // Already logged out
	}

	m.logger.Info().Msg("Logging out from Milestone")

	endpoint := fmt.Sprintf("%s/api/rest/v1/login", m.baseURL)
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.token))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.Warn().Err(err).Msg("Failed to logout from Milestone")
		// Clear token anyway
	} else {
		defer resp.Body.Close()
	}

	m.token = ""
	m.tokenExpiry = time.Time{}

	m.logger.Info().Msg("Logged out from Milestone")
	return nil
}

// RefreshToken refreshes the authentication token
func (m *MilestoneClient) RefreshToken(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info().Msg("Refreshing Milestone token")

	endpoint := fmt.Sprintf("%s/api/rest/v1/login/refresh", m.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.token))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Warn().Int("status_code", resp.StatusCode).Msg("Token refresh failed, re-logging in")
		// Token expired, need to re-login
		m.mu.Unlock()
		err := m.Login(ctx)
		m.mu.Lock()
		return err
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode refresh response: %w", err)
	}

	m.token = loginResp.Token
	m.tokenExpiry = time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)

	m.logger.Info().Time("token_expiry", m.tokenExpiry).Msg("Token refreshed")
	return nil
}

// ensureAuthenticated ensures we have a valid token, refreshing or re-logging if needed
func (m *MilestoneClient) ensureAuthenticated(ctx context.Context) error {
	m.mu.RLock()
	hasToken := m.token != ""
	isExpired := time.Now().After(m.tokenExpiry.Add(-5 * time.Minute)) // Refresh 5 min before expiry
	m.mu.RUnlock()

	if !hasToken {
		return m.Login(ctx)
	}

	if isExpired {
		return m.RefreshToken(ctx)
	}

	return nil
}

// ============================================================================
// Camera Discovery Methods
// ============================================================================

// MilestoneCamera represents a camera from Milestone XProtect
type MilestoneCamera struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Enabled          bool                   `json:"enabled"`
	Recording        bool                   `json:"recording"`
	RecordingServer  string                 `json:"recordingServer"`
	LiveStreamURL    string                 `json:"liveStreamUrl"`
	PTZCapabilities  *PTZCapabilities       `json:"ptzCapabilities,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// PTZCapabilities represents PTZ control capabilities
type PTZCapabilities struct {
	Pan      bool    `json:"pan"`
	Tilt     bool    `json:"tilt"`
	Zoom     bool    `json:"zoom"`
	Focus    bool    `json:"focus"`
	MaxSpeed float64 `json:"maxSpeed,omitempty"`
}

// ListCamerasOptions contains options for listing cameras
type ListCamerasOptions struct {
	Limit     int
	Offset    int
	Enabled   *bool // nil means no filter
	Recording *bool // nil means no filter
}

// CameraList represents paginated list of cameras
type CameraList struct {
	Cameras []*MilestoneCamera `json:"cameras"`
	Total   int                `json:"total"`
}

// ListCameras retrieves all cameras from Milestone
func (m *MilestoneClient) ListCameras(ctx context.Context, opts ListCamerasOptions) (*CameraList, error) {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if opts.Limit == 0 {
		opts.Limit = 100
	}

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras?limit=%d&offset=%d",
		m.baseURL, opts.Limit, opts.Offset)

	if opts.Enabled != nil {
		endpoint += fmt.Sprintf("&enabled=%t", *opts.Enabled)
	}
	if opts.Recording != nil {
		endpoint += fmt.Sprintf("&recording=%t", *opts.Recording)
	}

	m.logger.Debug().
		Str("endpoint", endpoint).
		Msg("Listing cameras from Milestone")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var cameraList CameraList
	if err := json.NewDecoder(resp.Body).Decode(&cameraList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	m.logger.Info().
		Int("count", len(cameraList.Cameras)).
		Int("total", cameraList.Total).
		Msg("Retrieved cameras from Milestone")

	return &cameraList, nil
}

// GetCamera retrieves a single camera by ID
func (m *MilestoneClient) GetCamera(ctx context.Context, cameraID string) (*MilestoneCamera, error) {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras/%s", m.baseURL, cameraID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("camera not found: %s", cameraID)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var camera MilestoneCamera
	if err := json.NewDecoder(resp.Body).Decode(&camera); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &camera, nil
}

// ============================================================================
// Recording Control Methods
// ============================================================================

// StartRecordingRequest represents a request to start manual recording
type StartRecordingRequest struct {
	CameraID        string `json:"camera_id"`
	DurationSeconds int    `json:"durationSeconds"`
	TriggerBy       string `json:"triggerBy"`
	Description     string `json:"description,omitempty"`
}

// RecordingSession represents an active recording session
type RecordingSession struct {
	RecordingID       string    `json:"recordingId"`
	CameraID          string    `json:"cameraId"`
	StartTime         time.Time `json:"startTime"`
	EstimatedEndTime  time.Time `json:"estimatedEndTime"`
	DurationSeconds   int       `json:"durationSeconds"`
	Status            string    `json:"status"` // recording, stopped, failed
}

// StartRecording initiates manual recording for a camera
func (m *MilestoneClient) StartRecording(ctx context.Context, req StartRecordingRequest) (*RecordingSession, error) {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	m.logger.Info().
		Str("camera_id", req.CameraID).
		Int("duration_seconds", req.DurationSeconds).
		Msg("Starting Milestone recording")

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras/%s/recordings/start", m.baseURL, req.CameraID)

	body, err := json.Marshal(map[string]interface{}{
		"durationSeconds": req.DurationSeconds,
		"triggerBy":       req.TriggerBy,
		"description":     req.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var session RecordingSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	m.logger.Info().
		Str("recording_id", session.RecordingID).
		Str("camera_id", req.CameraID).
		Time("estimated_end_time", session.EstimatedEndTime).
		Msg("Recording started successfully")

	return &session, nil
}

// StopRecording stops an active manual recording
func (m *MilestoneClient) StopRecording(ctx context.Context, cameraID, recordingID string) error {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	m.logger.Info().
		Str("camera_id", cameraID).
		Str("recording_id", recordingID).
		Msg("Stopping Milestone recording")

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras/%s/recordings/stop", m.baseURL, cameraID)

	body, err := json.Marshal(map[string]string{
		"recordingId": recordingID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	m.logger.Info().
		Str("camera_id", cameraID).
		Str("recording_id", recordingID).
		Msg("Recording stopped successfully")

	return nil
}

// RecordingStatus represents the current recording status
type RecordingStatus struct {
	IsRecording      bool              `json:"isRecording"`
	CurrentRecording *RecordingSession `json:"currentRecording,omitempty"`
	LastRecording    *RecordingSession `json:"lastRecording,omitempty"`
}

// GetRecordingStatus retrieves current recording status for a camera
func (m *MilestoneClient) GetRecordingStatus(ctx context.Context, cameraID string) (*RecordingStatus, error) {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras/%s/recordings/status", m.baseURL, cameraID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var status RecordingStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// ============================================================================
// Recording Query Methods
// ============================================================================

// SequenceQueryRequest represents a request to query recording sequences
type SequenceQueryRequest struct {
	CameraID  string    `json:"cameraId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// RecordingSequence represents a continuous recording segment
type RecordingSequence struct {
	SequenceID      string    `json:"sequenceId"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	DurationSeconds int       `json:"durationSeconds"`
	Available       bool      `json:"available"`
	SizeBytes       int64     `json:"sizeBytes,omitempty"`
}

// RecordingGap represents a gap in recording
type RecordingGap struct {
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	DurationSeconds int       `json:"durationSeconds"`
}

// SequenceList represents query result with sequences and gaps
type SequenceList struct {
	CameraID  string               `json:"cameraId"`
	Sequences []*RecordingSequence `json:"sequences"`
	Gaps      []*RecordingGap      `json:"gaps"`
}

// QuerySequences queries available recording sequences for a time range
func (m *MilestoneClient) QuerySequences(ctx context.Context, req SequenceQueryRequest) (*SequenceList, error) {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	m.logger.Debug().
		Str("camera_id", req.CameraID).
		Time("start_time", req.StartTime).
		Time("end_time", req.EndTime).
		Msg("Querying recording sequences from Milestone")

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras/%s/sequences?startTime=%s&endTime=%s",
		m.baseURL,
		req.CameraID,
		req.StartTime.Format(time.RFC3339),
		req.EndTime.Format(time.RFC3339))

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var sequenceList SequenceList
	if err := json.NewDecoder(resp.Body).Decode(&sequenceList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	m.logger.Info().
		Str("camera_id", req.CameraID).
		Int("sequences", len(sequenceList.Sequences)).
		Int("gaps", len(sequenceList.Gaps)).
		Msg("Retrieved recording sequences")

	return &sequenceList, nil
}

// RecordingMetadata represents metadata about recordings
type RecordingMetadata struct {
	CameraID     string    `json:"cameraId"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Available    bool      `json:"available"`
	SegmentCount int       `json:"segmentCount"`
	TotalSize    int64     `json:"totalSize"`
}

// GetRecordingMetadata retrieves metadata about recordings for a time range
func (m *MilestoneClient) GetRecordingMetadata(ctx context.Context, cameraID string, timeRange TimeRange) (*RecordingMetadata, error) {
	sequences, err := m.QuerySequences(ctx, SequenceQueryRequest{
		CameraID:  cameraID,
		StartTime: timeRange.Start,
		EndTime:   timeRange.End,
	})
	if err != nil {
		return nil, err
	}

	var totalSize int64
	for _, seq := range sequences.Sequences {
		totalSize += seq.SizeBytes
	}

	metadata := &RecordingMetadata{
		CameraID:     cameraID,
		StartTime:    timeRange.Start,
		EndTime:      timeRange.End,
		Available:    len(sequences.Sequences) > 0,
		SegmentCount: len(sequences.Sequences),
		TotalSize:    totalSize,
	}

	return metadata, nil
}

// ============================================================================
// Playback Methods
// ============================================================================

// VideoStreamRequest represents a request for video stream
type VideoStreamRequest struct {
	CameraID  string    `json:"cameraId"`
	Timestamp time.Time `json:"timestamp"`
	Speed     float64   `json:"speed"` // -8 to 8
	Format    string    `json:"format"` // mjpeg, h264, webrtc
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// GetVideoStream retrieves a video stream from Milestone (returns stream URL or reader)
func (m *MilestoneClient) GetVideoStream(ctx context.Context, req VideoStreamRequest) (io.ReadCloser, error) {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	m.logger.Debug().
		Str("camera_id", req.CameraID).
		Time("timestamp", req.Timestamp).
		Float64("speed", req.Speed).
		Str("format", req.Format).
		Msg("Getting video stream from Milestone")

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras/%s/video?time=%s&speed=%f&format=%s",
		m.baseURL,
		req.CameraID,
		req.Timestamp.Format(time.RFC3339),
		req.Speed,
		req.Format)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	m.logger.Info().
		Str("camera_id", req.CameraID).
		Msg("Video stream started")

	return resp.Body, nil
}

// GetSnapshot retrieves a snapshot at a specific timestamp
func (m *MilestoneClient) GetSnapshot(ctx context.Context, cameraID string, timestamp time.Time) ([]byte, error) {
	if err := m.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/rest/v1/cameras/%s/snapshots?time=%s",
		m.baseURL,
		cameraID,
		timestamp.Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.mu.RLock()
	token := m.token
	m.mu.RUnlock()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	snapshot, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	return snapshot, nil
}
