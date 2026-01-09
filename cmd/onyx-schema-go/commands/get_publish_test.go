package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyxclient"
)

func TestGetCommandWritesSchema(t *testing.T) {
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "schema.json")

	originalInit := initSchemaClient
	initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
		return &stubClient{schema: onyx.Schema{Tables: []onyx.Table{{Name: "User"}}}}, nil
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

func TestGetCommandPrintsSchema(t *testing.T) {
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "schema.json")

	originalInit := initSchemaClient
	initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
		return &stubClient{schema: onyx.Schema{Tables: []onyx.Table{{Name: "Role"}}}}, nil
	}
	defer func() { initSchemaClient = originalInit }()

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	cmd := &GetCommand{}
	if code := cmd.Run([]string{"--print", "--out", outPath}); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if !strings.Contains(out.String(), `"Role"`) {
		t.Fatalf("expected printed schema, got: %s", out.String())
	}

	if _, err := os.Stat(outPath); err == nil {
		t.Fatalf("expected no file to be written when --print is set")
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
	initSchemaClient = func(ctx context.Context, databaseID string) (onyx.Client, error) {
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
	schema        onyx.Schema
	publishCalled bool
}

func (s *stubClient) From(table string) onyx.Query                     { return nil }
func (s *stubClient) Cascade(spec onyx.CascadeSpec) onyx.CascadeClient { return nil }
func (s *stubClient) Save(ctx context.Context, table string, entity any, relationships []string) (map[string]any, error) {
	return nil, nil
}
func (s *stubClient) Delete(ctx context.Context, table, id string) error { return nil }
func (s *stubClient) BatchSave(ctx context.Context, table string, entities []any, batchSize int) error {
	return nil
}
func (s *stubClient) Schema(ctx context.Context) (onyx.Schema, error) { return s.schema, nil }
func (s *stubClient) GetSchema(ctx context.Context, tables []string) (onyx.Schema, error) {
	return s.schema, nil
}
func (s *stubClient) GetSchemaHistory(ctx context.Context) ([]onyx.Schema, error) {
	return nil, nil
}
func (s *stubClient) UpdateSchema(ctx context.Context, schema onyx.Schema, publish bool) error {
	s.publishCalled = publish
	s.schema = schema
	return nil
}
func (s *stubClient) ValidateSchema(ctx context.Context, schema onyx.Schema) error { return nil }
func (s *stubClient) PublishSchema(ctx context.Context, schema onyx.Schema) error {
	s.publishCalled = true
	s.schema = schema
	return nil
}
func (s *stubClient) Documents() onyx.DocumentClient { return nil }
func (s *stubClient) ListSecrets(ctx context.Context) ([]onyx.Secret, error) {
	return nil, nil
}
func (s *stubClient) GetSecret(ctx context.Context, key string) (onyx.Secret, error) {
	return onyx.Secret{}, nil
}
func (s *stubClient) PutSecret(ctx context.Context, secret onyx.Secret) (onyx.Secret, error) {
	return secret, nil
}
func (s *stubClient) DeleteSecret(ctx context.Context, key string) error { return nil }
func (s *stubClient) Typed() onyxclient.Client                           { return onyxclient.NewClient(s) }
