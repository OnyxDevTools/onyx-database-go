package onyx

import (
	"context"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/core"
)

type stubCoreClient struct{}

func (s *stubCoreClient) From(table string) core.Query { return nil }
func (s *stubCoreClient) Cascade(spec core.CascadeSpec) core.CascadeClient {
	return nil
}
func (s *stubCoreClient) Save(ctx context.Context, table string, entity any, relationships []string) (map[string]any, error) {
	return nil, nil
}
func (s *stubCoreClient) Delete(ctx context.Context, table, id string) error { return nil }
func (s *stubCoreClient) BatchSave(ctx context.Context, table string, entities []any, batchSize int) error {
	return nil
}
func (s *stubCoreClient) Schema(ctx context.Context) (core.Schema, error) { return core.Schema{}, nil }
func (s *stubCoreClient) GetSchema(ctx context.Context, tables []string) (core.Schema, error) {
	return core.Schema{}, nil
}
func (s *stubCoreClient) PublishSchema(ctx context.Context, schema core.Schema) error { return nil }
func (s *stubCoreClient) UpdateSchema(ctx context.Context, schema core.Schema, publish bool) error {
	return nil
}
func (s *stubCoreClient) ValidateSchema(ctx context.Context, schema core.Schema) error { return nil }
func (s *stubCoreClient) GetSchemaHistory(ctx context.Context) ([]core.Schema, error) { return nil, nil }
func (s *stubCoreClient) Documents() core.DocumentClient                           { return nil }
func (s *stubCoreClient) ListSecrets(ctx context.Context) ([]core.Secret, error)  { return nil, nil }
func (s *stubCoreClient) GetSecret(ctx context.Context, key string) (core.Secret, error) {
	return core.Secret{}, nil
}
func (s *stubCoreClient) PutSecret(ctx context.Context, secret core.Secret) (core.Secret, error) {
	return secret, nil
}
func (s *stubCoreClient) DeleteSecret(ctx context.Context, key string) error { return nil }

func TestTypedReturnsGeneratedClient(t *testing.T) {
	base := &stubCoreClient{}
	wrapped := client{Client: base}

	typed := wrapped.Typed()
	if typed.Core() != base {
		t.Fatalf("expected typed client to reuse core client")
	}
}
