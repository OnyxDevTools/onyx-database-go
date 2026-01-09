package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateOptionsDefaultsAndErrors(t *testing.T) {
	if err := ValidateOptions(nil); err == nil {
		t.Fatalf("expected nil options error")
	}

	opts := &Options{OutPath: "out.go", PackageName: "pkg", TimestampFormat: "bad"}
	if err := ValidateOptions(opts); err == nil {
		t.Fatalf("expected invalid timestamp format error")
	}

	opts = &Options{OutPath: "out.go", PackageName: "pkg", Source: "bad"}
	if err := ValidateOptions(opts); err == nil {
		t.Fatalf("expected invalid source error")
	}

	opts = &Options{OutPath: "out.go", PackageName: "pkg"}
	if err := ValidateOptions(opts); err != nil {
		t.Fatalf("expected defaults to pass, got %v", err)
	}
	if opts.TimestampFormat != "time" || opts.Source != "file" {
		t.Fatalf("expected defaults applied, got %+v", opts)
	}

	opts = &Options{OutPath: "out.go", PackageName: "pkg", Source: "API", TimestampFormat: "string"}
	if err := ValidateOptions(opts); err != nil {
		t.Fatalf("expected normalized valid values, got %v", err)
	}
	if opts.Source != "api" {
		t.Fatalf("expected source lowercased, got %s", opts.Source)
	}
}

func TestRunLoadsSchema(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","fields":[{"name":"id","type":"String"},{"name":"createdAt","type":"Timestamp","nullable":true}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	outDir := filepath.Join(tmp, "onyx")
	opts := Options{SchemaPath: schemaPath, OutPath: outDir, PackageName: "pkg"}
	if err := Run(opts); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(outDir, "models.go"))
	if err != nil {
		t.Fatalf("expected generated file, read err: %v", err)
	}
	content := string(out)
	if !strings.Contains(content, "package pkg") || !strings.Contains(content, "var Tables") {
		t.Fatalf("unexpected generated content:\n%s", content)
	}

	if !strings.Contains(content, "User: \"User\"") {
		t.Fatalf("expected table mapping, got:\n%s", content)
	}

	if strings.Contains(content, "FieldUser") {
		t.Fatalf("did not expect field constants, got:\n%s", content)
	}

	if !strings.Contains(content, "type User struct") || !strings.Contains(content, "CreatedAt *time.Time") {
		t.Fatalf("expected struct with timestamp mapping, got:\n%s", content)
	}

	if strings.Contains(content, "MarshalJSON") {
		t.Fatalf("did not expect custom marshal, got:\n%s", content)
	}

	clientPath := clientOutPath(outDir)
	clientContent, err := os.ReadFile(clientPath)
	if err != nil {
		t.Fatalf("expected client file, read err: %v", err)
	}
	if !strings.Contains(string(clientContent), "type DB struct") || !strings.Contains(string(clientContent), "Users() UsersClient") {
		t.Fatalf("expected client helpers, got:\n%s", string(clientContent))
	}
	if !strings.Contains(string(clientContent), "New(ctx context.Context, cfg Config) (DB, error)") {
		t.Fatalf("expected config-based constructor, got:\n%s", string(clientContent))
	}
	if !strings.Contains(string(clientContent), "Save(ctx context.Context, item User, cascades ...onyx.CascadeSpec) (User, error)") {
		t.Fatalf("expected typed save helper, got:\n%s", string(clientContent))
	}
	if !strings.Contains(string(clientContent), "DeleteByID(ctx context.Context, id string) error") {
		t.Fatalf("expected typed delete-by-id helper, got:\n%s", string(clientContent))
	}
	if !strings.Contains(string(clientContent), "Select(fields ...string) UsersMapClient") {
		t.Fatalf("expected select to return map client, got:\n%s", string(clientContent))
	}
	if !strings.Contains(string(clientContent), "List(ctx context.Context) ([]map[string]any, error)") {
		t.Fatalf("expected map list helper, got:\n%s", string(clientContent))
	}
}

func TestLoadSchemaRespectsSource(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User","resolvers":["roles"]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	fileSchema, err := loadSchema(context.Background(), Options{Source: "file", SchemaPath: schemaPath})
	if err != nil || len(fileSchema.Tables) != 1 {
		t.Fatalf("loadSchema file err: %v schema: %+v", err, fileSchema)
	}
	if len(fileSchema.Tables[0].Resolvers) != 1 || fileSchema.Tables[0].Resolvers[0].Name != "roles" {
		t.Fatalf("expected resolver parsed, got %+v", fileSchema.Tables[0].Resolvers)
	}

	called := false
	orig := initClient
	initClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
		called = true
		return &stubSchemaClient{}, nil
	}
	defer func() { initClient = orig }()

	_, err = loadSchema(context.Background(), Options{Source: "api"})
	if err != nil {
		t.Fatalf("loadSchema api err: %v", err)
	}
	if !called {
		t.Fatalf("expected api client to be used")
	}
}
