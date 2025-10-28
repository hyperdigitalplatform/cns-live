package soap

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"
)

// LoginRequest represents the SOAP request for authentication
type LoginRequest struct {
	XMLName    xml.Name `xml:"http://videoos.net/2/XProtectCSServerCommand Login"`
	InstanceId string   `xml:"instanceId"`
}

// LoginResponse represents the SOAP response from authentication
type LoginResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		LoginResult struct {
			Token      string `xml:"Token"`
			TimeToLive struct {
				MicroSeconds int64 `xml:"MicroSeconds"`
			} `xml:"TimeToLive"`
		} `xml:"LoginResponse>LoginResult"`
	} `xml:"Body"`
}

// Login authenticates with Milestone ServerCommandService and retrieves a SOAP token
func (c *Client) Login(ctx context.Context) error {
	loginURL := c.baseURL + "/ManagementServer/ServerCommandService.svc"
	soapAction := "http://videoos.net/2/XProtectCSServerCommand/IServerCommandService/Login"

	// Create login request with empty instance ID
	request := LoginRequest{
		InstanceId: "00000000-0000-0000-0000-000000000000",
	}

	// Create envelope for the request
	envelope := Envelope{
		Body: Body{Content: request},
	}

	xmlData, err := xml.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal login request: %w", err)
	}

	// Create HTTP request with Basic Auth
	req, err := c.newRequestWithBasicAuth(ctx, "POST", loginURL, xmlData)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var loginResp LoginResponse
	if err := xml.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("decode login response: %w", err)
	}

	// Extract token and expiry
	token := loginResp.Body.LoginResult.Token
	if token == "" {
		return fmt.Errorf("no token in login response")
	}

	ttlMicroSeconds := loginResp.Body.LoginResult.TimeToLive.MicroSeconds
	if ttlMicroSeconds == 0 {
		// Default to 4 hours if not provided
		ttlMicroSeconds = 14400000000
	}

	// Calculate expiry time
	ttlDuration := time.Duration(ttlMicroSeconds) * time.Microsecond

	// Store token and expiry (thread-safe)
	c.mu.Lock()
	c.token = token
	c.tokenExpiry = time.Now().Add(ttlDuration)
	c.mu.Unlock()

	return nil
}

// Logout invalidates the current SOAP token
func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simply clear the token - Milestone doesn't require explicit logout
	c.token = ""
	c.tokenExpiry = time.Time{}

	return nil
}
