package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCommandSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "schema.json")
	writeFile(t, path, `{
  "tables": [
    {"name": "users", "fields": [{"name": "id", "type": "string"}]}
  ]
}`)

	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &ValidateCommand{}
	exitCode := cmd.Run([]string{"--schema", path})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}

	if !strings.Contains(out.String(), "Schema is valid") {
		t.Fatalf("expected success message, got: %s", out.String())
	}
}

func TestValidateCommandDuplicateNames(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "schema.json")
	writeFile(t, path, `{
  "tables": [
    {"name": "users", "fields": [
      {"name": "id", "type": "string"},
      {"name": "id", "type": "int"}
    ]},
    {"name": "users", "fields": []}
  ]
}`)

	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &ValidateCommand{}
	exitCode := cmd.Run([]string{"--schema", path})
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", exitCode)
	}

	if !strings.Contains(out.String(), "duplicate") {
		t.Fatalf("expected duplicate error, got: %s", out.String())
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
