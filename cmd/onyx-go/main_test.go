package main

import (
	"bytes"
	"errors"
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

func TestRunGenHelpAndInitErrors(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	errWriter := errorWriter{err: errors.New("boom")}

	code := runGen([]string{"-h"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "Usage of onyx-go gen") {
		t.Fatalf("expected help output, got code %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	code = runGen([]string{"-h"}, errWriter, &stderr)
	if code != 0 || !strings.Contains(stderr.String(), "write usage: boom") {
		t.Fatalf("expected usage write error, got code %d stderr=%s", code, stderr.String())
	}
	stderr.Reset()

	code = runGen([]string{"--bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected parse error branch, got code %d stderr=%s", code, stderr.String())
	}
	stderr.Reset()

	code = runGen([]string{"--schema", filepath.Join(t.TempDir(), "missing.json")}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "no such file") {
		t.Fatalf("expected generator error branch, got code %d stderr=%s", code, stderr.String())
	}
	stderr.Reset()

	tmp := t.TempDir()
	// anchorPath points to directory so WriteFile should fail
	code = runGenInit([]string{"--file", tmp}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected failure writing anchor, got %d", code)
	}

	stdout.Reset()
	stderr.Reset()
	code = runGenInit([]string{"-h"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "Usage of onyx-go gen init") {
		t.Fatalf("expected init help, got code=%d stdout=%s", code, stdout.String())
	}

	code = runGenInit([]string{"--bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected parse error branch, got code %d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	origGetwd := getwd
	getwd = func() (string, error) {
		return "", errors.New("getwd fail")
	}
	t.Cleanup(func() { getwd = origGetwd })

	anchorPath := filepath.Join(t.TempDir(), "anchor.go")
	code = runGenInit([]string{"--file", anchorPath, "--schema", filepath.Join(t.TempDir(), "schema.json")}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "getwd") {
		t.Fatalf("expected getwd error branch, got code %d stderr=%s", code, stderr.String())
	}
}

func TestBuildAnchorFileIncludesAPIFlags(t *testing.T) {
	cfg := anchorConfig{
		Source:          "api",
		DatabaseID:      "db123",
		SchemaPath:      "/abs/schema.json",
		OutPath:         "/abs/out",
		PackageName:     "pkg",
		Tables:          []string{"users"},
		TimestampFormat: "string",
	}

	content := buildAnchorFile(cfg)
	assertContains := func(sub string) {
		if !strings.Contains(content, sub) {
			t.Fatalf("expected anchor content to include %q, got:\n%s", sub, content)
		}
	}
	assertContains("--source api")
	assertContains("--database-id db123")
	assertContains("--schema /abs/schema.json")
	assertContains("--out /abs/out")
	assertContains("--package pkg")
	assertContains("--timestamps string")
	assertContains("--tables users")
}

func TestPrintLayoutRelFallbacks(t *testing.T) {
	var out bytes.Buffer
	cfg := anchorConfig{
		SchemaPath: "/abs/api/onyx.schema.json",
		OutPath:    "gen/onyx",
	}

	printLayout(&out, "", "/abs/generate.go", cfg)
	content := out.String()
	if !strings.Contains(content, "/abs/generate.go") {
		t.Fatalf("expected absolute anchor path in output, got: %s", content)
	}
	if !strings.Contains(content, "onyx.schema.json") {
		t.Fatalf("expected schema path in output, got: %s", content)
	}
	if !strings.Contains(content, "gen/onyx") {
		t.Fatalf("expected cleaned relative out path, got: %s", content)
	}
}

type errorWriter struct {
	err error
}

func (e errorWriter) Write([]byte) (int, error) {
	return 0, e.err
}

type recordingErrorWriter struct {
	err error
	buf bytes.Buffer
}

func (w *recordingErrorWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	return 0, w.err
}

func TestWriteUsageBufferError(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("usage")
	err := writeUsageBuffer(&buf, errorWriter{err: errors.New("boom")})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error from writeUsageBuffer, got %v", err)
	}
}

func TestRunGenWriteUsageFailures(t *testing.T) {
	t.Run("parse error usage write fails", func(t *testing.T) {
		rec := &recordingErrorWriter{err: errors.New("fail")}
		code := runGen([]string{"--bogus"}, errorWriter{err: errors.New("stdout-fail")}, rec)
		if code != 2 {
			t.Fatalf("expected parse error code, got %d", code)
		}
		if !strings.Contains(rec.buf.String(), "write usage: fail") {
			t.Fatalf("expected usage error recorded, got %s", rec.buf.String())
		}
	})

	t.Run("generator error usage write fails", func(t *testing.T) {
		rec := &recordingErrorWriter{err: errors.New("fail")}
		code := runGen([]string{"--schema", filepath.Join(t.TempDir(), "missing.json")}, errorWriter{err: errors.New("stdout-fail")}, rec)
		if code != 1 {
			t.Fatalf("expected generator error code, got %d", code)
		}
		if !strings.Contains(rec.buf.String(), "write usage: fail") {
			t.Fatalf("expected usage error recorded, got %s", rec.buf.String())
		}
	})
}

func TestRunGenInitWriteUsageFailures(t *testing.T) {
	t.Run("usage write fails", func(t *testing.T) {
		recStdout := &recordingErrorWriter{err: errors.New("fail")}
		var stderr bytes.Buffer
		code := runGenInit([]string{"-h"}, recStdout, &stderr)
		if code != 0 {
			t.Fatalf("expected help code, got %d", code)
		}
		if !strings.Contains(stderr.String(), "write usage: fail") {
			t.Fatalf("expected usage error recorded, got %s", stderr.String())
		}
	})

	t.Run("parse error usage write fails", func(t *testing.T) {
		recStderr := &recordingErrorWriter{err: errors.New("fail")}
		code := runGenInit([]string{"--bogus"}, errorWriter{err: errors.New("stdout-fail")}, recStderr)
		if code != 2 {
			t.Fatalf("expected parse error code, got %d", code)
		}
		if !strings.Contains(recStderr.buf.String(), "write usage: fail") {
			t.Fatalf("expected usage error recorded, got %s", recStderr.buf.String())
		}
	})
}
