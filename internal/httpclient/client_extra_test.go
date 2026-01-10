package httpclient

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoStreamSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-onyx-key") != "k" {
			t.Fatalf("missing signer header")
		}
		if _, err := io.WriteString(w, "line1\n"); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, srv.Client(), Options{Signer: Signer{APIKey: "k", APISecret: "s"}})
	resp, err := c.DoStream(context.Background(), http.MethodGet, "/stream", nil)
	if err != nil {
		t.Fatalf("DoStream error: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(data), "line1") {
		t.Fatalf("unexpected stream body: %s", string(data))
	}
}

func TestParseErrorFallsBack(t *testing.T) {
	err := parseError(context.Background(), http.StatusTeapot, []byte("not json"))
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestDoStreamErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, srv.Client(), Options{Signer: Signer{}})
	_, err := c.DoStream(context.Background(), http.MethodGet, "/stream", nil)
	if err == nil {
		t.Fatalf("expected error on non-2xx stream")
	}
}

func TestDoStreamLogsResponses(t *testing.T) {
	buf := &strings.Builder{}
	logger := log.New(buf, "", 0)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("{}")); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, srv.Client(), Options{Logger: logger, LogResponses: true})
	resp, err := c.DoStream(context.Background(), http.MethodGet, "/stream", nil)
	if err != nil {
		t.Fatalf("DoStream error: %v", err)
	}
	resp.Body.Close()
	if !strings.Contains(buf.String(), "200") {
		t.Fatalf("expected response logs, got: %s", buf.String())
	}
}

func TestRedactedSecretLong(t *testing.T) {
	if got := redactedSecret("longsecret"); got != "long****" {
		t.Fatalf("unexpected redaction: %s", got)
	}
}

func TestDoJSONMarshalError(t *testing.T) {
	c := New("http://example.com", &http.Client{}, Options{})
	err := c.DoJSON(context.Background(), http.MethodPost, "/path", make(chan int), nil)
	if err == nil {
		t.Fatalf("expected marshal error")
	}
}

type errRoundTripper struct{}

func (errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}

func TestDoJSONTransportError(t *testing.T) {
	httpClient := &http.Client{Transport: errRoundTripper{}}
	c := New("http://example.com", httpClient, Options{})
	if err := c.DoJSON(context.Background(), http.MethodGet, "/fail", nil, nil); err == nil {
		t.Fatalf("expected transport error")
	}
}

func TestDoJSONNoBodyAndDefaultClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, nil, Options{})
	if err := c.DoJSON(context.Background(), http.MethodGet, "/", nil, &struct{}{}); err != nil {
		t.Fatalf("expected no error on empty body, got %v", err)
	}
}
