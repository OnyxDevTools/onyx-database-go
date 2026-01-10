package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func TestMapFieldTypeVariants(t *testing.T) {
	cases := []struct {
		field   onyx.Field
		format  string
		expects string
	}{
		{onyx.Field{Type: "string"}, "time", "string"},
		{onyx.Field{Type: "string", Nullable: true}, "time", "*string"},
		{onyx.Field{Type: "bool"}, "time", "bool"},
		{onyx.Field{Type: "bool", Nullable: true}, "time", "*bool"},
		{onyx.Field{Type: "integer"}, "time", "int64"},
		{onyx.Field{Type: "integer", Nullable: true}, "time", "*int64"},
		{onyx.Field{Type: "timestamp"}, "time", "time.Time"},
		{onyx.Field{Type: "timestamp", Nullable: true}, "time", "*time.Time"},
		{onyx.Field{Type: "timestamp", Nullable: true}, "string", "*string"},
		{onyx.Field{Type: "timestamp", Nullable: false}, "string", "string"},
		{onyx.Field{Type: "object"}, "time", "map[string]any"},
		{onyx.Field{Type: "unknown"}, "time", "any"},
	}

	for _, tc := range cases {
		if got := mapFieldType(tc.field, tc.format); got != tc.expects {
			t.Fatalf("mapFieldType(%+v,%s)=%s", tc.field, tc.format, got)
		}
	}
}

func TestSmallHelpersCoverage(t *testing.T) {
	table := onyx.Table{Fields: []onyx.Field{{Name: "Id"}, {Name: " name "}}}
	if !hasField(table, "name") || hasField(table, "missing") {
		t.Fatalf("hasField mismatch")
	}

	if plural := toPlural("users"); plural != "users" {
		t.Fatalf("expected trailing s preserved")
	}
	if plural := toPlural("role"); plural != "roles" {
		t.Fatalf("expected s appended")
	}

	if toFileName("") != "table" {
		t.Fatalf("expected fallback file name for empty input")
	}
	if cleaned := toFileName("Doc Name!"); !strings.HasPrefix(cleaned, "doc_name") {
		t.Fatalf("expected cleaned filename, got %s", cleaned)
	}

	if hasResolvers(onyx.Schema{}) {
		t.Fatalf("expected false when no resolvers present")
	}
	if !hasResolvers(onyx.Schema{Tables: []onyx.Table{{Resolvers: []onyx.Resolver{{Name: "r"}}}}}) {
		t.Fatalf("expected true when resolver exists")
	}
}

func TestRenderCommonResolverBranches(t *testing.T) {
	schema := onyx.Schema{
		Tables: []onyx.Table{
			{
				Name: "User",
				Resolvers: []onyx.Resolver{
					{Name: "roles"},
					{Name: "teams"},
				},
			},
			{Name: "Audit"},
		},
	}
	code := string(renderCommon(schema, "pkg"))
	if !strings.Contains(code, "Resolvers = map[string][]string") {
		t.Fatalf("expected resolvers map, got %s", code)
	}
}

func TestRenderGenerateIncludesOptions(t *testing.T) {
	opts := Options{
		Source:          "api",
		SchemaPath:      "schema.json",
		DatabaseID:      "db1",
		OutPath:         ".",
		PackageName:     "pkg",
		Tables:          []string{"User", "Role"},
		TimestampFormat: "string",
	}
	code := string(renderGenerate(opts))
	assertContains := func(sub string) {
		if !strings.Contains(code, sub) {
			t.Fatalf("expected renderGenerate to include %q, got:\n%s", sub, code)
		}
	}
	assertContains("--source api")
	assertContains("--schema schema.json")
	assertContains("--database-id db1")
	assertContains("--out .")
	assertContains("--package pkg")
	assertContains("--timestamps string")
	assertContains("--tables User,Role")
}

func TestLoadSchemaFailures(t *testing.T) {
	if _, err := loadSchemaFromFile("", nil); err == nil {
		t.Fatalf("expected schema path required error")
	}

	bad := filepath.Join(t.TempDir(), "missing.json")
	if _, err := loadSchemaFromFile(bad, nil); err == nil {
		t.Fatalf("expected file read error")
	}

	orig := initClient
	t.Cleanup(func() { initClient = orig })
	initClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
		return nil, context.DeadlineExceeded
	}
	if _, err := loadSchemaFromAPI(context.Background(), Options{DatabaseID: "db"}); err == nil {
		t.Fatalf("expected init error")
	}

	initClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
		return errSchemaClient{err: os.ErrInvalid}, nil
	}
	if _, err := loadSchemaFromAPI(context.Background(), Options{}); err == nil {
		t.Fatalf("expected schema fetch error")
	}
}

func (e errSchemaClient) Schema(context.Context) (onyx.Schema, error) {
	return onyx.Schema{}, e.err
}

type errSchemaClient struct{ err error }

func TestRunWithInvalidOptions(t *testing.T) {
	if err := Run(Options{}); err == nil {
		t.Fatalf("expected validation failure")
	}
}

func TestRunLoadsFromAPIWhenConfigured(t *testing.T) {
	tmp := t.TempDir()
	outDir := filepath.Join(tmp, "gen")
	orig := initClient
	t.Cleanup(func() { initClient = orig })
	initClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
		return &stubSchemaClient{schema: onyx.Schema{
			Tables: []onyx.Table{{Name: "User", Fields: []onyx.Field{{Name: "id"}}}},
		}}, nil
	}

	opts := Options{
		Source:      "api",
		DatabaseID:  "db1",
		OutPath:     outDir,
		PackageName: "pkg",
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run api err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "common.go")); err != nil {
		t.Fatalf("expected generated files: %v", err)
	}
}
