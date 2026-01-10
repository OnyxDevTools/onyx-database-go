package impl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func TestInitReusesHTTPClientForSameConfig(t *testing.T) {
	ClearConfigCache()
	ctx := context.Background()
	cfg := Config{
		DatabaseID:      "db",
		DatabaseBaseURL: "https://api.example.com",
		APIKey:          "key",
		APISecret:       "secret",
	}

	c1, err := Init(ctx, cfg)
	if err != nil {
		t.Fatalf("init 1 err: %v", err)
	}
	c2, err := Init(ctx, cfg)
	if err != nil {
		t.Fatalf("init 2 err: %v", err)
	}

	if c1.(*client).httpClient != c2.(*client).httpClient {
		t.Fatalf("expected http client to be cached and reused")
	}
}

func TestClearConfigCacheClearsHTTPClientCache(t *testing.T) {
	ClearConfigCache()
	ctx := context.Background()
	cfg := Config{
		DatabaseID:      "db2",
		DatabaseBaseURL: "https://api.example.com",
		APIKey:          "key2",
		APISecret:       "secret2",
	}

	c1, err := Init(ctx, cfg)
	if err != nil {
		t.Fatalf("init 1 err: %v", err)
	}
	ClearConfigCache()
	c2, err := Init(ctx, cfg)
	if err != nil {
		t.Fatalf("init 2 err: %v", err)
	}

	if c1.(*client).httpClient == c2.(*client).httpClient {
		t.Fatalf("expected http client cache to be cleared")
	}
}

func TestDeleteUsesDefaultPartition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "partition=p1" {
			t.Fatalf("expected partition query param, got %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := &client{
		cfg: resolver.ResolvedConfig{
			DatabaseID: "db",
			Partition:  "p1",
		},
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
	}
	if err := c.Delete(context.Background(), "users", "id1"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
}
