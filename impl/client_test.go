package impl

import (
	"context"
	"testing"
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
