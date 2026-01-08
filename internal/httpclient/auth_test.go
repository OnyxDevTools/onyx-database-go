package httpclient

import (
	"net/http"
	"testing"
)

func TestSignerSetsHeaders(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	s := Signer{APIKey: "key", APISecret: "secret"}
	if err := s.Sign(req, nil); err != nil {
		t.Fatalf("sign returned error: %v", err)
	}
	if req.Header.Get("x-onyx-key") != "key" || req.Header.Get("x-onyx-secret") != "secret" {
		t.Fatalf("headers not set: %+v", req.Header)
	}
}
