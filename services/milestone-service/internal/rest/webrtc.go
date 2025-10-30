package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WebRTCPlaybackRequest represents a request to start WebRTC playback
type WebRTCPlaybackRequest struct {
	DeviceID     string  `json:"deviceId"`
	PlaybackTime string  `json:"playbackTime"` // ISO 8601 format
	SkipGaps     bool    `json:"skipGaps"`
	Speed        float64 `json:"speed"`
	StreamID     string  `json:"streamId,omitempty"`
	IncludeAudio bool    `json:"includeAudio"`
}

// WebRTCSession represents the WebRTC session response from Milestone
type WebRTCSession struct {
	SessionID string `json:"sessionId"`
	OfferSDP  string `json:"offerSDP"`
}

// ICECandidate represents an ICE candidate
type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

// ICECandidatesResponse represents the response from getting ICE candidates
// Milestone returns candidates as an array of JSON strings, not objects
type ICECandidatesResponse struct {
	Candidates []string `json:"candidates"`
}

// CreateWebRTCPlaybackSession initiates a WebRTC playback session with Milestone
// This follows the WebRTC signaling protocol documented in:
// mipsdk-samples-protocol/WebRTC_JavaScript/README.md
func (c *Client) CreateWebRTCPlaybackSession(ctx context.Context, req WebRTCPlaybackRequest) (*WebRTCSession, error) {
	// Build request body according to Milestone WebRTC API spec
	body := map[string]interface{}{
		"deviceId":     req.DeviceID,
		"resolution":   "notInUse", // Required but not used
		"includeAudio": req.IncludeAudio,
		"playbackTimeNode": map[string]interface{}{
			"playbackTime": req.PlaybackTime,
			"skipGaps":     req.SkipGaps,
			"speed":        req.Speed,
		},
	}

	// Add optional streamId if provided
	if req.StreamID != "" {
		body["streamId"] = req.StreamID
	}

	// POST to Milestone WebRTC Session endpoint
	resp, err := c.doRequest(ctx, "POST", "/API/REST/v1/WebRTC/Session", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebRTC session: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("milestone returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var session WebRTCSession
	if err := json.Unmarshal(respBody, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	// Validate response
	if session.SessionID == "" {
		return nil, fmt.Errorf("invalid response: missing sessionId")
	}
	if session.OfferSDP == "" {
		return nil, fmt.Errorf("invalid response: missing offerSDP")
	}

	return &session, nil
}

// UpdateWebRTCAnswer sends the answer SDP back to Milestone
// This is step 3 of the WebRTC signaling process
func (c *Client) UpdateWebRTCAnswer(ctx context.Context, sessionID, answerSDP string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is required")
	}
	if answerSDP == "" {
		return fmt.Errorf("answerSDP is required")
	}

	body := map[string]interface{}{
		"answerSDP": answerSDP,
	}

	// PATCH to update the session with answer - sessionId goes in URL path
	path := fmt.Sprintf("/API/REST/v1/WebRTC/Session/%s", sessionID)
	resp, err := c.doRequest(ctx, "PATCH", path, body)
	if err != nil {
		return fmt.Errorf("failed to update answer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("milestone returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SendICECandidate sends an ICE candidate to Milestone
// ICE candidates are exchanged to establish the peer-to-peer connection
func (c *Client) SendICECandidate(ctx context.Context, sessionID string, candidate interface{}) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is required")
	}

	// Convert candidate to JSON string (Milestone expects candidates as JSON strings in an array)
	candidateJSON, err := json.Marshal(candidate)
	if err != nil {
		return fmt.Errorf("failed to marshal candidate: %w", err)
	}

	body := map[string]interface{}{
		"candidates": []string{string(candidateJSON)},
	}

	// POST ICE candidate - sessionId in URL path
	path := fmt.Sprintf("/API/REST/v1/WebRTC/IceCandidates/%s", sessionID)
	resp, err := c.doRequest(ctx, "POST", path, body)
	if err != nil {
		return fmt.Errorf("failed to send ICE candidate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("milestone returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetICECandidates retrieves ICE candidates from Milestone
// The client should poll this endpoint to get server-side ICE candidates
func (c *Client) GetICECandidates(ctx context.Context, sessionID string) ([]ICECandidate, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is required")
	}

	// GET ICE candidates - sessionId in URL path
	path := fmt.Sprintf("/API/REST/v1/WebRTC/IceCandidates/%s", sessionID)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get ICE candidates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("milestone returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var candidatesResp ICECandidatesResponse
	if err := json.Unmarshal(respBody, &candidatesResp); err != nil {
		return nil, fmt.Errorf("failed to parse candidates response: %w", err)
	}

	// Parse JSON strings into ICECandidate objects
	var candidates []ICECandidate
	for _, candStr := range candidatesResp.Candidates {
		var cand ICECandidate
		if err := json.Unmarshal([]byte(candStr), &cand); err != nil {
			return nil, fmt.Errorf("failed to parse candidate string: %w", err)
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

// CloseWebRTCSession closes an active WebRTC session
// This should be called when playback is stopped
func (c *Client) CloseWebRTCSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is required")
	}

	path := fmt.Sprintf("/API/REST/v1/WebRTC/Session?sessionId=%s", sessionID)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("milestone returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
