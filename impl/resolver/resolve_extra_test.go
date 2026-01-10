package resolver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveReadsConfigFile(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte(`{"databaseId":"file-db","databaseBaseUrl":"http://file","apiKey":"k","apiSecret":"s"}`), 0o644); err != nil {
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
}
