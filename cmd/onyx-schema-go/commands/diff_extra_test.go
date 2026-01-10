package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	schemas "github.com/OnyxDevTools/onyx-database-go/impl/schema"
	internalSchemas "github.com/OnyxDevTools/onyx-database-go/internal/schema"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func TestDiffCommandIdenticalSchemas(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "schema.json")
	writeFile(t, path, `{"tables":[{"name":"users","fields":[{"name":"id","type":"string"}]}]}`)

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--a", path, "--b", path}); code != 0 {
		t.Fatalf("expected success, got %d", code)
	}
	if !strings.Contains(out.String(), "Schemas are identical.") {
		t.Fatalf("expected identical message, got %s", out.String())
	}
}

func TestDiffCommandFetchError(t *testing.T) {
	orig := schemaClientFactoryHandler
	t.Cleanup(func() { schemaClientFactoryHandler = orig })
	schemaClientFactoryHandler = func(ctx context.Context, databaseID string) (schemaClient, error) {
		return schemaClientFunc(func(context.Context) (onyx.Schema, error) {
			return onyx.Schema{}, os.ErrNotExist
		}), nil
	}

	tmp := t.TempDir()
	path := filepath.Join(tmp, "schema.json")
	writeFile(t, path, `{"tables":[]}`)

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--a", path, "--database-id", "db"}); code != 1 {
		t.Fatalf("expected fetch error exit code 1, got %d", code)
	}
	if !strings.Contains(out.String(), "failed to fetch schema from API") {
		t.Fatalf("expected fetch error output, got %s", out.String())
	}
}

func TestDiffCommandFlagError(t *testing.T) {
	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--unknown"}); code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
}

func TestSummarizeDiffFullSections(t *testing.T) {
	diff := schemas.SchemaDiff{
		AddedTables:   []onyx.Table{{Name: "New"}},
		RemovedTables: []onyx.Table{{Name: "Old"}},
		TableDiffs: []schemas.TableDiff{{
			Name:          "Users",
			AddedFields:   []onyx.Field{{Name: "email", Type: "string"}},
			RemovedFields: []onyx.Field{{Name: "age", Type: "int"}},
			ModifiedFields: []schemas.FieldDiff{{
				Name: "id",
				From: onyx.Field{Name: "id", Type: "string"},
				To:   onyx.Field{Name: "id", Type: "uuid", Nullable: true},
			}},
			ModifiedResolvers: []internalSchemas.ResolverDiff{{
				Name: "roles",
				From: onyx.Resolver{Name: "roles", Resolver: "old", Meta: map[string]any{"x": 1}},
				To:   onyx.Resolver{Name: "roles", Resolver: "new", Meta: map[string]any{"x": 2}},
			}},
			AddedResolvers:   []string{"profiles"},
			RemovedResolvers: []string{"perms"},
		}},
	}

	summary := summarizeDiff(diff)
	assertContains := func(sub string) {
		if !strings.Contains(summary, sub) {
			t.Fatalf("expected %q in summary, got %s", sub, summary)
		}
	}
	assertContains("Tables only in updated schema")
	assertContains("Tables only in base schema")
	assertContains("Added fields")
	assertContains("Removed fields")
	assertContains("Modified fields")
	assertContains("Modified resolvers")
	assertContains("Added resolvers")
	assertContains("Removed resolvers")
}
