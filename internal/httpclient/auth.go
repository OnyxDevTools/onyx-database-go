package httpclient

import (
	"net/http"
)

// Signer applies API authentication headers.
type Signer struct {
	APIKey    string
	APISecret string
}

// Sign mutates the request in place by adding authentication headers.
func (s Signer) Sign(req *http.Request, body []byte) error {
	req.Header.Set("x-onyx-key", s.APIKey)
	req.Header.Set("x-onyx-secret", s.APISecret)
	req.Header.Set("Accept", "application/json")
	return nil
}
