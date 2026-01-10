package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	schemaCmds "github.com/OnyxDevTools/onyx-database-go/cmd/onyx-schema-go/commands"
)

func TestDispatchScenarios(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id","type":"string"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	outDir := filepath.Join(tmp, "gen")
	anchorPath := filepath.Join(tmp, "generate.go")

	tests := []struct {
		name          string
		args          []string
		wantCode      int
		wantStdoutSub []string
		wantStderrSub []string
		check         func(t *testing.T, stdout string)
	}{
		{
			name:          "usage",
			args:          []string{},
			wantCode:      2,
			wantStdoutSub: []string{"Usage: onyx-go"},
		},
		{
			name:          "help",
			args:          []string{"--help"},
			wantCode:      0,
			wantStdoutSub: []string{"Subcommands:"},
		},
		{
			name:          "unknown",
			args:          []string{"wat"},
			wantCode:      2,
			wantStderrSub: []string{"unknown subcommand"},
		},
		{
			name:          "schema help",
			args:          []string{"schema", "--help"},
			wantCode:      0,
			wantStdoutSub: []string{"Available commands"},
		},
		{
			name: "gen success",
			args: []string{
				"gen",
				"--schema", schemaPath,
				"--out", outDir,
				"--package", "pkg",
			},
			wantCode:      0,
			wantStdoutSub: []string{"Generated code from"},
		},
		{
			name: "gen init",
			args: []string{
				"gen",
				"init",
				"--file", anchorPath,
				"--schema", schemaPath,
				"--out", outDir,
				"--package", "pkg",
				"--timestamps", "string",
				"--tables", "users,accounts",
			},
			wantCode:      0,
			wantStdoutSub: []string{"Wrote go:generate anchor"},
			check: func(t *testing.T, stdout string) {
				data, err := os.ReadFile(anchorPath)
				if err != nil {
					t.Fatalf("read anchor: %v", err)
				}
				content := string(data)
				if !strings.Contains(content, "//go:generate onyx-go gen") {
					t.Fatalf("expected go:generate line, got: %s", content)
				}
				if !strings.Contains(content, "--schema "+schemaPath) {
					t.Fatalf("expected schema flag in anchor, got: %s", content)
				}
				if !strings.Contains(content, "--out "+outDir) {
					t.Fatalf("expected out flag in anchor, got: %s", content)
				}
				if !strings.Contains(content, "--package pkg") {
					t.Fatalf("expected package flag in anchor, got: %s", content)
				}
				if !strings.Contains(content, "--timestamps string") {
					t.Fatalf("expected timestamps flag in anchor, got: %s", content)
				}
				if !strings.Contains(content, "--tables users,accounts") {
					t.Fatalf("expected tables flag in anchor, got: %s", content)
				}
			},
		},
		{
			name:          "gen error",
			args:          []string{"gen", "--schema", filepath.Join(tmp, "missing.json")},
			wantCode:      1,
			wantStderrSub: []string{"no such file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			oldArgs := os.Args
			os.Args = append([]string{"onyx-go"}, tt.args...)
			defer func() { os.Args = oldArgs }()

			code := dispatch(os.Args[1:], &stdout, &stderr)
			schemaCmds.Stdout = os.Stdout
			schemaCmds.Stderr = os.Stderr
			if code != tt.wantCode {
				t.Fatalf("expected code %d, got %d (stdout=%q, stderr=%q)", tt.wantCode, code, stdout.String(), stderr.String())
			}

			for _, sub := range tt.wantStdoutSub {
				if !strings.Contains(stdout.String(), sub) {
					t.Fatalf("expected stdout to contain %q, got %q", sub, stdout.String())
				}
			}
			for _, sub := range tt.wantStderrSub {
				if !strings.Contains(stderr.String(), sub) {
					t.Fatalf("expected stderr to contain %q, got %q", sub, stderr.String())
				}
			}

			if tt.check != nil {
				tt.check(t, stdout.String())
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

func TestMainExitCodes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "help", args: []string{"--help"}, want: 0},
		{name: "unknown", args: []string{"nope"}, want: 2},
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
			os.Args = append([]string{"onyx-go"}, tt.args...)

			main()
			if got != tt.want {
				t.Fatalf("expected exit %d, got %d", tt.want, got)
			}
		})
	}
}
