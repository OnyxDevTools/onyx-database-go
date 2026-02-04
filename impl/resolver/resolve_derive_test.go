package resolver

import (
	"context"
	"os"
	"testing"
)

func isolateResolverEnv(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("ONYX_CONFIG_PATH", "")
	t.Setenv("ONYX_DATABASE_ID", "")
	t.Setenv("ONYX_DATABASE_BASE_URL", "")
	t.Setenv("ONYX_DATABASE_API_KEY", "")
	t.Setenv("ONYX_DATABASE_API_SECRET", "")
	t.Setenv("ONYX_AI_BASE_URL", "")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

func TestResolveDefaultsDatabaseBaseURL(t *testing.T) {
	isolateResolverEnv(t)
	ClearCache()
	cfg, meta, err := Resolve(context.Background(), Config{
		DatabaseID: "db",
		APIKey:     "k",
		APISecret:  "s",
	})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if cfg.DatabaseBaseURL != defaultDatabaseBaseURL {
		t.Fatalf("expected default base url %s, got %s", defaultDatabaseBaseURL, cfg.DatabaseBaseURL)
	}
	if meta.Sources.DatabaseBaseURL != SourceDefault {
		t.Fatalf("expected default source, got %s", meta.Sources.DatabaseBaseURL)
	}
}

func TestResolveDerivesDatabaseIDFromAPIKey(t *testing.T) {
	isolateResolverEnv(t)
	ClearCache()
	key := "key_123e4567-e89b-12d3-a456-426614174000"
	cfg, meta, err := Resolve(context.Background(), Config{
		APIKey:    key,
		APISecret: "secret",
	})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if cfg.DatabaseID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("expected derived database id, got %s", cfg.DatabaseID)
	}
	if meta.Sources.DatabaseID != SourceDerived {
		t.Fatalf("expected derived source, got %s", meta.Sources.DatabaseID)
	}
}

func TestResolveDerivesDatabaseIDViaLookup(t *testing.T) {
	isolateResolverEnv(t)
	ClearCache()
	original := lookupDatabaseIDFromAPIKeyFn
	lookupDatabaseIDFromAPIKeyFn = func(ctx context.Context, baseURL, apiKey, apiSecret string) (string, error) {
		if baseURL != defaultDatabaseBaseURL {
			t.Fatalf("expected default base url %s, got %s", defaultDatabaseBaseURL, baseURL)
		}
		if apiKey != "key_plain" || apiSecret != "secret" {
			t.Fatalf("unexpected credentials: %s/%s", apiKey, apiSecret)
		}
		return "db-from-api", nil
	}
	t.Cleanup(func() { lookupDatabaseIDFromAPIKeyFn = original })

	cfg, meta, err := Resolve(context.Background(), Config{
		APIKey:    "key_plain",
		APISecret: "secret",
	})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if cfg.DatabaseID != "db-from-api" {
		t.Fatalf("expected resolved database id, got %s", cfg.DatabaseID)
	}
	if meta.Sources.DatabaseID != SourceDerived {
		t.Fatalf("expected derived source, got %s", meta.Sources.DatabaseID)
	}
}
