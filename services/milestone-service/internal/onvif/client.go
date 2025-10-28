package onvif

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an ONVIF client
type Client struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new ONVIF client
func NewClient(endpoint, username, password string) *Client {
	return &Client{
		endpoint: endpoint,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // For cameras with self-signed certificates
				},
			},
		},
	}
}

// SOAPEnvelope represents a generic SOAP envelope
type SOAPEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  string   `xml:"Header"`
	Body    string   `xml:"Body"`
}

// sendSOAPRequest sends a SOAP request with WS-Security authentication
func (c *Client) sendSOAPRequest(ctx context.Context, endpoint, soapAction, body string) ([]byte, error) {
	// Generate WS-Security header
	wsSecurity, err := GenerateWSSecurity(c.username, c.password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate WS-Security: %w", err)
	}

	// Build complete SOAP envelope
	soapEnvelope := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
            xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
            xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
            xmlns:tds="http://www.onvif.org/ver10/device/wsdl"
            xmlns:trt="http://www.onvif.org/ver10/media/wsdl"
            xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Header>
		%s
	</s:Header>
	<s:Body>
		%s
	</s:Body>
</s:Envelope>`, wsSecurity.BuildSecurityHeader(), body)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString(soapEnvelope))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	if soapAction != "" {
		req.Header.Set("SOAPAction", soapAction)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for SOAP faults
	if bytes.Contains(responseBody, []byte("Fault")) {
		return nil, fmt.Errorf("SOAP fault: %s", string(responseBody))
	}

	return responseBody, nil
}
