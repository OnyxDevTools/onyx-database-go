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

	common, err := os.ReadFile(filepath.Join(outDir, "common.go"))
	if err != nil {
		t.Fatalf("expected generated common file, read err: %v", err)
	}
	if !strings.Contains(string(common), "package pkg") || !strings.Contains(string(common), "var Tables") {
		t.Fatalf("unexpected common content:\n%s", string(common))
	}
	if !strings.Contains(string(common), "User: \"User\"") {
		t.Fatalf("expected table mapping, got:\n%s", string(common))
	}
	if !strings.Contains(string(common), "type DB struct") || !strings.Contains(string(common), "New(ctx context.Context, cfg Config) (DB, error)") {
		t.Fatalf("expected DB helpers, got:\n%s", string(common))
	}

	userContent, err := os.ReadFile(filepath.Join(outDir, "user.go"))
	if err != nil {
		t.Fatalf("expected user file, read err: %v", err)
	}
	if !strings.Contains(string(userContent), "type User struct") || !strings.Contains(string(userContent), "CreatedAt *time.Time") {
		t.Fatalf("expected struct with timestamp mapping, got:\n%s", string(userContent))
	}
	if strings.Contains(string(userContent), "MarshalJSON") {
		t.Fatalf("did not expect custom marshal, got:\n%s", string(userContent))
	}
	if !strings.Contains(string(userContent), "Users() UsersClient") {
		t.Fatalf("expected typed accessor, got:\n%s", string(userContent))
	}
	if !strings.Contains(string(userContent), "Save(ctx context.Context, item User, cascades ...onyx.CascadeSpec) (User, error)") {
		t.Fatalf("expected typed save helper, got:\n%s", string(userContent))
	}
	if !strings.Contains(string(userContent), "DeleteByID(ctx context.Context, id string) (int, error)") {
		t.Fatalf("expected typed delete-by-id helper, got:\n%s", string(userContent))
	}
	if !strings.Contains(string(userContent), "Select(fields ...string) UsersMapClient") {
		t.Fatalf("expected select to return map client, got:\n%s", string(userContent))
	}
	if !strings.Contains(string(userContent), "List(ctx context.Context) ([]map[string]any, error)") {
		t.Fatalf("expected map list helper, got:\n%s", string(userContent))
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
