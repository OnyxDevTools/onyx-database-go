package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestGetCommandWritesSchema(t *testing.T) {
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "schema.json")

	originalInit := initSchemaClient
	initSchemaClient = func(ctx context.Context, databaseID string) (contract.Client, error) {
		return &stubClient{schema: contract.Schema{Tables: []contract.Table{{Name: "User"}}}}, nil
	}
	defer func() { initSchemaClient = originalInit }()

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	cmd := &GetCommand{}
	if code := cmd.Run([]string{"--database-id", "db_123", "--out", outPath}); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	contents, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if !strings.Contains(string(contents), "User") {
		t.Fatalf("unexpected contents: %s", string(contents))
	}
}

func TestPublishCommandSendsSchema(t *testing.T) {
	tmp := t.TempDir()
	schemaPath := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"tables":[{"name":"User"}]}`), 0o644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	stub := &stubClient{}
	originalInit := initSchemaClient
	initSchemaClient = func(ctx context.Context, databaseID string) (contract.Client, error) {
		return stub, nil
	}
	defer func() { initSchemaClient = originalInit }()

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	cmd := &PublishCommand{}
	if code := cmd.Run([]string{"--schema", schemaPath}); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if !stub.publishCalled {
		t.Fatalf("expected publish to be invoked")
	}
}

type stubClient struct {
	schema        contract.Schema
	publishCalled bool
}

func (s *stubClient) From(table string) contract.Query                         { return nil }
func (s *stubClient) Cascade(spec contract.CascadeSpec) contract.CascadeClient { return nil }
func (s *stubClient) Save(ctx context.Context, table string, entity any) error { return nil }
func (s *stubClient) Delete(ctx context.Context, table, id string) error       { return nil }
func (s *stubClient) BatchSave(ctx context.Context, table string, entities []any, batchSize int) error {
	return nil
}
func (s *stubClient) Schema(ctx context.Context) (contract.Schema, error) { return s.schema, nil }
func (s *stubClient) PublishSchema(ctx context.Context, schema contract.Schema) error {
	s.publishCalled = true
	s.schema = schema
	return nil
}
func (s *stubClient) Documents() contract.DocumentClient { return nil }
func (s *stubClient) Secrets() contract.SecretClient     { return nil }
