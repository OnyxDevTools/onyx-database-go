package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpFlagDisplaysUsage(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	if err := run([]string{"--help"}, stdout, stderr); err != nil {
		t.Fatalf("expected help to succeed, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage of onyx-gen-go") {
		t.Fatalf("expected usage output, got: %s", output)
	}
}

func TestRequiredFlagsAreValidated(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := run([]string{}, stdout, stderr)
	if err == nil {
		t.Fatalf("expected error when required flags missing")
	}

	if !strings.Contains(err.Error(), "missing required flag(s): --out, --package") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestTimestampValidation(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := run([]string{"--out", "./out.go", "--package", "models", "--timestamps", "invalid"}, stdout, stderr)
	if err == nil {
		t.Fatalf("expected timestamp validation error")
	}

	if !strings.Contains(err.Error(), "timestamps") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidArguments(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables": []}`), 0o644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	args := []string{"--out", "./out.go", "--package", "models", "--schema", schemaPath}

	if err := run(args, stdout, stderr); err != nil {
		t.Fatalf("expected valid arguments to pass, got %v", err)
	}
}

func TestInvalidSource(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := run([]string{"--out", "./out.go", "--package", "models", "--source", "unknown"}, stdout, stderr)
	if err == nil {
		t.Fatalf("expected invalid source to fail")
	}

	if !strings.Contains(err.Error(), "--source") {
		t.Fatalf("unexpected error for invalid source: %v", err)
	}
}

func TestParseTables(t *testing.T) {
	tests := []struct {
		raw      string
		expected []string
	}{
		{"", nil},
		{"Users", []string{"Users"}},
		{"Users,Orders", []string{"Users", "Orders"}},
		{" Users , Orders , ", []string{"Users", "Orders"}},
	}

	for _, tt := range tests {
		result := parseTables(tt.raw)
		if len(result) != len(tt.expected) {
			t.Fatalf("parseTables(%q) len mismatch: %d != %d", tt.raw, len(result), len(tt.expected))
		}

		for i := range result {
			if result[i] != tt.expected[i] {
				t.Fatalf("parseTables(%q) index %d mismatch: %q != %q", tt.raw, i, result[i], tt.expected[i])
			}
		}
	}
}
