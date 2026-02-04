package resolver

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveReadsConfigFile(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte(`{"databaseId":"file-db","databaseBaseUrl":"http://file","apiKey":"k","apiSecret":"s","partition":"part"}`), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	ClearCache()
	resolved, meta, err := Resolve(context.Background(), Config{ConfigPath: cfgPath, CacheTTL: time.Minute})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if resolved.DatabaseID != "file-db" || meta.Sources.DatabaseID != SourceFile {
		t.Fatalf("expected file source, got %+v %+v", resolved, meta)
	}
	if resolved.Partition != "part" {
		t.Fatalf("expected partition from file, got %q", resolved.Partition)
	}
}

func TestResolveCacheHitSkipsRecompute(t *testing.T) {
	ClearCache()
	readConfigEnvOriginal := readConfigEnv
	defer func() { readConfigEnv = readConfigEnvOriginal }()

	// First resolve populates cache.
	if _, _, err := Resolve(context.Background(), Config{
		DatabaseID:      "db",
		DatabaseBaseURL: "http://base",
		APIKey:          "k",
		APISecret:       "s",
	}); err != nil {
		t.Fatalf("initial resolve err: %v", err)
	}

	// Force resolver to panic if it tries to read env again; cache path should bypass.
	readConfigEnv = func(string) string {
		t.Fatalf("expected cached result, readConfigEnv called")
		return ""
	}

	if _, _, err := Resolve(context.Background(), Config{
		DatabaseID:      "db",
		DatabaseBaseURL: "http://base",
		APIKey:          "k",
		APISecret:       "s",
	}); err != nil {
		t.Fatalf("cached resolve err: %v", err)
	}
}

func TestResolveMissingRequired(t *testing.T) {
	isolateResolverEnv(t)
	ClearCache()
	if _, _, err := Resolve(context.Background(), Config{}); err == nil {
		t.Fatalf("expected error for missing required values")
	}
}

func TestResolveLogsSourcesWhenDebugEnabled(t *testing.T) {
	ClearCache()
	t.Setenv("ONYX_DEBUG", "true")
	buf := &bytes.Buffer{}
	orig := log.Writer()
	log.SetOutput(buf)
	defer log.SetOutput(orig)

	_, _, err := Resolve(context.Background(), Config{
		DatabaseID:      "db",
		DatabaseBaseURL: "http://base",
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if !strings.Contains(buf.String(), "onyx resolver") {
		t.Fatalf("expected debug log, got %s", buf.String())
	}
}

func TestResolveExplicitNoFilePathAndNoPartition(t *testing.T) {
	ClearCache()
	resolved, meta, err := Resolve(context.Background(), Config{
		DatabaseID:      "db-explicit",
		DatabaseBaseURL: "http://base",
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if meta.FilePath != "" {
		t.Fatalf("expected empty filepath, got %q", meta.FilePath)
	}
	if resolved.Partition != "" {
		t.Fatalf("expected empty partition, got %q", resolved.Partition)
	}
}

func TestResolveExplicitPartitionWins(t *testing.T) {
	ClearCache()
	resolved, _, err := Resolve(context.Background(), Config{
		DatabaseID:      "db-explicit",
		DatabaseBaseURL: "http://base",
		APIKey:          "k",
		APISecret:       "s",
		Partition:       " p-1 ",
	})
	if err != nil {
		t.Fatalf("resolve err: %v", err)
	}
	if resolved.Partition != "p-1" {
		t.Fatalf("expected trimmed explicit partition, got %q", resolved.Partition)
	}
}
