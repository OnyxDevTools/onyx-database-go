package contract

import "context"

// Client defines the SDK surface for interacting with an Onyx database.
type Client interface {
	From(table string) Query
	Cascade(spec CascadeSpec) CascadeClient

	Save(ctx context.Context, table string, entity any, relationships []string) (map[string]any, error)
	Delete(ctx context.Context, table, id string) error
	BatchSave(ctx context.Context, table string, entities []any, batchSize int) error

	Schema(ctx context.Context) (Schema, error)
	GetSchema(ctx context.Context, tables []string) (Schema, error)
	PublishSchema(ctx context.Context, schema Schema) error
	UpdateSchema(ctx context.Context, schema Schema, publish bool) error
	ValidateSchema(ctx context.Context, schema Schema) error
	GetSchemaHistory(ctx context.Context) ([]Schema, error)

	Documents() OnyxDocumentsClient

	ListSecrets(ctx context.Context) ([]OnyxSecret, error)
	GetSecret(ctx context.Context, key string) (OnyxSecret, error)
	PutSecret(ctx context.Context, secret OnyxSecret) (OnyxSecret, error)
	DeleteSecret(ctx context.Context, key string) error
}
