package resolver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveCacheHitReturnsCachedValues(t *testing.T) {
	ClearCache()
	cfg := Config{
		DatabaseID:      "id",
		DatabaseBaseURL: "https://base",
		APIKey:          "key",
		APISecret:       "secret",
	}

	first, meta, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("first resolve err: %v", err)
	}
	if meta.Sources.DatabaseID != SourceExplicit {
		t.Fatalf("expected explicit source, got %+v", meta.Sources)
	}

	// Mutate env to ensure cached path is exercised (would fail without cache).
	t.Setenv("ONYX_DATABASE_ID", "")
	second, meta2, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("cached resolve err: %v", err)
	}
	if second != first || meta2 != meta {
		t.Fatalf("expected cached values to match")
	}
}

func TestResolveMissingRequiredFields(t *testing.T) {
	isolateResolverEnv(t)
	ClearCache()
	if _, _, err := Resolve(context.Background(), Config{}); err == nil {
		t.Fatalf("expected error for missing required fields")
	}
}

func TestResolveContextCanceledDuringFileLoad(t *testing.T) {
	ClearCache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := Resolve(ctx, Config{ConfigPath: "any"})
	if err == nil {
		t.Fatalf("expected ctx error")
	}
}

func TestResolveFilePathAbsAndEvalFallbacks(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte(`{"databaseId":"db","databaseBaseUrl":"http://base","apiKey":"k","apiSecret":"s"}`), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}

	restore := func() {
		pathAbs = filepath.Abs
		evalSymlinks = filepath.EvalSymlinks
	}
	t.Cleanup(restore)

	pathAbs = func(string) (string, error) {
		return "", errors.New("abs fail")
	}
	_, meta, err := Resolve(context.Background(), Config{ConfigPath: cfgPath})
	if err != nil {
		t.Fatalf("resolve with abs fail err: %v", err)
	}
	if meta.FilePath != cfgPath {
		t.Fatalf("expected meta filepath to be cfg path, got %s", meta.FilePath)
	}

	ClearCache()
	pathAbs = filepath.Abs
	evalSymlinks = func(p string) (string, error) {
		return "", errors.New("eval fail")
	}
	meta = Meta{}
	resolved, meta2, err := Resolve(context.Background(), Config{ConfigPath: cfgPath})
	if err != nil || resolved.DatabaseID == "" {
		t.Fatalf("resolve with eval fail err: %v", err)
	}
	absPath, _ := filepath.Abs(cfgPath)
	if meta2.FilePath != absPath {
		t.Fatalf("expected abs path when eval fails, got %s", meta2.FilePath)
	}
}

func TestResolveFromFilesUnmarshalAndBaseURLFallback(t *testing.T) {
	ClearCache()
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	// Invalid JSON to hit unmarshal continue.
	if err := os.WriteFile(filepath.Join(configDir, "onyx-database.json"), []byte("{"), 0o644); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}
	// Valid root-specific fallback using baseUrl.
	rootSpecific := filepath.Join(tmp, "onyx-database-db1.json")
	if err := os.WriteFile(rootSpecific, []byte(`{"databaseId":"db1","baseUrl":"https://legacy","apiKey":"k","apiSecret":"s"}`), 0o644); err != nil {
		t.Fatalf("write root specific: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	resolved, meta, err := Resolve(context.Background(), Config{DatabaseID: "db1"})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if resolved.DatabaseBaseURL != "https://legacy" {
		t.Fatalf("expected baseUrl fallback, got %+v", resolved)
	}
	if meta.FilePath == "" {
		t.Fatalf("expected file path recorded")
	}
}

func TestResolveFromFilesContextCancel(t *testing.T) {
	ClearCache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := resolveFromFiles(ctx, Config{ConfigPath: "any"}, &ResolvedConfig{}, &Meta{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestResolveUsesEnvOverrideWhenConfigMissing(t *testing.T) {
	ClearCache()
	// Ensure no files used.
	t.Setenv("ONYX_CONFIG_PATH", "")
	t.Setenv("ONYX_DATABASE_ID", "env-id")
	t.Setenv("ONYX_DATABASE_BASE_URL", "http://env")
	t.Setenv("ONYX_DATABASE_API_KEY", "k")
	t.Setenv("ONYX_DATABASE_API_SECRET", "s")
	resolved, _, err := Resolve(context.Background(), Config{})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if resolved.DatabaseBaseURL != "http://env" {
		t.Fatalf("expected env base url, got %+v", resolved)
	}
}

func TestResolveRecordsEmptyFilePathWhenNoConfigUsed(t *testing.T) {
	ClearCache()
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	t.Setenv("ONYX_CONFIG_PATH", "")
	t.Setenv("ONYX_DATABASE_ID", "env-id")
	t.Setenv("ONYX_DATABASE_BASE_URL", "http://env")
	t.Setenv("ONYX_DATABASE_API_KEY", "k")
	t.Setenv("ONYX_DATABASE_API_SECRET", "s")

	_, meta, err := Resolve(context.Background(), Config{})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if meta.FilePath != "" {
		t.Fatalf("expected empty file path, got %s", meta.FilePath)
	}
}

func TestResolveFilePathEmptyBranchViaStub(t *testing.T) {
	ClearCache()
	restore := func() { resolveFromFilesFn = resolveFromFiles }
	resolveFromFilesFn = func(context.Context, Config, *ResolvedConfig, *Meta) (string, error) {
		return "", nil
	}
	t.Cleanup(restore)

	cfg := Config{
		DatabaseID:      "id",
		DatabaseBaseURL: "http://base",
		APIKey:          "k",
		APISecret:       "s",
	}
	_, meta, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if meta.FilePath != "" {
		t.Fatalf("expected empty filepath from stub, got %s", meta.FilePath)
	}
}

func TestResolveHonorsCacheExpiry(t *testing.T) {
	ClearCache()
	now := time.Unix(0, 0)
	defaultCache.now = func() time.Time { return now }
	defer func() { defaultCache.now = time.Now }()

	cfg := Config{
		DatabaseID:      "id",
		DatabaseBaseURL: "http://base",
		APIKey:          "k",
		APISecret:       "s",
		CacheTTL:        time.Millisecond,
	}
	if _, _, err := Resolve(context.Background(), cfg); err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	now = now.Add(2 * time.Millisecond)
	if _, _, ok := defaultCache.get(cacheKey(cfg)); ok {
		t.Fatalf("expected cache entry expired")
	}
}

func TestResolveFromFilesAppliesAIBaseURLAndPartition(t *testing.T) {
	ClearCache()
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "custom.json")
	if err := os.WriteFile(cfgPath, []byte(`{
		"databaseId":"db",
		"databaseBaseUrl":"http://base",
		"apiKey":"k",
		"apiSecret":"s",
		"aiBaseUrl":"http://ai",
		"partition":"p1"
	}`), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}

	resolved := &ResolvedConfig{}
	meta := &Meta{}
	chosen, err := resolveFromFiles(context.Background(), Config{ConfigPath: cfgPath}, resolved, meta)
	if err != nil {
		t.Fatalf("resolveFromFiles err: %v", err)
	}
	if chosen != cfgPath {
		t.Fatalf("expected chosen path %s, got %s", cfgPath, chosen)
	}
	if resolved.AIBaseURL != "http://ai" || resolved.Partition != "p1" {
		t.Fatalf("expected ai base url and partition applied, got %+v", resolved)
	}
	if meta.Sources.AIBaseURL != SourceFile || meta.Sources.DatabaseID != SourceFile {
		t.Fatalf("expected file sources recorded, got %+v", meta.Sources)
	}
}
