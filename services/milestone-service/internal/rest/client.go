package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// Client represents a REST API client for Milestone XProtect
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	token      string
	tokenMutex sync.RWMutex
	tokenExp   time.Time
}

// NewClient creates a new Milestone REST API client
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // For self-signed certificates
				},
			},
		},
	}
}

// NewClientFromEnv creates a client from environment variables
func NewClientFromEnv() *Client {
	baseURL := os.Getenv("MILESTONE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://192.168.1.11"
	}

	username := os.Getenv("MILESTONE_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("MILESTONE_PASSWORD")
	if password == "" {
		password = "password"
	}

	return NewClient(baseURL, username, password)
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// authenticate gets an OAuth access token
func (c *Client) authenticate(ctx context.Context) error {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Check if token is still valid
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return nil
	}

	// Request new token
	tokenURL := fmt.Sprintf("%s/IDP/connect/token", c.baseURL)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", c.username)
	data.Set("password", c.password)
	data.Set("client_id", "GrantValidatorClient")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.token = tokenResp.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// getToken returns the current access token (thread-safe)
func (c *Client) getToken() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.token
}

// ensureAuthenticated ensures we have a valid token
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	c.tokenMutex.RLock()
	needsRefresh := c.token == "" || time.Now().After(c.tokenExp)
	c.tokenMutex.RUnlock()

	if needsRefresh {
		return c.authenticate(ctx)
	}
	return nil
}

// doRequest performs an authenticated HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.getToken()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
