package commands

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCommandFlagError(t *testing.T) {
	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &NormalizeCommand{}
	if code := cmd.Run([]string{"--unknown"}); code != 2 {
		t.Fatalf("expected usage error code, got %d", code)
	}
}

func TestNormalizeCommandReadError(t *testing.T) {
	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &NormalizeCommand{}
	if code := cmd.Run([]string{"--schema", "missing.json"}); code != 1 {
		t.Fatalf("expected read error code, got %d", code)
	}
}

func TestNormalizeCommandParseError(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte("{"), 0o644); err != nil {
		t.Fatalf("write invalid schema: %v", err)
	}

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &NormalizeCommand{}
	if code := cmd.Run([]string{"--schema", schemaPath}); code != 1 {
		t.Fatalf("expected parse error code, got %d", code)
	}
}

func TestNormalizeCommandWriteStdoutError(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	errBuf := &failWriter{err: os.ErrClosed}
	Stdout, Stderr = errBuf, &bytes.Buffer{}
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &NormalizeCommand{}
	if code := cmd.Run([]string{"--schema", schemaPath}); code != 1 {
		t.Fatalf("expected write error code, got %d", code)
	}
}

func TestNormalizeCommandWriteFilePaths(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"users","fields":[{"name":"id","type":"string"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	outFile := filepath.Join(tmp, "out.json")
	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &NormalizeCommand{}
	if code := cmd.Run([]string{"--schema", schemaPath, "--out", outFile}); code != 0 {
		t.Fatalf("expected success writing file, got %d", code)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read out file: %v", err)
	}
	if !bytes.Contains(data, []byte("\n")) {
		t.Fatalf("expected trailing newline in output")
	}

	if code := cmd.Run([]string{"--schema", schemaPath, "--out", tmp}); code != 1 {
		t.Fatalf("expected error when out path is directory, got %d", code)
	}
}

func TestNormalizeCommandMarshalError(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	orig := jsonMarshalIndent
	jsonMarshalIndent = func(v any, prefix, indent string) ([]byte, error) {
		return nil, errors.New("encode fail")
	}
	t.Cleanup(func() { jsonMarshalIndent = orig })

	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &NormalizeCommand{}
	if code := cmd.Run([]string{"--schema", schemaPath}); code != 1 {
		t.Fatalf("expected marshal error code, got %d", code)
	}
}

type failWriter struct{ err error }

func (f *failWriter) Write([]byte) (int, error) { return 0, f.err }
