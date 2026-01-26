package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	schemaCmds "github.com/OnyxDevTools/onyx-database-go/cmd/onyx-schema-go/commands"
)

func TestDispatchScenarios(t *testing.T) {
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
