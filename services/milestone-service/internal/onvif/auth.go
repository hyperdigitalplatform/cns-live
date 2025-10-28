package onvif

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"
)

// WSSecurity represents WS-Security authentication header
type WSSecurity struct {
	Username       string
	Password       string
	Nonce          string
	Created        string
	PasswordDigest string
}

// GenerateWSSecurity creates WS-Security authentication header for ONVIF
func GenerateWSSecurity(username, password string) (*WSSecurity, error) {
	// Generate random nonce (16 bytes)
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)

	// Generate timestamp in ISO 8601 format
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	// Create password digest: Base64(SHA1(Nonce + Timestamp + Password))
	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(timestamp))
	h.Write([]byte(password))
	passwordDigest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return &WSSecurity{
		Username:       username,
		Password:       password,
		Nonce:          nonceBase64,
		Created:        timestamp,
		PasswordDigest: passwordDigest,
	}, nil
}

// BuildSecurityHeader builds the WS-Security SOAP header
func (ws *WSSecurity) BuildSecurityHeader() string {
	return fmt.Sprintf(`
		<wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
		               xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
		               s:mustUnderstand="true">
			<wsse:UsernameToken>
				<wsse:Username>%s</wsse:Username>
				<wsse:Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">%s</wsse:Password>
				<wsse:Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">%s</wsse:Nonce>
				<wsu:Created>%s</wsu:Created>
			</wsse:UsernameToken>
		</wsse:Security>
	`, ws.Username, ws.PasswordDigest, ws.Nonce, ws.Created)
}
