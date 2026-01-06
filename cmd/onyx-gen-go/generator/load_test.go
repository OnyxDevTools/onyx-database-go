package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestLoadSchemaFromFileNormalizesAndFilters(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	raw := `{"tables": [{"name": "Z", "fields": [{"name": "b"}, {"name": "a"}]}, {"name": "A", "fields": [{"name": "id"}]}]}`

	if err := os.WriteFile(schemaPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	schema, err := loadSchemaFromFile(schemaPath, []string{"A"})
	if err != nil {
		t.Fatalf("loadSchemaFromFile returned error: %v", err)
	}

	if len(schema.Tables) != 1 || schema.Tables[0].Name != "A" {
		t.Fatalf("unexpected tables after filter: %+v", schema.Tables)
	}

	fields := schema.Tables[0].Fields
	if len(fields) != 1 || fields[0].Name != "id" {
		t.Fatalf("unexpected fields after normalization: %+v", fields)
	}
}

func TestLoadSchemaFromAPIUsesResolver(t *testing.T) {
	called := false
	originalInit := initClient
	initClient = func(ctx context.Context, databaseID string) (schemaClient, error) {
		called = true
		if databaseID != "db_123" {
			t.Fatalf("unexpected database id: %s", databaseID)
		}
		return &stubSchemaClient{schema: contract.Schema{Tables: []contract.Table{{
			Name:   "B",
			Fields: []contract.Field{{Name: "z"}, {Name: "a"}},
		}}}}, nil
	}
	defer func() { initClient = originalInit }()

	schema, err := loadSchemaFromAPI(context.Background(), Options{DatabaseID: "db_123", Tables: []string{"B"}})
	if err != nil {
		t.Fatalf("loadSchemaFromAPI returned error: %v", err)
	}

	if !called {
		t.Fatalf("expected resolver to be invoked")
	}

	if len(schema.Tables) != 1 || len(schema.Tables[0].Fields) != 2 || schema.Tables[0].Fields[0].Name != "a" {
		t.Fatalf("schema not normalized as expected: %+v", schema.Tables)
	}
}

type stubSchemaClient struct {
	schema contract.Schema
}

func (s *stubSchemaClient) Schema(ctx context.Context) (contract.Schema, error) {
	return s.schema, nil
}
