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
		if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
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
	if !strings.Contains(logs, "GET") || !strings.Contains(logs, "200") {
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
		if _, err := w.Write([]byte("{}\n{}")); err != nil {
			t.Fatalf("write response: %v", err)
		}
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

func TestParseErrorAndRedaction(t *testing.T) {
	err := parseError(context.Background(), http.StatusBadRequest, []byte(`{"code":"bad","message":"nope"}`))
	if cerr, ok := err.(*contract.Error); !ok || cerr.Code != "bad" || cerr.Meta["status"].(int) != http.StatusBadRequest {
		t.Fatalf("unexpected structured error: %v", err)
	}

	err = parseError(context.Background(), http.StatusTeapot, []byte("oops"))
	if cerr, ok := err.(*contract.Error); !ok || cerr.Meta["body"] != "oops" {
		t.Fatalf("unexpected fallback error: %v", err)
	}

	if redactedSecret("abcd") != "****" {
		t.Fatalf("redaction mismatch for short secret")
	}
	if redactedSecret("abc") != "****" {
		t.Fatalf("short secret should be fully redacted")
	}
}

func TestErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"code":"bad","message":"bad things","meta":{"hint":"x"}}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
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
	signer := Signer{APIKey: "key", APISecret: "secret"}

	req, _ := http.NewRequest(http.MethodPost, "https://example.com/path", strings.NewReader("body"))
	_ = signer.Sign(req, []byte("body"))

	if req.Header.Get("x-onyx-key") != "key" {
		t.Fatalf("missing api key header")
	}
	if req.Header.Get("x-onyx-secret") != "secret" {
		t.Fatalf("missing secret header")
	}
}
