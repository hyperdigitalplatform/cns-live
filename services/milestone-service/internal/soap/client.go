package soap

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Client represents a Milestone SOAP API client
type Client struct {
	baseURL     string
	username    string
	password    string
	token       string
	tokenExpiry time.Time
	httpClient  *http.Client
	mu          sync.RWMutex
}

// Envelope represents a SOAP envelope
type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    Body
}

// Body represents a SOAP body
type Body struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Content interface{}
}

// SOAPFault represents a SOAP fault
type SOAPFault struct {
	XMLName     xml.Name `xml:"Fault"`
	FaultCode   string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
	Detail      struct {
		ErrorNumber    int `xml:"ErrorNumber"`
		SubErrorNumber int `xml:"SubErrorNumber"`
	} `xml:"detail"`
}

// NewClient creates a new Milestone SOAP client
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Accept self-signed certificates
				},
			},
		},
	}
}

// GetToken returns the current token (thread-safe)
func (c *Client) GetToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// newRequestWithBasicAuth creates an HTTP request with Basic Authentication
func (c *Client) newRequestWithBasicAuth(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(append([]byte(xml.Header), body...)))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.password)
	return req, nil
}

// IsTokenValid checks if the current token is valid
func (c *Client) IsTokenValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token != "" && time.Now().Before(c.tokenExpiry)
}

// ensureAuthenticated ensures the client has a valid token
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	c.mu.RLock()
	needsAuth := c.token == "" || time.Now().After(c.tokenExpiry.Add(-5*time.Minute))
	c.mu.RUnlock()

	if needsAuth {
		return c.Login(ctx)
	}
	return nil
}

// sendSOAPRequest sends a SOAP request and parses the response
func (c *Client) sendSOAPRequest(ctx context.Context, url, soapAction string, request interface{}, response interface{}) error {
	envelope := Envelope{
		Body: Body{Content: request},
	}

	xmlData, err := xml.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url,
		bytes.NewReader(append([]byte(xml.Header), xmlData...)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check for SOAP fault
	var faultEnvelope struct {
		Body struct {
			Fault SOAPFault `xml:"Fault"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(body, &faultEnvelope); err == nil {
		if faultEnvelope.Body.Fault.FaultString != "" {
			return fmt.Errorf("SOAP fault: %s (error %d:%d)",
				faultEnvelope.Body.Fault.FaultString,
				faultEnvelope.Body.Fault.Detail.ErrorNumber,
				faultEnvelope.Body.Fault.Detail.SubErrorNumber)
		}
	}

	// Parse actual response
	return xml.Unmarshal(body, response)
}
