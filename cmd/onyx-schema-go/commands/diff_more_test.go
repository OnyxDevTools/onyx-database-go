package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func TestDiffCommandAPISuccessWithoutDatabaseID(t *testing.T) {
	tmp := t.TempDir()
	pathA := filepath.Join(tmp, "a.json")
	writeFile(t, pathA, `{"tables":[{"name":"users","fields":[{"name":"id","type":"string"}]}]}`)

	orig := schemaClientFactoryHandler
	t.Cleanup(func() { schemaClientFactoryHandler = orig })
	schemaClientFactoryHandler = func(ctx context.Context, databaseID string) (schemaClient, error) {
		if databaseID != "" {
			t.Fatalf("expected empty database id, got %s", databaseID)
		}
		return schemaClientFunc(func(context.Context) (onyx.Schema, error) {
			return onyx.Schema{
				Tables: []onyx.Table{{Name: "users", Fields: []onyx.Field{{Name: "id", Type: "string"}}}},
			}, nil
		}), nil
	}

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--a", pathA}); code != 0 {
		t.Fatalf("expected success, got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("Comparing schemas (base=API (configured credentials)")) {
		t.Fatalf("expected base source message, got %s", out.String())
	}
}

func TestDiffCommandLoadSchemaAError(t *testing.T) {
	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--a", "missing.json"}); code != 1 {
		t.Fatalf("expected read error code, got %d", code)
	}
}

func TestDiffCommandMarshalError(t *testing.T) {
	tmp := t.TempDir()
	pathA := filepath.Join(tmp, "a.json")
	pathB := filepath.Join(tmp, "b.json")
	writeFile(t, pathA, `{"tables":[]}`)
	writeFile(t, pathB, `{"tables":[]}`)

	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	orig := jsonMarshalIndent
	jsonMarshalIndent = func(v any, prefix, indent string) ([]byte, error) {
		return nil, errors.New("marshal fail")
	}
	t.Cleanup(func() { jsonMarshalIndent = orig })

	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--a", pathA, "--b", pathB, "--json"}); code != 1 {
		t.Fatalf("expected marshal error code, got %d", code)
	}
}

func TestDiffCommandSchemaBReadError(t *testing.T) {
	tmp := t.TempDir()
	pathA := filepath.Join(tmp, "a.json")
	writeFile(t, pathA, `{"tables":[]}`)

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--a", pathA, "--b", filepath.Join(tmp, "missing.json")}); code != 1 {
		t.Fatalf("expected read error code, got %d", code)
	}
}

func TestConnectSchemaClientDefaultBranches(t *testing.T) {
	origFactory := schemaClientFactoryHandler
	t.Cleanup(func() { schemaClientFactoryHandler = origFactory })

	var calledID, calledDefault bool
	schemaClientFactoryHandler = func(ctx context.Context, databaseID string) (schemaClient, error) {
		if databaseID != "" {
			calledID = true
		} else {
			calledDefault = true
		}
		return schemaClientFunc(func(context.Context) (onyx.Schema, error) {
			return onyx.Schema{}, nil
		}), nil
	}

	if _, err := connectSchemaClient(context.Background(), "db1"); err != nil {
		t.Fatalf("connectSchemaClient with id err: %v", err)
	}
	if _, err := connectSchemaClient(context.Background(), ""); err != nil {
		t.Fatalf("connectSchemaClient default err: %v", err)
	}
	if !calledID || !calledDefault {
		t.Fatalf("expected both branches to be called, id=%t default=%t", calledID, calledDefault)
	}
}

func TestFetchSchemaFromAPIErrorPath(t *testing.T) {
	orig := schemaClientFactoryHandler
	t.Cleanup(func() { schemaClientFactoryHandler = orig })
	schemaClientFactoryHandler = func(context.Context, string) (schemaClient, error) {
		return nil, errors.New("connect fail")
	}

	if _, err := fetchSchemaFromAPI(context.Background(), "db"); err == nil {
		t.Fatalf("expected error from fetchSchemaFromAPI")
	}
}

func TestSchemaClientFactoryHandlerBranches(t *testing.T) {
	restore := setEnvVars(map[string]string{
		"ONYX_DATABASE_ID":         "db",
		"ONYX_DATABASE_BASE_URL":   "http://example.com",
		"ONYX_DATABASE_API_KEY":    "k",
		"ONYX_DATABASE_API_SECRET": "s",
	})
	t.Cleanup(restore)

	if _, err := schemaClientFactoryHandler(context.Background(), "db"); err != nil {
		t.Fatalf("expected handler with database id to succeed, got %v", err)
	}
	if _, err := schemaClientFactoryHandler(context.Background(), ""); err != nil {
		t.Fatalf("expected handler default branch to succeed, got %v", err)
	}
}
