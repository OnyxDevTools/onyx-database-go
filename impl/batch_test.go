package impl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func TestBatchSaveBatchesAndRetries(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			http.Error(w, `{"code":"rate","message":"slow"}`, http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db"},
		sleep:      func(time.Duration) {}, // avoid delay in test
	}

	entities := []any{map[string]any{"id": 1}, map[string]any{"id": 2}, map[string]any{"id": 3}}
	if err := batchSave(context.Background(), c, "users", entities, 2); err != nil {
		t.Fatalf("batchSave err: %v", err)
	}
	if calls < 2 {
		t.Fatalf("expected retry on transient error, got %d calls", calls)
	}
}

func TestBatchSaveContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	httpClient := httpclient.New("http://example.com", &http.Client{Transport: cancelRoundTripper{}}, httpclient.Options{})
	c := &client{httpClient: httpClient, cfg: resolver.ResolvedConfig{DatabaseID: "db"}}
	cancel()

	err := batchSave(ctx, c, "users", []any{map[string]any{"id": 1}}, 1)
	if err == nil {
		t.Fatalf("expected context error")
	}
}

func TestBatchSaveNonTransientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db"},
	}

	err := batchSave(context.Background(), c, "users", []any{map[string]any{"id": 1}}, 1)
	if err == nil {
		t.Fatalf("expected non-transient error")
	}
}

type cancelRoundTripper struct{}

func (cancelRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, context.Canceled
}
