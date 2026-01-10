package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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
	outDir := filepath.Join(tmp, "out")

	var out bytes.Buffer
	args := []string{
		"--schema", schemaPath,
		"--out", outDir,
		"--package", "pkg",
	}
	if err := run(args, &out, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == "generate.go" {
			t.Fatalf("did not expect generate.go to be emitted")
		}
	}
	if !strings.Contains(out.String(), "Generated code from") {
		t.Fatalf("expected success output, got %s", out.String())
	}
}

func TestRunInvalidFlags(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	tests := []struct {
		name   string
		args   []string
		errSub string
	}{
		{
			name:   "invalid timestamps",
			args:   []string{"--schema", schemaPath, "--timestamps", "invalid"},
			errSub: "invalid --timestamps",
		},
		{
			name:   "invalid source",
			args:   []string{"--schema", schemaPath, "--source", "invalid"},
			errSub: "invalid --source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := run(tt.args, &out, &out)
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.errSub) {
				t.Fatalf("expected error to contain %q, got %v", tt.errSub, err)
			}
		})
	}
}

func TestRunFlagParseError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run([]string{"--bogus"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected parse error")
	}
	if stderr.Len() == 0 {
		t.Fatalf("expected usage output on parse error")
	}
}

func TestMainExitCodes(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "success", args: []string{"--schema", schemaPath, "--out", filepath.Join(tmp, "out")}, want: 0},
		{name: "error", args: []string{"--timestamps", "invalid"}, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origExit := exit
			origArgs := os.Args
			t.Cleanup(func() {
				exit = origExit
				os.Args = origArgs
			})

			var got int
			exit = func(code int) {
				got = code
			}
			os.Args = append([]string{"onyx-gen-go"}, tt.args...)

			main()
			if got != tt.want {
				t.Fatalf("expected exit %d, got %d", tt.want, got)
			}
		})
	}
}

func TestParseTables(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "empty", raw: " ", want: nil},
		{name: "trim", raw: " users,  accounts , ,", want: []string{"users", "accounts"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTables(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("expected %v, got %v", tt.want, got)
				}
			}
		})
	}
}
