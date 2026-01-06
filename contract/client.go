package contract

import "context"

// Client defines the high-level SDK entrypoints for interacting with Onyx Database.
type Client interface {
	From(table string) Query
	Cascade(spec CascadeSpec) CascadeClient
	Save(ctx context.Context, table string, entity any) error
	Delete(ctx context.Context, table, id string) error
	BatchSave(ctx context.Context, table string, entities []any, batchSize int) error
	Schema(ctx context.Context) (Schema, error)
}
