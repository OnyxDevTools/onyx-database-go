package resolver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveExplicitOverridesEnv(t *testing.T) {
	t.Setenv("ONYX_DATABASE_ID", "env-id")
	t.Setenv("ONYX_DATABASE_BASE_URL", "https://env.example")
	t.Setenv("ONYX_DATABASE_API_KEY", "env-key")
	t.Setenv("ONYX_DATABASE_API_SECRET", "env-secret")

	cfg, meta, err := Resolve(context.Background(), Config{
		DatabaseID:      "explicit-id",
		DatabaseBaseURL: "https://explicit",
		APIKey:          "explicit-key",
		APISecret:       "explicit-secret",
	})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	if cfg.DatabaseID != "explicit-id" || cfg.APIKey != "explicit-key" {
		t.Fatalf("explicit values were not used: %+v", cfg)
	}
	if meta.Sources.DatabaseID != SourceExplicit || meta.Sources.APIKey != SourceExplicit {
		t.Fatalf("meta sources incorrect: %+v", meta)
	}
}

func TestResolveEnvOnly(t *testing.T) {
	t.Setenv("ONYX_DATABASE_ID", "env-id")
	t.Setenv("ONYX_DATABASE_BASE_URL", "https://env.example")
	t.Setenv("ONYX_DATABASE_API_KEY", "env-key")
	t.Setenv("ONYX_DATABASE_API_SECRET", "env-secret")

	cfg, _, err := Resolve(context.Background(), Config{})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if cfg.DatabaseID != "env-id" || cfg.DatabaseBaseURL != "https://env.example" {
		t.Fatalf("env values not applied: %+v", cfg)
	}
}

func TestResolveFileSearchOrder(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	if err := os.MkdirAll(filepath.Join(home, ".onyx"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Setenv("HOME", home)

	specific := filepath.Join(dir, "onyx-database-db1.json")
	os.WriteFile(specific, []byte(`{"databaseId":"db1","databaseBaseUrl":"https://file","apiKey":"k","apiSecret":"s"}`), 0o644)
	generic := filepath.Join(dir, "onyx-database.json")
	os.WriteFile(generic, []byte(`{"databaseId":"db1","databaseBaseUrl":"https://generic","apiKey":"k2","apiSecret":"s2"}`), 0o644)

	cwd, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(cwd) })

	cfg, meta, err := Resolve(context.Background(), Config{DatabaseID: "db1"})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if cfg.DatabaseBaseURL != "https://file" {
		t.Fatalf("expected database-specific file to win: %+v", cfg)
	}
	if meta.Sources.DatabaseBaseURL != SourceFile {
		t.Fatalf("expected file source recorded: %+v", meta)
	}
}

func TestResolveHonorsConfigPath(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.json")
	os.WriteFile(custom, []byte(`{"databaseId":"dbX","databaseBaseUrl":"https://custom","apiKey":"k","apiSecret":"s"}`), 0o644)

	cfg, _, err := Resolve(context.Background(), Config{ConfigPath: custom})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if cfg.DatabaseBaseURL != "https://custom" {
		t.Fatalf("custom path not used: %+v", cfg)
	}
}

func TestResolveCacheTTL(t *testing.T) {
	ClearCache()
	defaultCache.now = func() time.Time { return time.Unix(0, 0) }
	defer func() { defaultCache.now = time.Now }()

	cfg := Config{
		DatabaseID:      "id",
		DatabaseBaseURL: "https://example",
		APIKey:          "key",
		APISecret:       "secret",
		CacheTTL:        time.Minute,
	}
	resolved, _, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if _, _, ok := defaultCache.get(cacheKey(cfg)); !ok {
		t.Fatalf("expected cache to be populated")
	}

	// advance beyond TTL
	defaultCache.now = func() time.Time { return time.Unix(int64(time.Minute.Seconds()*2), 0) }
	if _, _, ok := defaultCache.get(cacheKey(cfg)); ok {
		t.Fatalf("expected cache to expire")
	}
	_ = resolved
}
