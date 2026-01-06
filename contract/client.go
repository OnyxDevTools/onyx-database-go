package contract

import "context"

// Client defines the high-level SDK entrypoints for interacting with Onyx Database.
type Client interface {
	From(table string) Query
	Cascade(spec CascadeSpec) CascadeClient
	Save(ctx context.Context, table string, entity any, relationships []string) (map[string]any, error)
	Delete(ctx context.Context, table, id string) error
	BatchSave(ctx context.Context, table string, entities []any, batchSize int) error
	// Schema/PublishSchema are kept for legacy callers; prefer GetSchema/UpdateSchema/ValidateSchema.
	Schema(ctx context.Context) (Schema, error)
	PublishSchema(ctx context.Context, schema Schema) error
	GetSchema(ctx context.Context, tables []string) (Schema, error)
	GetSchemaHistory(ctx context.Context) ([]Schema, error)
	UpdateSchema(ctx context.Context, schema Schema, publish bool) error
	ValidateSchema(ctx context.Context, schema Schema) error
	Documents() DocumentClient
	ListSecrets(ctx context.Context) ([]Secret, error)
	GetSecret(ctx context.Context, key string) (Secret, error)
	PutSecret(ctx context.Context, secret Secret) (Secret, error)
	DeleteSecret(ctx context.Context, key string) error
}
