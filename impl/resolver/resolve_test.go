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

	configDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	configSpecific := filepath.Join(configDir, "onyx-database-db1.json")
	if err := os.WriteFile(configSpecific, []byte(`{"databaseId":"db1","databaseBaseUrl":"https://config-specific","apiKey":"k","apiSecret":"s"}`), 0o644); err != nil {
		t.Fatalf("write config specific: %v", err)
	}
	configGeneric := filepath.Join(configDir, "onyx-database.json")
	if err := os.WriteFile(configGeneric, []byte(`{"databaseId":"db1","databaseBaseUrl":"https://config-generic","apiKey":"k2","apiSecret":"s2"}`), 0o644); err != nil {
		t.Fatalf("write config generic: %v", err)
	}
	rootSpecific := filepath.Join(dir, "onyx-database-db1.json")
	if err := os.WriteFile(rootSpecific, []byte(`{"databaseId":"db1","databaseBaseUrl":"https://root-specific","apiKey":"k3","apiSecret":"s3"}`), 0o644); err != nil {
		t.Fatalf("write root specific: %v", err)
	}
	rootGeneric := filepath.Join(dir, "onyx-database.json")
	if err := os.WriteFile(rootGeneric, []byte(`{"databaseId":"db1","databaseBaseUrl":"https://root-generic","apiKey":"k4","apiSecret":"s4"}`), 0o644); err != nil {
		t.Fatalf("write root generic: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	cfg, meta, err := Resolve(context.Background(), Config{DatabaseID: "db1"})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if cfg.DatabaseBaseURL != "https://config-specific" {
		t.Fatalf("expected config dir database-specific file to win: %+v", cfg)
	}
	if meta.Sources.DatabaseBaseURL != SourceFile {
		t.Fatalf("expected file source recorded: %+v", meta)
	}
}

func TestResolveHonorsConfigPath(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.json")
	if err := os.WriteFile(custom, []byte(`{"databaseId":"dbX","databaseBaseUrl":"https://custom","apiKey":"k","apiSecret":"s"}`), 0o644); err != nil {
		t.Fatalf("write custom config: %v", err)
	}

	cfg, _, err := Resolve(context.Background(), Config{ConfigPath: custom})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if cfg.DatabaseBaseURL != "https://custom" {
		t.Fatalf("custom path not used: %+v", cfg)
	}
}

func TestResolveConfigDirectoryFallback(t *testing.T) {
	ClearCache()
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configFile := filepath.Join(configDir, "onyx-database.json")
	if err := os.WriteFile(configFile, []byte(`{"databaseId":"db1","databaseBaseUrl":"https://cfg","apiKey":"k","apiSecret":"s"}`), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	// ensure env vars do not override file resolution
	t.Setenv("ONYX_DATABASE_ID", "")
	t.Setenv("ONYX_DATABASE_BASE_URL", "")
	t.Setenv("ONYX_DATABASE_API_KEY", "")
	t.Setenv("ONYX_DATABASE_API_SECRET", "")

	cfg, meta, err := Resolve(context.Background(), Config{})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if cfg.DatabaseBaseURL != "https://cfg" {
		t.Fatalf("expected config dir file to be used, got %+v", cfg)
	}
	configFileAbs, err := filepath.Abs(configFile)
	if err != nil {
		t.Fatalf("abs config file: %v", err)
	}
	configFileReal, err := filepath.EvalSymlinks(configFileAbs)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	if filepath.Clean(meta.FilePath) != filepath.Clean(configFileReal) {
		t.Fatalf("expected file path %s, got %s", configFileReal, meta.FilePath)
	}
	if meta.Sources.DatabaseBaseURL != SourceFile {
		t.Fatalf("expected file source, got %+v", meta.Sources)
	}
}

func TestResolveHomeProfileFallback(t *testing.T) {
	ClearCache()
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(filepath.Join(home, ".onyx"), 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	t.Setenv("HOME", home)

	cfgPath := filepath.Join(home, ".onyx", "onyx-database.json")
	if err := os.WriteFile(cfgPath, []byte(`{"databaseId":"home-db","databaseBaseUrl":"https://home","apiKey":"h","apiSecret":"s","partition":"home-part"}`), 0o644); err != nil {
		t.Fatalf("write home cfg: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// ensure no local config
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	cfg, meta, err := Resolve(context.Background(), Config{})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if cfg.DatabaseBaseURL != "https://home" || cfg.Partition != "home-part" {
		t.Fatalf("expected home config applied, got %+v", cfg)
	}
	if meta.FilePath == "" {
		t.Fatalf("expected meta filepath to be set")
	}
	realCfg, _ := filepath.EvalSymlinks(cfgPath)
	realMeta, _ := filepath.EvalSymlinks(meta.FilePath)
	if filepath.Clean(realMeta) != filepath.Clean(realCfg) {
		t.Fatalf("expected meta filepath to be home config, got %s", meta.FilePath)
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
