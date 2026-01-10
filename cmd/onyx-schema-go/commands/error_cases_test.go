package commands

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	schemas "github.com/OnyxDevTools/onyx-database-go/impl/schema"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestValidateCommandErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     func(t *testing.T) []string
		wantCode int
		wantSub  string
	}{
		{
			name: "flag error",
			args: func(t *testing.T) []string {
				return []string{"--nope"}
			},
			wantCode: 2,
		},
		{
			name: "missing file",
			args: func(t *testing.T) []string {
				return []string{"--schema", filepath.Join(t.TempDir(), "missing.json")}
			},
			wantCode: 1,
			wantSub:  "failed to read schema",
		},
		{
			name: "invalid json",
			args: func(t *testing.T) []string {
				path := filepath.Join(t.TempDir(), "schema.json")
				if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
					t.Fatalf("write schema: %v", err)
				}
				return []string{"--schema", path}
			},
			wantCode: 1,
			wantSub:  "failed to parse schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			Stdout, Stderr = &out, &out
			defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

			cmd := &ValidateCommand{}
			code := cmd.Run(tt.args(t))
			if code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, code)
			}
			if tt.wantSub != "" && !strings.Contains(out.String(), tt.wantSub) {
				t.Fatalf("expected %q in output, got %s", tt.wantSub, out.String())
			}
		})
	}
}

func TestValidateSchemaReportsEmptyNames(t *testing.T) {
	tests := []struct {
		name    string
		schema  onyx.Schema
		wantSub string
	}{
		{
			name: "empty table name",
			schema: onyx.Schema{Tables: []onyx.Table{{
				Name: "",
			}}},
			wantSub: "table name cannot be empty",
		},
		{
			name: "empty field name",
			schema: onyx.Schema{Tables: []onyx.Table{{
				Name:   "users",
				Fields: []onyx.Field{{Name: ""}},
			}}},
			wantSub: "field with empty name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateSchema(tt.schema)
			if len(errs) == 0 {
				t.Fatalf("expected errors")
			}
			if !strings.Contains(errs[0].Error(), tt.wantSub) {
				t.Fatalf("expected %q, got %v", tt.wantSub, errs[0])
			}
		})
	}
}

func TestNormalizeCommandErrors(t *testing.T) {
	validSchema := func(t *testing.T) string {
		path := filepath.Join(t.TempDir(), "schema.json")
		if err := os.WriteFile(path, []byte(`{"tables":[{"name":"User","fields":[]}]} `), 0o644); err != nil {
			t.Fatalf("write schema: %v", err)
		}
		return path
	}

	tests := []struct {
		name     string
		args     func(t *testing.T) []string
		stdout   func() io.Writer
		wantCode int
		wantSub  string
	}{
		{
			name: "flag error",
			args: func(t *testing.T) []string {
				return []string{"--nope"}
			},
			stdout:   func() io.Writer { return &bytes.Buffer{} },
			wantCode: 2,
		},
		{
			name: "missing file",
			args: func(t *testing.T) []string {
				return []string{"--schema", filepath.Join(t.TempDir(), "missing.json")}
			},
			stdout:   func() io.Writer { return &bytes.Buffer{} },
			wantCode: 1,
			wantSub:  "failed to read schema",
		},
		{
			name: "invalid json",
			args: func(t *testing.T) []string {
				path := filepath.Join(t.TempDir(), "schema.json")
				if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
					t.Fatalf("write schema: %v", err)
				}
				return []string{"--schema", path}
			},
			stdout:   func() io.Writer { return &bytes.Buffer{} },
			wantCode: 1,
			wantSub:  "failed to parse schema",
		},
		{
			name: "stdout write error",
			args: func(t *testing.T) []string {
				return []string{"--schema", validSchema(t)}
			},
			stdout:   func() io.Writer { return errWriter{} },
			wantCode: 1,
			wantSub:  "failed to write output",
		},
		{
			name: "output file error",
			args: func(t *testing.T) []string {
				path := validSchema(t)
				outDir := filepath.Join(t.TempDir(), "out")
				if err := os.MkdirAll(outDir, 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				return []string{"--schema", path, "--out", outDir}
			},
			stdout:   func() io.Writer { return &bytes.Buffer{} },
			wantCode: 1,
			wantSub:  "failed to write output file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errBuf bytes.Buffer
			Stdout = tt.stdout()
			Stderr = &errBuf
			defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

			cmd := &NormalizeCommand{}
			code := cmd.Run(tt.args(t))
			if code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, code)
			}
			if tt.wantSub != "" && !strings.Contains(errBuf.String(), tt.wantSub) {
				t.Fatalf("expected %q in output, got %s", tt.wantSub, errBuf.String())
			}
		})
	}
}

func TestGetCommandErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     func(t *testing.T) []string
		setup    func(t *testing.T)
		wantCode int
		wantSub  string
	}{
		{
			name:     "flag error",
			args:     func(t *testing.T) []string { return []string{"--nope"} },
			wantCode: 2,
		},
		{
			name: "init error",
			args: func(t *testing.T) []string { return []string{"--database-id", "db"} },
			setup: func(t *testing.T) {
				initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
					return nil, errors.New("init failed")
				}
			},
			wantCode: 1,
			wantSub:  "init failed",
		},
		{
			name: "schema error",
			args: func(t *testing.T) []string { return []string{"--database-id", "db"} },
			setup: func(t *testing.T) {
				initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
					return &stubClient{schemaErr: errors.New("schema failed")}, nil
				}
			},
			wantCode: 1,
			wantSub:  "schema failed",
		},
		{
			name: "marshal error",
			args: func(t *testing.T) []string { return []string{"--database-id", "db"} },
			setup: func(t *testing.T) {
				badSchema := onyx.Schema{Tables: []onyx.Table{{Name: "bad", Meta: map[string]any{"bad": func() {}}}}}
				initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
					return &stubClient{schema: badSchema}, nil
				}
			},
			wantCode: 1,
			wantSub:  "unsupported type",
		},
		{
			name: "write error",
			args: func(t *testing.T) []string {
				tmp := t.TempDir()
				outDir := filepath.Join(tmp, "out")
				if err := os.MkdirAll(outDir, 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				return []string{"--database-id", "db", "--out", outDir}
			},
			setup: func(t *testing.T) {
				initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
					return &stubClient{schema: onyx.Schema{Tables: []onyx.Table{{Name: "User"}}}}, nil
				}
			},
			wantCode: 1,
			wantSub:  "failed to write schema",
		},
		{
			name: "mkdir error",
			args: func(t *testing.T) []string {
				tmp := t.TempDir()
				// parent path is a file, so MkdirAll should fail
				parent := filepath.Join(tmp, "file")
				if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil {
					t.Fatalf("write file: %v", err)
				}
				return []string{"--database-id", "db", "--out", filepath.Join(parent, "schema.json")}
			},
			setup: func(t *testing.T) {
				initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
					return &stubClient{schema: onyx.Schema{Tables: []onyx.Table{{Name: "User"}}}}, nil
				}
			},
			wantCode: 1,
			wantSub:  "failed to create schema directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
				return &stubClient{schema: onyx.Schema{Tables: []onyx.Table{{Name: "User"}}}}, nil
			}
			if tt.setup != nil {
				tt.setup(t)
			}
			defer func() { initSchemaClient = defaultInitSchemaClient }()

			var out bytes.Buffer
			Stdout, Stderr = &out, &out
			defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

			cmd := &GetCommand{}
			code := cmd.Run(tt.args(t))
			if code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, code)
			}
			if tt.wantSub != "" && !strings.Contains(out.String(), tt.wantSub) {
				t.Fatalf("expected %q in output, got %s", tt.wantSub, out.String())
			}
		})
	}
}

func TestGetCommandStdoutWriteError(t *testing.T) {
	initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
		return &stubClient{schema: onyx.Schema{Tables: []onyx.Table{{Name: "User"}}}}, nil
	}
	defer func() { initSchemaClient = defaultInitSchemaClient }()

	var errBuf bytes.Buffer
	Stdout = errWriter{}
	Stderr = &errBuf
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	cmd := &GetCommand{}
	if code := cmd.Run([]string{"--print"}); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "write failed") {
		t.Fatalf("expected write error, got %s", errBuf.String())
	}
}

func TestPublishCommandErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     func(t *testing.T) []string
		setup    func(t *testing.T)
		wantCode int
		wantSub  string
	}{
		{
			name:     "flag error",
			args:     func(t *testing.T) []string { return []string{"--nope"} },
			wantCode: 2,
		},
		{
			name: "missing file",
			args: func(t *testing.T) []string {
				return []string{"--schema", filepath.Join(t.TempDir(), "missing.json")}
			},
			wantCode: 1,
			wantSub:  "failed to read schema",
		},
		{
			name: "invalid json",
			args: func(t *testing.T) []string {
				path := filepath.Join(t.TempDir(), "schema.json")
				if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
					t.Fatalf("write schema: %v", err)
				}
				return []string{"--schema", path}
			},
			wantCode: 1,
			wantSub:  "failed to parse schema",
		},
		{
			name: "init error",
			args: func(t *testing.T) []string {
				path := filepath.Join(t.TempDir(), "schema.json")
				if err := os.WriteFile(path, []byte(`{"tables":[]}`), 0o644); err != nil {
					t.Fatalf("write schema: %v", err)
				}
				return []string{"--schema", path}
			},
			setup: func(t *testing.T) {
				initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
					return nil, errors.New("init failed")
				}
			},
			wantCode: 1,
			wantSub:  "init failed",
		},
		{
			name: "publish error",
			args: func(t *testing.T) []string {
				path := filepath.Join(t.TempDir(), "schema.json")
				if err := os.WriteFile(path, []byte(`{"tables":[]}`), 0o644); err != nil {
					t.Fatalf("write schema: %v", err)
				}
				return []string{"--schema", path}
			},
			setup: func(t *testing.T) {
				initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
					return &stubClient{publishErr: errors.New("publish failed")}, nil
				}
			},
			wantCode: 1,
			wantSub:  "publish failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() { initSchemaClient = defaultInitSchemaClient }()
			if tt.setup != nil {
				tt.setup(t)
			}

			var out bytes.Buffer
			Stdout, Stderr = &out, &out
			defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

			cmd := &PublishCommand{}
			code := cmd.Run(tt.args(t))
			if code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, code)
			}
			if tt.wantSub != "" && !strings.Contains(out.String(), tt.wantSub) {
				t.Fatalf("expected %q in output, got %s", tt.wantSub, out.String())
			}
		})
	}
}

func TestInfoCommandStatusAndErrors(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "onyx-database.json")
	if err := os.WriteFile(configPath, []byte(`{"databaseId":"file_db","databaseBaseUrl":"https://file.example.com","apiKey":"file_key","apiSecret":"file_secret"}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("ONYX_DATABASE_ID", "")
	t.Setenv("ONYX_DATABASE_BASE_URL", "")
	t.Setenv("ONYX_DATABASE_API_KEY", "")
	t.Setenv("ONYX_DATABASE_API_SECRET", "")

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	connectInfoClient = func(ctx context.Context, cfg resolver.ResolvedConfig) (schemaClient, error) {
		return nil, errors.New("connect failed")
	}
	defer func() { connectInfoClient = defaultConnectInfoClient }()

	cmd := &InfoCommand{}
	if code := cmd.Run([]string{"--config", configPath}); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "Connection : failed") {
		t.Fatalf("expected failure status, got %s", out.String())
	}

	out.Reset()
	if code := cmd.Run([]string{"--config", configPath, "--no-verify"}); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "Connection : skipped (--no-verify)") {
		t.Fatalf("expected skipped status, got %s", out.String())
	}
}

func TestInfoCommandResolveError(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "onyx-database.json")
	if err := os.WriteFile(configPath, []byte(`{"databaseId":""}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("ONYX_DATABASE_ID", "")
	t.Setenv("ONYX_DATABASE_BASE_URL", "")
	t.Setenv("ONYX_DATABASE_API_KEY", "")
	t.Setenv("ONYX_DATABASE_API_SECRET", "")

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	cmd := &InfoCommand{}
	if code := cmd.Run([]string{"--config", configPath}); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(out.String(), "failed to resolve config") {
		t.Fatalf("expected resolve error, got %s", out.String())
	}
}

func TestRedact(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "", want: "(empty)"},
		{input: "abc", want: "***"},
		{input: "abcdef", want: "ab...ef"},
	}

	for _, tt := range tests {
		if got := redact(tt.input); got != tt.want {
			t.Fatalf("expected %q, got %q", tt.want, got)
		}
	}
}

func TestDescribeHelpers(t *testing.T) {
	fieldTests := []struct {
		name string
		a    onyx.Field
		b    onyx.Field
		want string
	}{
		{name: "type change", a: onyx.Field{Name: "id", Type: "string"}, b: onyx.Field{Name: "id", Type: "int"}, want: "type string -> int"},
		{name: "nullable change", a: onyx.Field{Name: "id", Type: "string"}, b: onyx.Field{Name: "id", Type: "string", Nullable: true}, want: "nullable false -> true"},
		{name: "both", a: onyx.Field{Name: "id", Type: "string"}, b: onyx.Field{Name: "id", Type: "int", Nullable: true}, want: "type string -> int; nullable false -> true"},
		{name: "none", a: onyx.Field{Name: "id", Type: "string"}, b: onyx.Field{Name: "id", Type: "string"}, want: ""},
	}
	for _, tt := range fieldTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeFieldChange(tt.a, tt.b); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}

	resolverTests := []struct {
		name string
		a    onyx.Resolver
		b    onyx.Resolver
		want string
	}{
		{name: "unchanged", a: onyx.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, b: onyx.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, want: "unchanged"},
		{name: "resolver change", a: onyx.Resolver{Name: "r", Resolver: "sql"}, b: onyx.Resolver{Name: "r", Resolver: "sql_v2"}, want: "definition changed"},
		{name: "meta length", a: onyx.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, b: onyx.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v", "k2": "v2"}}, want: "meta changed"},
		{name: "meta value", a: onyx.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "v"}}, b: onyx.Resolver{Name: "r", Resolver: "sql", Meta: map[string]any{"k": "other"}}, want: "meta changed"},
	}
	for _, tt := range resolverTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeResolverChange(tt.a, tt.b); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestLoadSchemaError(t *testing.T) {
	if _, err := loadSchema(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatalf("expected error for missing schema")
	}
}

func TestSummarizeDiffIncludesResolvers(t *testing.T) {
	base := onyx.Schema{Tables: []onyx.Table{{
		Name:      "users",
		Fields:    []onyx.Field{{Name: "id", Type: "string"}},
		Resolvers: []onyx.Resolver{{Name: "old", Resolver: "sql"}, {Name: "change", Resolver: "sql"}},
	}}}
	updated := onyx.Schema{Tables: []onyx.Table{{
		Name:      "users",
		Fields:    []onyx.Field{{Name: "id", Type: "string"}},
		Resolvers: []onyx.Resolver{{Name: "new", Resolver: "sql"}, {Name: "change", Resolver: "sql_v2"}},
	}}}

	diff := schemas.DiffSchemas(base, updated)
	text := summarizeDiff(diff)
	for _, fragment := range []string{"Modified resolvers:", "Added resolvers:", "Removed resolvers:"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected %q in summary, got %s", fragment, text)
		}
	}
}

var defaultInitSchemaClient = initSchemaClient
