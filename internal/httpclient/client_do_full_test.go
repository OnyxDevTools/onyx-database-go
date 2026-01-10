package httpclient

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoHandlesLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	c := New("http://example.com", &http.Client{}, Options{Signer: Signer{}, Logger: logger, LogRequests: true})
	_ = c.DoJSON(context.Background(), http.MethodGet, "/x", map[string]string{"a": "b"}, nil)
	if !strings.Contains(buf.String(), "GET") || !strings.Contains(buf.String(), "Headers") {
		t.Fatalf("expected request log, got %s", buf.String())
	}
}

func TestDoNon2xxLogsResponse(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"nope"}`, http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, srv.Client(), Options{Logger: logger, LogResponses: true})
	if err := c.DoJSON(context.Background(), http.MethodGet, "/fail", nil, nil); err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(buf.String(), "400") {
		t.Fatalf("expected response log")
	}
}

func TestDoStreamLogsSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, srv.Client(), Options{Logger: logger, LogResponses: true})
	resp, err := c.DoStream(context.Background(), http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("DoStream err: %v", err)
	}
	resp.Body.Close()
	if !strings.Contains(buf.String(), "200") {
		t.Fatalf("expected 200 log")
	}
}

func TestParseErrorCoversStatusTextFallbacks(t *testing.T) {
	if err := parseError(context.Background(), 0, nil); err == nil {
		t.Fatalf("expected error with fallback message")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := parseError(ctx, http.StatusBadRequest, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation")
	}
}
