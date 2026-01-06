package httpclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"
)

// Signer applies API authentication headers.
type Signer struct {
	APIKey    string
	APISecret string
	Now       func() time.Time
	RequestID func() string
}

// Sign mutates the request in place by adding authentication headers.
func (s Signer) Sign(req *http.Request, body []byte) error {
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}

	timestamp := now.UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(s.APISecret))
	mac.Write([]byte(req.Method))
	mac.Write([]byte("\n"))
	mac.Write([]byte(req.URL.Path))
	mac.Write([]byte("\n"))
	mac.Write(body)
	mac.Write([]byte("\n"))
	mac.Write([]byte(timestamp))

	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	hash := sha256.Sum256(body)
	req.Header.Set("X-Onyx-Api-Key", s.APIKey)
	req.Header.Set("X-Onyx-Signature", signature)
	req.Header.Set("X-Onyx-Timestamp", timestamp)
	req.Header.Set("X-Onyx-Content-Sha256", hex.EncodeToString(hash[:]))
	if s.RequestID != nil {
		req.Header.Set("X-Onyx-Request-Id", s.RequestID())
	}
	return nil
}
