package generator

import (
	"bytes"
	"testing"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func TestNeedsTimeTable(t *testing.T) {
	table := onyx.Table{
		Name: "T",
		Fields: []onyx.Field{
			{Name: "createdAt", Type: "Timestamp"},
		},
	}
	if !needsTimeTable(table, "time") {
		t.Fatalf("expected needs time table")
	}
	if needsTimeTable(table, "string") {
		t.Fatalf("expected string timestamps to skip time import")
	}
	if needsTimeTable(onyx.Table{}, "time") {
		t.Fatalf("expected false when no timestamp fields present")
	}
}

func TestClientPaths(t *testing.T) {
	if got := clientOutPath("gen/onyx"); got != "gen/onyx/common.go" {
		t.Fatalf("clientOutPath mismatch: %s", got)
	}
	if got := ClientPathFor("gen/onyx"); got != "gen/onyx/common.go" {
		t.Fatalf("ClientPathFor mismatch: %s", got)
	}
}

func TestRenderGenerateUsesTimestampEnv(t *testing.T) {
	opts := Options{SchemaPath: "schema.json", OutPath: ".", PackageName: "pkg"}
	t.Setenv("ONYX_GEN_TIMESTAMP", time.Unix(0, 0).UTC().Format(time.RFC3339))
	code := renderGenerate(opts)
	if len(code) == 0 || !bytes.Contains(code, []byte("//go:generate")) {
		t.Fatalf("expected generate directive")
	}
}
