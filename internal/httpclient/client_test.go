package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestDoJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected content type")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body["ping"] != "pong" {
			t.Fatalf("unexpected body: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	var resp map[string]bool
	err := c.DoJSON(context.Background(), http.MethodPost, "/test", map[string]string{"ping": "pong"}, &resp)
	if err != nil {
		t.Fatalf("DoJSON error: %v", err)
	}
	if !resp["ok"] {
		t.Fatalf("response not decoded: %v", resp)
	}
}

func TestLoggingToggle(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{Logger: logger, LogRequests: true, LogResponses: true})
	_ = c.DoJSON(context.Background(), http.MethodGet, "/", nil, nil)
	logs := buf.String()
	if !strings.Contains(logs, "GET") || !strings.Contains(logs, "response") {
		t.Fatalf("expected logs, got %s", logs)
	}

	buf.Reset()
	c = New(srv.URL, srv.Client(), Options{Logger: logger})
	_ = c.DoJSON(context.Background(), http.MethodGet, "/", nil, nil)
	if buf.Len() != 0 {
		t.Fatalf("expected no logs")
	}
}

func TestDoStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("{}\n{}"))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	resp, err := c.DoStream(context.Background(), http.MethodGet, "/stream", nil)
	if err != nil {
		t.Fatalf("DoStream error: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if len(data) == 0 {
		t.Fatalf("expected stream body")
	}
}

func TestErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"bad","message":"bad things","meta":{"hint":"x"}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	err := c.DoJSON(context.Background(), http.MethodGet, "/", nil, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	cerr, ok := err.(*contract.Error)
	if !ok || cerr.Code != "bad" || cerr.Meta["status"].(int) != http.StatusBadRequest {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthSigning(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	signer := Signer{APIKey: "key", APISecret: "secret", Now: func() time.Time { return now }, RequestID: func() string { return "req" }}

	req, _ := http.NewRequest(http.MethodPost, "https://example.com/path", strings.NewReader("body"))
	_ = signer.Sign(req, []byte("body"))

	if req.Header.Get("X-Onyx-Api-Key") != "key" {
		t.Fatalf("missing api key header")
	}
	if req.Header.Get("X-Onyx-Request-Id") != "req" {
		t.Fatalf("missing request id")
	}
	if req.Header.Get("X-Onyx-Timestamp") != now.Format(time.RFC3339) {
		t.Fatalf("timestamp mismatch")
	}
}
