package postgres

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rta/cctv/vms-service/internal/domain"
)

// sendPTZViaONVIFSOAP sends PTZ command using ONVIF SOAP protocol (for TP-Link Tapo and other ONVIF cameras)
func (r *PostgresRepository) sendPTZViaONVIFSOAP(u *url.URL, cmd *domain.PTZCommand) error {
	host := u.Hostname()
	ports := []string{"2020", "8888", "80", "8080", "8000"} // TP-Link Tapo: 2020, Others: 8888

	// Get credentials
	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	for _, port := range ports {
		// Try unified service endpoint (TP-Link Tapo uses this)
		serviceURL := fmt.Sprintf("http://%s:%s/onvif/service", host, port)
		if err := r.sendONVIFCommand(serviceURL, username, password, cmd); err == nil {
			r.logger.Debug().
				Str("camera_host", host).
				Str("port", port).
				Str("endpoint", "service").
				Msg("ONVIF SOAP PTZ command succeeded")
			return nil
		}

		// Try PTZ service endpoint (TrueView and other ONVIF cameras)
		ptzURL := fmt.Sprintf("http://%s:%s/onvif/ptz_service", host, port)
		if err := r.sendONVIFCommand(ptzURL, username, password, cmd); err == nil {
			r.logger.Debug().
				Str("camera_host", host).
				Str("port", port).
				Str("endpoint", "ptz_service").
				Msg("ONVIF SOAP PTZ command succeeded")
			return nil
		}

		// Try alternate PTZ endpoint (some cameras use /onvif/ptz)
		ptzAltURL := fmt.Sprintf("http://%s:%s/onvif/ptz", host, port)
		if err := r.sendONVIFCommand(ptzAltURL, username, password, cmd); err == nil {
			r.logger.Debug().
				Str("camera_host", host).
				Str("port", port).
				Str("endpoint", "ptz").
				Msg("ONVIF SOAP PTZ command succeeded")
			return nil
		}

		// Try device service endpoint (some cameras use this for PTZ)
		deviceURL := fmt.Sprintf("http://%s:%s/onvif/device_service", host, port)
		if err := r.sendONVIFCommand(deviceURL, username, password, cmd); err == nil {
			r.logger.Debug().
				Str("camera_host", host).
				Str("port", port).
				Str("endpoint", "device_service").
				Msg("ONVIF SOAP PTZ command succeeded")
			return nil
		}
	}

	return fmt.Errorf("all ONVIF SOAP endpoints failed")
}

// generateWSSecurityHeader generates WS-Security UsernameToken header
func generateWSSecurityHeader(username, password string) string {
	// Generate random nonce (20 bytes per ONVIF spec)
	nonceBytes := make([]byte, 20)
	for i := range nonceBytes {
		nonceBytes[i] = byte((time.Now().UnixNano() + int64(i)) % 256)
	}
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)
	created := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	// Password Digest = Base64(SHA1(nonce_bytes + created + password))
	hash := sha1.New()
	hash.Write(nonceBytes) // Use raw nonce bytes, not base64
	hash.Write([]byte(created))
	hash.Write([]byte(password))
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return fmt.Sprintf(`<Security s:mustUnderstand="1" xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
    <UsernameToken>
      <Username>%s</Username>
      <Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">%s</Password>
      <Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">%s</Nonce>
      <Created xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">%s</Created>
    </UsernameToken>
  </Security>`, username, digest, nonce, created)
}

// sendONVIFCommand sends an ONVIF SOAP command
// Tries multiple profile tokens to support different camera brands
func (r *PostgresRepository) sendONVIFCommand(endpoint, username, password string, cmd *domain.PTZCommand) error {
	// Try most common profile tokens: PROFILE_000 (TrueView), profile_1 (Tapo)
	profileTokens := []string{"PROFILE_000", "profile_1"}

	for _, profileToken := range profileTokens {
		err := r.trySendONVIFCommandWithProfile(endpoint, username, password, cmd, profileToken)
		if err == nil {
			r.logger.Debug().
				Str("endpoint", endpoint).
				Str("profile_token", profileToken).
				Msg("ONVIF command succeeded with profile token")
			return nil // Success
		}
		// Continue trying next profile token
	}

	return fmt.Errorf("all profile tokens failed")
}

// trySendONVIFCommandWithProfile attempts to send ONVIF command with specific profile token
func (r *PostgresRepository) trySendONVIFCommandWithProfile(endpoint, username, password string, cmd *domain.PTZCommand, profileToken string) error {
	// Generate WS-Security header
	securityHeader := ""
	if username != "" {
		securityHeader = generateWSSecurityHeader(username, password)
	}

	var soapBody string

	switch cmd.Action {
	case domain.PTZActionMove:
		// ONVIF ContinuousMove command
		panSpeed := cmd.Pan
		tiltSpeed := cmd.Tilt
		zoomSpeed := cmd.Zoom

		soapBody = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
            xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"
            xmlns:tt="http://www.onvif.org/ver10/schema">
  <s:Header>
    %s
  </s:Header>
  <s:Body>
    <tptz:ContinuousMove>
      <tptz:ProfileToken>%s</tptz:ProfileToken>
      <tptz:Velocity>
        <tt:PanTilt x="%f" y="%f" space="http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace"/>
        <tt:Zoom x="%f" space="http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace"/>
      </tptz:Velocity>
    </tptz:ContinuousMove>
  </s:Body>
</s:Envelope>`, securityHeader, profileToken, panSpeed, tiltSpeed, zoomSpeed)

	case domain.PTZActionStop:
		// ONVIF Stop command
		soapBody = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
            xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl">
  <s:Header>
    %s
  </s:Header>
  <s:Body>
    <tptz:Stop>
      <tptz:ProfileToken>%s</tptz:ProfileToken>
      <tptz:PanTilt>true</tptz:PanTilt>
      <tptz:Zoom>true</tptz:Zoom>
    </tptz:Stop>
  </s:Body>
</s:Envelope>`, securityHeader, profileToken)

	case domain.PTZActionGoToPreset:
		// ONVIF GotoPreset or GotoHomePosition
		if cmd.Preset == 1 {
			soapBody = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
            xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl">
  <s:Header>
    %s
  </s:Header>
  <s:Body>
    <tptz:GotoHomePosition>
      <tptz:ProfileToken>%s</tptz:ProfileToken>
    </tptz:GotoHomePosition>
  </s:Body>
</s:Envelope>`, securityHeader, profileToken)
		} else {
			soapBody = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
            xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl">
  <s:Header>
    %s
  </s:Header>
  <s:Body>
    <tptz:GotoPreset>
      <tptz:ProfileToken>%s</tptz:ProfileToken>
      <tptz:PresetToken>preset_%d</tptz:PresetToken>
    </tptz:GotoPreset>
  </s:Body>
</s:Envelope>`, securityHeader, profileToken, cmd.Preset)
		}

	default:
		return fmt.Errorf("unsupported PTZ action: %s", cmd.Action)
	}

	// Send SOAP request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(soapBody))
	if err != nil {
		return fmt.Errorf("failed to create SOAP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(soapBody)))

	// NOTE: Do NOT use HTTP Basic Auth with ONVIF
	// Authentication is done via WS-Security in the SOAP envelope

	// Send request with shorter timeout for faster fallback
	client := &http.Client{
		Timeout: 2 * time.Second, // 2 second timeout per request
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Check for SOAP fault
	if strings.Contains(bodyStr, "s:Fault") || strings.Contains(bodyStr, "soap:Fault") {
		return fmt.Errorf("ONVIF returned fault: %s", bodyStr)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("camera returned status %d: %s", resp.StatusCode, bodyStr)
	}

	return nil
}
