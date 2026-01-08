package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-h"}, &out, &out); err != nil {
		t.Fatalf("expected nil err on help, got %v", err)
	}
	if out.Len() == 0 {
		t.Fatalf("expected usage output")
	}
}

func TestRunSuccess(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	outFile := filepath.Join(tmp, "out.go")

	var out bytes.Buffer
	args := []string{
		"--schema", schemaPath,
		"--out", outFile,
		"--package", "pkg",
	}
	if err := run(args, &out, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}
