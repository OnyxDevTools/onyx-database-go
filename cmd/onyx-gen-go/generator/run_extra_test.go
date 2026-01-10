package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func TestRunMkdirError(t *testing.T) {
	tmp := t.TempDir()
	// Create a file where a directory is expected to force MkdirAll failure.
	outFile := filepath.Join(tmp, "out")
	if err := os.WriteFile(outFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	opts := Options{
		SchemaPath:  filepath.Join(tmp, "schema.json"),
		OutPath:     outFile,
		PackageName: "pkg",
	}
	if err := os.WriteFile(opts.SchemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id","type":"string"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	if err := Run(opts); err == nil {
		t.Fatalf("expected error when MkdirAll fails on file path")
	}
}

func TestRunWriteFileError(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id","type":"string"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	outDir := filepath.Join(tmp, "out")
	if err := os.MkdirAll(outDir, 0o500); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Remove write permission to trigger write failure.
	if err := os.Chmod(outDir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	opts := Options{SchemaPath: schemaPath, OutPath: outDir, PackageName: "pkg"}
	if err := Run(opts); err == nil {
		t.Fatalf("expected write error due to permissions")
	}
}

func TestRenderHelpersFullBranches(t *testing.T) {
	schema := onyx.Schema{
		Tables: []onyx.Table{{
			Name: "User",
			Fields: []onyx.Field{
				{Name: "id", Type: "String"},
				{Name: "email", Type: "String"},
				{Name: "isActive", Type: "Boolean"},
			},
			Resolvers: []onyx.Resolver{{Name: "roles"}},
		}},
	}
	common := renderCommon(schema, "pkg")
	if len(common) == 0 {
		t.Fatalf("expected common code")
	}

	table := schema.Tables[0]
	tableCode := renderTable(table, "pkg", "string")
	if len(tableCode) == 0 {
		t.Fatalf("expected table code")
	}

	if v := toExported("123"); v != "123" {
		t.Fatalf("toExported numeric prefix unexpected: %s", v)
	}
	if v := requiredStringFields(table); len(v) == 2 { // should exclude id and nullable fields
		t.Fatalf("requiredStringFields should skip id: %+v", v)
	}
}

func TestRunValidateOptionsError(t *testing.T) {
	err := Run(Options{TimestampFormat: "invalid"})
	if err == nil || !strings.Contains(err.Error(), "invalid --timestamps") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRunTableWriteError(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id","type":"string"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	outDir := filepath.Join(tmp, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	tablePath := filepath.Join(outDir, "user.go")
	if err := os.WriteFile(tablePath, []byte("existing"), 0o400); err != nil {
		t.Fatalf("prewrite table file: %v", err)
	}

	opts := Options{SchemaPath: schemaPath, OutPath: outDir, PackageName: "pkg"}
	if err := Run(opts); err == nil {
		t.Fatalf("expected table write error")
	}
}
