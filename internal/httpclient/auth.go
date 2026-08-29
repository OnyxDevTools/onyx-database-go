package httpclient

import "net/http"

// Signer applies API authentication headers.
type Signer struct {
	APIKey    string
	APISecret string
}

// Sign mutates the request in place by adding authentication headers.
func (s Signer) Sign(req *http.Request, body []byte) error {
	req.Header.Set("x-onyx-key", s.APIKey)
	req.Header.Set("x-onyx-secret", s.APISecret)
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return nil
}
